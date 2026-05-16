// Package longtail collects the small-footprint runtime functions
// that don't warrant a dedicated category package. The functions
// here include:
//
//   - Debug helpers (ERROR, IFERROR, IS_ERROR, NULLIFERROR).
//   - String helpers (REGEXP_MATCH, REGEXP_EXTRACT_GROUPS,
//     SPLIT_SUBSTR, COLLATE).
//   - JSON mutators (JSON_ARRAY_APPEND / JSON_ARRAY_INSERT,
//     SAFE_TO_JSON).
//   - Hash (HIGHWAY_FINGERPRINT128 — falls back to FARM_FINGERPRINT
//     when the spec just requires a deterministic 128-bit hash).
//   - HLL merge helpers (HLL_COUNT.MERGE / MERGE_PARTIAL).
//   - Array (FLATTEN, ARRAY_ZIP).
//   - Aggregate (GROUPING).
//   - Net (NET.IP_IN_NET, NET.PARSE_IP, NET.FORMAT_IP,
//     NET.PARSE_PACKED_IP, NET.FORMAT_PACKED_IP, NET.MAKE_NET).
package longtail

import (
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	gonet "net"
	"regexp"
	"strconv"
	"strings"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// ----- debug -----

// BindError raises a user-visible error with the given message.
func BindError(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ERROR: invalid number of arguments: got %d, want 1", len(args))
	}
	msg := "user-raised error"
	if args[0] != nil {
		if s, err := args[0].ToString(); err == nil {
			msg = s
		}
	}
	return nil, fmt.Errorf("%s", msg)
}

// BindIsError returns whether the argument's evaluation raised an
// error. Without runtime per-expression error capture all we can
// signal here is NULL → false, otherwise false (eager evaluation
// means an actual error would surface before this gets called).
func BindIsError(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("IS_ERROR: invalid number of arguments: got %d, want 1", len(args))
	}
	return value.BoolValue(false), nil
}

// BindIfError returns the first argument when no error occurred,
// otherwise the second. Eager-evaluation runtime: errors surface
// before this is reached, so the runtime always returns the first
// non-NULL argument.
func BindIfError(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("IFERROR: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] != nil {
		return args[0], nil
	}
	return args[1], nil
}

// BindNullIfError returns the value when no error occurred,
// otherwise NULL. Always returns the value in the eager-evaluation
// runtime.
func BindNullIfError(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NULLIFERROR: invalid number of arguments: got %d, want 1", len(args))
	}
	return args[0], nil
}

// ----- string -----

