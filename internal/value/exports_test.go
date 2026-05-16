package value_test

import (
	"strings"
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

// TestExportsParsers covers Parse{Date,Datetime,Time,Timestamp,Interval}
// plus the Is{Date,Datetime,Time,Timestamp,NullValue} classifiers.
func TestExportsParsers(t *testing.T) {
	t.Parallel()

	t.Run("ParseDate", func(t *testing.T) {
		got, err := value.ParseDate("2020-01-02")
		if err != nil {
			t.Fatal(err)
		}
		if got.IsZero() {
			t.Fatal("zero")
		}
		if _, err := value.ParseDate("bogus"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ParseDatetime both formats", func(t *testing.T) {
		// T-separator form
		if _, err := value.ParseDatetime("2020-01-02T03:04:05"); err != nil {
			t.Fatal(err)
		}
		// space-separator form
		if _, err := value.ParseDatetime("2020-01-02 03:04:05"); err != nil {
			t.Fatal(err)
		}
		if _, err := value.ParseDatetime("bogus"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ParseTime", func(t *testing.T) {
		if _, err := value.ParseTime("03:04:05"); err != nil {
			t.Fatal(err)
		}
		if _, err := value.ParseTime("bogus"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ParseTimestamp covers many formats", func(t *testing.T) {
		cases := []string{
			"2020-01-02T03:04:05Z",
			"2020-01-02T03:04:05+09:00",
			"2020-01-02T03:04:05-07",
			"2020-01-02T03:04:05",
			"2020-01-02 03:04:05",
			"2020-01-02 03:04:05+09:00",
			"2020-01-02 03:04:05-07",
			"2020-01-02 03:04:05.000000",
			"2020-01-02",
		}
		for _, c := range cases {
			t.Run(c, func(t *testing.T) {
				if _, err := value.ParseTimestamp(c, time.UTC); err != nil {
					t.Fatalf("input %q: %v", c, err)
				}
			})
		}
		if _, err := value.ParseTimestamp("bogus", time.UTC); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("ParseInterval", func(t *testing.T) {
		got, err := value.ParseInterval("0-0 1 0:0:0")
		if err != nil {
			t.Fatalf("ParseInterval: %v", err)
		}
		if got == nil {
			t.Fatal("nil interval")
		}
		if _, err := value.ParseInterval(""); err == nil {
			t.Fatal("expected error")
		}
		if _, err := value.ParseInterval("garbage"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("IsDate/IsDatetime/IsTime/IsTimestamp", func(t *testing.T) {
		if !value.IsDate("2020-01-02") {
			t.Fatal("IsDate true")
		}
		if value.IsDate("nope") {
			t.Fatal("IsDate false")
		}
		if !value.IsDatetime("2020-01-02 03:04:05") {
			t.Fatal("IsDatetime true space")
		}
		if !value.IsDatetime("2020-01-02T03:04:05") {
			t.Fatal("IsDatetime true T")
		}
		if value.IsDatetime("nope") {
			t.Fatal("IsDatetime false")
		}
		if !value.IsTime("03:04:05") {
			t.Fatal("IsTime true")
		}
		if value.IsTime("nope") {
			t.Fatal("IsTime false")
		}
		if !value.IsTimestamp("2020-01-02T03:04:05Z") {
			t.Fatal("IsTimestamp true")
		}
		if value.IsTimestamp("nope") {
			t.Fatal("IsTimestamp false")
		}
	})

	t.Run("IsNullValue", func(t *testing.T) {
		if !value.IsNullValue(nil) {
			t.Fatal("nil should be null")
		}
		if !value.IsNullValue([]byte(nil)) {
			t.Fatal("nil []byte should be null")
		}
		if value.IsNullValue([]byte{1}) {
			t.Fatal("non-nil []byte should not be null")
		}
		if value.IsNullValue(0) {
			t.Fatal("0 should not be null")
		}
	})

	t.Run("TimeFromUnixNano", func(t *testing.T) {
		got := value.TimeFromUnixNano(1_000_000_000)
		if got.IsZero() {
			t.Fatal("zero")
		}
	})

	t.Run("DateFromInt64Value/TimestampFromInt64Value/Float", func(t *testing.T) {
		if _, err := value.DateFromInt64Value(0); err != nil {
			t.Fatal(err)
		}
		if _, err := value.TimestampFromInt64Value(int64(time.Millisecond)); err != nil {
			t.Fatal(err)
		}
		if _, err := value.TimestampFromFloatValue(1.5); err != nil {
			t.Fatal(err)
		}
	})
}

// TestLocationCache exercises ToLocation across plain IANA names,
// numeric offsets, and named offsets, and verifies cache hits are
// consistent.
func TestLocationCache(t *testing.T) {
	t.Parallel()

	t.Run("UTC", func(t *testing.T) {
		loc, err := value.ToLocation("UTC")
		if err != nil {
			t.Fatal(err)
		}
		if loc == nil {
			t.Fatal("nil location")
		}
	})

	t.Run("offset full", func(t *testing.T) {
		loc, err := value.ToLocation("+09:00")
		if err != nil {
			t.Fatal(err)
		}
		if loc == nil {
			t.Fatal("nil location")
		}
	})

	t.Run("offset partial", func(t *testing.T) {
		loc, err := value.ToLocation("+09")
		if err != nil {
			t.Fatal(err)
		}
		if loc == nil {
			t.Fatal("nil location")
		}
	})

	t.Run("cache hit", func(t *testing.T) {
		// First call populates cache; second call uses cache path.
		loc1, _ := value.ToLocation("America/Los_Angeles")
		loc2, _ := value.ToLocation("America/Los_Angeles")
		if loc1.String() != loc2.String() {
			t.Fatalf("loc mismatch")
		}
	})

	t.Run("unknown returns error", func(t *testing.T) {
		_, err := value.ToLocation("Not/A/Real/Place")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "failed to load location") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestModifyTimeZone verifies that ModifyTimeZone reinterprets the
// wall-clock components in the target location.
func TestModifyTimeZone(t *testing.T) {
	t.Parallel()

	loc, _ := value.ToLocation("America/Los_Angeles")
	src := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	got, err := value.ModifyTimeZone(src, loc)
	if err != nil {
		t.Fatal(err)
	}
	// Wall-clock should preserve: 2020-01-02 03:04:05 in LA.
	if got.Year() != 2020 || got.Month() != 1 || got.Day() != 2 ||
		got.Hour() != 3 || got.Minute() != 4 || got.Second() != 5 {
		t.Fatalf("ModifyTimeZone wall-clock changed: %v", got)
	}
	if got.Location() != loc {
		t.Fatalf("Location not switched: got %v, want %v", got.Location(), loc)
	}
}
