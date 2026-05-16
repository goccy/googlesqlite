package proto

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestFilterFieldsNestedInclude: keep only field 2 inside the
// submessage at field 3 of the outer message.
func TestFilterFieldsNestedInclude(t *testing.T) {
	// outer = { field 1 = INT "kept", field 3 = submessage }
	// submessage = { field 1 = INT "drop", field 2 = INT "keep" }
	inner := []byte{}
	inner = append(inner, 0x08)
	inner = protowire.AppendVarint(inner, 11)
	inner = append(inner, 0x10)
	inner = protowire.AppendVarint(inner, 22)
	outer := []byte{}
	outer = append(outer, 0x08)
	outer = protowire.AppendVarint(outer, 1)
	outer = append(outer, 0x1a)
	outer = protowire.AppendVarint(outer, uint64(len(inner)))
	outer = append(outer, inner...)
	got, err := BindFilterFields(
		value.BytesValue(outer),
		value.StringValue("+3.2"),
	)
	if err != nil {
		t.Fatalf("BindFilterFields: %v", err)
	}
	gotBytes, _ := got.ToBytes()
	// Expected: outer field 1 dropped (not in include set); field 3
	// kept but pruned to only its field 2.
	wantInner := []byte{}
	wantInner = append(wantInner, 0x10)
	wantInner = protowire.AppendVarint(wantInner, 22)
	want := []byte{}
	want = append(want, 0x1a)
	want = protowire.AppendVarint(want, uint64(len(wantInner)))
	want = append(want, wantInner...)
	if !bytes.Equal(gotBytes, want) {
		t.Fatalf("nested include mismatch: got %x want %x", gotBytes, want)
	}
}

// TestReplaceFieldsNestedPath: replace field 2 inside the submessage
// at field 3 of the outer message.
func TestReplaceFieldsNestedPath(t *testing.T) {
	inner := []byte{}
	inner = append(inner, 0x10)
	inner = protowire.AppendVarint(inner, 100)
	outer := []byte{}
	outer = append(outer, 0x1a)
	outer = protowire.AppendVarint(outer, uint64(len(inner)))
	outer = append(outer, inner...)
	got, err := BindReplaceFields(
		value.BytesValue(outer),
		value.StringValue("3.2"),
		value.StringValue("int64"),
		value.IntValue(999),
	)
	if err != nil {
		t.Fatalf("BindReplaceFields: %v", err)
	}
	gotBytes, _ := got.ToBytes()
	// Walk the result to confirm the leaf value is now 999.
	tag1, _, n := protowire.ConsumeTag(gotBytes)
	if n < 0 || tag1 != 3 {
		t.Fatalf("expected outer field 3, got tag %d", tag1)
	}
	gotBytes = gotBytes[n:]
	innerBytes, n := protowire.ConsumeBytes(gotBytes)
	if n < 0 {
		t.Fatalf("malformed inner submessage")
	}
	tag2, _, n := protowire.ConsumeTag(innerBytes)
	if n < 0 || tag2 != 2 {
		t.Fatalf("expected inner field 2, got tag %d", tag2)
	}
	innerBytes = innerBytes[n:]
	v, _ := protowire.ConsumeVarint(innerBytes)
	if v != 999 {
		t.Fatalf("expected replaced value 999, got %d", v)
	}
}

// TestProtoModifyMapReturnsArray exercises the new BindProtoModifyMap
// signature `(parent_proto_bytes, map_field_tag, key_kind, val_kind,
// k1, v1, ...)`. Inserts one new (key=2 → value="b") entry into a
// parent message that already carries an entry under map field tag 7.
// Expected return is an ArrayValue whose two BytesValues are the
// inner wire-format payloads of each entry; the upstream `tag 0x3a`
// is now passed explicitly to the runtime rather than inferred from
// existing entries.
func TestProtoModifyMapReturnsArray(t *testing.T) {
	// Existing entry payload: { field 1 = 1, field 2 = "a" }.
	keyPayload := protowire.AppendVarint(nil, 1)
	valPayload := []byte{byte(len("a"))}
	valPayload = append(valPayload, "a"...)
	existingEntry := []byte{0x08}
	existingEntry = append(existingEntry, keyPayload...)
	existingEntry = append(existingEntry, 0x12)
	existingEntry = append(existingEntry, valPayload...)
	parent := []byte{0x3a}
	parent = protowire.AppendVarint(parent, uint64(len(existingEntry)))
	parent = append(parent, existingEntry...)

	got, err := BindProtoModifyMap(
		value.BytesValue(parent),
		value.IntValue(7),
		value.StringValue("int64"),
		value.StringValue("string"),
		value.IntValue(2),
		value.StringValue("b"),
	)
	if err != nil {
		t.Fatalf("BindProtoModifyMap: %v", err)
	}
	arr, err := got.ToArray()
	if err != nil {
		t.Fatalf("expected ArrayValue, got %T: %v", got, err)
	}
	if len(arr.Values) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(arr.Values))
	}
	// Second entry should carry key=2 (field 1, varint) and value="b" (field 2, bytes).
	inserted, _ := arr.Values[1].ToBytes()
	tag1, _, n := protowire.ConsumeTag(inserted)
	if n < 0 || tag1 != 1 {
		t.Fatalf("expected key field tag 1, got %d", tag1)
	}
	keyVal, _ := protowire.ConsumeVarint(inserted[n:])
	if keyVal != 2 {
		t.Fatalf("expected key 2, got %d", keyVal)
	}
}
