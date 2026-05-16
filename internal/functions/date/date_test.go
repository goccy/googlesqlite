package date

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func mkDate(y int, m time.Month, d int) value.Value {
	return value.DateValue(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func mustInterval(t *testing.T, s string) value.Value {
	t.Helper()
	iv, err := value.ParseInterval(s)
	if err != nil {
		t.Fatalf("ParseInterval(%q): %v", s, err)
	}
	return iv
}

// --- BindDate ---

func TestBindDateFromYMD(t *testing.T) {
	got, err := BindDate(value.IntValue(2024), value.IntValue(6), value.IntValue(15))
	if err != nil {
		t.Fatalf("BindDate: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Year() != 2024 || tt.Month() != 6 || tt.Day() != 15 {
		t.Fatalf("want 2024-06-15, got %v", tt)
	}
}

func TestBindDateFromTimestampZone(t *testing.T) {
	ts := value.TimestampValue(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC))
	got, err := BindDate(ts, value.StringValue("America/Los_Angeles"))
	if err != nil {
		t.Fatalf("BindDate: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Day() != 14 {
		t.Fatalf("LA time should be 2024-06-14, got %v", tt)
	}
}

func TestBindDateFromDate(t *testing.T) {
	got, _ := BindDate(mkDate(2024, 6, 15))
	tt, _ := got.ToTime()
	if tt.Day() != 15 {
		t.Fatalf("want day 15, got %d", tt.Day())
	}
}

func TestBindDateNullPropagation(t *testing.T) {
	got, _ := BindDate(nil, value.IntValue(1), value.IntValue(2))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

// --- BindDateAdd ---

func TestBindDateAdd(t *testing.T) {
	d := mkDate(2024, 1, 15)
	cases := []struct {
		part string
		n    int64
		want time.Time
	}{
		{"DAY", 5, time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		{"WEEK", 2, time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC)},
		{"MONTH", 2, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{"YEAR", 1, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		got, err := BindDateAdd(d, value.IntValue(tc.n), value.StringValue(tc.part))
		if err != nil {
			t.Errorf("BindDateAdd(%s): %v", tc.part, err)
			continue
		}
		tt, _ := got.ToTime()
		if !tt.Equal(tc.want) {
			t.Errorf("DATE_ADD %s -> %v, want %v", tc.part, tt, tc.want)
		}
	}
}

func TestBindDateAddInvalid(t *testing.T) {
	if _, err := BindDateAdd(mkDate(2024, 1, 1), value.IntValue(1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ := BindDateAdd(nil, value.IntValue(1), value.StringValue("DAY"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDateAdd(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDateSub ---

func TestBindDateSub(t *testing.T) {
	d := mkDate(2024, 1, 15)
	cases := []struct {
		part string
		n    int64
		want time.Time
	}{
		{"DAY", 5, time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
		{"WEEK", 2, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"MONTH", 2, time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC)},
		{"YEAR", 1, time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		got, err := BindDateSub(d, value.IntValue(tc.n), value.StringValue(tc.part))
		if err != nil {
			t.Errorf("BindDateSub(%s): %v", tc.part, err)
			continue
		}
		tt, _ := got.ToTime()
		if !tt.Equal(tc.want) {
			t.Errorf("DATE_SUB %s -> %v, want %v", tc.part, tt, tc.want)
		}
	}

	if _, err := BindDateSub(d, value.IntValue(1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ := BindDateSub(nil, value.IntValue(1), value.StringValue("DAY"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDateSub(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDateDiff ---

func TestBindDateDiff(t *testing.T) {
	a := mkDate(2024, 12, 31)
	b := mkDate(2024, 1, 1)
	cases := []struct {
		part string
		want int64
	}{
		{"DAY", 365},
		{"MONTH", 11},
		{"YEAR", 0},
		{"QUARTER", 3},
	}
	for _, tc := range cases {
		got, err := BindDateDiff(a, b, value.StringValue(tc.part))
		if err != nil {
			t.Errorf("BindDateDiff(%s): %v", tc.part, err)
			continue
		}
		n, _ := got.ToInt64()
		if n != tc.want {
			t.Errorf("DATE_DIFF %s = %d, want %d", tc.part, n, tc.want)
		}
	}
}

func TestBindDateDiffWeekVariants(t *testing.T) {
	a := mkDate(2024, 1, 15)
	b := mkDate(2024, 1, 1)
	for _, part := range []string{"WEEK", "ISOWEEK"} {
		got, err := BindDateDiff(a, b, value.StringValue(part))
		if err != nil {
			t.Errorf("BindDateDiff(%s): %v", part, err)
		}
		if got == nil {
			t.Errorf("DATE_DIFF(%s) returned nil", part)
		}
	}
}

func TestBindDateDiffInvalid(t *testing.T) {
	if _, err := BindDateDiff(mkDate(2024, 1, 1), mkDate(2024, 1, 1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ := BindDateDiff(nil, mkDate(2024, 1, 1), value.StringValue("DAY"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDateDiff(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDateTrunc ---

func TestBindDateTrunc(t *testing.T) {
	d := mkDate(2024, 6, 15)
	cases := []struct {
		part string
		want time.Time
	}{
		{"DAY", time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)},
		{"MONTH", time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
		{"YEAR", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"QUARTER", time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		got, err := BindDateTrunc(d, value.StringValue(tc.part))
		if err != nil {
			t.Errorf("BindDateTrunc(%s): %v", tc.part, err)
			continue
		}
		tt, _ := got.ToTime()
		if !tt.Equal(tc.want) {
			t.Errorf("DATE_TRUNC %s -> %v, want %v", tc.part, tt, tc.want)
		}
	}
}

func TestBindDateTruncWeek(t *testing.T) {
	d := mkDate(2024, 6, 15) // a Saturday
	got, err := BindDateTrunc(d, value.StringValue("WEEK"))
	if err != nil {
		t.Fatalf("BindDateTrunc WEEK: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Weekday() != time.Sunday {
		t.Fatalf("WEEK truncated to %v, want Sunday", tt.Weekday())
	}

	got, err = BindDateTrunc(d, value.StringValue("ISOWEEK"))
	if err != nil {
		t.Fatalf("BindDateTrunc ISOWEEK: %v", err)
	}
	tt, _ = got.ToTime()
	if tt.Weekday() != time.Monday {
		t.Fatalf("ISOWEEK truncated to %v, want Monday", tt.Weekday())
	}
}

func TestBindDateTruncInvalid(t *testing.T) {
	if _, err := BindDateTrunc(mkDate(2024, 1, 1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ := BindDateTrunc(nil, value.StringValue("DAY"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDateTrunc(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindDateFromUnixDate / BindUnixDate ---

func TestBindDateFromUnixDate(t *testing.T) {
	// Unix epoch day 0 → 1970-01-01.
	got, err := BindDateFromUnixDate(value.IntValue(0))
	if err != nil {
		t.Fatalf("BindDateFromUnixDate: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Year() != 1970 || tt.Month() != 1 || tt.Day() != 1 {
		t.Fatalf("want 1970-01-01, got %v", tt)
	}

	got, _ = BindDateFromUnixDate(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDateFromUnixDate(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindUnixDateRoundTrip(t *testing.T) {
	got, err := BindUnixDate(mkDate(1970, 1, 2))
	if err != nil {
		t.Fatalf("BindUnixDate: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 1 {
		t.Fatalf("want 1 unix day, got %d", n)
	}

	got, _ = BindUnixDate(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindUnixDate(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindLastDay ---

func TestBindLastDay(t *testing.T) {
	d := mkDate(2024, 2, 15)
	got, _ := BindLastDay(d) // default MONTH
	tt, _ := got.ToTime()
	if tt.Day() != 29 || tt.Month() != 2 {
		t.Fatalf("Feb 2024 last day = %v, want 29", tt)
	}

	got, _ = BindLastDay(d, value.StringValue("YEAR"))
	tt, _ = got.ToTime()
	if tt.Year() != 2024 || tt.Month() != 12 || tt.Day() != 31 {
		t.Fatalf("YEAR last day = %v, want 2024-12-31", tt)
	}

	for _, part := range []string{"WEEK", "WEEK_MONDAY", "WEEK_TUESDAY", "WEEK_WEDNESDAY", "WEEK_THURSDAY", "WEEK_FRIDAY", "WEEK_SATURDAY", "ISOWEEK", "ISOYEAR"} {
		if _, err := BindLastDay(d, value.StringValue(part)); err != nil {
			t.Errorf("LAST_DAY(%s): %v", part, err)
		}
	}

	if _, err := BindLastDay(d, value.StringValue("QUARTER")); err == nil {
		t.Fatalf("QUARTER is documented as unimplemented and should error")
	}
	if _, err := BindLastDay(d, value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ = BindLastDay(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindLastDay(); err == nil {
		t.Fatalf("arity error expected (0)")
	}
}

// --- BindDateBucket ---

func TestBindDateBucket(t *testing.T) {
	target := mkDate(2024, 6, 15)
	iv := mustInterval(t, "0-0 7 0:0:0") // 7 days
	got, err := BindDateBucket(target, iv)
	if err != nil {
		t.Fatalf("BindDateBucket: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil bucket")
	}

	// With origin.
	got, err = BindDateBucket(target, iv, mkDate(2024, 1, 1))
	if err != nil {
		t.Fatalf("BindDateBucket with origin: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil bucket")
	}
}

func TestBindDateBucketErrors(t *testing.T) {
	target := mkDate(2024, 6, 15)
	// Non-INTERVAL bucket width.
	if _, err := BindDateBucket(target, value.IntValue(1)); err == nil {
		t.Fatalf("non-INTERVAL should error")
	}
	// Interval with time parts.
	if _, err := BindDateBucket(target, mustInterval(t, "0-0 0 1:0:0")); err == nil {
		t.Fatalf("interval with HOUR should error")
	}
	got, _ := BindDateBucket(nil, mustInterval(t, "0-0 7 0:0:0"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindDateBucket(target); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindFormatDate / BindParseDate ---

func TestBindFormatAndParseDate(t *testing.T) {
	got, err := BindFormatDate(value.StringValue("%Y-%m-%d"), mkDate(2024, 6, 15))
	if err != nil {
		t.Fatalf("BindFormatDate: %v", err)
	}
	s, _ := got.ToString()
	if s != "2024-06-15" {
		t.Fatalf("format = %q, want 2024-06-15", s)
	}

	got, err = BindParseDate(value.StringValue("%Y-%m-%d"), value.StringValue("2024-06-15"))
	if err != nil {
		t.Fatalf("BindParseDate: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.Year() != 2024 || tt.Day() != 15 {
		t.Fatalf("parse -> %v", tt)
	}
}

func TestBindFormatDateNullAndArity(t *testing.T) {
	got, _ := BindFormatDate(nil, mkDate(2024, 1, 1))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindFormatDate(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindParseDateError(t *testing.T) {
	if _, err := BindParseDate(value.StringValue("%Y"), value.StringValue("nope")); err == nil {
		t.Fatalf("invalid input should error")
	}
	got, _ := BindParseDate(nil, value.StringValue("2024"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindParseDate(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindCurrentDate ---

func TestBindCurrentDate(t *testing.T) {
	got, err := BindCurrentDate()
	if err != nil {
		t.Fatalf("BindCurrentDate: %v", err)
	}
	if _, ok := got.(value.DateValue); !ok {
		t.Fatalf("want DateValue, got %T", got)
	}
}

func TestBindCurrentDateWithZone(t *testing.T) {
	got, err := BindCurrentDate(value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindCurrentDate: %v", err)
	}
	if _, ok := got.(value.DateValue); !ok {
		t.Fatalf("want DateValue, got %T", got)
	}
}

func TestBindCurrentDateWithUnixNano(t *testing.T) {
	when := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).UnixNano()
	got, _ := BindCurrentDate(value.IntValue(when))
	tt, _ := got.ToTime()
	utc := tt.UTC()
	if utc.Year() != 2024 || utc.Month() != 6 {
		t.Fatalf("want 2024-06-..UTC, got %v", tt)
	}
}

func TestBindCurrentDateUnixNanoZone(t *testing.T) {
	when := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).UnixNano()
	got, _ := BindCurrentDate(value.IntValue(when), value.StringValue("UTC"))
	if got == nil {
		t.Fatalf("expected non-nil result")
	}
}

func TestBindCurrentDateInvalidArg(t *testing.T) {
	if _, err := BindCurrentDate(value.BoolValue(true)); err == nil {
		t.Fatalf("invalid arg type should error")
	}
	if _, err := BindCurrentDate(value.StringValue("Bad/Zone")); err == nil {
		t.Fatalf("invalid zone should error")
	}
}

// --- BindExtract ---

func TestBindExtract(t *testing.T) {
	d := mkDate(2024, 6, 15)
	cases := []struct {
		part string
		want int64
	}{
		{"YEAR", 2024},
		{"MONTH", 6},
		{"DAY", 15},
		{"QUARTER", 2},
		{"DAYOFYEAR", 167},
	}
	for _, tc := range cases {
		got, err := BindExtract(d, value.StringValue(tc.part))
		if err != nil {
			t.Errorf("BindExtract(%s): %v", tc.part, err)
			continue
		}
		n, _ := got.ToInt64()
		if n != tc.want {
			t.Errorf("EXTRACT %s = %d, want %d", tc.part, n, tc.want)
		}
	}
}

func TestBindExtractTimestamp(t *testing.T) {
	ts := value.TimestampValue(time.Date(2024, 6, 15, 12, 30, 45, 123_000_000, time.UTC))
	for _, part := range []string{"YEAR", "HOUR", "MINUTE", "SECOND", "MILLISECOND", "MICROSECOND", "ISOYEAR", "ISOWEEK", "WEEK"} {
		if _, err := BindExtract(ts, value.StringValue(part)); err != nil {
			t.Errorf("EXTRACT %s: %v", part, err)
		}
	}
	// EXTRACT DATE/DATETIME/TIME branches return component values.
	for _, part := range []string{"DATE", "DATETIME", "TIME"} {
		if _, err := BindExtract(ts, value.StringValue(part)); err != nil {
			t.Errorf("EXTRACT %s: %v", part, err)
		}
	}
}

func TestBindExtractInterval(t *testing.T) {
	iv := mustInterval(t, "1-2 3 4:5:6")
	for _, part := range []string{"YEAR", "MONTH", "DAY", "HOUR", "MINUTE", "SECOND", "MILLISECOND", "MICROSECOND"} {
		if _, err := BindExtract(iv, value.StringValue(part)); err != nil {
			t.Errorf("EXTRACT %s of INTERVAL: %v", part, err)
		}
	}
	if _, err := BindExtract(iv, value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part on INTERVAL should error")
	}
}

func TestBindExtractTimestampZone(t *testing.T) {
	ts := value.TimestampValue(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC))
	got, err := BindExtract(ts, value.StringValue("DAY"), value.StringValue("America/Los_Angeles"))
	if err != nil {
		t.Fatalf("BindExtract zone: %v", err)
	}
	n, _ := got.ToInt64()
	if n != 14 {
		t.Fatalf("LA day = %d, want 14", n)
	}
}

func TestBindExtractInvalid(t *testing.T) {
	d := mkDate(2024, 1, 1)
	if _, err := BindExtract(d, value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	if _, err := BindExtract(value.IntValue(1), value.StringValue("YEAR")); err == nil {
		t.Fatalf("non-date/datetime should error")
	}
	got, _ := BindExtract(nil, value.StringValue("YEAR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindExtractDate(t *testing.T) {
	ts := value.TimestampValue(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC))
	got, err := BindExtractDate(ts)
	if err != nil {
		t.Fatalf("BindExtractDate: %v", err)
	}
	if _, ok := got.(value.DateValue); !ok {
		t.Fatalf("want DateValue, got %T", got)
	}

	got, _ = BindExtractDate(ts, value.StringValue("UTC"))
	if got == nil {
		t.Fatalf("with zone, expected non-nil")
	}

	got, _ = BindExtractDate(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

// --- BindGenerateDateArray ---

func TestBindGenerateDateArray(t *testing.T) {
	start := mkDate(2024, 1, 1)
	end := mkDate(2024, 1, 5)
	got, err := BindGenerateDateArray(start, end)
	if err != nil {
		t.Fatalf("BindGenerateDateArray: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 5 {
		t.Fatalf("want 5 dates (Jan 1..5), got %d", len(arr.Values))
	}
}

func TestBindGenerateDateArrayCustomStep(t *testing.T) {
	start := mkDate(2024, 1, 1)
	end := mkDate(2024, 1, 7)
	got, err := BindGenerateDateArray(start, end, value.IntValue(2), value.StringValue("DAY"))
	if err != nil {
		t.Fatalf("BindGenerateDateArray: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 4 {
		t.Fatalf("want 4 dates by 2-day step, got %d", len(arr.Values))
	}
}

func TestBindGenerateDateArrayArity(t *testing.T) {
	if _, err := BindGenerateDateArray(mkDate(2024, 1, 1)); err == nil {
		t.Fatalf("arity error expected")
	}
}

// Smoke check CURRENT_DATE_WITH_TIME directly.
func TestCurrentDateWithTime(t *testing.T) {
	now := time.Now()
	got, _ := CURRENT_DATE_WITH_TIME(now)
	tt, _ := got.ToTime()
	if !tt.Equal(now) {
		t.Fatalf("want %v, got %v", now, tt)
	}
}
