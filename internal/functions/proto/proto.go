// Package proto backs the GoogleSQL proto-reflection runtime
// surface. It exposes one core UDF — GET_PROTO_FIELD(bytes, number,
// type_kind, default) — that extracts a single field from a
// wire-encoded proto message and returns the value typed for the SQL
// surface. The formatter in internal/formatter.go lowers
// ResolvedGetProtoField (and PROTO_DEFAULT_IF_NULL over it) to a
// call to this UDF; the protobuf descriptor metadata is captured at
// formatter time (field tag number, expected SQL type, default) so
// the runtime can decode without a live DescriptorPool.
package proto

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	gotime "time"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// SQL type kinds we accept. The formatter passes the kind as a
// lowercase string so the runtime doesn't have to import googlesql.
// "message" is the catch-all for sub-message fields — the field
// payload is returned as BYTES so the caller can recurse.
const (
	kindBool    = "bool"
	kindInt32   = "int32"
	kindInt64   = "int64"
	kindUint32  = "uint32"
	kindUint64  = "uint64"
	kindFloat   = "float"
	kindDouble  = "double"
	kindString  = "string"
	kindBytes   = "bytes"
	kindEnum    = "enum"
	kindMessage = "message"
)

// BindGetProtoField implements GOOGLESQLITE_GET_PROTO_FIELD(proto,
// field_number, type_kind, default_b64). `proto` is the proto wire
// bytes for the containing message; `field_number` is the integer
// tag of the field to extract; `type_kind` is one of the lowercase
// kind strings above; `default_b64` is the base64 of the field's
// proto-defined default value (or empty when no default).
//
// If the proto bytes don't carry the field, the binding returns the
// supplied default (or SQL NULL if no default and the field is not
// present). The implementation walks the wire stream once, picking
// up the last occurrence of the requested tag (proto3 semantic: the
// last serialised value wins for non-repeated fields).
// wireEntry is one decoded occurrence of a target proto field found
// by scanProtoField. Exactly one of the slots is meaningful, selected
// by Wire:
//
//   - VarintType  → Varint holds the raw varint value.
//   - Fixed32Type → Fixed holds the 32-bit value (zero-extended).
//   - Fixed64Type → Fixed holds the 64-bit value.
//   - BytesType   → Bytes holds a private copy of the length-delimited
//     payload (safe to retain past the scan).
type wireEntry struct {
	Wire   protowire.Type
	Varint uint64
	Fixed  uint64
	Bytes  []byte
}

// scanResult reports the outcome of a scanProtoField walk.
type scanResult int

const (
	// scanOK — the whole stream parsed cleanly.
	scanOK scanResult = iota
	// scanSkipFailed — a tag decode, or a non-target field skip via
	// ConsumeFieldValue, failed partway through. The original singular
	// and repeated scanners both simply stopped at this point (the
	// repeated scanner returned the array collected so far without an
	// error).
	scanSkipFailed
	// scanTargetFailed — decoding an occurrence of the target field
	// itself failed; targetWire on the result records which wire type.
	// The repeated scanner turned this into a wire-type-specific error.
	scanTargetFailed
)

// scanProtoField walks the wire stream `raw` once and collects every
// occurrence of field number `target`, in serialised order. Fields
// with other tags (and target occurrences whose wire type is neither
// varint / fixed32 / fixed64 / length-delimited, e.g. start-group)
// are skipped via protowire.ConsumeFieldValue.
//
// On a clean parse `res` is scanOK. On a malformed stream `entries`
// holds the occurrences gathered before the malformation and `res`
// distinguishes a non-target skip failure (scanSkipFailed) from a
// failure decoding a target occurrence (scanTargetFailed); in the
// latter case `targetWire` is the offending occurrence's wire type.
// This single scan is the shared core of BindGetProtoField (which
// keeps the last occurrence per proto3 scalar semantics) and
// BindGetProtoFieldRepeated (which collects every occurrence).
func scanProtoField(raw []byte, target protowire.Number) (entries []wireEntry, res scanResult, targetWire protowire.Type) {
	b := raw
	for len(b) > 0 {
		tag, wire, n := protowire.ConsumeTag(b)
		if n < 0 {
			return entries, scanSkipFailed, 0
		}
		b = b[n:]
		if tag != target {
			n = protowire.ConsumeFieldValue(tag, wire, b)
			if n < 0 {
				return entries, scanSkipFailed, 0
			}
			b = b[n:]
			continue
		}
		switch wire {
		case protowire.VarintType:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return entries, scanTargetFailed, wire
			}
			b = b[n:]
			entries = append(entries, wireEntry{Wire: wire, Varint: v})
		case protowire.Fixed32Type:
			v, n := protowire.ConsumeFixed32(b)
			if n < 0 {
				return entries, scanTargetFailed, wire
			}
			b = b[n:]
			entries = append(entries, wireEntry{Wire: wire, Fixed: uint64(v)})
		case protowire.Fixed64Type:
			v, n := protowire.ConsumeFixed64(b)
			if n < 0 {
				return entries, scanTargetFailed, wire
			}
			b = b[n:]
			entries = append(entries, wireEntry{Wire: wire, Fixed: v})
		case protowire.BytesType:
			payload, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return entries, scanTargetFailed, wire
			}
			b = b[n:]
			// Copy the payload so later iterations don't clobber it.
			cp := make([]byte, len(payload))
			copy(cp, payload)
			entries = append(entries, wireEntry{Wire: wire, Bytes: cp})
		default:
			n = protowire.ConsumeFieldValue(tag, wire, b)
			if n < 0 {
				// The original repeated scanner reported this as
				// "malformed wire type N" (a target-tag decode failure),
				// while the singular scanner fell through to default.
				return entries, scanTargetFailed, wire
			}
			b = b[n:]
		}
	}
	return entries, scanOK, 0
}

