package timestamp

import (
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func mkDate(y int, m time.Month, d int) value.Value {
	return value.DateValue(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func mkDatetime(y int, m time.Month, d, hh, mm, ss int) value.Value {
	return value.DatetimeValue(time.Date(y, m, d, hh, mm, ss, 0, time.UTC))
}

func mkTimestamp(y int, m time.Month, d, hh, mm, ss int) value.Value {
	return value.TimestampValue(time.Date(y, m, d, hh, mm, ss, 0, time.UTC))
}

func mustInterval(t *testing.T, s string) value.Value {
	t.Helper()
	iv, err := value.ParseInterval(s)
	if err != nil {
		t.Fatalf("ParseInterval(%q): %v", s, err)
	}
	return iv
}

// --- BindTimestamp ---

func TestBindTimestampFromDate(t *testing.T) {
	got, err := BindTimestamp(mkDate(2024, 6, 15))
	if err != nil {
		t.Fatalf("BindTimestamp: %v", err)
	}
	if _, ok := got.(value.TimestampValue); !ok {
		t.Fatalf("want TimestampValue, got %T", got)
	}
}

func TestBindTimestampFromDatetime(t *testing.T) {
	got, err := BindTimestamp(mkDatetime(2024, 6, 15, 12, 0, 0))
	if err != nil {
		t.Fatalf("BindTimestamp: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}
}

func TestBindTimestampFromString(t *testing.T) {
	got, err := BindTimestamp(value.StringValue("2024-06-15 00:00:00"))
	if err != nil {
		t.Fatalf("BindTimestamp string: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}
}

func TestBindTimestampStringZone(t *testing.T) {
	got, err := BindTimestamp(value.StringValue("2024-06-15 00:00:00"), value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindTimestamp zone: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}
}

func TestBindTimestampInvalidZone(t *testing.T) {
	if _, err := BindTimestamp(mkDate(2024, 1, 1), value.StringValue("Bad/Zone")); err == nil {
		t.Fatalf("invalid zone should error")
	}
}

func TestBindTimestampNullAndArity(t *testing.T) {
	got, _ := BindTimestamp(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimestamp(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindTimestamp(value.IntValue(1)); err == nil {
		t.Fatalf("non-date/string/datetime should error")
	}
}

// --- BindTimestampSeconds / Millis / Micros ---

func TestBindTimestampSecondsMillisMicros(t *testing.T) {
	got, _ := BindTimestampSeconds(value.IntValue(100))
	tt, _ := got.ToTime()
	if tt.Unix() != 100 {
		t.Fatalf("TIMESTAMP_SECONDS(100) unix = %d", tt.Unix())
	}

	got, _ = BindTimestampMillis(value.IntValue(1000))
	tt, _ = got.ToTime()
	if tt.UnixMilli() != 1000 {
		t.Fatalf("TIMESTAMP_MILLIS(1000) unix_ms = %d", tt.UnixMilli())
	}

	got, _ = BindTimestampMicros(value.IntValue(1_000_000))
	tt, _ = got.ToTime()
	if tt.UnixMicro() != 1_000_000 {
		t.Fatalf("TIMESTAMP_MICROS = %d", tt.UnixMicro())
	}

	for _, fn := range []func(...value.Value) (value.Value, error){
		BindTimestampSeconds, BindTimestampMillis, BindTimestampMicros,
	} {
		got, _ := fn(nil)
		if got != nil {
			t.Errorf("NULL input must produce NULL output")
		}
		if _, err := fn(); err == nil {
			t.Errorf("arity error expected")
		}
	}
}

// --- BindUnixSeconds / Millis / Micros / MysqlUnixTimestamp ---

func TestBindUnixVariants(t *testing.T) {
	ts := mkTimestamp(1970, 1, 1, 0, 0, 100)
	got, _ := BindUnixSeconds(ts)
	if n, _ := got.ToInt64(); n != 100 {
		t.Fatalf("UNIX_SECONDS = %d", n)
	}

	got, _ = BindUnixMillis(ts)
	if n, _ := got.ToInt64(); n != 100_000 {
		t.Fatalf("UNIX_MILLIS = %d", n)
	}

	got, _ = BindUnixMicros(ts)
	if n, _ := got.ToInt64(); n != 100_000_000 {
		t.Fatalf("UNIX_MICROS = %d", n)
	}

	got, _ = BindMysqlUnixTimestamp(ts)
	if n, _ := got.ToInt64(); n != 100 {
		t.Fatalf("MYSQL_UNIX_TIMESTAMP = %d", n)
	}

	for _, fn := range []func(...value.Value) (value.Value, error){
		BindUnixSeconds, BindUnixMillis, BindUnixMicros,
	} {
		got, _ := fn(nil)
		if got != nil {
			t.Errorf("NULL input must produce NULL output")
		}
		if _, err := fn(); err == nil {
			t.Errorf("arity error expected")
		}
	}

	// BindMysqlUnixTimestamp accepts 0 args (returns "now"); only test
	// NULL propagation in the 1-arg form.
	got, _ = BindMysqlUnixTimestamp(nil)
	if got != nil {
		t.Errorf("BindMysqlUnixTimestamp(NULL) must produce NULL output")
	}
	// 0-arg form returns a current-instant INT64.
	if v, err := BindMysqlUnixTimestamp(); err != nil || v == nil {
		t.Errorf("BindMysqlUnixTimestamp(): err=%v v=%v, want non-nil int", err, v)
	}
}

// --- BindTimestampAdd / Sub / Diff / Trunc ---

func TestBindTimestampAddSub(t *testing.T) {
	ts := mkTimestamp(2024, 6, 15, 12, 0, 0)
	for _, part := range []string{"MICROSECOND", "MILLISECOND", "SECOND", "MINUTE", "HOUR", "DAY"} {
		if _, err := BindTimestampAdd(ts, value.IntValue(1), value.StringValue(part)); err != nil {
			t.Errorf("TIMESTAMP_ADD(%s): %v", part, err)
		}
		if _, err := BindTimestampSub(ts, value.IntValue(1), value.StringValue(part)); err != nil {
			t.Errorf("TIMESTAMP_SUB(%s): %v", part, err)
		}
	}

	if _, err := BindTimestampAdd(ts, value.IntValue(1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	if _, err := BindTimestampSub(ts, value.IntValue(1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}

	got, _ := BindTimestampAdd(nil, value.IntValue(1), value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimestampAdd(); err == nil {
		t.Fatalf("arity error expected")
	}
	if _, err := BindTimestampSub(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindTimestampDiff(t *testing.T) {
	a := mkTimestamp(2024, 6, 15, 12, 0, 0)
	b := mkTimestamp(2024, 6, 15, 10, 0, 0)
	for _, part := range []string{"MICROSECOND", "MILLISECOND", "SECOND", "MINUTE", "HOUR", "DAY"} {
		got, err := BindTimestampDiff(a, b, value.StringValue(part))
		if err != nil {
			t.Errorf("TIMESTAMP_DIFF(%s): %v", part, err)
			continue
		}
		if got == nil {
			t.Errorf("TIMESTAMP_DIFF(%s) returned nil", part)
		}
	}
	if _, err := BindTimestampDiff(a, b, value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ := BindTimestampDiff(nil, b, value.StringValue("HOUR"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimestampDiff(); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestBindTimestampTrunc(t *testing.T) {
	ts := value.TimestampValue(time.Date(2024, 6, 15, 12, 34, 56, 123_000_000, time.UTC))
	for _, part := range []string{"MICROSECOND", "MILLISECOND", "SECOND", "MINUTE", "HOUR", "DAY", "MONTH", "YEAR"} {
		if _, err := BindTimestampTrunc(ts, value.StringValue(part)); err != nil {
			t.Errorf("TIMESTAMP_TRUNC(%s): %v", part, err)
		}
	}

	// With zone argument.
	got, err := BindTimestampTrunc(ts, value.StringValue("DAY"), value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("TIMESTAMP_TRUNC DAY UTC: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}

	if _, err := BindTimestampTrunc(ts, value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
	got, _ = BindTimestampTrunc(nil, value.StringValue("DAY"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimestampTrunc(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindTimestampBucket ---

func TestBindTimestampBucket(t *testing.T) {
	target := mkTimestamp(2024, 6, 15, 5, 0, 0)
	iv := mustInterval(t, "0-0 1 0:0:0")
	got, err := BindTimestampBucket(target, iv)
	if err != nil {
		t.Fatalf("BindTimestampBucket: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil bucket")
	}

	// With origin.
	got, err = BindTimestampBucket(target, iv, mkTimestamp(2024, 1, 1, 0, 0, 0))
	if err != nil {
		t.Fatalf("BindTimestampBucket origin: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil bucket")
	}

	// Non-INTERVAL bucket width.
	if _, err := BindTimestampBucket(target, value.IntValue(1)); err == nil {
		t.Fatalf("non-INTERVAL should error")
	}
	got, _ = BindTimestampBucket(nil, iv)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindTimestampBucket(target); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindFormatTimestamp / BindParseTimestamp / BindString ---

func TestBindFormatAndParseTimestamp(t *testing.T) {
	ts := mkTimestamp(2024, 6, 15, 12, 0, 0)
	got, err := BindFormatTimestamp(value.StringValue("%Y-%m-%d %H:%M:%S"), ts)
	if err != nil {
		t.Fatalf("BindFormatTimestamp: %v", err)
	}
	s, _ := got.ToString()
	if s == "" {
		t.Fatalf("expected non-empty format")
	}

	// With zone arg.
	got, err = BindFormatTimestamp(value.StringValue("%Y-%m-%d"), ts, value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindFormatTimestamp zone: %v", err)
	}
	s, _ = got.ToString()
	if s != "2024-06-15" {
		t.Fatalf("format = %q, want 2024-06-15", s)
	}

	// Parse round-trip.
	got, err = BindParseTimestamp(value.StringValue("%Y-%m-%d %H:%M:%S"), value.StringValue("2024-06-15 12:00:00"))
	if err != nil {
		t.Fatalf("BindParseTimestamp: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}

	got, _ = BindFormatTimestamp(nil, ts)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindFormatTimestamp(); err == nil {
		t.Fatalf("arity error expected")
	}

	got, _ = BindParseTimestamp(nil, value.StringValue("2024"))
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindParseTimestampError(t *testing.T) {
	if _, err := BindParseTimestamp(value.StringValue("%Y"), value.StringValue("garbage")); err == nil {
		t.Fatalf("invalid input should error")
	}
}

// TestBindParseTimestampWithZone covers the 3-arg form that routes
// through PARSE_TIMESTAMP_WITH_TIMEZONE.
func TestBindParseTimestampWithZone(t *testing.T) {
	got, err := BindParseTimestamp(
		value.StringValue("%Y-%m-%d %H:%M:%S"),
		value.StringValue("2024-06-15 00:00:00"),
		value.StringValue("UTC"),
	)
	if err != nil {
		t.Fatalf("BindParseTimestamp 3-arg: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}

	// Invalid zone surfaces an error.
	if _, err := BindParseTimestamp(
		value.StringValue("%Y-%m-%d %H:%M:%S"),
		value.StringValue("2024-06-15 00:00:00"),
		value.StringValue("Bad/Zone"),
	); err == nil {
		t.Fatalf("invalid zone should error")
	}
}

func TestBindString(t *testing.T) {
	ts := mkTimestamp(2024, 6, 15, 12, 0, 0)
	got, err := BindString(ts)
	if err != nil {
		t.Fatalf("BindString: %v", err)
	}
	s, _ := got.ToString()
	if s == "" {
		t.Fatalf("expected non-empty STRING representation")
	}

	// With zone arg.
	got, err = BindString(ts, value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindString zone: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}

	got, _ = BindString(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindString(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// --- BindCurrentTimestamp ---

func TestBindCurrentTimestamp(t *testing.T) {
	got, err := BindCurrentTimestamp()
	if err != nil {
		t.Fatalf("BindCurrentTimestamp: %v", err)
	}
	if _, ok := got.(value.TimestampValue); !ok {
		t.Fatalf("want TimestampValue, got %T", got)
	}
}

func TestBindCurrentTimestampWithZone(t *testing.T) {
	if _, err := BindCurrentTimestamp(value.StringValue("UTC")); err != nil {
		t.Fatalf("BindCurrentTimestamp UTC: %v", err)
	}
	if _, err := BindCurrentTimestamp(value.StringValue("Bad/Zone")); err == nil {
		t.Fatalf("invalid zone should error")
	}
}

func TestBindCurrentTimestampWithUnixNano(t *testing.T) {
	when := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).UnixNano()
	got, err := BindCurrentTimestamp(value.IntValue(when))
	if err != nil {
		t.Fatalf("BindCurrentTimestamp: %v", err)
	}
	tt, _ := got.ToTime()
	if tt.UTC().Year() != 2024 {
		t.Fatalf("want year 2024, got %d", tt.Year())
	}
}

func TestBindCurrentTimestampUnixNanoZone(t *testing.T) {
	when := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC).UnixNano()
	got, err := BindCurrentTimestamp(value.IntValue(when), value.StringValue("UTC"))
	if err != nil {
		t.Fatalf("BindCurrentTimestamp: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil")
	}
}

// --- BindGenerateTimestampArray ---

func TestBindGenerateTimestampArray(t *testing.T) {
	start := mkTimestamp(2024, 6, 15, 12, 0, 0)
	end := mkTimestamp(2024, 6, 15, 15, 0, 0)
	got, err := BindGenerateTimestampArray(start, end, value.IntValue(1), value.StringValue("HOUR"))
	if err != nil {
		t.Fatalf("BindGenerateTimestampArray: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 4 {
		t.Fatalf("want 4 entries (12,13,14,15), got %d", len(arr.Values))
	}
}

func TestBindGenerateTimestampArrayArity(t *testing.T) {
	if _, err := BindGenerateTimestampArray(); err == nil {
		t.Fatalf("arity error expected")
	}
}

// Smoke check CURRENT_TIMESTAMP_WITH_TIME directly.
func TestCurrentTimestampWithTime(t *testing.T) {
	now := time.Now()
	got, _ := CURRENT_TIMESTAMP_WITH_TIME(now)
	tt, _ := got.ToTime()
	if !tt.Equal(now) {
		t.Fatalf("want %v, got %v", now, tt)
	}
}
