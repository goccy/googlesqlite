// Package spanner contains the per-spec runtime UDFs that back
// Spanner's `mysql.*` compatibility namespace. The analyzer
// resolves `mysql.<name>` through a sub-catalog registered in
// internal/catalog.go; the formatter then routes the call to one
// of the bind functions defined here.
//
// Most aliases are thin wrappers over the equivalent GoogleSQL
// builtin (CHAR_LENGTH, EXTRACT(<part> FROM TIMESTAMP), etc.)
// — the alias exists only so a Spanner-flavoured query can use
// the MySQL spelling.
package spanner

import (
	"fmt"
	gomath "math"
	"strings"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// extractTimestampPart returns the requested time component from
// a TIMESTAMP/DATETIME/DATE value as INT64. Used by DAY / MONTH
// / YEAR / HOUR / MINUTE / SECOND / etc. aliases.
func extractTimestampPart(args []value.Value, part string) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.%s: invalid number of arguments: got %d, want 1", strings.ToLower(part), len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	// EXTRACT defaults to UTC for TIMESTAMP per GoogleSQL semantics.
	// DATE and DATETIME values are already in absolute calendar fields,
	// but ToTime() may return them with a non-UTC location; force UTC
	// here so the answer is independent of the host machine's tz.
	t = t.UTC()
	switch part {
	case "DAY", "DAYOFMONTH":
		return value.IntValue(int64(t.Day())), nil
	case "MONTH":
		return value.IntValue(int64(t.Month())), nil
	case "YEAR":
		return value.IntValue(int64(t.Year())), nil
	case "HOUR":
		return value.IntValue(int64(t.Hour())), nil
	case "MINUTE":
		return value.IntValue(int64(t.Minute())), nil
	case "SECOND":
		return value.IntValue(int64(t.Second())), nil
	case "MICROSECOND":
		return value.IntValue(int64(t.Nanosecond()) / 1000), nil
	case "QUARTER":
		return value.IntValue(int64((int(t.Month())-1)/3) + 1), nil
	case "DAYOFWEEK":
		// MySQL convention: 1=Sunday, 2=Monday, ..., 7=Saturday.
		return value.IntValue(int64(t.Weekday()) + 1), nil
	case "DAYOFYEAR":
		return value.IntValue(int64(t.YearDay())), nil
	case "WEEK", "WEEKOFYEAR":
		_, w := t.ISOWeek()
		return value.IntValue(int64(w)), nil
	case "WEEKDAY":
		// MySQL convention: 0=Monday, ..., 6=Sunday.
		wd := int(t.Weekday())
		if wd == 0 {
			wd = 6
		} else {
			wd = wd - 1
		}
		return value.IntValue(int64(wd)), nil
	}
	return nil, fmt.Errorf("mysql.%s: unsupported part", strings.ToLower(part))
}

func BindDay(args ...value.Value) (value.Value, error)   { return extractTimestampPart(args, "DAY") }
func BindMonth(args ...value.Value) (value.Value, error) { return extractTimestampPart(args, "MONTH") }
func BindYear(args ...value.Value) (value.Value, error)  { return extractTimestampPart(args, "YEAR") }
func BindHour(args ...value.Value) (value.Value, error)  { return extractTimestampPart(args, "HOUR") }
func BindMinute(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "MINUTE")
}
func BindSecond(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "SECOND")
}
func BindMicrosecond(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "MICROSECOND")
}
func BindQuarter(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "QUARTER")
}
func BindDayOfWeek(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "DAYOFWEEK")
}
func BindDayOfYear(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "DAYOFYEAR")
}
func BindDayOfMonth(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "DAYOFMONTH")
}
func BindWeek(args ...value.Value) (value.Value, error) { return extractTimestampPart(args, "WEEK") }
func BindWeekOfYear(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "WEEKOFYEAR")
}
func BindWeekday(args ...value.Value) (value.Value, error) {
	return extractTimestampPart(args, "WEEKDAY")
}

// BindBitLength returns the bit count of a STRING in bytes * 8.
func BindBitLength(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.bit_length: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return value.IntValue(int64(len(s)) * 8), nil
}

// BindHex hex-encodes BYTES as STRING.
func BindHex(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.hex: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	const hexdigits = "0123456789ABCDEF"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hexdigits[c>>4]
		out[i*2+1] = hexdigits[c&0x0F]
	}
	return value.StringValue(string(out)), nil
}

// BindSpace returns a STRING composed of `n` spaces.
func BindSpace(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.space: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	n, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	if n < 0 {
		n = 0
	}
	count, err := helper.SafeInt(n)
	if err != nil {
		return nil, err
	}
	return value.StringValue(strings.Repeat(" ", count)), nil
}

// BindPosition returns the 1-based offset of `needle` within
// `haystack`, or 0 when not found.
func BindPosition(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.position: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	needle, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	haystack, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	idx := strings.Index(haystack, needle)
	if idx < 0 {
		return value.IntValue(0), nil
	}
	return value.IntValue(int64(idx) + 1), nil
}

// BindDegrees converts radians to degrees.
func BindDegrees(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.degrees: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	f, err := args[0].ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(f * 180 / gomath.Pi), nil
}

// BindRadians converts degrees to radians.
func BindRadians(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.radians: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	f, err := args[0].ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(f * gomath.Pi / 180), nil
}

// BindLog2 returns log base 2 of x.
func BindLog2(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.log2: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	f, err := args[0].ToFloat64()
	if err != nil {
		return nil, err
	}
	return value.FloatValue(gomath.Log2(f)), nil
}

// BindTruncate truncates a FLOAT64 toward zero to the given
// decimal places.
func BindTruncate(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.truncate: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	f, err := args[0].ToFloat64()
	if err != nil {
		return nil, err
	}
	d, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	scale := gomath.Pow(10, float64(d))
	return value.FloatValue(gomath.Trunc(f*scale) / scale), nil
}

// BindUnixTimestampNow returns the current Unix timestamp in
// seconds; used for `mysql.unix_timestamp()` with no args.
func BindUnixTimestampNow(args ...value.Value) (value.Value, error) {
	return value.IntValue(time.Now().Unix()), nil
}
