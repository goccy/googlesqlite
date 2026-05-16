package value_test

import (
	"encoding/base64"
	"testing"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"
)

// envelope wraps a malformed (header, body) pair and returns the
// base64-encoded representation DecodeValue accepts.
func envelope(t *testing.T, header value.ValueType, body string) string {
	t.Helper()
	b, err := json.Marshal(value.ValueLayout{Header: header, Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// TestDecodeValueMalformedBodies feeds DecodeValue base64-of-JSON
// envelopes whose body is wrong for the declared header. Each table
// row exercises a distinct error branch inside decodeFromValueLayout.
func TestDecodeValueMalformedBodies(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		header value.ValueType
		body   string
	}{
		{
			// bytes header with a body that isn't valid base64.
			name:   "bytes_with_invalid_base64",
			header: value.BytesValueType,
			body:   "!!!not-base64!!!",
		},
		{
			// date header with a body that isn't an ISO date.
			name:   "date_with_invalid_iso",
			header: value.DateValueType,
			body:   "not-a-date",
		},
		{
			// datetime header with a body that isn't an ISO datetime.
			name:   "datetime_with_invalid_iso",
			header: value.DatetimeValueType,
			body:   "not-a-datetime",
		},
		{
			// time header with a body that isn't a valid HH:MM:SS.
			name:   "time_with_invalid_format",
			header: value.TimeValueType,
			body:   "99:99:99",
		},
		{
			// timestamp body must parse as an int64 microsecond value.
			name:   "timestamp_with_non_integer_body",
			header: value.TimestampValueType,
			body:   "not-an-int",
		},
		{
			// interval body must parse via the interval parser.
			name:   "interval_with_garbage_body",
			header: value.IntervalValueType,
			body:   "not-an-interval",
		},
		{
			// array body must be valid JSON. Truncate.
			name:   "array_with_truncated_json",
			header: value.ArrayValueType,
			body:   "[1, 2,",
		},
		{
			// struct body must be valid JSON.
			name:   "struct_with_truncated_json",
			header: value.StructValueType,
			body:   "{\"keys\":[\"a\"]",
		},
		{
			// geography body must parse via GeographyFromWKT.
			name:   "geography_with_invalid_wkt",
			header: value.GeographyValueType,
			body:   "not-a-wkt",
		},
		{
			// range body must be valid JSON.
			name:   "range_with_truncated_json",
			header: value.RangeValueType,
			body:   "{\"elem\":\"date\"",
		},
		{
			// Unknown header surfaces the trailing default error.
			name:   "unknown_header",
			header: value.ValueType("totally-bogus"),
			body:   "ignored",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := envelope(t, tc.header, tc.body)
			if _, err := value.DecodeValue(env); err == nil {
				t.Fatalf("expected error for %s header / body=%q", tc.header, tc.body)
			}
		})
	}
}

// TestDecodeValueEmptyHeaderFallsThroughToString covers the
// `layout.Header == ""` branch, which returns the raw input as a
// StringValue rather than a typed Value.
func TestDecodeValueEmptyHeaderFallsThroughToString(t *testing.T) {
	t.Parallel()

	env := envelope(t, "", "anything")
	got, err := value.DecodeValue(env)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(value.StringValue); !ok {
		t.Fatalf("expected StringValue, got %T", got)
	}
}

// TestDecodeValueArrayElementErrorPropagates makes the array decoder
// recurse into an element whose JSON body is itself malformed,
// exercising the inner DecodeValue error-propagation branch.
func TestDecodeValueArrayElementErrorPropagates(t *testing.T) {
	t.Parallel()

	// An array body whose element is a base64-of-JSON envelope with
	// a bogus header.
	badElement := envelope(t, value.ValueType("totally-bogus"), "ignored")
	bodyJSON, err := json.Marshal([]any{badElement})
	if err != nil {
		t.Fatal(err)
	}
	env := envelope(t, value.ArrayValueType, string(bodyJSON))
	if _, err := value.DecodeValue(env); err == nil {
		t.Fatal("expected error from malformed array element")
	}
}

// TestDecodeValueStructElementErrorPropagates does the same for the
// struct value decoder.
func TestDecodeValueStructElementErrorPropagates(t *testing.T) {
	t.Parallel()

	badElement := envelope(t, value.ValueType("totally-bogus"), "ignored")
	bodyJSON, err := json.Marshal(value.StructValueLayout{
		Keys:   []string{"x"},
		Values: []any{badElement},
	})
	if err != nil {
		t.Fatal(err)
	}
	env := envelope(t, value.StructValueType, string(bodyJSON))
	if _, err := value.DecodeValue(env); err == nil {
		t.Fatal("expected error from malformed struct element")
	}
}

// TestDecodeValueRangeBoundErrorPropagates feeds a range whose start
// bound is a malformed envelope.
func TestDecodeValueRangeBoundErrorPropagates(t *testing.T) {
	t.Parallel()

	badBound := envelope(t, value.ValueType("totally-bogus"), "ignored")
	bodyJSON, err := json.Marshal(struct {
		Elem  string `json:"elem"`
		Start any    `json:"start"`
		End   any    `json:"end"`
	}{
		Elem:  string(value.DateValueType),
		Start: badBound,
	})
	if err != nil {
		t.Fatal(err)
	}
	env := envelope(t, value.RangeValueType, string(bodyJSON))
	if _, err := value.DecodeValue(env); err == nil {
		t.Fatal("expected error from malformed range start bound")
	}
}

// TestDecodeValueRangeEndErrorPropagates does the same for the End
// bound.
func TestDecodeValueRangeEndErrorPropagates(t *testing.T) {
	t.Parallel()

	badBound := envelope(t, value.ValueType("totally-bogus"), "ignored")
	bodyJSON, err := json.Marshal(struct {
		Elem  string `json:"elem"`
		Start any    `json:"start"`
		End   any    `json:"end"`
	}{
		Elem: string(value.DateValueType),
		End:  badBound,
	})
	if err != nil {
		t.Fatal(err)
	}
	env := envelope(t, value.RangeValueType, string(bodyJSON))
	if _, err := value.DecodeValue(env); err == nil {
		t.Fatal("expected error from malformed range end bound")
	}
}

// TestDecodeValueGeographyInverted exercises the "INVERTED " prefix
// branch of the geography decoder.
func TestDecodeValueGeographyInverted(t *testing.T) {
	t.Parallel()

	env := envelope(t, value.GeographyValueType, "INVERTED POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))")
	got, err := value.DecodeValue(env)
	if err != nil {
		t.Fatal(err)
	}
	gv, ok := got.(*value.GeographyValue)
	if !ok {
		t.Fatalf("type: %T", got)
	}
	if !gv.Inverted() {
		t.Fatal("expected Inverted() true")
	}
}

// TestDecodeValueNumericAndBigNumeric round-trips the NUMERIC /
// BIGNUMERIC branches via EncodeValue.
func TestDecodeValueNumericAndBigNumeric(t *testing.T) {
	t.Parallel()

	// NUMERIC envelope.
	env := envelope(t, value.NumericValueType, "123/1")
	got, err := value.DecodeValue(env)
	if err != nil {
		t.Fatal(err)
	}
	nv, ok := got.(*value.NumericValue)
	if !ok {
		t.Fatalf("type: %T", got)
	}
	if nv.IsBigNumeric {
		t.Fatal("expected IsBigNumeric=false")
	}

	// BIGNUMERIC envelope.
	bigEnv := envelope(t, value.BigNumericValueType, "456/1")
	got, err = value.DecodeValue(bigEnv)
	if err != nil {
		t.Fatal(err)
	}
	bnv, ok := got.(*value.NumericValue)
	if !ok {
		t.Fatalf("type: %T", got)
	}
	if !bnv.IsBigNumeric {
		t.Fatal("expected IsBigNumeric=true")
	}
}

// TestDecodeValueIntervalRoundtrip drives the INTERVAL branch.
func TestDecodeValueIntervalRoundtrip(t *testing.T) {
	t.Parallel()

	// "1 0 0 0:0:0" is the predecessor's canonical zero-extended form.
	env := envelope(t, value.IntervalValueType, "0-0 1 0:0:0")
	got, err := value.DecodeValue(env)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(*value.IntervalValue); !ok {
		t.Fatalf("type: %T", got)
	}
}