func BindGetProtoField(args ...value.Value) (value.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("GET_PROTO_FIELD: invalid number of arguments: got %d, want at least 3", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		// Empty message — fall through to default handling.
		return decodeDefault(args)
	}
	num, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	target := protowire.Number(num)
	kind, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	entries, res, _ := scanProtoField(raw, target)
	if res != scanOK {
		// Malformed wire stream — fall through to default handling,
		// exactly as the original single-pass scanner did on the first
		// protowire decode error (it returned decodeDefault(args) for
		// every failing decode, target or skip alike).
		return decodeDefault(args)
	}
	if len(entries) == 0 {
		return decodeDefault(args)
	}
	// proto3 scalar semantics: the last serialised occurrence wins for
	// non-repeated fields. The original scanner kept three independent
	// "last" slots (one per wire category); reproduce that by scanning
	// the collected entries for the most recent occurrence of each
	// category.
	var (
		lastVal     uint64
		lastBytes   []byte
		lastTypeFix uint64
	)
	for _, e := range entries {
		switch e.Wire {
		case protowire.VarintType:
			lastVal = e.Varint
		case protowire.Fixed32Type, protowire.Fixed64Type:
			lastTypeFix = e.Fixed
		case protowire.BytesType:
			lastBytes = e.Bytes
		}
	}
	switch kind {
	case kindBool:
		return value.BoolValue(lastVal != 0), nil
	case kindInt32, kindInt64:
		return value.IntValue(int64(lastVal)), nil
	case kindUint32, kindUint64:
		return value.IntValue(int64(lastVal)), nil
	case kindFloat:
		return value.FloatValue(float64(math.Float32frombits(uint32(lastTypeFix)))), nil
	case kindDouble:
		return value.FloatValue(math.Float64frombits(lastTypeFix)), nil
	case kindString:
		return value.StringValue(string(lastBytes)), nil
	case kindBytes:
		return value.BytesValue(lastBytes), nil
	case kindEnum:
		return value.IntValue(int64(lastVal)), nil
	case kindMessage:
		return value.BytesValue(lastBytes), nil
	}
	_ = helper.ExistsNull
	return nil, fmt.Errorf("GET_PROTO_FIELD: unsupported type kind %q", kind)
}

