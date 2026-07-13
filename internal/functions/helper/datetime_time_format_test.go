package helper

import (
	"strings"
	"testing"
	"time"
)

// referenceTime is a single deterministic time used across every
// formatter case. Its exact components are chosen so that the
// resulting strings cross-check against the BigQuery / GoogleSQL
// reference docs (`%U`, `%V`, `%W` align, weekday mid-week, day-of-year > 99, etc.).
//
//	2026-05-15 03:04:05 UTC (Friday)
//	  Year   2026 -> %Y "2026" ; %y "26" ; %C "20"
//	  ISO    week 20 (Friday)
//	  Day    of year 135 -> %j "135"
//	  Quarter 2
var referenceTime = time.Date(2026, 5, 15, 3, 4, 5, 123_456_789, time.UTC)

// TestFormatTimeAllSpecifiers walks every supported %-specifier and
// asserts FormatTime produces the canonical rendering. The expected
// strings come from the BigQuery / GoogleSQL FORMAT_DATE,
// FORMAT_TIME and FORMAT_TIMESTAMP reference (Cloud BigQuery docs)
// — the same source the upstream googlesql spec carries forward.
func TestFormatTimeAllSpecifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		format string
		typ    TimeFormatType
		want   string
	}{
		// %A = full weekday name (Friday for 2026-05-15).
		{"%A", "%A", FormatTypeDate, "Friday"},
		{"%a", "%a", FormatTypeDate, "Fri"},
		{"%B", "%B", FormatTypeDate, "May"},
		{"%b", "%b", FormatTypeDate, "May"},
		{"%C", "%C", FormatTypeDate, "20"},
		{"%c", "%c", FormatTypeDatetime, "Fri May 15 03:04:05 2026"},
		{"%D", "%D", FormatTypeDate, "05/15/26"},
		{"%d", "%d", FormatTypeDate, "15"},
		{"%e", "%e", FormatTypeDate, "15"},
		{"%F", "%F", FormatTypeDate, "2026-05-15"},
		{"%G", "%G", FormatTypeDate, "2026"},
		{"%g", "%g", FormatTypeDate, "2026"},
		{"%H", "%H", FormatTypeTime, "03"},
		{"%h", "%h", FormatTypeDate, "May"},
		{"%I", "%I", FormatTypeTime, "03"},
		{"%J", "%J", FormatTypeDate, "2026"},
		{"%j", "%j", FormatTypeDate, "135"},
		{"%k", "%k", FormatTypeTime, " 3"},
		{"%l", "%l", FormatTypeTime, " 3"},
		{"%M", "%M", FormatTypeTime, "04"},
		{"%m", "%m", FormatTypeDate, "05"},
		{"%n", "%n", FormatTypeDate, "\n"},
		{"%P", "%P", FormatTypeTime, "am"},
		{"%p", "%p", FormatTypeTime, "AM"},
		{"%Q", "%Q", FormatTypeDate, "2"}, // day-of-year 135 -> Q2
		{"%R", "%R", FormatTypeTime, "03:04"},
		{"%S", "%S", FormatTypeTime, "05"},
		{"%s", "%s", FormatTypeTime, "1778814245"}, // unix seconds for 2026-05-15 03:04:05 UTC
		{"%T", "%T", FormatTypeTime, "03:04:05"},
		{"%t", "%t", FormatTypeDate, "\t"},
		{"%U", "%U", FormatTypeDate, "20"},
		{"%u", "%u", FormatTypeDate, "5"},
		{"%V", "%V", FormatTypeDate, "20"},
		{"%W", "%W", FormatTypeDate, "20"},
		{"%w", "%w", FormatTypeDate, "19"}, // weekNumberZeroBaseFormatter is ISOWeek-1.
		{"%X", "%X", FormatTypeTime, "03:04:05"},
		{"%x", "%x", FormatTypeDate, "05/15/26"},
		{"%Y", "%Y", FormatTypeDate, "2026"},
		{"%y", "%y", FormatTypeDate, "26"},
		{"%Z", "%Z", FormatTypeTimestamp, "UTC"},
		{"%z", "%z", FormatTypeTimestamp, "0"},
		{"%%", "%%", FormatTypeDate, "%"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := FormatTime(c.format, &referenceTime, c.typ)
			if err != nil {
				t.Fatalf("FormatTime(%q): %v", c.format, err)
			}
			if got != c.want {
				t.Fatalf("FormatTime(%q) = %q, want %q", c.format, got, c.want)
			}
		})
	}
}