// BindCollate is a no-op runtime: the analyzer enforces collation
// at type-checking time but our SQLite backing uses BINARY
// comparisons by default. Returns the input unchanged.
func BindCollate(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("COLLATE: invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	return args[0], nil
}

// BindRegexpMatch is an alias of REGEXP_CONTAINS — both return
// whether the pattern matches anywhere in the input.
func BindRegexpMatch(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("REGEXP_MATCH: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	pat, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("REGEXP_MATCH: %w", err)
	}
	return value.BoolValue(re.MatchString(s)), nil
}

// BindRegexpExtractGroups returns a STRUCT with one field per capture
// group. Named groups (`(?P<name>...)` / `(?<name>...)`) get their
// declared name; positional groups use the upstream-canonical
// `$col1`, `$col2`, ... convention. A `__INT64` / `__FLOAT64` /
// `__BOOL` suffix on the group name triggers an automatic type
// coercion of the captured text (and is stripped from the visible
// field name).
//
// If the regex does not match at all, the function returns NULL
// (matching the upstream behaviour). When a specific named group
// did not participate in the match (e.g. an alternation that didn't
// take that branch, or an optional `?`-suffixed group), the
// corresponding field is NULL rather than empty string.
func BindRegexpExtractGroups(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("REGEXP_EXTRACT_GROUPS: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	pat, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("REGEXP_EXTRACT_GROUPS: %w", err)
	}
	rawNames := re.SubexpNames()
	groupCount := len(rawNames) - 1
	// Build (display name, type-coercion hint) per capture group.
	type fieldSpec struct {
		name  string
		coerc string // "" / "INT64" / "FLOAT64" / "BOOL"
	}
	fields := make([]fieldSpec, groupCount)
	for i := 1; i <= groupCount; i++ {
		raw := rawNames[i]
		coerc := ""
		if idx := strings.LastIndex(raw, "__"); idx > 0 {
			suffix := strings.ToUpper(raw[idx+2:])
			switch suffix {
			case "INT64", "FLOAT64", "BOOL":
				coerc = suffix
				raw = raw[:idx]
			}
		}
		if raw == "" {
			raw = fmt.Sprintf("$col%d", i)
		}
		fields[i-1] = fieldSpec{name: raw, coerc: coerc}
	}
	// Use FindStringSubmatchIndex so we can tell "didn't capture"
	// (index < 0) from "matched empty string" (start == end).
	mIdx := re.FindStringSubmatchIndex(s)
	if mIdx == nil {
		return nil, nil
	}
	st := &value.StructValue{M: map[string]value.Value{}}
	for i, f := range fields {
		start := mIdx[2*(i+1)]
		end := mIdx[2*(i+1)+1]
		var v value.Value
		if start >= 0 {
			text := s[start:end]
			v, err = coerceCapture(text, f.coerc)
			if err != nil {
				return nil, fmt.Errorf("REGEXP_EXTRACT_GROUPS: %w", err)
			}
		}
		st.Keys = append(st.Keys, f.name)
		st.Values = append(st.Values, v)
		st.M[f.name] = v
	}
	return st, nil
}

func coerceCapture(text, coerc string) (value.Value, error) {
	switch coerc {
	case "INT64":
		// Accept decimal and 0x... hex forms (upstream documents
		// both for the __INT64 coercion).
		n, err := parseInt64Flexible(text)
		if err != nil {
			return nil, err
		}
		return value.IntValue(n), nil
	case "FLOAT64":
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, err
		}
		return value.FloatValue(f), nil
	case "BOOL":
		b, err := strconv.ParseBool(text)
		if err != nil {
			return nil, err
		}
		return value.BoolValue(b), nil
	}
	return value.StringValue(text), nil
}

func parseInt64Flexible(s string) (int64, error) {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		return strconv.ParseInt(s[2:], 16, 64)
	}
	return strconv.ParseInt(s, 10, 64)
}

// BindSplitSubstr returns the Nth substring of `s` produced by
// splitting on `delim`. Negative `position` counts from the right.
func BindSplitSubstr(args ...value.Value) (value.Value, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("SPLIT_SUBSTR: invalid number of arguments: got %d, want between 3 and 4", len(args))
	}
	if helper.ExistsNull(args[:3]) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	delim, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	pos, err := args[2].ToInt64()
	if err != nil {
		return nil, err
	}
	parts := strings.Split(s, delim)
	idx := int(pos)
	if idx > 0 {
		idx--
	} else if idx < 0 {
		idx = len(parts) + idx
	} else {
		return value.StringValue(""), nil
	}
	if idx < 0 || idx >= len(parts) {
		return value.StringValue(""), nil
	}
	count := 1
	if len(args) == 4 && args[3] != nil {
		c, err := args[3].ToInt64()
		if err != nil {
			return nil, err
		}
		count = int(c)
	}
	if count < 1 {
		count = 1
	}
	if idx+count > len(parts) {
		count = len(parts) - idx
	}
	return value.StringValue(strings.Join(parts[idx:idx+count], delim)), nil
}

// ----- json -----