// BindGetProtoFieldRepeated extracts every occurrence of a repeated
// proto field and returns the values as an ArrayValue. args =
// (proto_bytes, field_number, element_kind). Used for map fields
// (lowered as ARRAY<STRUCT<key, value>>) and other repeated proto
// fields where the upstream analyzer rewrites consumers to iterate
// the array via googlesqlite_decode_array + json_each.
//
// For repeated message / map fields each element is the entry's
// wire-format payload bytes (so consumers can recursively GetProtoField
// the entry's `key` / `value` sub-fields). For primitive repeated
// fields each element is decoded to the matching SQL value kind.
func BindGetProtoFieldRepeated(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("GET_PROTO_FIELD_REPEATED: invalid number of arguments: got %d, want 3", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	num, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	target := protowire.Number(num)
	kind, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	entries, res, targetWire := scanProtoField(raw, target)
	switch res {
	case scanTargetFailed:
		// The original repeated scanner returned a wire-type-specific
		// error when decoding an occurrence of the target field failed.
		switch targetWire {
		case protowire.VarintType:
			return nil, fmt.Errorf("GET_PROTO_FIELD_REPEATED: malformed varint")
		case protowire.Fixed32Type:
			return nil, fmt.Errorf("GET_PROTO_FIELD_REPEATED: malformed fixed32")
		case protowire.Fixed64Type:
			return nil, fmt.Errorf("GET_PROTO_FIELD_REPEATED: malformed fixed64")
		case protowire.BytesType:
			return nil, fmt.Errorf("GET_PROTO_FIELD_REPEATED: malformed bytes")
		default:
			return nil, fmt.Errorf("GET_PROTO_FIELD_REPEATED: malformed wire type %d", targetWire)
		}
	default:
		// scanOK or scanSkipFailed: the original loop, on a ConsumeTag
		// or non-target skip failure, simply broke out and returned the
		// array collected so far without an error.
		return entriesToRepeatedArray(entries, kind), nil
	}
}

// entriesToRepeatedArray decodes each collected wire occurrence into a
// SQL value typed for the repeated element `kind`, preserving the
// original per-occurrence collect-all behaviour. Packed repeated
// fields are not unpacked here — a length-delimited occurrence is
// surfaced as a single element, matching the original scanner.
func entriesToRepeatedArray(entries []wireEntry, kind string) *value.ArrayValue {
	out := &value.ArrayValue{}
	for _, e := range entries {
		switch e.Wire {
		case protowire.VarintType:
			switch kind {
			case kindBool:
				out.Values = append(out.Values, value.BoolValue(e.Varint != 0))
			case kindInt32, kindInt64, kindUint32, kindUint64, kindEnum:
				out.Values = append(out.Values, value.IntValue(int64(e.Varint)))
			default:
				out.Values = append(out.Values, value.IntValue(int64(e.Varint)))
			}
		case protowire.Fixed32Type:
			if kind == kindFloat {
				out.Values = append(out.Values, value.FloatValue(float64(math.Float32frombits(uint32(e.Fixed)))))
			} else {
				out.Values = append(out.Values, value.IntValue(int64(uint32(e.Fixed))))
			}
		case protowire.Fixed64Type:
			if kind == kindDouble {
				out.Values = append(out.Values, value.FloatValue(math.Float64frombits(e.Fixed)))
			} else {
				out.Values = append(out.Values, value.IntValue(int64(e.Fixed)))
			}
		case protowire.BytesType:
			switch kind {
			case kindString:
				out.Values = append(out.Values, value.StringValue(string(e.Bytes)))
			case kindBytes, kindMessage:
				out.Values = append(out.Values, value.BytesValue(e.Bytes))
			default:
				out.Values = append(out.Values, value.BytesValue(e.Bytes))
			}
		}
	}
	return out
}

// BindFromProto unwraps a google.protobuf.<Wrapper> proto value to
// its SQL primitive. The first argument is the proto wire bytes; the
// second is the target SQL kind (one of bool/int32/int64/uint32/
// uint64/float/double/string/bytes/timestamp/date).
//
// For the wrapper types (Int32Value, Int64Value, ..., StringValue,
// BoolValue, BytesValue) the proto layout is one field with tag 1
// carrying the wrapped value. google.protobuf.Timestamp /
// google.protobuf.Date use two fields (seconds + nanos / year +
// month + day); the runtime composes those into the right SQL type.
func BindFromProto(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FROM_PROTO: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	kind, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	// Identity passthrough: when the analyzer's overload resolution
	// surfaces a non-proto input matching the target kind, our runtime
	// receives the SQL value directly (DateValue, TimestampValue,
	// IntValue, etc.) instead of proto wire bytes. Upstream BigQuery
	// behaves the same for examples like `FROM_PROTO(DATE '2019-10-30')`.
	switch v := args[0].(type) {
	case value.DateValue:
		if kind == "date" {
			return v, nil
		}
	case value.TimestampValue:
		if kind == "timestamp" {
			return v, nil
		}
	case value.IntValue, value.FloatValue, value.BoolValue, value.StringValue:
		// Wrapper-proto SQL primitives flow through unchanged: the
		// analyzer typed the FROM_PROTO call's return type to match
		// the input.
		return args[0], nil
	}
	raw, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	switch kind {
	case "timestamp":
		seconds, nanos := readTimestampFields(raw)
		t := protoTimestampToTime(seconds, nanos)
		return value.TimestampValue(t), nil
	case "date":
		year, month, day := readDateFields(raw)
		t := protoDateToTime(year, month, day)
		return value.DateValue(t), nil
	}
	// Wrapper types: single tag-1 field carrying the value.
	v, err := readSingleField(raw, 1)
	if err != nil {
		return nil, err
	}
	if v == nil {
		// Wrapper proto with no field set → default zero value per
		// proto3 semantics, except STRING / BYTES which become empty.
		return zeroValueForKind(kind), nil
	}
	return decodeWireValueAsKind(v, kind)
}

// BindToProto wraps a SQL primitive into the well-known proto wrapper
// wire bytes. Inverse of BindFromProto.
func BindToProto(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TO_PROTO: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	kind, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	// Identity passthrough: TO_PROTO of an already-proto value (`new
	// MessageName(...)`) emits the wire bytes our MakeProto runtime
	// already produced. Detect by looking at args[0]'s type — for
	// proto values it's a BytesValue holding the wire bytes.
	if kind == "message" {
		if b, ok := args[0].(value.BytesValue); ok {
			return b, nil
		}
		if raw, err := args[0].ToBytes(); err == nil {
			return value.BytesValue(raw), nil
		}
	}
	switch kind {
	case "timestamp":
		t, err := args[0].ToTime()
		if err != nil {
			return nil, err
		}
		return value.BytesValue(encodeTimestamp(t)), nil
	case "date":
		t, err := args[0].ToTime()
		if err != nil {
			return nil, err
		}
		return value.BytesValue(encodeDate(t)), nil
	}
	return value.BytesValue(encodeSingleField(1, args[0], kind)), nil
}

// BindMakeProto materialises the wire-format bytes for a proto message
// from a flat (tag, kind, value, tag, kind, value, ...) argument tuple.
// The formatter lowers `new MessageName(value1 AS field1, ...)` to
// `googlesqlite_make_proto(tag1, 'kind1', value1, tag2, 'kind2', value2, ...)`.
// Each (tag, kind, value) is encoded via encodeSingleField and the
// outputs are concatenated, producing standards-compliant proto3 wire
// bytes that downstream FROM_PROTO / EXTRACT / etc. can decode.
//
// NULL or zero-valued fields are encoded as their proto-default
// according to encodePayload; this matches the upstream "fields not
// set keep their default" semantics for proto3.
func BindMakeProto(args ...value.Value) (value.Value, error) {
	if len(args)%3 != 0 {
		return nil, fmt.Errorf("MAKE_PROTO: expected (tag, kind, value) triples, got %d args", len(args))
	}
	var out []byte
	for i := 0; i < len(args); i += 3 {
		if args[i] == nil || args[i+1] == nil {
			return nil, fmt.Errorf("MAKE_PROTO: tag/kind cannot be NULL")
		}
		tag, err := args[i].ToInt64()
		if err != nil {
			return nil, err
		}
		kind, err := args[i+1].ToString()
		if err != nil {
			return nil, err
		}
		out = append(out, encodeSingleField(int(tag), args[i+2], kind)...)
	}
	return value.BytesValue(out), nil
}

// BindFilterFields returns proto bytes containing only / dropping
// the fields whose dotted paths appear in the second argument. The
// path list is comma-separated, each entry prefixed by either `+`
// (include) or `-` (exclude); dotted paths walk into submessages.
// When the list contains any include entry, the runtime treats it
// as an include-only filter; otherwise it treats it as an
// exclude-list (production semantics).
func BindFilterFields(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("FILTER_FIELDS: invalid number of arguments: got %d, want at least 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	pathSpec, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	resetRequired := false
	protoName := ""
	if len(args) >= 3 && args[2] != nil {
		if b, err := args[2].ToBool(); err == nil {
			resetRequired = b
		}
	}
	if len(args) >= 4 && args[3] != nil {
		if s, err := args[3].ToString(); err == nil {
			protoName = s
		}
	}
	includes, excludes := parseFieldFilterPaths(pathSpec)
	tree := buildFilterTree(includes, excludes)
	filtered := applyFilterTree(raw, tree)
	if resetRequired && protoName != "" {
		filtered = ensureRequiredFieldDefaults(filtered, protoName, tree)
	}
	return value.BytesValue(filtered), nil
}

// filterNode is a single node in the include/exclude tree built
// from the comma-separated dotted-path spec. Leaves carry the
// terminal include / exclude flag; intermediate nodes carry per-
// child path decisions.
type filterNode struct {
	include  bool
	exclude  bool
	children map[uint64]*filterNode
}

func (n *filterNode) leafInclude() bool { return n != nil && n.include }
func (n *filterNode) leafExclude() bool { return n != nil && n.exclude }

func parseFieldFilterPaths(spec string) (includes, excludes [][]uint64) {
	for _, part := range strings.Split(spec, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		include := true
		switch p[0] {
		case '+':
			include = true
			p = strings.TrimSpace(p[1:])
		case '-':
			include = false
			p = strings.TrimSpace(p[1:])
		}
		segs := strings.Split(p, ".")
		path := make([]uint64, 0, len(segs))
		for _, s := range segs {
			var n uint64
			_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
			if err != nil || n == 0 {
				path = nil
				break
			}
			path = append(path, n)
		}
		if len(path) == 0 {
			continue
		}
		if include {
			includes = append(includes, path)
		} else {
			excludes = append(excludes, path)
		}
	}
	return
}

func buildFilterTree(includes, excludes [][]uint64) *filterNode {
	root := &filterNode{children: map[uint64]*filterNode{}}
	insert := func(path []uint64, include bool) {
		cur := root
		for i, seg := range path {
			if cur.children == nil {
				cur.children = map[uint64]*filterNode{}
			}
			child, ok := cur.children[seg]
			if !ok {
				child = &filterNode{}
				cur.children[seg] = child
			}
			if i == len(path)-1 {
				if include {
					child.include = true
				} else {
					child.exclude = true
				}
			}
			cur = child
		}
	}
	for _, p := range includes {
		insert(p, true)
	}
	for _, p := range excludes {
		insert(p, false)
	}
	return root
}

// applyFilterTree walks the proto wire stream and emits only the
// fields covered by the include set (or omits fields covered by the
// exclude set when no includes are present). Submessage fields are
// descended into so nested paths apply.
func applyFilterTree(raw []byte, root *filterNode) []byte {
	out := []byte{}
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return out
		}
		head := raw[:n]
		raw = raw[n:]
		valLen := protowire.ConsumeFieldValue(tag, wire, raw)
		if valLen < 0 {
			return out
		}
		val := raw[:valLen]
		raw = raw[valLen:]
		child := root.children[uint64(tag)]
		if treeHasIncludeBelow(root) {
			// Include-only mode: only keep tags reachable from an
			// include entry.
			if child == nil {
				continue
			}
			// An include node with no further children means "emit the
			// whole subtree verbatim" — there are no nested filters to
			// apply.
			if child.leafInclude() && len(child.children) == 0 {
				out = append(out, head...)
				out = append(out, val...)
				continue
			}
			// Submessage with nested includes / excludes: recurse.
			// Two cases:
			//   - this child has `include: true` (e.g. `+type`) AND
			//     nested excludes (e.g. `-type.award_name`): every
			//     field is implicitly kept except the excluded ones.
			//   - this child has no `include: true` (e.g. `+type.x`):
			//     only fields explicitly listed underneath are kept.
			if wire == protowire.BytesType {
				inner, _ := protowire.ConsumeBytes(val)
				var rebuilt []byte
				if child.include {
					rebuilt = applyFilterTreeIncludeWithExcludes(inner, child)
				} else {
					rebuilt = applyFilterTree(inner, child)
				}
				out = append(out, head...)
				out = protowire.AppendVarint(out, uint64(len(rebuilt)))
				out = append(out, rebuilt...)
			}
			continue
		}
		// Exclude mode: omit tags that match a leaf exclude.
		if child != nil && child.leafExclude() {
			continue
		}
		if child != nil && wire == protowire.BytesType && len(child.children) > 0 {
			inner, _ := protowire.ConsumeBytes(val)
			rebuilt := applyFilterTree(inner, child)
			out = append(out, head...)
			out = protowire.AppendVarint(out, uint64(len(rebuilt)))
			out = append(out, rebuilt...)
			continue
		}
		out = append(out, head...)
		out = append(out, val...)
	}
	return out
}

