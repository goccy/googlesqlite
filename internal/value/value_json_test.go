package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestJsonValue covers JsonValue conversions, Type discrimination, and
// the all-unsupported arithmetic / comparison stubs.
func TestJsonValue(t *testing.T) {
	t.Parallel()

	t.Run("arithmetic/comparison all unsupported", func(t *testing.T) {
		jv := value.JsonValue("1")
		if _, err := jv.Add(jv); err == nil {
			t.Fatal("Add")
		}
		if _, err := jv.Sub(jv); err == nil {
			t.Fatal("Sub")
		}
		if _, err := jv.Mul(jv); err == nil {
			t.Fatal("Mul")
		}
		if _, err := jv.Div(jv); err == nil {
			t.Fatal("Div")
		}
		if _, err := jv.EQ(jv); err == nil {
			t.Fatal("EQ")
		}
		if _, err := jv.GT(jv); err == nil {
			t.Fatal("GT")
		}
		if _, err := jv.GTE(jv); err == nil {
			t.Fatal("GTE")
		}
		if _, err := jv.LT(jv); err == nil {
			t.Fatal("LT")
		}
		if _, err := jv.LTE(jv); err == nil {
			t.Fatal("LTE")
		}
	})

	t.Run("ToInt64", func(t *testing.T) {
		if got, _ := value.JsonValue("42").ToInt64(); got != 42 {
			t.Fatalf("42: %d", got)
		}
		if _, err := value.JsonValue("xx").ToInt64(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToFloat64", func(t *testing.T) {
		if got, _ := value.JsonValue("1.5").ToFloat64(); got != 1.5 {
			t.Fatalf("1.5: %f", got)
		}
		if _, err := value.JsonValue("xx").ToFloat64(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToBool", func(t *testing.T) {
		if v, _ := value.JsonValue("true").ToBool(); !v {
			t.Fatal("expected true")
		}
		if _, err := value.JsonValue("xx").ToBool(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToString/ToBytes/ToJSON", func(t *testing.T) {
		jv := value.JsonValue(`{"x":1}`)
		s, _ := jv.ToString()
		if s != `{"x":1}` {
			t.Fatalf("ToString: %s", s)
		}
		b, _ := jv.ToBytes()
		if string(b) != `{"x":1}` {
			t.Fatalf("ToBytes: %s", b)
		}
		j, _ := jv.ToJSON()
		if j != `{"x":1}` {
			t.Fatalf("ToJSON: %s", j)
		}
	})

	t.Run("ToArray/ToStruct/ToTime unsupported", func(t *testing.T) {
		jv := value.JsonValue("1")
		if _, err := jv.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := jv.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
		if _, err := jv.ToTime(); err == nil {
			t.Fatal("ToTime")
		}
	})

	t.Run("ToRat", func(t *testing.T) {
		r, err := value.JsonValue("123").ToRat()
		if err != nil {
			t.Fatal(err)
		}
		if r.Num().Int64() != 123 {
			t.Fatalf("ToRat: %s", r.String())
		}
		if _, err := value.JsonValue("xx").ToRat(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Format/Interface", func(t *testing.T) {
		jv := value.JsonValue(`{"x":1}`)
		if jv.Format('t') != `{"x":1}` {
			t.Fatalf("Format: %s", jv.Format('t'))
		}
		// Interface unmarshals the JSON.
		got := jv.Interface()
		if got == nil {
			t.Fatal("Interface nil for valid JSON")
		}
		// Invalid JSON returns nil.
		bogus := value.JsonValue("not valid json")
		if bogus.Interface() != nil {
			t.Fatalf("invalid JSON Interface: %v", bogus.Interface())
		}
	})

	t.Run("Type", func(t *testing.T) {
		cases := []struct {
			in   string
			want string
		}{
			{"42", "number"},
			{"1.5", "number"},
			{`"hi"`, "string"},
			{"true", "boolean"},
			{"[1,2]", "array"},
			{`{"a":1}`, "object"},
			{"null", "null"},
		}
		for _, c := range cases {
			t.Run(c.in, func(t *testing.T) {
				if got := value.JsonValue(c.in).Type(); got != c.want {
					t.Fatalf("%s -> %s, want %s", c.in, got, c.want)
				}
			})
		}
	})
}