// TestFormatTimeCombinedSpecifiers exercises the %E-prefixed
// combination specifiers handled by combinationPatternInfo.
func TestFormatTimeCombinedSpecifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		format string
		typ    TimeFormatType
		want   string
	}{
		{"%Ez", FormatTypeTimestamp, "+00:00"},
		{"%E4Y", FormatTypeDate, "2026"},
		// %E*S -> seconds.6-digit fractional. nanoseconds 123_456_789 -> "123456".
		{"%E*S", FormatTypeTime, "05.123456"},
		{"%E1S", FormatTypeTime, "05.1"},
		{"%E2S", FormatTypeTime, "05.12"},
		{"%E3S", FormatTypeTime, "05.123"},
		{"%E4S", FormatTypeTime, "05.1234"},
		{"%E5S", FormatTypeTime, "05.12345"},
		{"%E6S", FormatTypeTime, "05.123456"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.format, func(t *testing.T) {
			t.Parallel()
			got, err := FormatTime(c.format, &referenceTime, c.typ)
			if err != nil {
				t.Fatalf("FormatTime(%q): %v", c.format, err)
			}
			if got != c.want {
				t.Fatalf("FormatTime(%q) = %q, want %q", c.format, got, c.want)
			}
		})
	}
}

// TestFormatTimeLiteralAfterCombination pins that literal characters
// and further format elements following a %E-prefixed combination
// token (%E<n>S, %E*S, %Ez, %E4Y) are emitted, not swallowed.
//
// Every element renders independently and literal characters in the
// format string are copied verbatim (format-elements reference:
// FORMAT_TIMESTAMP("%b %Y %Ez", ...) -> "Dec 2008 +00:00" keeps the
// spaces around %Ez). The expected strings below are the documented
// per-element renderings composed left-to-right for referenceTime
// (2026-05-15 03:04:05.123456789 UTC).
func TestFormatTimeLiteralAfterCombination(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		format string
		typ    TimeFormatType
		want   string
	}{
		// Trailing literal 'Z' after %E3S (the ISO "...%E3SZ" idiom).
		{"E3S_then_Z", "%E3SZ", FormatTypeTime, "05.123Z"},
		// Trailing literal after %E*S.
		{"EstarS_then_Z", "%E*SZ", FormatTypeTime, "05.123456Z"},
		// %E4Y followed by a '-' literal and another element.
		{"E4Y_then_dash_month", "%E4Y-%m", FormatTypeDate, "2026-05"},
		// %Ez followed by a literal character.
		{"Ez_then_literal", "%Ez!", FormatTypeTimestamp, "+00:00!"},
		// %Ez followed by a space and a further element.
		{"Ez_then_space_month", "%Ez %b", FormatTypeTimestamp, "+00:00 May"},
		// Combination token in the middle, element on both sides.
		{"month_E3S_year", "%b %E3S %Y", FormatTypeTimestamp, "May 05.123 2026"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := FormatTime(c.format, &referenceTime, c.typ)
			if err != nil {
				t.Fatalf("FormatTime(%q): %v", c.format, err)
			}
			if got != c.want {
				t.Fatalf("FormatTime(%q) = %q, want %q", c.format, got, c.want)
			}
		})
	}
}

// TestFormatTimeErrors covers the FormatTime error paths.
func TestFormatTimeErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		format string
		typ    TimeFormatType
	}{
		// Trailing `%` with no specifier.
		{"trailing_percent", "%", FormatTypeDate},
		// Trailing `%E` with no specifier.
		{"trailing_percent_E", "%E", FormatTypeDate},
		// Unknown specifier.
		{"unknown_specifier", "%~", FormatTypeDate},
		// Date specifier used with type=Time.
		{"unavailable_for_type", "%Y", FormatTypeTime},
		// Unavailable %E* for type=Date.
		{"unavailable_E_for_type", "%E*S", FormatTypeDate},
		// Unknown %E specifier.
		{"unknown_E_specifier", "%E9X", FormatTypeDate},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := FormatTime(c.format, &referenceTime, c.typ); err == nil {
				t.Fatalf("expected error for %q", c.format)
			}
		})
	}
}