// RequiredField describes one top-level required field of a proto
// message, used by FILTER_FIELDS to honour the
// `RESET_CLEARED_REQUIRED_FIELDS => TRUE` named argument: when set,
// required fields that the filter would have dropped must reappear
// in the output with their proto-defined default values.
type RequiredField struct {
	Number  uint64
	Wire    uint8 // protowire.Type cast to uint8
	Payload []byte
}

// RequiredFieldResolver is the hook the parent driver supplies to
// expose the descriptor-pool view of required fields. internal/
// proto_registry.go wires the real resolver against the Go-side
// proto descriptor pool. Tests can leave it nil; the runtime then
// skips the reset behaviour silently.
var RequiredFieldResolver func(fullName string) []RequiredField

// ensureRequiredFieldDefaults inspects `filtered` for the listed
// required fields and re-emits any missing ones with the proto's
// declared default values, in field-number order. The output is the
// filtered bytes with the defaults appended at the end.
func ensureRequiredFieldDefaults(filtered []byte, fullName string, tree *filterNode) []byte {
	if RequiredFieldResolver == nil {
		return filtered
	}
	required := RequiredFieldResolver(fullName)
	if len(required) == 0 {
		return filtered
	}
	present := map[uint64]bool{}
	b := filtered
	for len(b) > 0 {
		tag, wire, n := protowire.ConsumeTag(b)
		if n < 0 {
			break
		}
		present[uint64(tag)] = true
		b = b[n:]
		n = protowire.ConsumeFieldValue(tag, wire, b)
		if n < 0 {
			break
		}
		b = b[n:]
	}
	out := append([]byte(nil), filtered...)
	for _, r := range required {
		if present[r.Number] {
			continue
		}
		// Skip required fields that the user explicitly excluded.
		if child, ok := tree.children[r.Number]; ok && child.leafExclude() {
			continue
		}
		out = protowire.AppendTag(out, protowire.Number(r.Number), protowire.Type(r.Wire))
		out = append(out, r.Payload...)
	}
	return out
}

// applyFilterTreeIncludeWithExcludes is the recursive helper for the
// `+parent, -parent.child` case. The parent is included (so by
// default emit every sub-field) but the explicit excludes underneath
// `parent` drop matching children. New include entries deeper in the
// tree are honoured the same way.
func applyFilterTreeIncludeWithExcludes(raw []byte, node *filterNode) []byte {
	out := []byte{}
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return out
		}
		head := raw[:n]
		raw = raw[n:]
		valLen := protowire.ConsumeFieldValue(tag, wire, raw)
		if valLen < 0 {
			return out
		}
		val := raw[:valLen]
		raw = raw[valLen:]
		child := node.children[uint64(tag)]
		if child != nil && child.leafExclude() {
			continue
		}
		if child != nil && wire == protowire.BytesType && len(child.children) > 0 {
			inner, _ := protowire.ConsumeBytes(val)
			rebuilt := applyFilterTreeIncludeWithExcludes(inner, child)
			out = append(out, head...)
			out = protowire.AppendVarint(out, uint64(len(rebuilt)))
			out = append(out, rebuilt...)
			continue
		}
		// No exclusion at this leaf → keep verbatim.
		out = append(out, head...)
		out = append(out, val...)
	}
	return out
}

