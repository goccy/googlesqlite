package datetime

import (
	"testing"
	gotime "time"

	"github.com/goccy/googlesqlite/internal/value"
)

func mkDate(y int, m gotime.Month, d int) value.Value {
	return value.DateValue(gotime.Date(y, m, d, 0, 0, 0, 0, gotime.UTC))
}

func mkDatetime(y int, m gotime.Month, d, hh, mm, ss int) value.Value {
	return value.DatetimeValue(gotime.Date(y, m, d, hh, mm, ss, 0, gotime.UTC))
}

func mkTimestamp(y int, m gotime.Month, d, hh, mm, ss int) value.Value {
	return value.TimestampValue(gotime.Date(y, m, d, hh, mm, ss, 0, gotime.UTC))
}

func mustInterval(t *testing.T, s string) value.Value {
	t.Helper()
	iv, err := value.ParseInterval(s)
	if err != nil {
		t.Fatalf("ParseInterval(%q): %v", s, err)
	}
	return iv
}

// --- BindDatetime ---

func TestBindDatetimeFromYMDHMS(t *testing.T) {
	got, err := BindDatetime(
		value.IntValue(2024), value.IntValue(6), value.IntValue(15),
		value.IntValue(10), value.IntValue(30), value.IntValue(45),
	)
	if err != nil {
		t.Fatalf("BindDatetime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Year() != 2024 || tt.Month() != 6 || tt.Hour() != 10 {
		t.Fatalf("want 2024-06-15 10:..., got %v", tt)
	}
}

func TestBindDatetimeFromDate(t *testing.T) {
	got, _ := BindDatetime(mkDate(2024, 6, 15))
	tt, _ := got.ToTime()
	if tt.Year() != 2024 || tt.Month() != 6 || tt.Day() != 15 {
		t.Fatalf("want 2024-06-15, got %v", tt)
	}
}

func TestBindDatetimeFromDatePlusTime(t *testing.T) {
	tv := value.TimeValue(gotime.Date(1970, 1, 1, 9, 0, 0, 0, gotime.UTC))
	got, _ := BindDatetime(mkDate(2024, 6, 15), tv)
	tt, _ := got.ToTime()
	if tt.Year() != 2024 || tt.Day() != 15 || tt.Hour() != 9 {
		t.Fatalf("want 2024-06-15 09:00, got %v", tt)
	}
}

func TestBindDatetimeFromTimestamp(t *testing.T) {
	got, _ := BindDatetime(mkTimestamp(2024, 6, 15, 12, 0, 0))
	tt, _ := got.ToTime()
	if tt.Year() != 2024 || tt.Day() != 15 || tt.Hour() != 12 {
		t.Fatalf("want 2024-06-15 12:00, got %v", tt)
	}
}

func TestBindDatetimeFromTimestampZone(t *testing.T) {
	got, err := BindDatetime(
		mkTimestamp(2024, 6, 15, 0, 0, 0),
		value.StringValue("America/Los_Angeles"),
	)
	if err != nil {
		t.Fatalf("BindDatetime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() == 0 {
		t.Fatalf("zone conversion did not apply, hour stayed 0")
	}
}

func TestBindDatetimeNullPropagation(t *testing.T) {
	got, _ := BindDatetime(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindDatetimeArityAndType(t *testing.T) {
	if _, err := BindDatetime(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindDatetime(value.StringValue("nope")); err == nil {
		t.Fatalf("invalid first arg type should error")
	}
}

// --- BindDatetimeAdd ---

func TestBindDatetimeAdd(t *testing.T) {
	dt := mkDatetime(2024, 6, 15, 10, 0, 0)
	cases := []struct {
		part string
		n    int64
		want int
	}{
		{"HOUR", 2, 12},
		{"MINUTE", 30, 10},
		{"SECOND", 15, 10},
		{"MILLISECOND", 1000, 10},
		{"MICROSECOND", 1000, 10},
	}
	for _, tc := range cases {
		got, err := BindDatetimeAdd(dt, value.IntValue(tc.n), value.StringValue(tc.part))
		if err != nil {
			t.Errorf("BindDatetimeAdd(%s): %v", tc.part, err)
			continue
		}
		tt, _ := got.ToTime()
		if tt.Hour() != tc.want {
			t.Errorf("DATETIME_ADD %s -> hour %d, want %d", tc.part, tt.Hour(), tc.want)
		}
	}

	// DAY routes to DATE_ADD.
	got, _ := BindDatetimeAdd(dt, value.IntValue(2), value.StringValue("DAY"))
	tt, _ := got.ToTime()
	if tt.Day() != 17 {
		t.Errorf("DATETIME_ADD DAY -> day %d, want 17", tt.Day())
	}
}

func TestBindDatetimeAddInvalidPart(t *testing.T) {
	dt := mkDatetime(2024, 6, 15, 10, 0, 0)
	if _, err := BindDatetimeAdd(dt, value.IntValue(1), value.StringValue("FORTNIGHT")); err == nil {
		t.Fatalf("invalid part should error")
	}
}

func TestBindDatetimeAddNullAndArity(t *testing.T) {
	got, _ := BindDatetimeAdd(nil, value.IntValue(1), value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDatetimeAdd(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDatetimeSub ---

func TestBindDatetimeSub(t *testing.T) {
	dt := mkDatetime(2024, 6, 15, 10, 0, 0)
	for _, part := range []string{"HOUR", "MINUTE", "SECOND", "MILLISECOND", "MICROSECOND"} {
		got, err := BindDatetimeSub(dt, value.IntValue(1), value.StringValue(part))
		if err != nil {
			t.Errorf("BindDatetimeSub(%s): %v", part, err)
		}
		if got == nil {
			t.Errorf("BindDatetimeSub(%s) returned nil", part)
		}
	}

	got, _ := BindDatetimeSub(dt, value.IntValue(5), value.StringValue("DAY"))
	tt, _ := got.ToTime()
	if tt.Day() != 10 {
		t.Fatalf("DATETIME_SUB 5 DAY -> day %d, want 10", tt.Day())
	}

	got, _ = BindDatetimeSub(dt, value.IntValue(1), value.StringValue("MONTH"))
	tt, _ = got.ToTime()
	if tt.Month() != 5 {
		t.Fatalf("DATETIME_SUB 1 MONTH -> month %d, want 5", tt.Month())
	}

	if _, err := BindDatetimeSub(dt, value.IntValue(1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ = BindDatetimeSub(nil, value.IntValue(1), value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDatetimeSub(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDatetimeDiff ---

func TestBindDatetimeDiff(t *testing.T) {
	a := mkDatetime(2024, 6, 15, 12, 0, 0)
	b := mkDatetime(2024, 6, 15, 10, 0, 0)
	got, _ := BindDatetimeDiff(a, b, value.StringValue("HOUR"))
	n, _ := got.ToInt64()
	if n != 2 {
		t.Fatalf("DATETIME_DIFF HOUR = %d, want 2", n)
	}

	got, _ = BindDatetimeDiff(a, b, value.StringValue("MINUTE"))
	n, _ = got.ToInt64()
	if n != 120 {
		t.Fatalf("DATETIME_DIFF MINUTE = %d, want 120", n)
	}

	got, _ = BindDatetimeDiff(a, b, value.StringValue("SECOND"))
	n, _ = got.ToInt64()
	if n != 7200 {
		t.Fatalf("DATETIME_DIFF SECOND = %d, want 7200", n)
	}

	got, _ = BindDatetimeDiff(a, b, value.StringValue("MILLISECOND"))
	n, _ = got.ToInt64()
	if n != 7200*1000 {
		t.Fatalf("DATETIME_DIFF MILLISECOND = %d", n)
	}

	got, _ = BindDatetimeDiff(a, b, value.StringValue("MICROSECOND"))
	if got == nil {
		t.Fatalf("MICROSECOND should produce a value")
	}

	// DAY routes to DATE_DIFF.
	a2 := mkDatetime(2024, 6, 20, 0, 0, 0)
	got, _ = BindDatetimeDiff(a2, b, value.StringValue("DAY"))
	n, _ = got.ToInt64()
	if n != 5 {
		t.Fatalf("DATETIME_DIFF DAY = %d, want 5", n)
	}

	if _, err := BindDatetimeDiff(a, b, value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ = BindDatetimeDiff(nil, b, value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDatetimeDiff(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDatetimeTrunc ---

func TestBindDatetimeTrunc(t *testing.T) {
	dt := value.DatetimeValue(gotime.Date(2024, 6, 15, 12, 34, 56, 123_000_000, gotime.UTC))

	got, _ := BindDatetimeTrunc(dt, value.StringValue("HOUR"))
	tt, _ := got.ToTime()
	if tt.Hour() != 12 || tt.Minute() != 0 {
		t.Fatalf("TRUNC HOUR -> %02d:%02d, want 12:00", tt.Hour(), tt.Minute())
	}

	got, _ = BindDatetimeTrunc(dt, value.StringValue("DAY"))
	tt, _ = got.ToTime()
	if tt.Hour() != 0 || tt.Day() != 15 {
		t.Fatalf("TRUNC DAY -> hour %d / day %d, want 0/15", tt.Hour(), tt.Day())
	}

	for _, part := range []string{"MINUTE", "SECOND", "MILLISECOND", "MICROSECOND"} {
		if _, err := BindDatetimeTrunc(dt, value.StringValue(part)); err != nil {
			t.Errorf("DATETIME_TRUNC(%s): %v", part, err)
		}
	}

	if _, err := BindDatetimeTrunc(dt, value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ = BindDatetimeTrunc(nil, value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDatetimeTrunc(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDatetimeBucket ---

func TestBindDatetimeBucket(t *testing.T) {
	target := mkDatetime(2024, 6, 15, 5, 0, 0)
	iv := mustInterval(t, "0-0 1 0:0:0") // 1 day
	got, err := BindDatetimeBucket(target, iv)
	if err != nil {
		t.Fatalf("BindDatetimeBucket: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() != 0 || tt.Year() != 2024 || tt.Day() != 15 {
		t.Fatalf("bucket floor -> %v, want 2024-06-15 00:00", tt)
	}

	// With explicit origin.
	origin := mkDatetime(2024, 1, 1, 0, 0, 0)
	got, err = BindDatetimeBucket(target, iv, origin)
	if err != nil {
		t.Fatalf("BindDatetimeBucket with origin: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil bucket")
	}
}

func TestBindDatetimeBucketErrors(t *testing.T) {
	target := mkDatetime(2024, 6, 15, 5, 0, 0)
	if _, err := BindDatetimeBucket(target, value.IntValue(1)); err == nil {
		t.Fatalf("non-INTERVAL bucket width should error")
	}
	got, _ := BindDatetimeBucket(nil, mustInterval(t, "0-0 1 0:0:0"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDatetimeBucket(target); err == nil {
		t.Fatalf("arity error expected (1)")
	}
}

// --- BindFormatDatetime / BindParseDatetime ---

func TestBindFormatAndParseDatetime(t *testing.T) {
	dt := mkDatetime(2024, 6, 15, 9, 5, 7)
	got, err := BindFormatDatetime(value.StringValue("%Y-%m-%d %H:%M:%S"), dt)
	if err != nil {
		t.Fatalf("BindFormatDatetime: %v", err)
	}
	s, _ := got.ToString()
	if s != "2024-06-15 09:05:07" {
		t.Fatalf("format = %q, want 2024-06-15 09:05:07", s)
	}

	got, err = BindParseDatetime(value.StringValue("%Y-%m-%d %H:%M:%S"), value.StringValue("2024-06-15 09:05:07"))
	if err != nil {
		t.Fatalf("BindParseDatetime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Year() != 2024 || tt.Hour() != 9 {
		t.Fatalf("parse -> %v", tt)
	}
}

func TestBindFormatDatetimeNullAndArity(t *testing.T) {
	got, _ := BindFormatDatetime(nil, mkDatetime(2024, 1, 1, 0, 0, 0))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindFormatDatetime(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindParseDatetimeError(t *testing.T) {
	if _, err := BindParseDatetime(value.StringValue("%Y"), value.StringValue("not-a-date")); err == nil {
		t.Fatalf("invalid input should error")
	}
	got, _ := BindParseDatetime(nil, value.StringValue("2024-01-01"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindParseDatetime(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindCurrentDatetime ---

func TestBindCurrentDatetimeNoArgs(t *testing.T) {
	got, err := BindCurrentDatetime()
	if err != nil {
		t.Fatalf("BindCurrentDatetime: %v", err)
	}
	if _, ok := got.(value.DatetimeValue); !ok {
		t.Fatalf("want DatetimeValue, got %T", got)
	}
}

func TestBindCurrentDatetimeWithZone(t *testing.T) {
	got, err := BindCurrentDatetime(value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindCurrentDatetime: %v", err)
	}
	if _, ok := got.(value.DatetimeValue); !ok {
		t.Fatalf("want DatetimeValue, got %T", got)
	}
}

func TestBindCurrentDatetimeWithUnixNano(t *testing.T) {
	when := gotime.Date(2024, 6, 15, 12, 0, 0, 0, gotime.UTC)
	got, _ := BindCurrentDatetime(value.IntValue(when.UnixNano()))
	tt, _ := got.ToTime()
	utc := tt.UTC()
	if utc.Year() != 2024 || utc.Hour() != 12 {
		t.Fatalf("want 2024 12:..UTC, got %v", tt)
	}
}

func TestBindCurrentDatetimeUnixNanoZone(t *testing.T) {
	when := gotime.Date(2024, 6, 15, 12, 0, 0, 0, gotime.UTC).UnixNano()
	got, _ := BindCurrentDatetime(value.IntValue(when), value.StringValue("UTC"))
	tt, _ := got.ToTime()
	if tt.Year() != 2024 {
		t.Fatalf("want year 2024, got %d", tt.Year())
	}
}

func TestBindCurrentDatetimeInvalidZone(t *testing.T) {
	if _, err := BindCurrentDatetime(value.StringValue("Not/A_Real_Zone")); err == nil {
		t.Fatalf("invalid zone should error")
	}
}

func TestBindCurrentDatetimeInvalidType(t *testing.T) {
	if _, err := BindCurrentDatetime(value.BoolValue(true)); err == nil {
		t.Fatalf("invalid arg type should error")
	}
}

// Smoke check CURRENT_DATETIME_WITH_TIME directly.
func TestCurrentDatetimeWithTime(t *testing.T) {
	now := gotime.Now()
	got, _ := CURRENT_DATETIME_WITH_TIME(now)
	tt, _ := got.ToTime()
	if !tt.Equal(now) {
		t.Fatalf("want %v, got %v", now, tt)
	}
}