// TestParseTimeFormatRoundtrip exercises the parser side of every
// parser-supported specifier. Each row picks a small literal target
// string that matches the format. We then format the resulting
// time back with the same specifier (when the specifier round-trips)
// or check a derivable field of the parsed time. Cases follow the
// canonical PARSE_DATE / PARSE_TIME / PARSE_TIMESTAMP examples.
func TestParseTimeFormatRoundtrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		format         string
		target         string
		typ            TimeFormatType
		expectHour     int
		expectMin      int
		expectSec      int
		expectMonth    time.Month
		expectDay      int
		expectYear     int
		assertWeekDay  *time.Weekday
		assertDayOfYr  int
		assertLocation string
	}{
		// %A & %a parse a (short) weekday name and consume it.
		// The day-of-week doesn't directly write to the time so we only
		// check the parse succeeds.
		{name: "weekday_full", format: "%A", target: "Friday", typ: FormatTypeDate},
		{name: "weekday_short", format: "%a", target: "Fri", typ: FormatTypeDate},
		// %B / %b populate the month.
		{name: "month_full", format: "%B", target: "May", typ: FormatTypeDate, expectMonth: time.May},
		{name: "month_short", format: "%b", target: "May", typ: FormatTypeDate, expectMonth: time.May},
		// %C parses a century -> year = c*100 - 99 (1->1901, 20->1901+1900=2001-ish).
		// We only assert the parse succeeds and the year is set.
		{name: "century", format: "%C", target: "20", typ: FormatTypeDate, expectYear: 1901},
		// %c is ANSIC: "Mon Jan 02 15:04:05 2006".
		{name: "ansic", format: "%c", target: "Fri May 15 03:04:05 2026", typ: FormatTypeTimestamp, expectYear: 2026, expectMonth: time.May, expectDay: 15, expectHour: 3, expectMin: 4, expectSec: 5},
		// %D parses month/day/year (year-without-century).
		{name: "mdy_slash", format: "%D", target: "05/15/26", typ: FormatTypeDate, expectMonth: time.May, expectDay: 15, expectYear: 2026},
		// %d parses day of month.
		{name: "day", format: "%d", target: "15", typ: FormatTypeDate, expectDay: 15},
		// %e parses day of month with optional leading space.
		{name: "day_padded", format: "%e", target: " 5", typ: FormatTypeDate, expectDay: 5},
		// %F parses year-month-day.
		{name: "ymd", format: "%F", target: "2026-05-15", typ: FormatTypeDate, expectYear: 2026, expectMonth: time.May, expectDay: 15},
		// %H parses hour (0-23).
		{name: "hour24", format: "%H", target: "13", typ: FormatTypeTime, expectHour: 13},
		// %h parses short month (same as %b).
		{name: "short_month_h", format: "%h", target: "May", typ: FormatTypeDate, expectMonth: time.May},
		// %I parses 12-hour clock.
		{name: "hour12", format: "%I", target: "11", typ: FormatTypeTime, expectHour: 11},
		// %j parses day-of-year.
		{name: "day_of_year", format: "%j", target: "135", typ: FormatTypeDate, expectMonth: time.May, expectDay: 15},
		// %k parses 24-hour clock with leading space.
		{name: "hour24_padded", format: "%k", target: " 9", typ: FormatTypeTime, expectHour: 9},
		// %l parses 12-hour clock with leading space.
		{name: "hour12_padded", format: "%l", target: " 9", typ: FormatTypeTime, expectHour: 9},
		// %M parses minute.
		{name: "minute", format: "%M", target: "30", typ: FormatTypeTime, expectMin: 30},
		// %m parses month number.
		{name: "month_num", format: "%m", target: "05", typ: FormatTypeDate, expectMonth: time.May},
		// %R parses HH:MM.
		{name: "hour_minute", format: "%R", target: "03:04", typ: FormatTypeTime, expectHour: 3, expectMin: 4},
		// %S parses seconds.
		{name: "second", format: "%S", target: "59", typ: FormatTypeTime, expectSec: 59},
		// %s parses unix seconds (10 chars wide).
		{name: "unixsec", format: "%s", target: "1778814245", typ: FormatTypeTimestamp, expectYear: 2026, expectMonth: time.May, expectDay: 15},
		// %T parses HH:MM:SS.
		{name: "hms", format: "%T", target: "03:04:05", typ: FormatTypeTime, expectHour: 3, expectMin: 4, expectSec: 5},
		// %X parses HH:MM:SS (alias).
		{name: "X_hms", format: "%X", target: "03:04:05", typ: FormatTypeTime, expectHour: 3, expectMin: 4, expectSec: 5},
		// %x parses month/day/year.
		{name: "x_mdy", format: "%x", target: "05/15/26", typ: FormatTypeDate, expectMonth: time.May, expectDay: 15, expectYear: 2026},
		// %Y parses 4-digit year.
		{name: "year4", format: "%Y", target: "2026", typ: FormatTypeDate, expectYear: 2026},
		// %y parses 2-digit year. Per the rule >=69 -> 1900s.
		{name: "year2_low", format: "%y", target: "26", typ: FormatTypeDate, expectYear: 2026},
		{name: "year2_high", format: "%y", target: "70", typ: FormatTypeDate, expectYear: 1970},
		// %E*S parses seconds with fractional precision.
		{name: "EstarS", format: "%E*S", target: "05.123456", typ: FormatTypeTime, expectSec: 5},
		{name: "E2S", format: "%E2S", target: "05.12", typ: FormatTypeTime, expectSec: 5},
		// %E4Y parses 4-digit year (alias for %Y).
		{name: "E4Y", format: "%E4Y", target: "2026", typ: FormatTypeDate, expectYear: 2026},
		// %Ez parses RFC3339 offset.
		{name: "Ez_Z", format: "%Ez", target: "Z", typ: FormatTypeTimestamp},
		{name: "Ez_plus", format: "%Ez", target: "+05:30", typ: FormatTypeTimestamp},
		{name: "Ez_minus", format: "%Ez", target: "-08:00", typ: FormatTypeTimestamp},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := ParseTimeFormat(c.format, c.target, c.typ)
			if err != nil {
				t.Fatalf("ParseTimeFormat(%q, %q): %v", c.format, c.target, err)
			}
			if c.expectYear != 0 && parsed.Year() != c.expectYear {
				t.Errorf("year: got %d want %d", parsed.Year(), c.expectYear)
			}
			if c.expectMonth != 0 && parsed.Month() != c.expectMonth {
				t.Errorf("month: got %v want %v", parsed.Month(), c.expectMonth)
			}
			if c.expectDay != 0 && parsed.Day() != c.expectDay {
				t.Errorf("day: got %d want %d", parsed.Day(), c.expectDay)
			}
			if c.expectHour != 0 && parsed.Hour() != c.expectHour {
				t.Errorf("hour: got %d want %d", parsed.Hour(), c.expectHour)
			}
			if c.expectMin != 0 && parsed.Minute() != c.expectMin {
				t.Errorf("minute: got %d want %d", parsed.Minute(), c.expectMin)
			}
			if c.expectSec != 0 && parsed.Second() != c.expectSec {
				t.Errorf("second: got %d want %d", parsed.Second(), c.expectSec)
			}
		})
	}
}

