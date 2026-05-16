package time

import (
	"strings"
	"testing"
	gotime "time"

	"github.com/goccy/googlesqlite/internal/value"
)

// mkTime constructs a TimeValue at hh:mm:ss in UTC.
func mkTime(t *testing.T, hh, mm, ss int) value.TimeValue {
	t.Helper()
	return value.TimeValue(gotime.Date(1970, 1, 1, hh, mm, ss, 0, gotime.UTC))
}

func mustTimeString(t *testing.T, v value.Value) string {
	t.Helper()
	s, err := v.ToString()
	if err != nil {
		t.Fatalf("ToString: %v", err)
	}
	return s
}

// --- BindTime ---

// TestBindTimeFromHMS builds a TIME from (hour, minute, second).
func TestBindTimeFromHMS(t *testing.T) {
	got, err := BindTime(value.IntValue(15), value.IntValue(30), value.IntValue(45))
	if err != nil {
		t.Fatalf("BindTime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() != 15 || tt.Minute() != 30 || tt.Second() != 45 {
		t.Fatalf("want 15:30:45, got %02d:%02d:%02d", tt.Hour(), tt.Minute(), tt.Second())
	}
}

// TestBindTimeFromTimestamp extracts a TIME from a TIMESTAMP value.
func TestBindTimeFromTimestamp(t *testing.T) {
	ts := value.TimestampValue(gotime.Date(2024, 6, 15, 9, 0, 0, 0, gotime.UTC))
	got, err := BindTime(ts)
	if err != nil {
		t.Fatalf("BindTime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() != 9 {
		t.Fatalf("want hour 9, got %d", tt.Hour())
	}
}

// TestBindTimeFromTimestampZone honours an explicit time zone.
func TestBindTimeFromTimestampZone(t *testing.T) {
	ts := value.TimestampValue(gotime.Date(2024, 6, 15, 0, 0, 0, 0, gotime.UTC))
	got, err := BindTime(ts, value.StringValue("America/Los_Angeles"))
	if err != nil {
		t.Fatalf("BindTime: %v", err)
	}
	tt, _ := got.ToTime()
	// LA is UTC-7 (PDT) or UTC-8 — anything not 0 confirms we honoured the zone.
	if tt.Hour() == 0 {
		t.Fatalf("zone conversion did not apply, hour stayed at %d", tt.Hour())
	}
}

// TestBindTimeFromDatetime drops the date portion of a DATETIME.
func TestBindTimeFromDatetime(t *testing.T) {
	dt := value.DatetimeValue(gotime.Date(2024, 6, 15, 12, 34, 56, 0, gotime.UTC))
	got, err := BindTime(dt)
	if err != nil {
		t.Fatalf("BindTime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() != 12 || tt.Minute() != 34 {
		t.Fatalf("want 12:34, got %02d:%02d", tt.Hour(), tt.Minute())
	}
}

// TestBindTimeNullPropagation: any NULL argument returns SQL NULL.
func TestBindTimeNullPropagation(t *testing.T) {
	got, err := BindTime(nil, value.IntValue(0), value.IntValue(0))
	if err != nil {
		t.Fatalf("BindTime: %v", err)
	}
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindTimeArityError(t *testing.T) {
	if _, err := BindTime(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindTime(value.IntValue(0), value.IntValue(0), value.IntValue(0), value.IntValue(0)); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindTimeInvalidFirstArg(t *testing.T) {
	if _, err := BindTime(value.StringValue("nope")); err == nil {
		t.Fatalf("invalid first arg type should error")
	}
}

// --- BindTimeAdd ---

func TestBindTimeAdd(t *testing.T) {
	tv := mkTime(t, 10, 0, 0)
	cases := []struct {
		part string
		n    int64
		hh   int
		mm   int
		ss   int
	}{
		{"HOUR", 2, 12, 0, 0},
		{"MINUTE", 30, 10, 30, 0},
		{"SECOND", 15, 10, 0, 15},
	}
	for _, tc := range cases {
		got, err := BindTimeAdd(tv, value.IntValue(tc.n), value.StringValue(tc.part))
		if err != nil {
			t.Fatalf("BindTimeAdd(%s): %v", tc.part, err)
		}
		tt, _ := got.ToTime()
		if tt.Hour() != tc.hh || tt.Minute() != tc.mm || tt.Second() != tc.ss {
			t.Errorf("TIME_ADD %d %s -> %02d:%02d:%02d, want %02d:%02d:%02d",
				tc.n, tc.part, tt.Hour(), tt.Minute(), tt.Second(), tc.hh, tc.mm, tc.ss)
		}
	}
}

func TestBindTimeAddInvalidPart(t *testing.T) {
	tv := mkTime(t, 10, 0, 0)
	if _, err := BindTimeAdd(tv, value.IntValue(1), value.StringValue("DAY")); err == nil {
		t.Fatalf("invalid part should error")
	}
}

func TestBindTimeAddNullAndArity(t *testing.T) {
	got, _ := BindTimeAdd(nil, value.IntValue(1), value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimeAdd(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindTimeSub ---

func TestBindTimeSub(t *testing.T) {
	tv := mkTime(t, 10, 0, 0)
	got, err := BindTimeSub(tv, value.IntValue(2), value.StringValue("HOUR"))
	if err != nil {
		t.Fatalf("BindTimeSub: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() != 8 {
		t.Fatalf("want hour 8, got %d", tt.Hour())
	}

	got, _ = BindTimeSub(tv, value.IntValue(1), value.StringValue("MILLISECOND"))
	tt, _ = got.ToTime()
	if tt.Hour() != 9 || tt.Minute() != 59 || tt.Second() != 59 {
		t.Fatalf("want 09:59:59.x, got %02d:%02d:%02d", tt.Hour(), tt.Minute(), tt.Second())
	}

	got, _ = BindTimeSub(tv, value.IntValue(1), value.StringValue("MICROSECOND"))
	if got == nil {
		t.Fatalf("MICROSECOND should produce a value")
	}
	got, _ = BindTimeSub(tv, value.IntValue(1), value.StringValue("MINUTE"))
	if got == nil {
		t.Fatalf("MINUTE should produce a value")
	}
}

func TestBindTimeSubInvalidPart(t *testing.T) {
	tv := mkTime(t, 10, 0, 0)
	if _, err := BindTimeSub(tv, value.IntValue(1), value.StringValue("DAY")); err == nil {
		t.Fatalf("invalid part should error")
	}
}

func TestBindTimeSubNullAndArity(t *testing.T) {
	got, _ := BindTimeSub(nil, value.IntValue(1), value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimeSub(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// TestBindTimeAddAllParts hits the MICROSECOND and MILLISECOND paths.
func TestBindTimeAddAllParts(t *testing.T) {
	tv := mkTime(t, 10, 0, 0)
	for _, part := range []string{"MICROSECOND", "MILLISECOND"} {
		got, err := BindTimeAdd(tv, value.IntValue(1000), value.StringValue(part))
		if err != nil {
			t.Fatalf("BindTimeAdd(%s): %v", part, err)
		}
		if got == nil {
			t.Fatalf("BindTimeAdd(%s) returned nil", part)
		}
	}
}

// --- BindTimeDiff ---

func TestBindTimeDiff(t *testing.T) {
	a := mkTime(t, 12, 30, 0)
	b := mkTime(t, 12, 0, 0)
	cases := []struct {
		part string
		want int64
	}{
		{"HOUR", 0},
		{"MINUTE", 30},
		{"SECOND", 1800},
		{"MILLISECOND", 1800 * 1000},
		{"MICROSECOND", 1800 * 1000 * 1000},
	}
	for _, tc := range cases {
		got, err := BindTimeDiff(a, b, value.StringValue(tc.part))
		if err != nil {
			t.Fatalf("BindTimeDiff(%s): %v", tc.part, err)
		}
		n, _ := got.ToInt64()
		if n != tc.want {
			t.Errorf("TIME_DIFF %s = %d, want %d", tc.part, n, tc.want)
		}
	}
}

func TestBindTimeDiffInvalidPart(t *testing.T) {
	a := mkTime(t, 12, 30, 0)
	b := mkTime(t, 12, 0, 0)
	if _, err := BindTimeDiff(a, b, value.StringValue("DAY")); err == nil {
		t.Fatalf("invalid part should error")
	}
}

func TestBindTimeDiffNullAndArity(t *testing.T) {
	a := mkTime(t, 12, 30, 0)
	got, _ := BindTimeDiff(a, nil, value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimeDiff(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindTimeTrunc ---

func TestBindTimeTrunc(t *testing.T) {
	tv := value.TimeValue(gotime.Date(1970, 1, 1, 12, 34, 56, 123_000_000, gotime.UTC))
	cases := []struct {
		part string
		hh   int
		mm   int
		ss   int
	}{
		{"HOUR", 12, 0, 0},
		{"MINUTE", 12, 34, 0},
		{"SECOND", 12, 34, 0},
	}
	for _, tc := range cases {
		got, err := BindTimeTrunc(tv, value.StringValue(tc.part))
		if err != nil {
			t.Fatalf("BindTimeTrunc(%s): %v", tc.part, err)
		}
		tt, _ := got.ToTime()
		if tt.Hour() != tc.hh || tt.Minute() != tc.mm {
			t.Errorf("TIME_TRUNC %s -> %02d:%02d, want %02d:%02d",
				tc.part, tt.Hour(), tt.Minute(), tc.hh, tc.mm)
		}
	}

	// MICROSECOND / MILLISECOND branches exist but are effectively
	// passthrough for our integer-second TIME values; just confirm
	// they don't error.
	for _, part := range []string{"MICROSECOND", "MILLISECOND"} {
		if _, err := BindTimeTrunc(tv, value.StringValue(part)); err != nil {
			t.Errorf("TIME_TRUNC(%s): %v", part, err)
		}
	}
}

func TestBindTimeTruncInvalidPart(t *testing.T) {
	tv := mkTime(t, 12, 0, 0)
	if _, err := BindTimeTrunc(tv, value.StringValue("DAY")); err == nil {
		t.Fatalf("invalid part should error")
	}
}

func TestBindTimeTruncNullAndArity(t *testing.T) {
	got, _ := BindTimeTrunc(nil, value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimeTrunc(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindFormatTime / BindParseTime ---

func TestBindFormatTime(t *testing.T) {
	tv := mkTime(t, 9, 5, 7)
	got, err := BindFormatTime(value.StringValue("%H:%M:%S"), tv)
	if err != nil {
		t.Fatalf("BindFormatTime: %v", err)
	}
	if s := mustTimeString(t, got); s != "09:05:07" {
		t.Fatalf("want '09:05:07', got %q", s)
	}
}

func TestBindFormatTimeNullAndArity(t *testing.T) {
	got, _ := BindFormatTime(nil, mkTime(t, 1, 2, 3))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindFormatTime(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindParseTime(t *testing.T) {
	got, err := BindParseTime(value.StringValue("%H:%M:%S"), value.StringValue("13:45:30"))
	if err != nil {
		t.Fatalf("BindParseTime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() != 13 || tt.Minute() != 45 || tt.Second() != 30 {
		t.Fatalf("want 13:45:30, got %v", tt)
	}
}

func TestBindParseTimeError(t *testing.T) {
	if _, err := BindParseTime(value.StringValue("%H:%M:%S"), value.StringValue("not-a-time")); err == nil {
		t.Fatalf("invalid input should error")
	}
}

func TestBindParseTimeNullAndArity(t *testing.T) {
	got, _ := BindParseTime(nil, value.StringValue("13:00:00"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindParseTime(value.StringValue("%H")); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindCurrentTime ---

// TestBindCurrentTimeNoArgs just confirms the no-argument variant
// returns a TimeValue (the value itself is non-deterministic).
func TestBindCurrentTimeNoArgs(t *testing.T) {
	got, err := BindCurrentTime()
	if err != nil {
		t.Fatalf("BindCurrentTime: %v", err)
	}
	if _, ok := got.(value.TimeValue); !ok {
		t.Fatalf("want TimeValue, got %T", got)
	}
}

func TestBindCurrentTimeWithZone(t *testing.T) {
	got, err := BindCurrentTime(value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindCurrentTime: %v", err)
	}
	if _, ok := got.(value.TimeValue); !ok {
		t.Fatalf("want TimeValue, got %T", got)
	}
}

// TestBindCurrentTimeWithUnixNano: the 1-arg INT64 form converts the
// nanosecond instant via the runtime's local time helper; we only
// assert that the wall-clock fields round-trip through UTC.
func TestBindCurrentTimeWithUnixNano(t *testing.T) {
	when := gotime.Date(2024, 6, 15, 12, 34, 56, 0, gotime.UTC)
	got, err := BindCurrentTime(value.IntValue(when.UnixNano()))
	if err != nil {
		t.Fatalf("BindCurrentTime: %v", err)
	}
	tt, _ := got.ToTime()
	utc := tt.UTC()
	if utc.Hour() != 12 || utc.Minute() != 34 || utc.Second() != 56 {
		t.Fatalf("want 12:34:56 in UTC, got %v", tt)
	}
}

// TestBindCurrentTimeUnixNanoZone: 2-arg form takes (unix_nano, zone).
func TestBindCurrentTimeUnixNanoZone(t *testing.T) {
	when := gotime.Date(2024, 6, 15, 12, 0, 0, 0, gotime.UTC).UnixNano()
	got, err := BindCurrentTime(value.IntValue(when), value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindCurrentTime: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Hour() != 12 {
		t.Fatalf("want hour 12, got %d", tt.Hour())
	}
}

func TestBindCurrentTimeInvalidArg(t *testing.T) {
	if _, err := BindCurrentTime(value.BoolValue(true)); err == nil {
		t.Fatalf("invalid first arg type should error")
	}
}

// TestBindCurrentTimeInvalidZone: an unparseable IANA time zone must
// surface an error.
func TestBindCurrentTimeInvalidZone(t *testing.T) {
	if _, err := BindCurrentTime(value.StringValue("Not/A_Real_Zone")); err == nil {
		t.Fatalf("invalid zone should error")
	}
}

// Smoke check the CURRENT_TIME_WITH_TIME helper directly.
func TestCurrentTimeWithTime(t *testing.T) {
	now := gotime.Now()
	got, err := CURRENT_TIME_WITH_TIME(now)
	if err != nil {
		t.Fatalf("CURRENT_TIME_WITH_TIME: %v", err)
	}
	tt, _ := got.ToTime()
	if !tt.Equal(now) {
		t.Fatalf("want %v, got %v", now, tt)
	}
}

// silence unused import warnings if helpers above are trimmed later.
var _ = strings.Contains
