package interval

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/intervalvalue"

	"github.com/goccy/googlesqlite/internal/value"
)

func mkIV(years, months, days, hours, minutes, seconds, subSecondNanos int32) *value.IntervalValue {
	return &value.IntervalValue{
		IntervalValue: &intervalvalue.IntervalValue{
			Years: years, Months: months, Days: days,
			Hours: hours, Minutes: minutes, Seconds: seconds,
			SubSecondNanos: subSecondNanos,
		},
	}
}

// --- BindInterval ---

func TestBindInterval(t *testing.T) {
	cases := []struct {
		part string
		v    int64
		want func(*value.IntervalValue) bool
	}{
		{"YEAR", 3, func(iv *value.IntervalValue) bool { return iv.Years == 3 }},
		{"MONTH", 5, func(iv *value.IntervalValue) bool { return iv.Months == 5 }},
		{"DAY", 7, func(iv *value.IntervalValue) bool { return iv.Days == 7 }},
		{"HOUR", 2, func(iv *value.IntervalValue) bool { return iv.Hours == 2 }},
		{"MINUTE", 30, func(iv *value.IntervalValue) bool { return iv.Minutes == 30 }},
		{"SECOND", 45, func(iv *value.IntervalValue) bool { return iv.Seconds == 45 }},
		{"NANOSECOND", 999, func(iv *value.IntervalValue) bool { return iv.SubSecondNanos == 999 }},
	}
	for _, tc := range cases {
		got, err := BindInterval(value.IntValue(tc.v), value.StringValue(tc.part))
		if err != nil {
			t.Errorf("BindInterval(%s): %v", tc.part, err)
			continue
		}
		iv := got.(*value.IntervalValue)
		if !tc.want(iv) {
			t.Errorf("BindInterval(%d, %s) wrong field: %+v", tc.v, tc.part, *iv.IntervalValue)
		}
	}

	if _, err := BindInterval(value.IntValue(1), value.StringValue("BAD")); err == nil {
		t.Fatalf("invalid part should error")
	}
}

// --- BindMakeInterval ---

func TestBindMakeInterval(t *testing.T) {
	got, err := BindMakeInterval(
		value.IntValue(1), value.IntValue(2), value.IntValue(3),
		value.IntValue(4), value.IntValue(5), value.IntValue(6),
	)
	if err != nil {
		t.Fatalf("BindMakeInterval: %v", err)
	}
	iv := got.(*value.IntervalValue)
	if iv.Years != 1 || iv.Months != 2 || iv.Days != 3 ||
		iv.Hours != 4 || iv.Minutes != 5 || iv.Seconds != 6 {
		t.Fatalf("MAKE_INTERVAL fields wrong: %+v", *iv.IntervalValue)
	}
}

// --- BindJustifyDays ---

func TestBindJustifyDays(t *testing.T) {
	iv := mkIV(0, 0, 35, 0, 0, 0, 0)
	got, err := BindJustifyDays(iv)
	if err != nil {
		t.Fatalf("BindJustifyDays: %v", err)
	}
	out := got.(*value.IntervalValue)
	if out.Days != 5 || out.Months != 1 {
		t.Fatalf("JUSTIFY_DAYS(35 days) = %+v, want 1 month + 5 days", *out.IntervalValue)
	}

	// Negative path.
	iv = mkIV(0, 0, -35, 0, 0, 0, 0)
	got, _ = BindJustifyDays(iv)
	out = got.(*value.IntervalValue)
	if out.Days != -5 || out.Months != -1 {
		t.Fatalf("JUSTIFY_DAYS(-35 days) = %+v, want -1 month - 5 days", *out.IntervalValue)
	}

	// Months overflow to year.
	iv = mkIV(0, 13, 0, 0, 0, 0, 0)
	got, _ = BindJustifyDays(iv)
	out = got.(*value.IntervalValue)
	if out.Months != 1 || out.Years != 1 {
		t.Fatalf("JUSTIFY_DAYS(13 months) = %+v, want 1 year + 1 month", *out.IntervalValue)
	}

	// Months negative overflow.
	iv = mkIV(0, -13, 0, 0, 0, 0, 0)
	got, _ = BindJustifyDays(iv)
	out = got.(*value.IntervalValue)
	if out.Months != -1 || out.Years != -1 {
		t.Fatalf("JUSTIFY_DAYS(-13 months) = %+v, want -1 year - 1 month", *out.IntervalValue)
	}

	if _, err := BindJustifyDays(value.IntValue(1)); err == nil {
		t.Fatalf("non-INTERVAL should error")
	}
}

// --- BindJustifyHours ---

func TestBindJustifyHours(t *testing.T) {
	iv := mkIV(0, 0, 0, 25, 0, 0, 0)
	got, err := BindJustifyHours(iv)
	if err != nil {
		t.Fatalf("BindJustifyHours: %v", err)
	}
	out := got.(*value.IntervalValue)
	if out.Hours != 1 || out.Days != 1 {
		t.Fatalf("JUSTIFY_HOURS(25h) = %+v, want 1d 1h", *out.IntervalValue)
	}

	// Negative path.
	iv = mkIV(0, 0, 0, -25, 0, 0, 0)
	got, _ = BindJustifyHours(iv)
	out = got.(*value.IntervalValue)
	if out.Hours != -1 || out.Days != -1 {
		t.Fatalf("JUSTIFY_HOURS(-25h) = %+v, want -1d -1h", *out.IntervalValue)
	}

	// Minutes overflow.
	iv = mkIV(0, 0, 0, 0, 75, 0, 0)
	got, _ = BindJustifyHours(iv)
	out = got.(*value.IntervalValue)
	if out.Minutes != 15 || out.Hours != 1 {
		t.Fatalf("JUSTIFY_HOURS(75m) = %+v, want 1h 15m", *out.IntervalValue)
	}

	// Seconds overflow.
	iv = mkIV(0, 0, 0, 0, 0, 75, 0)
	got, _ = BindJustifyHours(iv)
	out = got.(*value.IntervalValue)
	if out.Seconds != 15 || out.Minutes != 1 {
		t.Fatalf("JUSTIFY_HOURS(75s) = %+v, want 1m 15s", *out.IntervalValue)
	}

	// Negative seconds.
	iv = mkIV(0, 0, 0, 0, 0, -75, 0)
	got, _ = BindJustifyHours(iv)
	out = got.(*value.IntervalValue)
	if out.Seconds != -15 || out.Minutes != -1 {
		t.Fatalf("JUSTIFY_HOURS(-75s) = %+v, want -1m -15s", *out.IntervalValue)
	}

	if _, err := BindJustifyHours(value.IntValue(1)); err == nil {
		t.Fatalf("non-INTERVAL should error")
	}
}

// --- BindJustifyInterval ---

func TestBindJustifyInterval(t *testing.T) {
	iv := mkIV(0, 0, 35, 25, 0, 0, 0)
	got, err := BindJustifyInterval(iv)
	if err != nil {
		t.Fatalf("BindJustifyInterval: %v", err)
	}
	out := got.(*value.IntervalValue)
	// 25h -> 1d 1h, then 35+1 = 36 days -> 1 month 6 days.
	if out.Months != 1 || out.Days != 6 || out.Hours != 1 {
		t.Fatalf("JUSTIFY_INTERVAL(35d 25h) = %+v, want 1mo 6d 1h", *out.IntervalValue)
	}

	if _, err := BindJustifyInterval(value.IntValue(1)); err == nil {
		t.Fatalf("non-INTERVAL should error")
	}
}