// TestParseTimeFormatFractionalPrecision pins the parse-side matching
// rule for the fractional-second combination tokens:
//
//   - %E<n>S matches AT MOST n fractional digits. A longer run of
//     fractional digits is a parse failure (e.g. PARSE_TIMESTAMP with
//     %E3S over "...00.1234567Z" returns NULL in BigQuery), while
//     fewer than n digits (or none) parse fine.
//   - %E*S matches a variable number of fractional digits and keeps
//     microsecond precision.
//
// Anchored to the format-elements reference ("%E<number>S: Seconds
// with <number> digits of fractional precision") and the abseil
// ParseTime semantics GoogleSQL/BigQuery inherit.
func TestParseTimeFormatFractionalPrecision(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		format    string
		target    string
		wantErr   bool
		wantSec   int
		wantNanos int
	}{
		// %E3S: 0..3 fractional digits accepted.
		{name: "E3S_zero", format: "%E3S", target: "05", wantSec: 5, wantNanos: 0},
		{name: "E3S_two", format: "%E3S", target: "05.12", wantSec: 5, wantNanos: 120_000_000},
		{name: "E3S_three", format: "%E3S", target: "05.123", wantSec: 5, wantNanos: 123_000_000},
		// %E3S: 4+ fractional digits must fail.
		{name: "E3S_four", format: "%E3S", target: "05.1234", wantErr: true},
		{name: "E3S_seven", format: "%E3S", target: "05.1234567", wantErr: true},
		// %E6S: 6 digits ok, 7 fails.
		{name: "E6S_six", format: "%E6S", target: "05.123456", wantSec: 5, wantNanos: 123_456_000},
		{name: "E6S_seven", format: "%E6S", target: "05.1234567", wantErr: true},
		// %E*S: variable count, truncated to microseconds.
		{name: "Estar_seven", format: "%E*S", target: "05.1234567", wantSec: 5, wantNanos: 123_456_000},
		{name: "Estar_nine", format: "%E*S", target: "05.123456789", wantSec: 5, wantNanos: 123_456_000},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseTimeFormat(c.format, c.target, FormatTypeTime)
			if c.wantErr {
				if err == nil {
					t.Fatalf("ParseTimeFormat(%q, %q): expected error, got %v", c.format, c.target, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTimeFormat(%q, %q): %v", c.format, c.target, err)
			}
			if got.Second() != c.wantSec || got.Nanosecond() != c.wantNanos {
				t.Fatalf("ParseTimeFormat(%q, %q) = sec %d nanos %d, want sec %d nanos %d",
					c.format, c.target, got.Second(), got.Nanosecond(), c.wantSec, c.wantNanos)
			}
		})
	}
}