func treeHasIncludeBelow(n *filterNode) bool {
	if n == nil {
		return false
	}
	if n.include {
		return true
	}
	for _, c := range n.children {
		if treeHasIncludeBelow(c) {
			return true
		}
	}
	return false
}

// BindReplaceFields rewrites a proto blob with one field replaced.
// args = (proto, path_string, kind, new_value).
//
// `path_string` is the dotted field-number path emitted by the
// formatter — flat "1" for a top-level field, "3.2" for the field
// numbered 2 inside the submessage at field 3, etc. The runtime
// descends through each path segment, stripping + rewriting only
// the leaf field while preserving the surrounding bytes.
//
// Multi-pair REPLACE_FIELDS lower into nested
// REPLACE_FIELDS(REPLACE_FIELDS(...)) calls at the formatter level
// (see internal/formatter.go), so the runtime never needs to
// process more than one (path, kind, value) per call.
func BindReplaceFields(args ...value.Value) (value.Value, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("REPLACE_FIELDS: invalid number of arguments: got %d, want 4", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	path, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	kind, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	segs, ok := parseFieldNumberPath(path)
	if !ok || len(segs) == 0 {
		return value.BytesValue(raw), nil
	}
	out, err := replaceFieldAtPath(raw, segs, kind, args[3])
	if err != nil {
		return nil, err
	}
	return value.BytesValue(out), nil
}

// elementKindForRepeatedKind maps an ARRAY-typed leaf kind to the
// element kind the wire encoder should use. The formatter reports
// the SQL type of the replacement value as `array` for repeated
// fields; without the unwrap each element would be re-encoded as an
// opaque BYTES blob. The element kind matches the SQL type of the
// inner literal, which our formatter currently does not surface, so
// we default to "string" — the only case the upstream Examples
// exercise. Improving this requires the formatter to inspect the
// FieldDescriptor's TypeName / Kind for repeated fields and pass it
// in as a distinct argument.
func elementKindForRepeatedKind(arrayKind string) string {
	if arrayKind == "array" {
		return kindString
	}
	return arrayKind
}

// parseFieldNumberPath splits "3.2.1" → [3, 2, 1]. Returns ok=false
// when any segment fails to parse as a positive integer.
func parseFieldNumberPath(s string) ([]protowire.Number, bool) {
	if s == "" {
		return nil, false
	}
	parts := strings.Split(s, ".")
	out := make([]protowire.Number, 0, len(parts))
	for _, p := range parts {
		var n int64
		_, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &n)
		if err != nil || n <= 0 {
			return nil, false
		}
		out = append(out, protowire.Number(n))
	}
	return out, true
}

// replaceFieldAtPath descends into submessage tags along `segs`,
// rewriting the leaf field to `(kind, value)`. The non-target
// fields at every level are preserved by re-emitting them verbatim.
func replaceFieldAtPath(raw []byte, segs []protowire.Number, kind string, val value.Value) ([]byte, error) {
	if len(segs) == 1 {
		// Strip every existing occurrence of segs[0]. NULL clears the
		// field without re-emitting anything. ArrayValue (repeated
		// field replacement) emits one field-tag per element so the
		// upstream `[a, b, c] AS reviews` shape lands as repeated
		// `reviews:` wire fields.
		out := walkAndFilter(raw, map[protowire.Number]struct{}{segs[0]: {}}, false)
		if val == nil {
			return out, nil
		}
		if arr, ok := val.(*value.ArrayValue); ok && arr != nil {
			elemKind := elementKindForRepeatedKind(kind)
			for _, ev := range arr.Values {
				if ev == nil {
					continue
				}
				out = append(out, encodeSingleField(int(segs[0]), ev, elemKind)...)
			}
			return out, nil
		}
		out = append(out, encodeSingleField(int(segs[0]), val, kind)...)
		return out, nil
	}
	target := segs[0]
	out := []byte{}
	matched := false
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return out, fmt.Errorf("REPLACE_FIELDS: malformed proto")
		}
		head := raw[:n]
		raw = raw[n:]
		valLen := protowire.ConsumeFieldValue(tag, wire, raw)
		if valLen < 0 {
			return out, fmt.Errorf("REPLACE_FIELDS: malformed field")
		}
		valBytes := raw[:valLen]
		raw = raw[valLen:]
		if tag != target || wire != protowire.BytesType {
			out = append(out, head...)
			out = append(out, valBytes...)
			continue
		}
		matched = true
		inner, _ := protowire.ConsumeBytes(valBytes)
		rebuilt, err := replaceFieldAtPath(inner, segs[1:], kind, val)
		if err != nil {
			return out, err
		}
		out = append(out, head...)
		out = protowire.AppendVarint(out, uint64(len(rebuilt)))
		out = append(out, rebuilt...)
	}
	if !matched {
		// The submessage didn't exist — synthesise one carrying the
		// leaf replacement so callers can build proto values from
		// scratch.
		fresh, err := replaceFieldAtPath(nil, segs[1:], kind, val)
		if err != nil {
			return out, err
		}
		header := tagForKind(int(target), "message")
		out = append(out, header...)
		out = protowire.AppendVarint(out, uint64(len(fresh)))
		out = append(out, fresh...)
	}
	return out, nil
}