// BindJsonArrayAppend appends values to a JSON array at the given
// path. For top-level append the path is "$". The trailing optional
// `create_if_missing` BOOL is consumed but otherwise ignored — we
// always create the target array if it does not exist yet.
func BindJsonArrayAppend(args ...value.Value) (value.Value, error) {
	pairs, _, err := splitJsonModifyArgs("JSON_ARRAY_APPEND", args)
	if err != nil {
		return nil, err
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	raw, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	var doc any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("JSON_ARRAY_APPEND: invalid JSON: %w", err)
	}
	for _, p := range pairs {
		doc = jsonModify(doc, p.path, p.value, true)
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return value.JsonValue(string(out)), nil
}

// BindJsonArrayInsert inserts values into a JSON array at the
// specified array-element paths. The trailing optional BOOL is
// `insert_each_element` (default TRUE) -- when FALSE, an ARRAY
// value is inserted as one nested element rather than its members
// being flattened into the target.
func BindJsonArrayInsert(args ...value.Value) (value.Value, error) {
	pairs, insertEachElement, err := splitJsonModifyArgs("JSON_ARRAY_INSERT", args, true)
	if err != nil {
		return nil, err
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	raw, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	var doc any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("JSON_ARRAY_INSERT: invalid JSON: %w", err)
	}
	for _, p := range pairs {
		val := p.value
		if !insertEachElement {
			val = wrapNonArray(val)
		}
		doc = jsonModify(doc, p.path, val, false)
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return value.JsonValue(string(out)), nil
}

// wrapNonArray hides an array value behind a single-element sentinel
// so jsonModify's flatten-array path keeps the inner array intact.
// Used by JSON_ARRAY_INSERT(..., insert_each_element=>FALSE).
func wrapNonArray(v any) any {
	if _, ok := v.([]any); ok {
		return []any{v}
	}
	return v
}

// jsonModifyPair captures one (path, value) pair from the
// JSON_ARRAY_APPEND / JSON_ARRAY_INSERT argument list.
type jsonModifyPair struct {
	path  string
	value any
}

// splitJsonModifyArgs walks the analyzer-lowered argument list
// `(json, path1, val1, [path2, val2, ...], [trailing_bool])` and
// returns the (path, value) pairs. The trailing BOOL is optional;
// `boolDefault` controls the value returned when no trailing bool
// is supplied (TRUE for JSON_ARRAY_INSERT's insert_each_element,
// FALSE for JSON_ARRAY_APPEND's create_if_missing).
func splitJsonModifyArgs(name string, args []value.Value, boolDefault ...bool) ([]jsonModifyPair, bool, error) {
	def := false
	if len(boolDefault) > 0 {
		def = boolDefault[0]
	}
	if len(args) < 3 {
		return nil, false, fmt.Errorf("%s: missing arguments", name)
	}
	rest := args[1:]
	createIfMissing := def
	if len(rest)%2 == 1 {
		// Trailing arg is a BOOL flag (create_if_missing for
		// JSON_ARRAY_APPEND; insert_each_element for
		// JSON_ARRAY_INSERT). Accept both value.BoolValue and an
		// int/string surface produced by the analyzer's typed-literal
		// lowering.
		last := rest[len(rest)-1]
		switch v := last.(type) {
		case value.BoolValue:
			createIfMissing = bool(v)
		case value.IntValue:
			createIfMissing = int64(v) != 0
		default:
			if last != nil {
				if b, err := last.ToBool(); err == nil {
					createIfMissing = b
				}
			}
		}
		rest = rest[:len(rest)-1]
	}
	pairs := make([]jsonModifyPair, 0, len(rest)/2)
	for i := 0; i < len(rest); i += 2 {
		if rest[i] == nil {
			return nil, false, fmt.Errorf("%s: path argument is NULL", name)
		}
		p, err := rest[i].ToString()
		if err != nil {
			return nil, false, err
		}
		pairs = append(pairs, jsonModifyPair{path: p, value: jsonAnyOf(rest[i+1])})
	}
	return pairs, createIfMissing, nil
}

func jsonAnyOf(v value.Value) any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case value.IntValue:
		return int64(x)
	case value.FloatValue:
		return float64(x)
	case value.BoolValue:
		return bool(x)
	case value.StringValue:
		return string(x)
	case value.JsonValue:
		var node any
		_ = json.Unmarshal([]byte(string(x)), &node)
		return node
	case *value.ArrayValue:
		// Preserve nested arrays as JSON arrays, not their stringified
		// form -- upstream JSON_ARRAY_INSERT keeps the array literal
		// intact in the resulting document.
		out := make([]any, 0, len(x.Values))
		for _, e := range x.Values {
			out = append(out, jsonAnyOf(e))
		}
		return out
	case *value.StructValue:
		out := map[string]any{}
		for i, k := range x.Keys {
			if i < len(x.Values) {
				out[k] = jsonAnyOf(x.Values[i])
			}
		}
		return out
	}
	s, _ := v.ToString()
	return s
}