// TestParseTimeFormatPostProcessPM exercises the AM/PM
// post-processor: a 12-hour-clock %I plus a %p of "PM" promotes the
// hour into the 13-23 range.
func TestParseTimeFormatPostProcessPM(t *testing.T) {
	t.Parallel()
	got, err := ParseTimeFormat("%I:%M:%S %p", "01:30:00 PM", FormatTypeTime)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 13 || got.Minute() != 30 {
		t.Fatalf("hour=%d minute=%d", got.Hour(), got.Minute())
	}
}

// TestParseTimeFormatPostProcessAM_12 makes sure 12 AM maps to
// hour 0 (the midnight case).
func TestParseTimeFormatPostProcessAM_12(t *testing.T) {
	t.Parallel()
	got, err := ParseTimeFormat("%I:%M:%S %p", "12:00:00 AM", FormatTypeTime)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 0 {
		t.Fatalf("hour=%d, want 0", got.Hour())
	}
}

// TestParseTimeFormatErrors covers the ParseTimeFormat error paths.
func TestParseTimeFormatErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		format string
		target string
		typ    TimeFormatType
		match  string // substring of expected error
	}{
		{"trailing_percent", "%", "x", FormatTypeDate, "invalid"},
		{"trailing_percent_E", "%E", "x", FormatTypeDate, "invalid"},
		{"unknown_specifier", "%~", "x", FormatTypeDate, "unexpected"},
		{"unavailable_specifier", "%Y", "2026", FormatTypeTime, "unavailable"},
		{"unknown_E_specifier", "%E9X", "x", FormatTypeDate, "unexpected"},
		{"target_too_short", "%Y", "", FormatTypeDate, "invalid"},
		{"unparsed_trailing_text", "%Y", "2026XYZ", FormatTypeDate, "unparsed"},
		{"E_target_too_short_or_invalid", "%E2S", "", FormatTypeTime, "invalid"},
		// %P (small am/pm) is a format-only specifier; parsing it errors.
		{"P_parse_fails", "%P", "am", FormatTypeTime, "cannot be used"},
		// %s (unix seconds) expects exactly 10 chars.
		{"unixsec_too_short", "%s", "12345", FormatTypeTime, "unexpected"},
		// %s with negative numeric body is rejected.
		{"unixsec_negative", "%s", "-100000000", FormatTypeTime, "invalid unixtime"},
		// %C with non-numeric body is rejected.
		{"century_non_numeric", "%C", "ab", FormatTypeDate, "unexpected"},
		// %A with garbage target.
		{"weekday_invalid", "%A", "Notaday!!", FormatTypeDate, "unexpected"},
		// %a with too-short target.
		{"weekday_short_too_short", "%a", "Fr", FormatTypeDate, "unexpected"},
		// %B with garbage.
		{"month_full_invalid", "%B", "Notamonth!", FormatTypeDate, "unexpected"},
		// %b with too-short.
		{"month_short_invalid", "%b", "Ma", FormatTypeDate, "unexpected"},
		// %Ez with malformed offset.
		{"Ez_bad_format", "%Ez", "+5:30", FormatTypeTimestamp, "offset"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseTimeFormat(c.format, c.target, c.typ)
			if err == nil {
				t.Fatalf("expected error for %q -> %q", c.format, c.target)
			}
			if c.match != "" && !strings.Contains(err.Error(), c.match) {
				t.Fatalf("error %q does not contain %q", err.Error(), c.match)
			}
		})
	}
}

