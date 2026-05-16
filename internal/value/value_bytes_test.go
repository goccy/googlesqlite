package value_test

import (
	"bytes"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestBytesValue covers append/compare/convert paths and the
// printable-character branch of BytesValue.Format.
func TestBytesValue(t *testing.T) {
	t.Parallel()

	t.Run("Add appends", func(t *testing.T) {
		got, err := value.BytesValue([]byte("ab")).Add(value.BytesValue([]byte("cd")))
		if err != nil {
			t.Fatal(err)
		}
		b, _ := got.ToBytes()
		if !bytes.Equal(b, []byte("abcd")) {
			t.Fatalf("Add: %q", b)
		}
	})

	t.Run("Sub/Mul/Div unsupported", func(t *testing.T) {
		bv := value.BytesValue([]byte("x"))
		if _, err := bv.Sub(bv); err == nil {
			t.Fatal("Sub")
		}
		if _, err := bv.Mul(bv); err == nil {
			t.Fatal("Mul")
		}
		if _, err := bv.Div(bv); err == nil {
			t.Fatal("Div")
		}
	})

	t.Run("EQ", func(t *testing.T) {
		a := value.BytesValue([]byte{1, 2, 3})
		b := value.BytesValue([]byte{1, 2, 3})
		c := value.BytesValue([]byte{1, 2, 4})
		if ok, _ := a.EQ(b); !ok {
			t.Fatal("expected equal")
		}
		if ok, _ := a.EQ(c); ok {
			t.Fatal("expected not equal")
		}
	})

	t.Run("GT GTE LT LTE", func(t *testing.T) {
		a := value.BytesValue([]byte("a"))
		b := value.BytesValue([]byte("b"))
		if v, _ := b.GT(a); !v {
			t.Fatal("GT")
		}
		if v, _ := a.GTE(a); !v {
			t.Fatal("GTE")
		}
		if v, _ := a.LT(b); !v {
			t.Fatal("LT")
		}
		if v, _ := a.LTE(a); !v {
			t.Fatal("LTE")
		}
	})

	t.Run("ToInt64 empty 0", func(t *testing.T) {
		got, err := value.BytesValue(nil).ToInt64()
		if err != nil || got != 0 {
			t.Fatalf("empty -> %d / err=%v", got, err)
		}
	})

	t.Run("ToInt64 parsable", func(t *testing.T) {
		got, err := value.BytesValue([]byte("42")).ToInt64()
		if err != nil || got != 42 {
			t.Fatalf("42 -> %d / err=%v", got, err)
		}
	})

	t.Run("ToInt64 invalid", func(t *testing.T) {
		if _, err := value.BytesValue([]byte("abc")).ToInt64(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToString base64", func(t *testing.T) {
		// "hi" -> base64
		got, _ := value.BytesValue([]byte("hi")).ToString()
		if got != "aGk=" {
			t.Fatalf("ToString: %s", got)
		}
	})

	t.Run("ToFloat64 empty 0", func(t *testing.T) {
		got, _ := value.BytesValue(nil).ToFloat64()
		if got != 0 {
			t.Fatalf("empty -> %f", got)
		}
	})

	t.Run("ToFloat64 parsable", func(t *testing.T) {
		got, _ := value.BytesValue([]byte("1.5")).ToFloat64()
		if got != 1.5 {
			t.Fatalf("ToFloat64: %f", got)
		}
	})

	t.Run("ToBool", func(t *testing.T) {
		if v, _ := value.BytesValue(nil).ToBool(); v != false {
			t.Fatalf("empty -> %v", v)
		}
		if v, _ := value.BytesValue([]byte("true")).ToBool(); v != true {
			t.Fatalf("true -> %v", v)
		}
		if _, err := value.BytesValue([]byte("nope")).ToBool(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToArray/ToStruct errors", func(t *testing.T) {
		bv := value.BytesValue([]byte("x"))
		if _, err := bv.ToArray(); err == nil {
			t.Fatal("ToArray")
		}
		if _, err := bv.ToStruct(); err == nil {
			t.Fatal("ToStruct")
		}
	})

	t.Run("ToJSON wraps base64 in quotes", func(t *testing.T) {
		got, _ := value.BytesValue([]byte("hi")).ToJSON()
		if got != `"aGk="` {
			t.Fatalf("ToJSON: %s", got)
		}
	})

	t.Run("ToTime", func(t *testing.T) {
		got, err := value.BytesValue([]byte("2020-01-02")).ToTime()
		if err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Fatal("zero")
		}
		// other branches
		if _, err := value.BytesValue([]byte("2020-01-02 03:04:05")).ToTime(); err != nil {
			t.Fatal(err)
		}
		if _, err := value.BytesValue([]byte("03:04:05")).ToTime(); err != nil {
			t.Fatal(err)
		}
		if _, err := value.BytesValue([]byte("2020-01-02T03:04:05Z")).ToTime(); err != nil {
			t.Fatal(err)
		}
		if _, err := value.BytesValue([]byte("garbage")).ToTime(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ToRat", func(t *testing.T) {
		r, err := value.BytesValue([]byte("1/2")).ToRat()
		if err != nil {
			t.Fatal(err)
		}
		if r.Cmp(r) != 0 {
			t.Fatal("compare")
		}
	})

	t.Run("Format t printable and nonprintable", func(t *testing.T) {
		bv := value.BytesValue([]byte{'a', 0x01, 'b'})
		got := bv.Format('t')
		if got != "a\\x01b" {
			t.Fatalf("Format t: %q", got)
		}
		got2 := bv.Format('T')
		if got2 != `b"a\x01b"` {
			t.Fatalf("Format T: %q", got2)
		}
		// default verb falls back to ToString (base64).
		def := bv.Format('x')
		if def == "" {
			t.Fatalf("Format default: %q", def)
		}
	})

	t.Run("PrintableChar", func(t *testing.T) {
		if !value.PrintableChar('a') {
			t.Fatal("a should be printable")
		}
		if value.PrintableChar(0x01) {
			t.Fatal("0x01 should not be printable")
		}
		if !value.PrintableChar(0x20) {
			t.Fatal("0x20 (space) should be printable")
		}
		if !value.PrintableChar(0x7e) {
			t.Fatal("0x7e should be printable")
		}
		if value.PrintableChar(0x7f) {
			t.Fatal("0x7f should not be printable")
		}
	})

	t.Run("Interface", func(t *testing.T) {
		bv := value.BytesValue([]byte("hi"))
		if got, ok := bv.Interface().([]byte); !ok || string(got) != "hi" {
			t.Fatalf("Interface: %v (%T)", bv.Interface(), bv.Interface())
		}
	})
}