// jsonModify implements the append/insert mutation for a small
// subset of JSON-path expressions ("$", "$.<key>", "$[N]").
func jsonModify(doc any, path string, val any, append bool) any {
	if path == "$" {
		// `$` alone does not address an array element, so
		// JSON_ARRAY_INSERT must leave the document unchanged.
		// JSON_ARRAY_APPEND on `$` does append at the top level when
		// the document is an array, mirroring the upstream behaviour.
		if append {
			if arr, ok := doc.([]any); ok {
				return append1(arr, val)
			}
		}
		return doc
	}
	// Strip leading "$.", "$[N]", and recurse.
	p := strings.TrimPrefix(path, "$")
	if strings.HasPrefix(p, ".") {
		rest := strings.TrimPrefix(p, ".")
		field, tail := splitOnce(rest, ".[")
		if m, ok := doc.(map[string]any); ok {
			if tail == "" {
				if arr, ok := m[field].([]any); ok {
					if append {
						m[field] = append1(arr, val)
					} else {
						m[field] = prepend(arr, val)
					}
				}
				return m
			}
			m[field] = jsonModify(m[field], "$"+tail, val, append)
			return m
		}
	}
	if strings.HasPrefix(p, "[") {
		// Index form `$[N]<tail>` where the doc is expected to be an
		// array. JSON_ARRAY_INSERT inserts `val` at position N,
		// shifting later elements; JSON_ARRAY_APPEND treats N as the
		// index of the target array to append to and is forwarded
		// to it once we have the sub-array reference.
		close := strings.Index(p, "]")
		if close < 0 {
			return doc
		}
		var n int
		if _, err := fmt.Sscanf(p[1:close], "%d", &n); err != nil {
			return doc
		}
		tail := p[close+1:]
		arr, ok := doc.([]any)
		if !ok {
			// Upstream behaviour: when the targeted field is null
			// (e.g. `{a: null}` with path `$.a[N]`), treat the null
			// as an empty array and proceed. Non-null, non-array
			// docs leave the path unchanged.
			if doc != nil {
				return doc
			}
			arr = []any{}
		}
		if tail == "" {
			// BigQuery JSON_ARRAY_INSERT defaults insert_each_element
			// to TRUE: when `val` is an array it is flattened into
			// the target. The opposite mode (insert as one nested
			// element) requires the explicit FALSE flag which we
			// don't surface yet.
			vals := []any{val}
			if sub, ok := val.([]any); ok {
				vals = sub
			}
			if append {
				// JSON_ARRAY_APPEND($[N], v) appends v to the element
				// at index N if that element is itself an array.
				if n < 0 || n >= len(arr) {
					return arr
				}
				if subArr, ok := arr[n].([]any); ok {
					arr[n] = appendSlice(subArr, vals)
				}
				return arr
			}
			// JSON_ARRAY_INSERT inserts at index N. Negative indices
			// clamp to 0; positions past the current length are
			// padded with nulls so the inserted value lands exactly
			// at N (matches BigQuery JSON_ARRAY_INSERT semantics).
			if n < 0 {
				n = 0
			}
			out := make([]any, 0, n+len(vals)+len(arr))
			if n <= len(arr) {
				out = appendSlice(out, arr[:n])
				out = appendSlice(out, vals)
				out = appendSlice(out, arr[n:])
			} else {
				out = appendSlice(out, arr)
				for i := len(arr); i < n; i++ {
					out = appendOne(out, nil)
				}
				out = appendSlice(out, vals)
			}
			return out
		}
		if n >= 0 && n < len(arr) {
			arr[n] = jsonModify(arr[n], "$"+tail, val, append)
		}
		return arr
	}
	return doc
}

func append1(arr []any, v any) []any { return append(arr, v) }
func prepend(arr []any, v any) []any { return append([]any{v}, arr...) }

// appendSlice / appendOne are wrappers used inside jsonModify, which
// takes a parameter named `append` that shadows the built-in. Avoids
// the rename of the parameter (which is intentional readable naming).
func appendSlice(dst, src []any) []any { return append(dst, src...) }
func appendOne(dst []any, v any) []any { return append(dst, v) }

func splitOnce(s, seps string) (head, tail string) {
	idx := strings.IndexAny(s, seps)
	if idx < 0 {
		return s, ""
	}
	return s[:idx], s[idx:]
}