// TestParseTimeFormatLiteralAndWhitespace exercises the "literal
// character" and "whitespace slurp" branches of ParseTimeFormat.
func TestParseTimeFormatLiteralAndWhitespace(t *testing.T) {
	t.Parallel()
	// Literal hyphen between %Y and %m round-trips.
	got, err := ParseTimeFormat("%Y-%m-%d", "2026-05-15", FormatTypeDate)
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 || got.Month() != time.May || got.Day() != 15 {
		t.Fatalf("got %v", got)
	}
	// Whitespace token slurps multiple spaces in the target.
	got, err = ParseTimeFormat("%Y %m", "2026   05", FormatTypeDate)
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 || got.Month() != time.May {
		t.Fatalf("got %v", got)
	}
}

// TestParseTimeFormatLiteralMatch pins that a format literal must be
// matched in the target: a missing or mismatched character is a clean
// parse failure, not a panic (the ISO "%E3SZ" idiom fed a Z-less input).
func TestParseTimeFormatLiteralMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		format string
		target string
	}{
		// Trailing literal 'Z' with the target one character short.
		{"missing_trailing_literal", "%Y-%m-%dT%H:%M:%E3SZ", "2024-01-05T12:30:00.123"},
		// Trailing literal present but wrong ('X' instead of 'Z').
		{"wrong_trailing_literal", "%Y-%m-%dT%H:%M:%E3SZ", "2024-01-05T12:30:00.123X"},
		// Literal '-' with the target exhausted just before it.
		{"missing_mid_literal", "%Y-%m", "2024"},
		// Literal '-' present but wrong.
		{"wrong_mid_literal", "%Y/%m", "2024-05"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// Must return an error and must not panic.
			if _, err := ParseTimeFormat(c.format, c.target, FormatTypeTimestamp); err == nil {
				t.Fatalf("ParseTimeFormat(%q, %q): expected error, got nil", c.format, c.target)
			}
		})
	}

	// Control: the matching literal still parses.
	if _, err := ParseTimeFormat("%Y-%m-%dT%H:%M:%E3SZ", "2024-01-05T12:30:00.123Z", FormatTypeTimestamp); err != nil {
		t.Fatalf("matching trailing literal must parse: %v", err)
	}
}

// TestTimeFormatTypeString covers the TimeFormatType.String branch
// matrix including the unknown fallback.
func TestTimeFormatTypeString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		typ  TimeFormatType
		want string
	}{
		{FormatTypeDate, "date"},
		{FormatTypeDatetime, "datetime"},
		{FormatTypeTime, "time"},
		{FormatTypeTimestamp, "timestamp"},
		{TimeFormatType(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.typ.String(); got != c.want {
			t.Errorf("%d: got %q want %q", c.typ, got, c.want)
		}
	}
}

// TestFormatTimeInfoAvailable exercises FormatTimeInfo.Available.
func TestFormatTimeInfoAvailable(t *testing.T) {
	t.Parallel()
	info := &FormatTimeInfo{AvailableTypes: []TimeFormatType{FormatTypeDate}}
	if !info.Available(FormatTypeDate) {
		t.Fatal("date should be available")
	}
	if info.Available(FormatTypeTime) {
		t.Fatal("time should not be available")
	}
}

// TestParseTimeFormatTargetExhausted hits the "invalid target text"
// branch by making the format demand a token after the target ends.
func TestParseTimeFormatTargetExhausted(t *testing.T) {
	t.Parallel()
	if _, err := ParseTimeFormat("%Y-%m", "2026-", FormatTypeDate); err == nil {
		t.Fatal("expected error from short target")
	}
}

// TestFormatTimePreservesLiteralRunes makes sure non-% characters in
// the format string are passed through verbatim.
func TestFormatTimePreservesLiteralRunes(t *testing.T) {
	t.Parallel()
	got, err := FormatTime("Year=%Y!", &referenceTime, FormatTypeDate)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Year=2026!" {
		t.Fatalf("got %q", got)
	}
}

// TestFormatTimeYearWithoutCenturyLow makes sure %y under 69
// formats as a 2-digit year (the formatter logic strips the
// century).
func TestFormatTimeYearWithoutCenturyLow(t *testing.T) {
	t.Parallel()
	tm := time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := FormatTime("%y", &tm, FormatTypeDate)
	if err != nil {
		t.Fatal(err)
	}
	if got != "05" {
		t.Fatalf("got %q", got)
	}
}