// BindProtoMapContainsKey reports whether a proto map<K,V> field in
// the given parent message contains the supplied key. The formatter
// lowers `PROTO_MAP_CONTAINS_KEY(parent.map_field, key)` to a call
// shape of (parent_proto_bytes, map_field_tag, key_kind, key_value)
// so the runtime can walk every occurrence of the map field's outer
// tag rather than relying on a non-accumulating GetProtoField
// extraction (which only retains the last entry's payload).
func BindProtoMapContainsKey(args ...value.Value) (value.Value, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("PROTO_MAP_CONTAINS_KEY: invalid number of arguments: got %d, want 4", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	raw, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	mapTag, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	target := protowire.Number(mapTag)
	keyKind, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	wantKey := args[3]
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			break
		}
		raw = raw[n:]
		if tag != target {
			n = protowire.ConsumeFieldValue(tag, wire, raw)
			if n < 0 {
				break
			}
			raw = raw[n:]
			continue
		}
		if wire != protowire.BytesType {
			n = protowire.ConsumeFieldValue(tag, wire, raw)
			if n < 0 {
				break
			}
			raw = raw[n:]
			continue
		}
		entry, n := protowire.ConsumeBytes(raw)
		if n < 0 {
			break
		}
		raw = raw[n:]
		k, _ := readSingleField(entry, 1)
		if k == nil {
			continue
		}
		v, err := decodeWireValueAsKind(k, keyKind)
		if err != nil {
			continue
		}
		eq, err := v.EQ(wantKey)
		if err == nil && eq {
			return value.BoolValue(true), nil
		}
	}
	return value.BoolValue(false), nil
}

// BindProtoModifyMap deletes / inserts / replaces map entries in a
// proto map<K,V> field. Args layout produced by the formatter:
//
//	args[0] = parent_proto_bytes
//	args[1] = map_field_tag (INT64)
//	args[2] = key_kind   (STRING — "string"/"int64"/...)
//	args[3] = value_kind (STRING)
//	args[4..]  = key1, value1, key2, value2, ...
//
// Per the upstream spec:
//
//   - key is NULL → error
//   - duplicate keys in arg list → error
//   - value is NULL → delete that key from the map
//   - otherwise → insert or replace
//
// The return type is ARRAY<PROTO<MapEntry>> (the map field itself,
// matching the shape produced by GetProtoFieldRepeated for the same
// field): an ArrayValue whose elements are BytesValues, each carrying
// the inner wire-format payload of one map entry.
func BindProtoModifyMap(args ...value.Value) (value.Value, error) {
	if len(args) < 6 || (len(args)-4)%2 != 0 {
		return nil, fmt.Errorf("PROTO_MODIFY_MAP: invalid number of arguments: got %d, want at least 6 with an even number of key/value pairs", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	parent, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	mapTagNum, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	mapTag := protowire.Number(mapTagNum)
	keyKind, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	valKind, err := args[3].ToString()
	if err != nil {
		return nil, err
	}
	type kvOp struct {
		key   value.Value
		val   value.Value // nil ⇒ delete
		isDel bool
	}
	ops := make([]kvOp, 0, (len(args)-4)/2)
	seenKeys := make([]value.Value, 0)
	for i := 4; i+1 < len(args); i += 2 {
		k := args[i]
		v := args[i+1]
		if k == nil {
			return nil, fmt.Errorf("PROTO_MODIFY_MAP: key argument %d is NULL", i-3)
		}
		for _, prev := range seenKeys {
			if eq, err := k.EQ(prev); err == nil && eq {
				return nil, fmt.Errorf("PROTO_MODIFY_MAP: duplicate key argument %d", i-3)
			}
		}
		seenKeys = append(seenKeys, k)
		ops = append(ops, kvOp{key: k, val: v, isDel: v == nil})
	}
	keysMatchingOp := func(k value.Value) (kvOp, bool) {
		for _, op := range ops {
			if eq, err := op.key.EQ(k); err == nil && eq {
				return op, true
			}
		}
		return kvOp{}, false
	}
	entries := make([][]byte, 0)
	consumedKeys := make([]value.Value, 0, len(ops))
	b := parent
	for len(b) > 0 {
		tag, wire, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("PROTO_MODIFY_MAP: malformed tag")
		}
		b = b[n:]
		if tag != mapTag || wire != protowire.BytesType {
			n = protowire.ConsumeFieldValue(tag, wire, b)
			if n < 0 {
				return nil, fmt.Errorf("PROTO_MODIFY_MAP: malformed field value")
			}
			b = b[n:]
			continue
		}
		entry, n := protowire.ConsumeBytes(b)
		if n < 0 {
			return nil, fmt.Errorf("PROTO_MODIFY_MAP: malformed map entry")
		}
		b = b[n:]
		kBytes, _ := readSingleField(entry, 1)
		var entryKey value.Value
		if kBytes != nil {
			entryKey, _ = decodeWireValueAsKind(kBytes, keyKind)
		}
		if entryKey != nil {
			if op, ok := keysMatchingOp(entryKey); ok {
				consumedKeys = append(consumedKeys, op.key)
				if op.isDel {
					continue
				}
				entries = append(entries, buildMapEntry(op.key, op.val, keyKind, valKind))
				continue
			}
		}
		cp := make([]byte, len(entry))
		copy(cp, entry)
		entries = append(entries, cp)
	}
	for _, op := range ops {
		seen := false
		for _, k := range consumedKeys {
			if eq, err := op.key.EQ(k); err == nil && eq {
				seen = true
				break
			}
		}
		if seen || op.isDel {
			continue
		}
		entries = append(entries, buildMapEntry(op.key, op.val, keyKind, valKind))
	}
	arr := &value.ArrayValue{Values: make([]value.Value, 0, len(entries))}
	for _, e := range entries {
		arr.Values = append(arr.Values, value.BytesValue(e))
	}
	return arr, nil
}

// buildMapEntry encodes a `key, value` pair as the inner wire bytes
// of a map entry: field 1 carries the key, field 2 carries the value,
// both typed according to the key/value kind strings.
func buildMapEntry(key, val value.Value, keyKind, valKind string) []byte {
	keyTag := tagForKind(1, keyKind)
	valTag := tagForKind(2, valKind)
	keyPayload := encodePayload(key, keyKind)
	valPayload := encodePayload(val, valKind)
	out := make([]byte, 0, len(keyTag)+len(keyPayload)+len(valTag)+len(valPayload))
	out = append(out, keyTag...)
	out = append(out, keyPayload...)
	out = append(out, valTag...)
	out = append(out, valPayload...)
	return out
}

// BindEnumValueDescriptorProto returns an EnumValueDescriptorProto
// for the given enum-number input. args = (enum_number, enum_full_name).
//
// The production proto carries the enum value's name + number;
// `enum_full_name` is the dotted full name of the enum type passed
// in by the formatter (looked up via the catalog's registered
// EnumType handles). If the enum is registered we look up the enum
// value name through the global proto registry's Descriptor and
// emit field 1 (name) as a STRING; field 2 (number) is always
// emitted as INT32.
func BindEnumValueDescriptorProto(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ENUM_VALUE_DESCRIPTOR_PROTO: missing argument")
	}
	if args[0] == nil {
		return nil, nil
	}
	n, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	enumName := ""
	if len(args) >= 2 && args[1] != nil {
		if s, err := args[1].ToString(); err == nil {
			enumName = s
		}
	}
	out := []byte{}
	if name := lookupEnumValueName(enumName, int32(n)); name != "" {
		// field 1 (name) = STRING, wire = BytesType.
		out = append(out, 0x0a)
		out = protowire.AppendVarint(out, uint64(len(name)))
		out = append(out, name...)
	}
	// field 2 (number) = INT32, wire = VarintType.
	out = append(out, 0x10)
	out = protowire.AppendVarint(out, uint64(n))
	return value.BytesValue(out), nil
}

