package spanner

import (
	gobytes "bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// -------- mysql_string extras --------

// BindChar returns the STRING formed by interpreting each INT64
// argument as a Unicode code point.
func BindChar(args ...value.Value) (value.Value, error) {
	if helper.ExistsNull(args) {
		return nil, nil
	}
	var b strings.Builder
	for _, a := range args {
		n, err := a.ToInt64()
		if err != nil {
			return nil, err
		}
		b.WriteRune(rune(n))
	}
	return value.StringValue(b.String()), nil
}

// BindConcatWs joins args[1:] using args[0] as a separator,
// skipping NULL arguments after the separator.
func BindConcatWs(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("mysql.concat_ws: at least 1 argument required")
	}
	if args[0] == nil {
		return nil, nil
	}
	sep, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	parts := make([]string, 0, len(args)-1)
	for _, a := range args[1:] {
		if a == nil {
			continue
		}
		s, err := a.ToString()
		if err != nil {
			return nil, err
		}
		parts = append(parts, s)
	}
	return value.StringValue(strings.Join(parts, sep)), nil
}

// BindInsert replaces a substring of len `length` starting at
// 1-based position `pos` with `newstr`. Out-of-range positions
// fall back to MySQL's behaviour: pos<=0 or pos>len(src) returns
// src unchanged; length<0 deletes through end-of-string.
func BindInsert(args ...value.Value) (value.Value, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("mysql.insert: invalid number of arguments: got %d, want 4", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	src, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	pos, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	length, err := args[2].ToInt64()
	if err != nil {
		return nil, err
	}
	newstr, err := args[3].ToString()
	if err != nil {
		return nil, err
	}
	if pos <= 0 || int(pos) > len(src) {
		return value.StringValue(src), nil
	}
	start := int(pos - 1)
	end := start + int(length)
	if length < 0 || end > len(src) {
		end = len(src)
	}
	return value.StringValue(src[:start] + newstr + src[end:]), nil
}

// BindLocate returns the 1-based offset of the first occurrence
// of `needle` in `haystack`, optionally starting from `pos`. 0
// when not found.
func BindLocate(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("mysql.locate: invalid number of arguments: got %d, want between 2 and 3", len(args))
	}
	if helper.ExistsNull(args[:2]) {
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
	start := 0
	if len(args) == 3 && args[2] != nil {
		s, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		start = int(s) - 1
		if start < 0 {
			start = 0
		}
		if start > len(haystack) {
			return value.IntValue(0), nil
		}
	}
	idx := strings.Index(haystack[start:], needle)
	if idx < 0 {
		return value.IntValue(0), nil
	}
	return value.IntValue(int64(start+idx) + 1), nil
}

// BindMid is an alias of SUBSTRING. Variadic 2- or 3-arg form.
func BindMid(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("mysql.mid: invalid number of arguments: got %d, want between 2 and 3", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	pos, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	start := int(pos) - 1
	if pos < 0 {
		start = len(s) + int(pos)
	}
	if start < 0 {
		start = 0
	}
	if start > len(s) {
		return value.StringValue(""), nil
	}
	end := len(s)
	if len(args) == 3 {
		l, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		if int(l) < 0 {
			return value.StringValue(""), nil
		}
		end = start + int(l)
		if end > len(s) {
			end = len(s)
		}
	}
	return value.StringValue(s[start:end]), nil
}

// BindOct converts an integer to its octal string representation.
func BindOct(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.oct: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	n, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	return value.StringValue(strconv.FormatInt(n, 8)), nil
}

// BindOrd returns the code-point of the first character of the
// input string, or 0 for the empty string.
func BindOrd(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.ord: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	if s == "" {
		return value.IntValue(0), nil
	}
	for _, r := range s {
		return value.IntValue(int64(r)), nil
	}
	return value.IntValue(0), nil
}

// BindQuote wraps a STRING in single quotes and escapes embedded
// single quotes, NUL, backslash, CR, LF.
func BindQuote(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.quote: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return value.StringValue("NULL"), nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteByte('\'')
	for _, r := range s {
		switch r {
		case '\'':
			b.WriteString(`\'`)
		case '\\':
			b.WriteString(`\\`)
		case 0:
			b.WriteString(`\0`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('\'')
	return value.StringValue(b.String()), nil
}

// BindRegexpLike returns whether `s` matches the regular
// expression `pattern`.
func BindRegexpLike(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.regexp_like: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	pat, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("mysql.regexp_like: invalid pattern: %w", err)
	}
	return value.BoolValue(re.MatchString(s)), nil
}

// BindStrcmp returns -1, 0, or 1 per Go strings.Compare.
func BindStrcmp(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.strcmp: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	a, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	b, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	return value.IntValue(int64(strings.Compare(a, b))), nil
}

// BindSubstringIndex returns the substring of `s` before the
// `count`-th occurrence of `delim`. Negative `count` counts
// occurrences from the right.
func BindSubstringIndex(args ...value.Value) (value.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("mysql.substring_index: invalid number of arguments: got %d, want 3", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	delim, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	count, err := args[2].ToInt64()
	if err != nil {
		return nil, err
	}
	if count == 0 || delim == "" {
		return value.StringValue(""), nil
	}
	parts := strings.Split(s, delim)
	if count > 0 {
		if int(count) >= len(parts) {
			return value.StringValue(s), nil
		}
		return value.StringValue(strings.Join(parts[:count], delim)), nil
	}
	n := int(-count)
	if n >= len(parts) {
		return value.StringValue(s), nil
	}
	return value.StringValue(strings.Join(parts[len(parts)-n:], delim)), nil
}

// BindUnhex decodes a hex STRING to BYTES. Returns NULL on
// invalid input (MySQL semantics).
func BindUnhex(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.unhex: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, nil
	}
	return value.BytesValue(b), nil
}

// -------- mysql_datetime extras (continued) --------

// BindDayName returns the English name of the day-of-week.
func BindDayName(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.dayname: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	return value.StringValue(t.UTC().Weekday().String()), nil
}

// BindMonthName returns the English name of the month.
func BindMonthName(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.monthname: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	return value.StringValue(t.UTC().Month().String()), nil
}

// MySQL's to_days uses 0000-01-01 (proleptic Gregorian with a year-0)
// as its origin. Go's time.Duration is int64 nanoseconds and overflows
// for spans wider than ~292 years, so we cannot compute the span via
// t.Sub(epoch). Instead, do all the arithmetic in seconds-since-Unix-
// epoch (int64) and add a constant offset that maps the Unix epoch
// (1970-01-01) to its MySQL-day-number 719528.
//
// Verified against MySQL: TO_DAYS('1970-01-01') == 719528.
const mysqlDaysFromYear0ToUnixEpoch int64 = 719528

func daysSinceMysqlEpoch(t time.Time) int64 {
	// Floor-divide Unix seconds by 86400 so dates before 1970 still
	// yield the correct calendar day.
	secs := t.UTC().Unix()
	days := secs / 86400
	if secs%86400 < 0 {
		days--
	}
	return days + mysqlDaysFromYear0ToUnixEpoch
}

var mysqlUnixEpochAsTime = time.Unix(0, 0).UTC()

// BindFromDays returns a DATE for the given day count since
// 0000-01-01 (MySQL convention).
func BindFromDays(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.from_days: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	n, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	d := mysqlUnixEpochAsTime.AddDate(0, 0, int(n-mysqlDaysFromYear0ToUnixEpoch))
	return value.DateValue(d), nil
}

// BindToDays returns the day count since 0000-01-01 for the
// given DATE/TIMESTAMP.
func BindToDays(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.to_days: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	return value.IntValue(daysSinceMysqlEpoch(t)), nil
}

// BindToSeconds returns seconds since year 0 for the given
// TIMESTAMP/DATE.
func BindToSeconds(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.to_seconds: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	days := daysSinceMysqlEpoch(t)
	t = t.UTC()
	tod := int64(t.Hour())*3600 + int64(t.Minute())*60 + int64(t.Second())
	return value.IntValue(days*86400 + tod), nil
}

// BindMakeDate returns a DATE built from year and day-of-year.
func BindMakeDate(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.makedate: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	year, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	doy, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	if doy < 1 {
		return nil, nil
	}
	d := time.Date(int(year), 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(doy-1))
	return value.DateValue(d), nil
}

// BindFromUnixtime converts a UNIX seconds value to TIMESTAMP.
func BindFromUnixtime(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.from_unixtime: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	return value.TimestampValue(time.Unix(s, 0).UTC()), nil
}

// BindSysDate returns the current TIMESTAMP at call time.
func BindSysDate(args ...value.Value) (value.Value, error) {
	return value.TimestampValue(time.Now().UTC()), nil
}

// BindUtcDate returns today's DATE in UTC.
func BindUtcDate(args ...value.Value) (value.Value, error) {
	now := time.Now().UTC()
	d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return value.DateValue(d), nil
}

// BindUtcTimestamp returns the current TIMESTAMP in UTC.
func BindUtcTimestamp(args ...value.Value) (value.Value, error) {
	return value.TimestampValue(time.Now().UTC()), nil
}

// BindPeriodAdd adds `months` to a YYYYMM period number.
func BindPeriodAdd(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.period_add: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	p, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	n, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	y, m := splitPeriod(p)
	tm := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC).AddDate(0, int(n), 0)
	return value.IntValue(int64(tm.Year())*100 + int64(tm.Month())), nil
}

// BindPeriodDiff returns months between two YYYYMM periods.
func BindPeriodDiff(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.period_diff: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	a, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	b, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	ay, am := splitPeriod(a)
	by, bm := splitPeriod(b)
	return value.IntValue(int64((ay-by)*12 + (am - bm))), nil
}

func splitPeriod(p int64) (int, int) {
	if p < 1 {
		return 0, 0
	}
	if p < 7000 {
		// MySQL convention: 2-digit year < 70 → 20xx, else 19xx.
		yy := int(p / 100)
		mm := int(p % 100)
		if yy < 70 {
			yy += 2000
		} else {
			yy += 1900
		}
		return yy, mm
	}
	return int(p / 100), int(p % 100)
}

// BindStrToDate parses a STRING using a MySQL date format string.
// Supports the common subset of MySQL format specifiers, with a
// fallback to a few RFC-ish formats.
func BindStrToDate(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.str_to_date: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	fmtStr, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	layout := mysqlFormatToGo(fmtStr)
	t, perr := time.Parse(layout, s)
	if perr != nil {
		return nil, nil
	}
	return value.TimestampValue(t.UTC()), nil
}

// BindDateFormat formats a TIMESTAMP/DATE/DATETIME using a MySQL
// format string.
func BindDateFormat(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.date_format: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	t, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	fmtStr, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	layout := mysqlFormatToGo(fmtStr)
	return value.StringValue(t.UTC().Format(layout)), nil
}

// mysqlFormatToGo translates a subset of MySQL DATE_FORMAT
// specifiers to Go reference layout. Unrecognised %X codes pass
// through unchanged.
func mysqlFormatToGo(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '%' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'Y':
			b.WriteString("2006")
		case 'y':
			b.WriteString("06")
		case 'm':
			b.WriteString("01")
		case 'c':
			b.WriteString("1")
		case 'd':
			b.WriteString("02")
		case 'e':
			b.WriteString("2")
		case 'H':
			b.WriteString("15")
		case 'h', 'I':
			b.WriteString("03")
		case 'i':
			b.WriteString("04")
		case 'S', 's':
			b.WriteString("05")
		case 'p':
			b.WriteString("PM")
		case 'M':
			b.WriteString("January")
		case 'b':
			b.WriteString("Jan")
		case 'W':
			b.WriteString("Monday")
		case 'a':
			b.WriteString("Mon")
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// -------- mysql_timestamp --------

// BindDateDiff returns the day count between two DATEs/TIMESTAMPs
// (MySQL: arg1 - arg2 in days).
func BindDateDiff(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.datediff: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	a, err := args[0].ToTime()
	if err != nil {
		return nil, err
	}
	b, err := args[1].ToTime()
	if err != nil {
		return nil, err
	}
	a = time.Date(a.UTC().Year(), a.UTC().Month(), a.UTC().Day(), 0, 0, 0, 0, time.UTC)
	b = time.Date(b.UTC().Year(), b.UTC().Month(), b.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return value.IntValue(int64(a.Sub(b) / (24 * time.Hour))), nil
}

// -------- mysql_utility (network + uuid) --------

// BindInetAton converts an IPv4 dotted-decimal string to INT64.
func BindInetAton(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.inet_aton: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(s).To4()
	if ip == nil {
		return nil, nil
	}
	n := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	return value.IntValue(int64(n)), nil
}

// BindInetNtoa converts an INT64 to an IPv4 dotted-decimal string.
func BindInetNtoa(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.inet_ntoa: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	n, err := args[0].ToInt64()
	if err != nil {
		return nil, err
	}
	if n < 0 || n > 0xFFFFFFFF {
		return nil, nil
	}
	ip := net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	return value.StringValue(ip.String()), nil
}

// BindInet6Aton converts an IP string (v4 or v6) to BYTES.
func BindInet6Aton(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.inet6_aton: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, nil
	}
	if v4 := ip.To4(); v4 != nil {
		return value.BytesValue(v4), nil
	}
	return value.BytesValue(ip.To16()), nil
}

// BindInet6Ntoa converts BYTES to an IP string.
func BindInet6Ntoa(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.inet6_ntoa: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	ip := net.IP(b)
	if len(b) != 4 && len(b) != 16 {
		return nil, nil
	}
	return value.StringValue(ip.String()), nil
}

// BindIsIPv4 returns whether `s` parses as IPv4.
func BindIsIPv4(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.is_ipv4: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(s)
	return value.BoolValue(ip != nil && ip.To4() != nil && strings.Contains(s, ".")), nil
}

// BindIsIPv6 returns whether `s` parses as IPv6.
func BindIsIPv6(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.is_ipv6: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(s)
	return value.BoolValue(ip != nil && ip.To4() == nil && strings.Contains(s, ":")), nil
}

// BindIsIPv4Compat returns whether BYTES is an IPv4-compatible IPv6 address.
func BindIsIPv4Compat(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.is_ipv4_compat: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	if len(b) != 16 {
		return value.BoolValue(false), nil
	}
	zero := gobytes.Repeat([]byte{0}, 12)
	return value.BoolValue(gobytes.Equal(b[:12], zero)), nil
}

// BindIsIPv4Mapped returns whether BYTES is an IPv4-mapped IPv6 address.
func BindIsIPv4Mapped(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.is_ipv4_mapped: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	if len(b) != 16 {
		return value.BoolValue(false), nil
	}
	expect := append(gobytes.Repeat([]byte{0}, 10), 0xff, 0xff)
	return value.BoolValue(gobytes.Equal(b[:12], expect)), nil
}

var uuidRE = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// BindIsUUID returns whether `s` is a canonical UUID string.
func BindIsUUID(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.is_uuid: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return value.BoolValue(uuidRE.MatchString(s)), nil
}

// BindBinToUUID converts 16 BYTES to a canonical UUID STRING.
func BindBinToUUID(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.bin_to_uuid: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	if len(b) != 16 {
		return nil, nil
	}
	s := hex.EncodeToString(b)
	return value.StringValue(s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]), nil
}

// BindUUIDToBin converts a UUID STRING to 16 BYTES.
func BindUUIDToBin(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.uuid_to_bin: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	if !uuidRE.MatchString(s) {
		return nil, nil
	}
	b, err := hex.DecodeString(strings.ReplaceAll(s, "-", ""))
	if err != nil {
		return nil, nil
	}
	return value.BytesValue(b), nil
}

// -------- mysql_encryption --------

// BindSHA2 returns the SHA-256/384/512 digest as a lowercase hex
// STRING. Bit-length must be one of 0/224/256/384/512 (0 == 256).
func BindSHA2(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("mysql.sha2: invalid number of arguments: got %d, want 2", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	n, err := args[1].ToInt64()
	if err != nil {
		return nil, err
	}
	switch n {
	case 0, 256:
		h := sha256.Sum256([]byte(s))
		return value.StringValue(hex.EncodeToString(h[:])), nil
	case 224:
		h := sha256.Sum224([]byte(s))
		return value.StringValue(hex.EncodeToString(h[:])), nil
	case 384:
		h := sha512.Sum384([]byte(s))
		return value.StringValue(hex.EncodeToString(h[:])), nil
	case 512:
		h := sha512.Sum512([]byte(s))
		return value.StringValue(hex.EncodeToString(h[:])), nil
	}
	return nil, nil
}

// -------- mysql_json --------

// BindJsonQuote returns the JSON-encoded form of a STRING (i.e.
// wrapped in quotes with embedded special characters escaped).
func BindJsonQuote(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.json_quote: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return value.StringValue(jsonEncodeString(s)), nil
}

// BindJsonUnquote strips outer JSON quoting and unescapes the
// payload. Non-quoted inputs pass through unchanged.
func BindJsonUnquote(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mysql.json_unquote: invalid number of arguments: got %d, want 1", len(args))
	}
	if helper.ExistsNull(args) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return value.StringValue(s), nil
	}
	out, ok := jsonDecodeString(s[1 : len(s)-1])
	if !ok {
		return value.StringValue(s), nil
	}
	return value.StringValue(out), nil
}

func jsonEncodeString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(`\u`)
				b.WriteString(strconv.FormatUint(uint64(r), 16))
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func jsonDecodeString(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		if i+1 >= len(s) {
			return "", false
		}
		i++
		switch s[i] {
		case '"':
			b.WriteByte('"')
		case '\\':
			b.WriteByte('\\')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case '/':
			b.WriteByte('/')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'u':
			if i+4 >= len(s) {
				return "", false
			}
			n, err := strconv.ParseUint(s[i+1:i+5], 16, 32)
			if err != nil {
				return "", false
			}
			b.WriteRune(rune(n))
			i += 4
		default:
			return "", false
		}
	}
	return b.String(), true
}