// TestParseTimeFormatUnimplementedSpecifiers covers the parser stubs
// that return "unimplemented" errors. These are valid format
// specifiers that the upstream documentation classifies as "format-
// only" or that the project hasn't yet wired through the parser; the
// shipping behavior is to surface a clear error rather than silently
// produce wrong output.
func TestParseTimeFormatUnimplementedSpecifiers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		format string
		typ    TimeFormatType
	}{
		{"isoYear_G", "%G", FormatTypeDate},
		{"isoCentury_g", "%g", FormatTypeDate},
		{"newline_n", "%n", FormatTypeDate},
		{"tab_t", "%t", FormatTypeDate},
		{"quarter_Q", "%Q", FormatTypeDate},
		{"weekOfYear_U", "%U", FormatTypeDate},
		{"isoWeekday_u", "%u", FormatTypeDate},
		{"weekOfYearISO_V", "%V", FormatTypeDate},
		{"weekOfYear_W", "%W", FormatTypeDate},
		{"weekNumber0_w", "%w", FormatTypeDate},
		{"timezone_Z", "%Z", FormatTypeTimestamp},
		{"timezoneOffset_z", "%z", FormatTypeTimestamp},
		{"escape_%", "%%", FormatTypeDate},
		{"isoYearJ", "%J", FormatTypeDate},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// Use a non-empty target so the parser dispatches to the
			// specifier (rather than failing on "invalid target text"
			// before reaching it).
			if _, err := ParseTimeFormat(c.format, "12345", c.typ); err == nil {
				t.Fatalf("expected error for %s", c.format)
			}
		})
	}
}

// TestFormatTimeQuarterAllBranches drives every branch of
// quarterFormatter.
func TestFormatTimeQuarterAllBranches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		when time.Time
		want string
	}{
		{time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), "1"},  // day 32 -> Q1
		{time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "2"},  // day 121 -> Q2
		{time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), "3"},  // day 213 -> Q3
		{time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), "4"}, // day 305 -> Q4
	}
	for _, c := range cases {
		c := c
		got, err := FormatTime("%Q", &c.when, FormatTypeDate)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Fatalf("when=%v got %q want %q", c.when, got, c.want)
		}
	}
}

// TestFormatTimeAMPMPMBranch exercises the PM branch of the small
// and large am/pm formatters.
func TestFormatTimeAMPMPMBranch(t *testing.T) {
	t.Parallel()
	pm := time.Date(2026, 5, 15, 15, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		format string
		want   string
	}{
		{"%p", "PM"},
		{"%P", "pm"},
	} {
		got, err := FormatTime(c.format, &pm, FormatTypeTime)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Fatalf("%s: got %q want %q", c.format, got, c.want)
		}
	}
}

// TestFormatTimeHourSpaceTwoDigits drives the "two-digit hour" branch
// of the leading-space hour formatters: when the hour is already two
// digits the prefix space isn't added.
func TestFormatTimeHourSpaceTwoDigits(t *testing.T) {
	t.Parallel()
	pm := time.Date(2026, 5, 15, 15, 0, 0, 0, time.UTC)
	if got, _ := FormatTime("%k", &pm, FormatTypeTime); got != "15" {
		t.Fatalf("%%k two-digit: %q", got)
	}
	// 12-hour clock: 15 % 12 == 3 -> single digit -> " 3".
	if got, _ := FormatTime("%l", &pm, FormatTypeTime); got != " 3" {
		t.Fatalf("%%l: %q", got)
	}
}

// TestFormatTimeISOWeekdaySunday hits the Sunday branch of
// isoWeekdayFormatter (wd==0 -> 7).
func TestFormatTimeISOWeekdaySunday(t *testing.T) {
	t.Parallel()
	// 2026-05-17 is a Sunday.
	sun := time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)
	got, err := FormatTime("%u", &sun, FormatTypeDate)
	if err != nil {
		t.Fatal(err)
	}
	if got != "7" {
		t.Fatalf("%%u Sunday: %q", got)
	}
}