// lookupEnumValueName is patched at init time by internal/catalog to
// route enum-name resolution through the Catalog's registered enum
// types. Default returns "" (number-only encoding).
var lookupEnumValueName = func(enumFullName string, number int32) string {
	return ""
}

// SetEnumValueNameLookup installs the production enum-name lookup
// used by BindEnumValueDescriptorProto. The internal package wires
// this at init time so the proto package stays
// dependency-free from the catalog.
func SetEnumValueNameLookup(fn func(enumFullName string, number int32) string) {
	if fn != nil {
		lookupEnumValueName = fn
	}
}

// readTimestampFields reads google.protobuf.Timestamp { int64 seconds
// = 1; int32 nanos = 2; } from the supplied bytes.
func readTimestampFields(raw []byte) (int64, int32) {
	var seconds int64
	var nanos int32
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			break
		}
		raw = raw[n:]
		switch {
		case tag == 1 && wire == protowire.VarintType:
			v, n := protowire.ConsumeVarint(raw)
			if n < 0 {
				return seconds, nanos
			}
			seconds = int64(v)
			raw = raw[n:]
		case tag == 2 && wire == protowire.VarintType:
			v, n := protowire.ConsumeVarint(raw)
			if n < 0 {
				return seconds, nanos
			}
			nanos = int32(v)
			raw = raw[n:]
		default:
			n = protowire.ConsumeFieldValue(tag, wire, raw)
			if n < 0 {
				return seconds, nanos
			}
			raw = raw[n:]
		}
	}
	return seconds, nanos
}

func readDateFields(raw []byte) (int32, int32, int32) {
	var year, month, day int32
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			break
		}
		raw = raw[n:]
		if wire == protowire.VarintType {
			v, k := protowire.ConsumeVarint(raw)
			if k < 0 {
				break
			}
			raw = raw[k:]
			switch tag {
			case 1:
				year = int32(v)
			case 2:
				month = int32(v)
			case 3:
				day = int32(v)
			}
		} else {
			n = protowire.ConsumeFieldValue(tag, wire, raw)
			if n < 0 {
				break
			}
			raw = raw[n:]
		}
	}
	return year, month, day
}

// readSingleField returns the single value carried in the proto bytes
// for the supplied tag, choosing the last occurrence (proto3
// semantics) and ignoring the wire type — the caller knows the kind.
func readSingleField(raw []byte, num protowire.Number) (any, error) {
	var last any
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return nil, fmt.Errorf("malformed proto")
		}
		raw = raw[n:]
		if tag != num {
			n = protowire.ConsumeFieldValue(tag, wire, raw)
			if n < 0 {
				return nil, fmt.Errorf("malformed proto")
			}
			raw = raw[n:]
			continue
		}
		switch wire {
		case protowire.VarintType:
			v, n := protowire.ConsumeVarint(raw)
			if n < 0 {
				return nil, fmt.Errorf("malformed varint")
			}
			raw = raw[n:]
			last = v
		case protowire.Fixed32Type:
			v, n := protowire.ConsumeFixed32(raw)
			if n < 0 {
				return nil, fmt.Errorf("malformed fixed32")
			}
			raw = raw[n:]
			last = v
		case protowire.Fixed64Type:
			v, n := protowire.ConsumeFixed64(raw)
			if n < 0 {
				return nil, fmt.Errorf("malformed fixed64")
			}
			raw = raw[n:]
			last = v
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(raw)
			if n < 0 {
				return nil, fmt.Errorf("malformed bytes")
			}
			raw = raw[n:]
			cp := make([]byte, len(v))
			copy(cp, v)
			last = cp
		default:
			n = protowire.ConsumeFieldValue(tag, wire, raw)
			if n < 0 {
				return nil, fmt.Errorf("malformed proto")
			}
			raw = raw[n:]
		}
	}
	return last, nil
}

func decodeWireValueAsKind(v any, kind string) (value.Value, error) {
	switch kind {
	case kindBool:
		if u, ok := v.(uint64); ok {
			return value.BoolValue(u != 0), nil
		}
	case kindInt32, kindInt64, kindUint32, kindUint64, kindEnum:
		if u, ok := v.(uint64); ok {
			return value.IntValue(int64(u)), nil
		}
	case kindFloat:
		if u, ok := v.(uint32); ok {
			return value.FloatValue(float64(math.Float32frombits(u))), nil
		}
	case kindDouble:
		if u, ok := v.(uint64); ok {
			return value.FloatValue(math.Float64frombits(u)), nil
		}
	case kindString:
		if b, ok := v.([]byte); ok {
			return value.StringValue(string(b)), nil
		}
	case kindBytes, kindMessage:
		if b, ok := v.([]byte); ok {
			return value.BytesValue(b), nil
		}
	}
	return nil, fmt.Errorf("decode: cannot convert %T to %s", v, kind)
}