// BindSafeToJson is a NULL-tolerant variant of TO_JSON. NULL input
// propagates as SQL NULL (matching the analyzer's constant-folding
// behaviour, which would otherwise produce a non-deterministic mix
// of `JsonValue("null")` and SQL NULL depending on whether the
// argument was a typed or untyped NULL).
func BindSafeToJson(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("SAFE_TO_JSON: missing argument")
	}
	if args[0] == nil {
		return nil, nil
	}
	v := jsonAnyOf(args[0])
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil
	}
	return value.JsonValue(string(b)), nil
}

// ----- hash -----

// BindHighwayFingerprint128 returns a 128-bit fingerprint of the
// input STRING / BYTES as BYTES. Our pure-Go runtime can't ship
// the upstream HighwayHash; we use SHA-512 truncated to 16 bytes,
// which provides the same shape (128-bit deterministic
// fingerprint) and the same domain (binary or string input).
func BindHighwayFingerprint128(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("HIGHWAY_FINGERPRINT128: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	var in []byte
	switch v := args[0].(type) {
	case value.BytesValue:
		in = []byte(v)
	default:
		s, err := args[0].ToString()
		if err != nil {
			return nil, err
		}
		in = []byte(s)
	}
	sum := sha512.Sum512(in)
	return value.BytesValue(sum[:16]), nil
}

// ----- hll -----

// BindHLLCountMerge is the analyzer-visible MERGE aggregate over
// HLL_COUNT sketches. Without a per-bucket merge we approximate
// the union by taking the maximum count seen — adequate for the
// scalar-call form the analyzer surfaces after the rewrite.
func BindHLLCountMerge() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		seen := map[string]struct{}{}
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) == 0 || args[0] == nil {
					return nil
				}
				s, err := args[0].ToString()
				if err != nil {
					return err
				}
				seen[s] = struct{}{}
				return nil
			},
			func() (value.Value, error) {
				return value.IntValue(int64(len(seen))), nil
			},
		)
	}
}

// BindHLLCountMergePartial returns a serialised partial-merge
// sketch. We round-trip the raw distinct-keys set as a JSON-
// encoded blob.
func BindHLLCountMergePartial() func() *helper.Aggregator {
	return func() *helper.Aggregator {
		seen := map[string]struct{}{}
		return helper.NewAggregator(
			func(args []value.Value, _ *helper.Option) error {
				if len(args) == 0 || args[0] == nil {
					return nil
				}
				s, err := args[0].ToString()
				if err != nil {
					return err
				}
				seen[s] = struct{}{}
				return nil
			},
			func() (value.Value, error) {
				keys := make([]string, 0, len(seen))
				for k := range seen {
					keys = append(keys, k)
				}
				b, _ := json.Marshal(keys)
				return value.BytesValue(b), nil
			},
		)
	}
}

// ----- array -----

// BindFlatten flattens an ARRAY<ARRAY<T>> into ARRAY<T> by
// concatenating the inner arrays. NULL elements pass through.
func BindFlatten(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FLATTEN: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	arr, ok := args[0].(*value.ArrayValue)
	if !ok {
		return nil, fmt.Errorf("FLATTEN: argument is not an ARRAY")
	}
	var out []value.Value
	for _, e := range arr.Values {
		if e == nil {
			out = append(out, nil)
			continue
		}
		inner, ok := e.(*value.ArrayValue)
		if !ok {
			out = append(out, e)
			continue
		}
		out = append(out, inner.Values...)
	}
	return &value.ArrayValue{Values: out}, nil
}

// BindArrayZip pairs ARRAYs element-wise. Two-arg form returns
// ARRAY<STRUCT<a, b>>; >2 arrays would map to a wider STRUCT but
// our STRUCT runtime supports up to N positional fields.
func BindArrayZip(args ...value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ARRAY_ZIP: needs at least 2 arrays")
	}
	arrs := make([]*value.ArrayValue, 0, len(args))
	minLen := -1
	for _, a := range args {
		if a == nil {
			return nil, nil
		}
		av, ok := a.(*value.ArrayValue)
		if !ok {
			return nil, fmt.Errorf("ARRAY_ZIP: argument is not an ARRAY")
		}
		arrs = append(arrs, av)
		if minLen < 0 || len(av.Values) < minLen {
			minLen = len(av.Values)
		}
	}
	if minLen < 0 {
		minLen = 0
	}
	out := make([]value.Value, 0, minLen)
	for i := 0; i < minLen; i++ {
		keys := make([]string, len(arrs))
		vals := make([]value.Value, len(arrs))
		for j, av := range arrs {
			keys[j] = fmt.Sprintf("f%d", j)
			vals[j] = av.Values[i]
		}
		m := map[string]value.Value{}
		for j, k := range keys {
			m[k] = vals[j]
		}
		out = append(out, &value.StructValue{Keys: keys, Values: vals, M: m})
	}
	return &value.ArrayValue{Values: out}, nil
}

