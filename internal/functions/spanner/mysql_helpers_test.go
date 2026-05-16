package spanner

import (
	"math"
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

// Compatibility tests for the mysql.* aliases. Expected values
// come from the MySQL 8.0 reference manual (the source we mirror
// when implementing the aliases) and the Spanner mysql-namespace
// compliance fixtures. Each test exercises the Bind* function's
// stable public signature (variadic value.Value, returning
// value.Value or error) so it survives an internal refactor.

func mustString(t *testing.T, v value.Value) string {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	s, err := v.ToString()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustInt64(t *testing.T, v value.Value) int64 {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	n, err := v.ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func mustFloat64(t *testing.T, v value.Value) float64 {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	f, err := v.ToFloat64()
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func mustBool(t *testing.T, v value.Value) bool {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	b, err := v.ToBool()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mustBytes(t *testing.T, v value.Value) []byte {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
	b, err := v.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBindExtractTimestampParts(t *testing.T) {
	t.Parallel()

	// 2023-06-14T05:23:45Z is a Wednesday (Weekday=3).
	ts := value.TimestampValue(time.Date(2023, 6, 14, 5, 23, 45, 123_456_000, time.UTC))

	cases := []struct {
		name string
		bind func(...value.Value) (value.Value, error)
		want int64
	}{
		{"day", BindDay, 14},
		{"month", BindMonth, 6},
		{"year", BindYear, 2023},
		{"hour", BindHour, 5},
		{"minute", BindMinute, 23},
		{"second", BindSecond, 45},
		{"microsecond", BindMicrosecond, 123_456},
		{"quarter", BindQuarter, 2},
		// MySQL DAYOFWEEK: Sunday=1...Saturday=7, so Wednesday=4.
		{"dayofweek", BindDayOfWeek, 4},
		// 2023-06-14 is the 165th day of the year.
		{"dayofyear", BindDayOfYear, 165},
		{"dayofmonth", BindDayOfMonth, 14},
		// MySQL WEEKDAY: Monday=0...Sunday=6, Wednesday=2.
		{"weekday", BindWeekday, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.bind(ts)
			if err != nil {
				t.Fatal(err)
			}
			if mustInt64(t, got) != c.want {
				t.Fatalf("%s: got %d want %d", c.name, mustInt64(t, got), c.want)
			}
		})
	}

	t.Run("null propagates", func(t *testing.T) {
		v, err := BindDay(nil)
		if err != nil || v != nil {
			t.Fatalf("expected nil/nil, got %v %v", v, err)
		}
	})
	t.Run("wrong arg count", func(t *testing.T) {
		if _, err := BindDay(); err == nil {
			t.Fatal("expected error")
		}
		if _, err := BindDay(value.IntValue(1), value.IntValue(2)); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBindBitLength(t *testing.T) {
	t.Parallel()

	// MySQL: BIT_LENGTH('text') = 32.
	got, err := BindBitLength(value.StringValue("text"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 32 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindBitLength(value.StringValue(""))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatal("expected 0")
	}

	if v, err := BindBitLength(nil); err != nil || v != nil {
		t.Fatalf("nil propagation failed: %v %v", v, err)
	}
	if _, err := BindBitLength(); err == nil {
		t.Fatal("expected error on 0 args")
	}
}

func TestBindHex(t *testing.T) {
	t.Parallel()

	// MySQL: HEX('abc') = '616263'. We uppercase.
	got, err := BindHex(value.StringValue("abc"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "616263" {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindHex(value.BytesValue{0xff, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "FF00" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindHex(nil); v != nil {
		t.Fatal("expected nil for null")
	}
	if _, err := BindHex(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpace(t *testing.T) {
	t.Parallel()

	// MySQL: SPACE(5) = '     '.
	got, err := BindSpace(value.IntValue(5))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "     " {
		t.Fatalf("got %q", mustString(t, got))
	}

	// MySQL: SPACE(0) = ''.
	got, err = BindSpace(value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty")
	}

	// Negative coerces to 0 (MySQL behavior).
	got, err = BindSpace(value.IntValue(-3))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty for negative")
	}

	if v, _ := BindSpace(nil); v != nil {
		t.Fatal("expected nil")
	}
	if _, err := BindSpace(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindPosition(t *testing.T) {
	t.Parallel()

	// MySQL: POSITION('bar' IN 'foobar') = 4.
	got, err := BindPosition(value.StringValue("bar"), value.StringValue("foobar"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 4 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindPosition(value.StringValue("xyz"), value.StringValue("foobar"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatal("expected 0 for not-found")
	}

	if v, _ := BindPosition(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null propagation")
	}
	if _, err := BindPosition(value.StringValue("a")); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindDegreesRadians(t *testing.T) {
	t.Parallel()

	// MySQL: DEGREES(PI()) = 180.
	got, err := BindDegrees(value.FloatValue(math.Pi))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(mustFloat64(t, got)-180) > 1e-9 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// MySQL: RADIANS(180) = PI.
	got, err = BindRadians(value.FloatValue(180))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(mustFloat64(t, got)-math.Pi) > 1e-9 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if v, _ := BindDegrees(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindRadians(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindDegrees(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindRadians(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindLog2(t *testing.T) {
	t.Parallel()

	got, err := BindLog2(value.FloatValue(8))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 3 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	got, err = BindLog2(value.FloatValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatal("expected 0")
	}

	if v, _ := BindLog2(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindLog2(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindTruncate(t *testing.T) {
	t.Parallel()

	// MySQL: TRUNCATE(1.999, 1) = 1.9.
	got, err := BindTruncate(value.FloatValue(1.999), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(mustFloat64(t, got)-1.9) > 1e-9 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	// MySQL: TRUNCATE(122.998, -2) = 100.
	got, err = BindTruncate(value.FloatValue(122.998), value.IntValue(-2))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 100 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if v, _ := BindTruncate(nil, value.IntValue(0)); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindTruncate(value.FloatValue(1)); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindWeekAliases(t *testing.T) {
	t.Parallel()

	// 2023-01-04 is in ISO week 1 of 2023.
	ts := value.TimestampValue(time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC))
	got, err := BindWeek(ts)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}
	got, err = BindWeekOfYear(ts)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if v, _ := BindWeek(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindWeekOfYear(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindWeek(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindWeekOfYear(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindUnixTimestampNow(t *testing.T) {
	t.Parallel()

	// Just verify it returns a reasonable INT64 in seconds since epoch.
	got, err := BindUnixTimestampNow()
	if err != nil {
		t.Fatal(err)
	}
	n := mustInt64(t, got)
	if n < 1_600_000_000 || n > 1<<33 {
		t.Fatalf("unexpected timestamp %d", n)
	}
}