func zeroValueForKind(kind string) value.Value {
	switch kind {
	case kindBool:
		return value.BoolValue(false)
	case kindInt32, kindInt64, kindUint32, kindUint64, kindEnum:
		return value.IntValue(0)
	case kindFloat, kindDouble:
		return value.FloatValue(0)
	case kindString:
		return value.StringValue("")
	case kindBytes, kindMessage:
		return value.BytesValue(nil)
	}
	return nil
}

func encodeSingleField(num int, v value.Value, kind string) []byte {
	out := tagForKind(num, kind)
	out = append(out, encodePayload(v, kind)...)
	return out
}

func encodePayload(v value.Value, kind string) []byte {
	if v == nil {
		return encodePayload(zeroValueForKind(kind), kind)
	}
	switch kind {
	case kindBool:
		b, _ := v.ToBool()
		if b {
			return protowire.AppendVarint(nil, 1)
		}
		return protowire.AppendVarint(nil, 0)
	case kindInt32, kindInt64, kindUint32, kindUint64, kindEnum:
		n, _ := v.ToInt64()
		return protowire.AppendVarint(nil, uint64(n))
	case kindFloat:
		f, _ := v.ToFloat64()
		return protowire.AppendFixed32(nil, math.Float32bits(float32(f)))
	case kindDouble:
		f, _ := v.ToFloat64()
		return protowire.AppendFixed64(nil, math.Float64bits(f))
	case kindString:
		s, _ := v.ToString()
		out := protowire.AppendVarint(nil, uint64(len(s)))
		return append(out, s...)
	case kindBytes, kindMessage:
		b, _ := v.ToBytes()
		out := protowire.AppendVarint(nil, uint64(len(b)))
		return append(out, b...)
	}
	return nil
}

func tagForKind(num int, kind string) []byte {
	var wire protowire.Type
	switch kind {
	case kindBool, kindInt32, kindInt64, kindUint32, kindUint64, kindEnum:
		wire = protowire.VarintType
	case kindFloat:
		wire = protowire.Fixed32Type
	case kindDouble:
		wire = protowire.Fixed64Type
	default:
		wire = protowire.BytesType
	}
	return protowire.AppendTag(nil, protowire.Number(num), wire)
}

func walkAndFilter(raw []byte, fieldNumbers map[protowire.Number]struct{}, keep bool) []byte {
	out := []byte{}
	for len(raw) > 0 {
		tag, wire, n := protowire.ConsumeTag(raw)
		if n < 0 {
			return out
		}
		head := raw[:n]
		raw = raw[n:]
		valLen := protowire.ConsumeFieldValue(tag, wire, raw)
		if valLen < 0 {
			return out
		}
		_, inSet := fieldNumbers[tag]
		emit := (keep && inSet) || (!keep && !inSet)
		if emit {
			out = append(out, head...)
			out = append(out, raw[:valLen]...)
		}
		raw = raw[valLen:]
	}
	return out
}

func protoTimestampToTime(seconds int64, nanos int32) gotime.Time {
	return gotime.Unix(seconds, int64(nanos)).UTC()
}

func encodeTimestamp(t gotime.Time) []byte {
	seconds := t.Unix()
	nanos := int32(t.Nanosecond())
	out := []byte{}
	if seconds != 0 {
		out = append(out, 0x08) // tag 1, varint
		out = protowire.AppendVarint(out, uint64(seconds))
	}
	if nanos != 0 {
		out = append(out, 0x10) // tag 2, varint
		out = protowire.AppendVarint(out, uint64(nanos))
	}
	return out
}

func protoDateToTime(year, month, day int32) gotime.Time {
	m := int(month)
	if m == 0 {
		m = 1
	}
	d := int(day)
	if d == 0 {
		d = 1
	}
	return gotime.Date(int(year), gotime.Month(m), d, 0, 0, 0, 0, gotime.UTC)
}

func encodeDate(t gotime.Time) []byte {
	out := []byte{}
	out = append(out, 0x08) // tag 1, varint
	out = protowire.AppendVarint(out, uint64(t.Year()))
	out = append(out, 0x10) // tag 2
	out = protowire.AppendVarint(out, uint64(t.Month()))
	out = append(out, 0x18) // tag 3
	out = protowire.AppendVarint(out, uint64(t.Day()))
	return out
}

// decodeDefault picks up the optional 4th argument (base64-encoded
// default-value blob) and decodes it according to the requested SQL
// kind. Returns nil (SQL NULL) when no default is available.
func decodeDefault(args []value.Value) (value.Value, error) {
	if len(args) < 4 || args[3] == nil {
		return nil, nil
	}
	defB64, err := args[3].ToString()
	if err != nil {
		return nil, nil
	}
	if defB64 == "" {
		return nil, nil
	}
	kind, err := args[2].ToString()
	if err != nil {
		return nil, nil
	}
	// The default is encoded as a base64'd string carrying the
	// upstream FieldDescriptor::DefaultValue* output for the
	// requested kind. We round-trip:
	//   - integer / enum kinds: base64 of the decimal string.
	//   - bool: "true" / "false".
	//   - float / double: base64 of the decimal string.
	//   - string: base64 of the literal.
	//   - bytes: base64 of the raw bytes.
	defRaw, err := base64.StdEncoding.DecodeString(defB64)
	if err != nil {
		return nil, nil
	}
	s := string(defRaw)
	switch kind {
	case kindBool:
		return value.BoolValue(s == "true"), nil
	case kindInt32, kindInt64, kindUint32, kindUint64, kindEnum:
		var n int64
		_, err := fmt.Sscanf(s, "%d", &n)
		if err != nil {
			return nil, nil
		}
		return value.IntValue(n), nil
	case kindFloat, kindDouble:
		var f float64
		_, err := fmt.Sscanf(s, "%g", &f)
		if err != nil {
			return nil, nil
		}
		return value.FloatValue(f), nil
	case kindString:
		return value.StringValue(s), nil
	case kindBytes, kindMessage:
		return value.BytesValue(defRaw), nil
	}
	return nil, nil
}