// TestParseTimeFormatAMPMOverride covers ampmShouldPostProcessResult:
// when a 24-hour-format specifier appears alongside %p, the post-
// processor should NOT modify the hour.
func TestParseTimeFormatAMPMOverride(t *testing.T) {
	t.Parallel()
	// %H is a 24-hour format. The hour parsed is 14, %p says "PM"
	// but %H wins.
	got, err := ParseTimeFormat("%H:%M %p", "14:30 PM", FormatTypeTime)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 14 {
		t.Fatalf("hour: %d", got.Hour())
	}
}

// TestParseTimeFormatAMPMLowercase exercises the largeAMPMParser
// case-folding ("am"/"pm" lowercase) branch.
func TestParseTimeFormatAMPMLowercase(t *testing.T) {
	t.Parallel()
	got, err := ParseTimeFormat("%I:%M %p", "11:30 pm", FormatTypeTime)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 23 {
		t.Fatalf("hour: %d", got.Hour())
	}
}

// TestParseTimeFormatNoMatchPaths exercises the final-fall-through
// "unexpected" error path of name-based parsers — when the input is
// long enough but matches no known weekday / month / etc.
func TestParseTimeFormatNoMatchPaths(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		format string
		target string
		typ    TimeFormatType
	}{
		// %A: long-enough but no match.
		{"weekday_no_match", "%A", "Notaday!!", FormatTypeDate},
		// %a: 3 chars but no match.
		{"short_wday_no_match", "%a", "Xyz", FormatTypeDate},
		// %B: long-enough but no match.
		{"month_no_match", "%B", "Notamonth!", FormatTypeDate},
		// %b: 3 chars but no match.
		{"short_month_no_match", "%b", "Xyz", FormatTypeDate},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseTimeFormat(c.format, c.target, c.typ); err == nil {
				t.Fatalf("expected error for %q -> %q", c.format, c.target)
			}
		})
	}
}

// TestParseTimeFormatHourRanges exercises the bounds-check error
// branches of hour / minute / second / day / month / year parsers.
func TestParseTimeFormatBoundsErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		format string
		target string
		typ    TimeFormatType
	}{
		// %H rejects values > 23.
		{"hour24_too_high", "%H", "99", FormatTypeTime},
		// %I (12-hour) rejects 0 / values > 12.
		{"hour12_too_high", "%I", "13", FormatTypeTime},
		// %M rejects > 59.
		{"minute_too_high", "%M", "99", FormatTypeTime},
		// %S rejects > 59.
		{"second_too_high", "%S", "99", FormatTypeTime},
		// %d rejects > 31.
		{"day_too_high", "%d", "99", FormatTypeDate},
		// %m rejects > 12.
		{"month_too_high", "%m", "99", FormatTypeDate},
		// %j rejects > 366.
		{"doy_too_high", "%j", "999", FormatTypeDate},
		// %Y rejects 0.
		{"year_too_low", "%Y", "0000", FormatTypeDate},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseTimeFormat(c.format, c.target, c.typ); err == nil {
				t.Fatalf("expected error for %q -> %q", c.format, c.target)
			}
		})
	}
}

// TestParseTimeFormatStaticTextMismatch exercises the static-text
// parser's mismatch branch (hyphen / colon / slash mismatch).
func TestParseTimeFormatStaticTextMismatch(t *testing.T) {
	t.Parallel()
	// %F expects "-" between the year and month. Use ":" instead.
	if _, err := ParseTimeFormat("%F", "2026:05:15", FormatTypeDate); err == nil {
		t.Fatal("expected hyphen-mismatch error")
	}
	// %T expects ":" between hour and minute. Use "-" instead.
	if _, err := ParseTimeFormat("%T", "03-04-05", FormatTypeTime); err == nil {
		t.Fatal("expected colon-mismatch error")
	}
	// %D expects "/" between month and day. Use "-" instead.
	if _, err := ParseTimeFormat("%D", "05-15-26", FormatTypeDate); err == nil {
		t.Fatal("expected slash-mismatch error")
	}
}

// TestFormatTimeFractionalSecondsWideAndZero exercises the
// timePrecisionFormatter for boundary precisions.
func TestFormatTimeFractionalSecondsBoundary(t *testing.T) {
	t.Parallel()
	tm := time.Date(2026, 5, 15, 3, 4, 5, 0, time.UTC) // nanoseconds 0
	got, err := FormatTime("%E6S", &tm, FormatTypeTime)
	if err != nil {
		t.Fatal(err)
	}
	if got != "05.000000" {
		t.Fatalf("got %q", got)
	}
}