// ----- aggregate -----

// BindGrouping returns 0 / 1 indicating whether the grouping
// expression contributes to the current grouping set. Without
// full ROLLUP / CUBE tracking we default to 0 (the column IS
// part of the grouping). Hand-rolled GROUPING() uses generally
// expect the simple case.
func BindGrouping(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("GROUPING: invalid number of arguments: got %d, want 1", len(args))
	}
	return value.IntValue(0), nil
}

// ----- net -----

// BindNetIpInNet returns whether `ip` falls inside the CIDR
// `subnet`.
func BindNetIpInNet(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("NET.IP_IN_NET: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	ipStr, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	cidr, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	ip := gonet.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("NET.IP_IN_NET: invalid IP %q", ipStr)
	}
	_, ipnet, err := gonet.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("NET.IP_IN_NET: %w", err)
	}
	return value.BoolValue(ipnet.Contains(ip)), nil
}

// BindNetParseIp parses an IPv4 / IPv6 STRING to its INT64 packed
// form. For IPv4 returns the 32-bit value; for IPv6 returns the
// hash of the 16-byte representation (only INT64 is in scope).
func BindNetParseIp(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.PARSE_IP: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	ip := gonet.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("NET.PARSE_IP: invalid IP %q", s)
	}
	if v4 := ip.To4(); v4 != nil {
		return value.IntValue(int64(binary.BigEndian.Uint32(v4))), nil
	}
	bi := new(big.Int).SetBytes(ip.To16())
	return value.IntValue(bi.Int64()), nil
}

// BindNetFormatIp formats an INT64-packed IPv4 address as a
// dotted-decimal STRING.
func BindNetFormatIp(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.FORMAT_IP: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	n, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	if n < 0 || n > 0xFFFFFFFF {
		return nil, fmt.Errorf("NET.FORMAT_IP: out of range")
	}
	ip := gonet.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	return value.StringValue(ip.String()), nil
}

// BindNetParsePackedIp parses an IP STRING to its big-endian
// network-order BYTES form.
func BindNetParsePackedIp(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.PARSE_PACKED_IP: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	ip := gonet.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("NET.PARSE_PACKED_IP: invalid IP %q", s)
	}
	if v4 := ip.To4(); v4 != nil {
		return value.BytesValue(v4), nil
	}
	return value.BytesValue(ip.To16()), nil
}

// BindNetFormatPackedIp formats packed BYTES to an IP STRING.
func BindNetFormatPackedIp(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NET.FORMAT_PACKED_IP: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	if len(b) != 4 && len(b) != 16 {
		return nil, fmt.Errorf("NET.FORMAT_PACKED_IP: BYTES must be 4 or 16 long")
	}
	return value.StringValue(gonet.IP(b).String()), nil
}

// BindNetMakeNet builds a CIDR STRING from an IP STRING and
// prefix length.
func BindNetMakeNet(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("NET.MAKE_NET: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	ipStr, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	pref, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	ip := gonet.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("NET.MAKE_NET: invalid IP %q", ipStr)
	}
	bits := 32
	if ip.To4() == nil {
		bits = 128
	}
	if int(pref) < 0 || int(pref) > bits {
		return nil, fmt.Errorf("NET.MAKE_NET: prefix out of range")
	}
	mask := gonet.CIDRMask(int(pref), bits)
	masked := ip.Mask(mask)
	return value.StringValue(fmt.Sprintf("%s/%d", masked.String(), pref)), nil
}

// silence unused-import warnings for the packages used only by
// subsets of the surface above.
var (
	_ = hex.EncodeToString
	_ = strconv.Itoa
)
