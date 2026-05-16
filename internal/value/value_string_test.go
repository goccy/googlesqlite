package value_test

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestStringValue covers concatenation, comparisons, conversions, and
// the time-parsing branches of StringValue.ToTime.
func TestStringValue(t *testing.T) {
	t.Parallel()

	t.Run("Add concatenates", func(t *testing.T) {
		got, err := value.StringValue("foo").Add(value.StringValue("bar"))
		if err != nil {
			t.Fatal(err)
		}
		if got != value.StringValue("foobar") {
			t.Fatalf("Add: %v", got)
		}
	})

	t.Run("Sub/Mul/Div unsupported", func(t *testing.T) {
		sv := value.StringValue("x")
		if _, err := sv.Sub(sv); err == nil {
			t.Fatal("Sub should be unsupported")
		}
		if _, err := sv.Mul(sv); err == nil {
			t.Fatal("Mul should be unsupported")
		}
		if _, err := sv.Div(sv); err == nil {
			t.Fatal("Div should be unsupported")
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		a := value.StringValue("a")
		b := value.StringValue("b")
		if eq, _ := a.EQ(a); !eq {
			t.Fatal("EQ self")
		}
		if eq, _ := a.EQ(b); eq {
			t.Fatal("EQ diff")
		}
		if gt, _ := b.GT(a); !gt {
			t.Fatal("GT")
		}
		if gte, _ := a.GTE(a); !gte {
			t.Fatal("GTE")
		}
		if lt, _ := a.LT(b); !lt {
			t.Fatal("LT")
		}
		if lte, _ := a.LTE(a); !lte {
			t.Fatal("LTE")
		}
	})

	t.Run("ToInt64 empty -> 0", func(t *testing.T) {
		got, err := value.StringValue("").ToInt64()
		if err != nil || got != 0 {
			t.Fatalf("empty -> %d / err=%v", got, err)
		}
	})

	t.Run("ToInt64 decimal", func(t *testing.T) {
		got, err := value.StringValue("123").ToInt64()
		if err != nil || got != 123 {
			t.Fatalf("123 -> %d / err=%v", got, err)
		}
	})

	t.Run("ToInt64 hex", func(t *testing.T) {
		got, err := value.StringValue("0xff").ToInt64()
		if err != nil || got != 255 {
			t.Fatalf("0xff -> %d / err=%v", got, err)
		}
	})

	t.Run("ToInt64 invalid", func(t *testing.T) {
		if _, err := value.StringValue("not-int").ToInt64(); err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("ToFloat64", func(t *testing.T) {
		if got, _ := value.StringValue("").ToFloat64(); got != 0 {
			t.Fatalf("empty -> %f", got)
		}
		if got, err := value.StringValue("3.14").ToFloat64(); err != nil || got != 3.14 {
			t.Fatalf("3.14 -> %f / err=%v", got, err)
		}
		if _, err := value.StringValue("nope").ToFloat64(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToBool", func(t *testing.T) {
		if v, _ := value.StringValue("").ToBool(); v != false {
			t.Fatalf("empty -> %v", v)
		}
		if v, _ := value.StringValue("true").ToBool(); v != true {
			t.Fatalf("true -> %v", v)
		}
		if _, err := value.StringValue("nope").ToBool(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToArray empty returns nil,nil", func(t *testing.T) {
		got, err := value.StringValue("").ToArray()
		if err != nil || got != nil {
			t.Fatalf("empty -> %v / err=%v", got, err)
		}
	})

	t.Run("ToArray non-empty err", func(t *testing.T) {
		if _, err := value.StringValue("x").ToArray(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToStruct empty returns nil,nil", func(t *testing.T) {
		got, err := value.StringValue("").ToStruct()
		if err != nil || got != nil {
			t.Fatalf("empty -> %v / err=%v", got, err)
		}
	})

	t.Run("ToStruct non-empty err", func(t *testing.T) {
		if _, err := value.StringValue("x").ToStruct(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToJSON quotes", func(t *testing.T) {
		got, err := value.StringValue(`hi"there`).ToJSON()
		if err != nil {
			t.Fatal(err)
		}
		if got != `"hi\"there"` {
			t.Fatalf("ToJSON: %s", got)
		}
	})

	t.Run("ToBytes", func(t *testing.T) {
		got, _ := value.StringValue("foo").ToBytes()
		if string(got) != "foo" {
			t.Fatalf("ToBytes: %s", got)
		}
	})

	t.Run("ToTime date/datetime/time/timestamp", func(t *testing.T) {
		cases := []string{
			"2020-01-02",
			"2020-01-02T03:04:05",
			"03:04:05",
			"2020-01-02 03:04:05",
			"2020-01-02T03:04:05Z",
		}
		for _, c := range cases {
			t.Run(c, func(t *testing.T) {
				got, err := value.StringValue(c).ToTime()
				if err != nil {
					t.Fatalf("input %q err: %v", c, err)
				}
				if got.IsZero() {
					t.Fatalf("input %q got zero time", c)
				}
			})
		}
	})

	t.Run("ToTime bogus", func(t *testing.T) {
		if _, err := value.StringValue("not-a-date").ToTime(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToRat", func(t *testing.T) {
		r, err := value.StringValue("1/2").ToRat()
		if err != nil {
			t.Fatal(err)
		}
		if r.Cmp(r) != 0 {
			t.Fatal("comparison failed")
		}
	})

	t.Run("Format", func(t *testing.T) {
		sv := value.StringValue(`he\"y`)
		if sv.Format('t') != `he\"y` {
			t.Fatalf("Format t: %s", sv.Format('t'))
		}
		// 'T' quotes with strconv.Quote
		if got := sv.Format('T'); got != `"he\\\"y"` {
			t.Fatalf("Format T: %s", got)
		}
		// fallback verb returns raw string
		if sv.Format('x') != `he\"y` {
			t.Fatalf("Format default: %s", sv.Format('x'))
		}
	})

	t.Run("Interface", func(t *testing.T) {
		sv := value.StringValue("hi")
		if got, ok := sv.Interface().(string); !ok || got != "hi" {
			t.Fatalf("Interface: %v (%T)", sv.Interface(), sv.Interface())
		}
	})
}
