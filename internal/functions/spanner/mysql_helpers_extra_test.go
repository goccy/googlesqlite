package spanner

import (
	"strings"
	"testing"
	"time"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestBindChar(t *testing.T) {
	t.Parallel()

	// MySQL: CHAR(77, 121, 83, 81, 76) -> 'MySQL'.
	got, err := BindChar(
		value.IntValue(77), value.IntValue(121),
		value.IntValue(83), value.IntValue(81), value.IntValue(76),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "MySQL" {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindChar()
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty")
	}

	if v, _ := BindChar(nil, value.IntValue(65)); v != nil {
		t.Fatal("expected null propagation")
	}
}

func TestBindConcatWs(t *testing.T) {
	t.Parallel()

	// MySQL: CONCAT_WS(',', 'a', 'b', 'c') = 'a,b,c'.
	got, err := BindConcatWs(value.StringValue(","), value.StringValue("a"), value.StringValue("b"), value.StringValue("c"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "a,b,c" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// MySQL: NULLs after the separator are skipped.
	got, err = BindConcatWs(value.StringValue("-"), value.StringValue("a"), nil, value.StringValue("c"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "a-c" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// MySQL: when the separator is NULL the result is NULL.
	v, err := BindConcatWs(nil, value.StringValue("a"))
	if err != nil || v != nil {
		t.Fatal("expected null separator -> null")
	}

	if _, err := BindConcatWs(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindInsert(t *testing.T) {
	t.Parallel()

	// MySQL: INSERT('Quadratic', 3, 4, 'What') = 'QuWhattic'.
	got, err := BindInsert(value.StringValue("Quadratic"), value.IntValue(3), value.IntValue(4), value.StringValue("What"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "QuWhattic" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Out-of-range pos returns input unchanged.
	got, err = BindInsert(value.StringValue("Quadratic"), value.IntValue(-1), value.IntValue(4), value.StringValue("X"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "Quadratic" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Negative length deletes through end.
	got, err = BindInsert(value.StringValue("Quadratic"), value.IntValue(3), value.IntValue(-1), value.StringValue("What"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "QuWhat" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if _, err := BindInsert(value.StringValue("a"), value.IntValue(1)); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindInsert(nil, value.IntValue(1), value.IntValue(1), value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindLocate(t *testing.T) {
	t.Parallel()

	// MySQL: LOCATE('bar', 'foobarbar') = 4.
	got, err := BindLocate(value.StringValue("bar"), value.StringValue("foobarbar"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 4 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// MySQL: LOCATE('bar', 'foobarbar', 5) = 7.
	got, err = BindLocate(value.StringValue("bar"), value.StringValue("foobarbar"), value.IntValue(5))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 7 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindLocate(value.StringValue("xyzzy"), value.StringValue("abc"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatal("expected 0 not-found")
	}

	if _, err := BindLocate(value.StringValue("a")); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindLocate(nil, value.StringValue("foo")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindMid(t *testing.T) {
	t.Parallel()

	// MySQL: MID('foobar', 2, 3) = 'oob'.
	got, err := BindMid(value.StringValue("foobar"), value.IntValue(2), value.IntValue(3))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "oob" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Negative pos counts from end: MID('foobar', -3) = 'bar'.
	got, err = BindMid(value.StringValue("foobar"), value.IntValue(-3))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "bar" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// pos beyond length returns empty.
	got, err = BindMid(value.StringValue("abc"), value.IntValue(10))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty")
	}

	// Negative length returns empty.
	got, err = BindMid(value.StringValue("foobar"), value.IntValue(2), value.IntValue(-1))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty")
	}

	if _, err := BindMid(value.StringValue("a")); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindMid(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindOct(t *testing.T) {
	t.Parallel()

	// MySQL: OCT(12) = '14'.
	got, err := BindOct(value.IntValue(12))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "14" {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindOct(value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "0" {
		t.Fatal("expected 0")
	}

	if v, _ := BindOct(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindOct(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindOrd(t *testing.T) {
	t.Parallel()

	// MySQL: ORD('A') = 65.
	got, err := BindOrd(value.StringValue("A"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 65 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindOrd(value.StringValue(""))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatal("expected 0 for empty")
	}

	if v, _ := BindOrd(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindOrd(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindQuote(t *testing.T) {
	t.Parallel()

	// MySQL: QUOTE('Don\'t') = '\'Don\\\'t\''.
	got, err := BindQuote(value.StringValue("Don't"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `'Don\'t'` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// MySQL: QUOTE(NULL) = 'NULL' (literal, no quotes).
	got, err = BindQuote(nil)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "NULL" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if _, err := BindQuote(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindRegexpLike(t *testing.T) {
	t.Parallel()

	got, err := BindRegexpLike(value.StringValue("abc123"), value.StringValue(`^[a-z]+[0-9]+$`))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	got, err = BindRegexpLike(value.StringValue("abc"), value.StringValue(`^[0-9]+$`))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}

	if _, err := BindRegexpLike(value.StringValue("x"), value.StringValue(`[`)); err == nil {
		t.Fatal("expected error on invalid regexp")
	}
	if _, err := BindRegexpLike(value.StringValue("x")); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindRegexpLike(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStrcmp(t *testing.T) {
	t.Parallel()

	// MySQL: STRCMP('a', 'b') = -1.
	got, err := BindStrcmp(value.StringValue("a"), value.StringValue("b"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != -1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindStrcmp(value.StringValue("ab"), value.StringValue("ab"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindStrcmp(value.StringValue("b"), value.StringValue("a"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 1 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if _, err := BindStrcmp(value.StringValue("a")); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSubstringIndex(t *testing.T) {
	t.Parallel()

	// MySQL: SUBSTRING_INDEX('www.mysql.com', '.', 2) = 'www.mysql'.
	got, err := BindSubstringIndex(value.StringValue("www.mysql.com"), value.StringValue("."), value.IntValue(2))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "www.mysql" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// MySQL: SUBSTRING_INDEX('www.mysql.com', '.', -2) = 'mysql.com'.
	got, err = BindSubstringIndex(value.StringValue("www.mysql.com"), value.StringValue("."), value.IntValue(-2))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "mysql.com" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Count == 0 yields empty.
	got, err = BindSubstringIndex(value.StringValue("a.b.c"), value.StringValue("."), value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty")
	}

	// Count >= len(parts) yields original.
	got, err = BindSubstringIndex(value.StringValue("a.b"), value.StringValue("."), value.IntValue(99))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "a.b" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if _, err := BindSubstringIndex(value.StringValue("a")); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindUnhex(t *testing.T) {
	t.Parallel()

	got, err := BindUnhex(value.StringValue("4D7953514C"))
	if err != nil {
		t.Fatal(err)
	}
	if string(mustBytes(t, got)) != "MySQL" {
		t.Fatalf("got %q", mustBytes(t, got))
	}

	// Invalid hex yields NULL (MySQL behaviour).
	v, err := BindUnhex(value.StringValue("not-hex"))
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatal("expected null")
	}

	if v, _ := BindUnhex(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindUnhex(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindDayMonthName(t *testing.T) {
	t.Parallel()

	ts := value.TimestampValue(time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC))
	got, err := BindDayName(ts)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "Wednesday" {
		t.Fatalf("got %q", mustString(t, got))
	}

	got, err = BindMonthName(ts)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "January" {
		t.Fatalf("got %q", mustString(t, got))
	}

	if v, _ := BindDayName(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindMonthName(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindDayName(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindMonthName(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindToDaysFromDays(t *testing.T) {
	t.Parallel()

	// MySQL: TO_DAYS('1970-01-01') = 719528.
	d := value.DateValue(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
	got, err := BindToDays(d)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 719528 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// MySQL: FROM_DAYS(719528) = '1970-01-01'.
	got, err = BindFromDays(value.IntValue(719528))
	if err != nil {
		t.Fatal(err)
	}
	dv, ok := got.(value.DateValue)
	if !ok {
		t.Fatalf("expected DateValue, got %T", got)
	}
	if !time.Time(dv).Equal(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("got %v", time.Time(dv))
	}

	if v, _ := BindToDays(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindFromDays(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindToSeconds(t *testing.T) {
	t.Parallel()

	// MySQL: TO_SECONDS('1970-01-01 00:00:00') = 62167219200.
	ts := value.TimestampValue(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
	got, err := BindToSeconds(ts)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 62167219200 {
		t.Fatalf("got %d", mustInt64(t, got))
	}
	if v, _ := BindToSeconds(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindToSeconds(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindMakeDate(t *testing.T) {
	t.Parallel()

	got, err := BindMakeDate(value.IntValue(2023), value.IntValue(60))
	if err != nil {
		t.Fatal(err)
	}
	dv, ok := got.(value.DateValue)
	if !ok {
		t.Fatalf("expected DateValue, got %T", got)
	}
	want := time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC) // day-of-year 60 in 2023
	if !time.Time(dv).Equal(want) {
		t.Fatalf("got %v want %v", time.Time(dv), want)
	}

	// doy < 1 -> NULL.
	v, err := BindMakeDate(value.IntValue(2023), value.IntValue(0))
	if err != nil || v != nil {
		t.Fatal("expected null for doy=0")
	}
	if _, err := BindMakeDate(value.IntValue(2023)); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindFromUnixtime(t *testing.T) {
	t.Parallel()

	got, err := BindFromUnixtime(value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	tv, ok := got.(value.TimestampValue)
	if !ok {
		t.Fatalf("expected TimestampValue, got %T", got)
	}
	if !time.Time(tv).Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("got %v", time.Time(tv))
	}

	if v, _ := BindFromUnixtime(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindFromUnixtime(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindSysDateUtcDateUtcTimestamp(t *testing.T) {
	t.Parallel()

	// Just verify each helper returns a value of the right shape.
	if v, err := BindSysDate(); err != nil || v == nil {
		t.Fatalf("BindSysDate: %v %v", v, err)
	}
	if v, err := BindUtcDate(); err != nil || v == nil {
		t.Fatalf("BindUtcDate: %v %v", v, err)
	}
	if v, err := BindUtcTimestamp(); err != nil || v == nil {
		t.Fatalf("BindUtcTimestamp: %v %v", v, err)
	}
}

func TestBindPeriodAdd(t *testing.T) {
	t.Parallel()

	// MySQL: PERIOD_ADD(200801, 2) = 200803.
	got, err := BindPeriodAdd(value.IntValue(200801), value.IntValue(2))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 200803 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// Two-digit year < 70 -> 20xx.
	got, err = BindPeriodAdd(value.IntValue(6912), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 207001 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if _, err := BindPeriodAdd(value.IntValue(1)); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindPeriodAdd(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindPeriodDiff(t *testing.T) {
	t.Parallel()

	// MySQL: PERIOD_DIFF(200802, 200703) = 11.
	got, err := BindPeriodDiff(value.IntValue(200802), value.IntValue(200703))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 11 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if _, err := BindPeriodDiff(value.IntValue(1)); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindPeriodDiff(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindStrToDateAndDateFormat(t *testing.T) {
	t.Parallel()

	// MySQL: STR_TO_DATE('2023-06-14', '%Y-%m-%d').
	got, err := BindStrToDate(value.StringValue("2023-06-14"), value.StringValue("%Y-%m-%d"))
	if err != nil {
		t.Fatal(err)
	}
	tv, ok := got.(value.TimestampValue)
	if !ok {
		t.Fatalf("expected TimestampValue, got %T", got)
	}
	want := time.Date(2023, 6, 14, 0, 0, 0, 0, time.UTC)
	if !time.Time(tv).Equal(want) {
		t.Fatalf("got %v want %v", time.Time(tv), want)
	}

	// DATE_FORMAT round-trip.
	ts := value.TimestampValue(time.Date(2023, 6, 14, 5, 23, 45, 0, time.UTC))
	gotS, err := BindDateFormat(ts, value.StringValue("%Y-%m-%d %H:%i:%s"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, gotS) != "2023-06-14 05:23:45" {
		t.Fatalf("got %q", mustString(t, gotS))
	}

	// Failed parse yields NULL.
	v, err := BindStrToDate(value.StringValue("garbage"), value.StringValue("%Y-%m-%d"))
	if err != nil || v != nil {
		t.Fatal("expected null on parse failure")
	}
	if _, err := BindStrToDate(value.StringValue("x")); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindDateFormat(value.StringValue("x")); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindDateDiff(t *testing.T) {
	t.Parallel()

	// MySQL: DATEDIFF('2023-06-14', '2023-06-10') = 4.
	a := value.DateValue(time.Date(2023, 6, 14, 0, 0, 0, 0, time.UTC))
	b := value.DateValue(time.Date(2023, 6, 10, 0, 0, 0, 0, time.UTC))
	got, err := BindDateDiff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 4 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	got, err = BindDateDiff(b, a)
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != -4 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	if _, err := BindDateDiff(a); err == nil {
		t.Fatal("expected error")
	}
	if v, _ := BindDateDiff(nil, a); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindInetAton(t *testing.T) {
	t.Parallel()

	// MySQL: INET_ATON('10.0.5.9') = 167773449.
	got, err := BindInetAton(value.StringValue("10.0.5.9"))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 167773449 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// Invalid input -> NULL.
	v, err := BindInetAton(value.StringValue("not-an-ip"))
	if err != nil || v != nil {
		t.Fatal("expected null for invalid")
	}
	if _, err := BindInetAton(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindInetNtoa(t *testing.T) {
	t.Parallel()

	// MySQL: INET_NTOA(167773449) = '10.0.5.9'.
	got, err := BindInetNtoa(value.IntValue(167773449))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "10.0.5.9" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Out-of-range -> NULL.
	v, err := BindInetNtoa(value.IntValue(-1))
	if err != nil || v != nil {
		t.Fatal("expected null for negative")
	}
	v, err = BindInetNtoa(value.IntValue(int64(1) << 33))
	if err != nil || v != nil {
		t.Fatal("expected null for >32-bit")
	}
}

func TestBindInet6AtonNtoa(t *testing.T) {
	t.Parallel()

	got, err := BindInet6Aton(value.StringValue("::1"))
	if err != nil {
		t.Fatal(err)
	}
	b := mustBytes(t, got)
	if len(b) != 16 {
		t.Fatalf("expected 16 bytes, got %d", len(b))
	}

	// IPv4 form yields 4 bytes.
	got, err = BindInet6Aton(value.StringValue("1.2.3.4"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mustBytes(t, got)) != 4 {
		t.Fatal("expected 4 bytes for ipv4")
	}

	// Round-trip: ntoa(aton(::1)) == "::1".
	bv, _ := BindInet6Aton(value.StringValue("::1"))
	s, err := BindInet6Ntoa(bv)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, s) != "::1" {
		t.Fatalf("got %q", mustString(t, s))
	}

	// Invalid byte lengths -> NULL.
	v, err := BindInet6Ntoa(value.BytesValue{1, 2, 3})
	if err != nil || v != nil {
		t.Fatal("expected null for invalid length")
	}
	if _, err := BindInet6Aton(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindInet6Ntoa(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindIsIPvN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		fn   func(...value.Value) (value.Value, error)
		in   string
		want bool
	}{
		{"is_ipv4 valid", BindIsIPv4, "10.0.0.1", true},
		{"is_ipv4 not v4", BindIsIPv4, "::1", false},
		{"is_ipv6 valid", BindIsIPv6, "::1", true},
		{"is_ipv6 not v6", BindIsIPv6, "10.0.0.1", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.fn(value.StringValue(c.in))
			if err != nil {
				t.Fatal(err)
			}
			if mustBool(t, got) != c.want {
				t.Fatalf("got %v want %v", mustBool(t, got), c.want)
			}
		})
	}

	if _, err := BindIsIPv4(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindIsIPv6(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindIsIPv4CompatMapped(t *testing.T) {
	t.Parallel()

	// ::1.2.3.4 (IPv4-compatible IPv6).
	v, _ := BindInet6Aton(value.StringValue("::1.2.3.4"))
	if v != nil {
		got, err := BindIsIPv4Compat(v)
		if err != nil {
			t.Fatal(err)
		}
		if !mustBool(t, got) {
			t.Fatal("expected true for ::1.2.3.4")
		}
	}

	// Build IPv4-mapped IPv6 manually: ::ffff:1.2.3.4 is 0x00*10, 0xff, 0xff, 1, 2, 3, 4.
	mapped := value.BytesValue{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 1, 2, 3, 4}
	got, err := BindIsIPv4Mapped(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	// Wrong length -> false.
	got, err = BindIsIPv4Compat(value.BytesValue{1, 2, 3, 4})
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false for 4-byte input")
	}

	if _, err := BindIsIPv4Compat(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindIsIPv4Mapped(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindIsUUID(t *testing.T) {
	t.Parallel()

	got, err := BindIsUUID(value.StringValue("550e8400-e29b-41d4-a716-446655440000"))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected true")
	}

	got, err = BindIsUUID(value.StringValue("not-uuid"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}
	if _, err := BindIsUUID(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindBinUUIDRoundTrip(t *testing.T) {
	t.Parallel()

	uuid := "550e8400-e29b-41d4-a716-446655440000"
	bin, err := BindUUIDToBin(value.StringValue(uuid))
	if err != nil {
		t.Fatal(err)
	}
	if len(mustBytes(t, bin)) != 16 {
		t.Fatalf("expected 16 bytes, got %d", len(mustBytes(t, bin)))
	}
	back, err := BindBinToUUID(bin)
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, back) != uuid {
		t.Fatalf("got %q", mustString(t, back))
	}

	// Invalid UUID -> NULL.
	v, err := BindUUIDToBin(value.StringValue("not-uuid"))
	if err != nil || v != nil {
		t.Fatal("expected null for invalid uuid")
	}
	v, err = BindBinToUUID(value.BytesValue{1, 2, 3})
	if err != nil || v != nil {
		t.Fatal("expected null for short bytes")
	}
	if _, err := BindUUIDToBin(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindBinToUUID(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindSHA2(t *testing.T) {
	t.Parallel()

	// MySQL: SHA2('abc', 256) = 'ba7816bf...0015ad'.
	got, err := BindSHA2(value.StringValue("abc"), value.IntValue(256))
	if err != nil {
		t.Fatal(err)
	}
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if mustString(t, got) != want {
		t.Fatalf("got %q", mustString(t, got))
	}

	// 0 means 256 in MySQL.
	got, err = BindSHA2(value.StringValue("abc"), value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != want {
		t.Fatal("expected same as 256")
	}

	// 224 / 384 / 512 produce different-length hex strings.
	for _, n := range []int64{224, 384, 512} {
		got, err = BindSHA2(value.StringValue(""), value.IntValue(n))
		if err != nil {
			t.Fatalf("SHA2(%d): %v", n, err)
		}
		if got == nil {
			t.Fatalf("SHA2(%d): expected non-nil", n)
		}
	}

	// Unknown bit-length -> NULL.
	v, err := BindSHA2(value.StringValue("abc"), value.IntValue(999))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSHA2(value.StringValue("x")); err == nil {
		t.Fatal("expected arg count error")
	}
}

// TestBindJsonQuoteUnquoteEscapes covers all JSON-escape cases in
// the project's hand-rolled decode/encode (\\/\n\r\t\b\f\u).
// TestBindDateFormatMicro exercises additional DATE_FORMAT specifiers
// that lift coverage on the format-translator branches.
func TestBindDateFormatMicro(t *testing.T) {
	t.Parallel()

	ts := value.TimestampValue(time.Date(2023, 6, 14, 5, 23, 45, 0, time.UTC))

	// Use a complex format to ensure no panic and all branches walk.
	got, err := BindDateFormat(ts, value.StringValue("%Y-%m-%d %H:%i:%s"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "2023-06-14 05:23:45" {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindJsonQuoteUnquoteEscapes(t *testing.T) {
	t.Parallel()

	// Unquote each escape recognised by jsonDecodeString.
	cases := []struct {
		in   string
		want string
	}{
		{`"\""`, `"`},
		{`"\\"`, `\`},
		{`"\/"`, `/`},
		{`"\n"`, "\n"},
		{`"\r"`, "\r"},
		{`"\t"`, "\t"},
		{`"\b"`, "\b"},
		{`"\f"`, "\f"},
		{`"A"`, "A"},
	}
	for _, c := range cases {
		got, err := BindJsonUnquote(value.StringValue(c.in))
		if err != nil {
			t.Fatalf("%s: %v", c.in, err)
		}
		if mustString(t, got) != c.want {
			t.Fatalf("%s: got %q want %q", c.in, mustString(t, got), c.want)
		}
	}

	// Bad escape -> pass through original string.
	got, err := BindJsonUnquote(value.StringValue(`"\z"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `"\z"` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Truncated \u sequence -> pass through.
	got, err = BindJsonUnquote(value.StringValue(`"\u00"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `"\u00"` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// jsonEncodeString control-char path: code < 0x20 (non-\n/\r/\t).
	got, err = BindJsonQuote(value.StringValue("\x01"))
	if err != nil {
		t.Fatal(err)
	}
	// Result must be a quoted JSON string containing a \u escape.
	if !strings.Contains(mustString(t, got), `\u`) {
		t.Fatalf("got %q", mustString(t, got))
	}
}

// TestBindDateFormatEscapes covers all MySQL DATE_FORMAT specifiers
// the implementation recognises plus the literal-`%` escape and the
// unknown-specifier pass-through.
func TestBindDateFormatEscapes(t *testing.T) {
	t.Parallel()

	// 2023-06-14 17:23:45 UTC, Wednesday, June.
	ts := value.TimestampValue(time.Date(2023, 6, 14, 17, 23, 45, 0, time.UTC))

	cases := []struct {
		format string
		want   string
	}{
		{"%Y/%y", "2023/23"},
		{"%m/%c", "06/6"},
		{"%d/%e", "14/14"},
		{"%h:%i:%S", "05:23:45"},
		{"%I %p", "05 PM"},
		{"%M %b", "June Jun"},
		{"%W %a", "Wednesday Wed"},
		{"%H%%", "17%"},
		// Unknown specifier passes through unchanged.
		{"%q", "%q"},
		// Trailing solitary `%` is preserved.
		{"%", "%"},
	}
	for _, c := range cases {
		got, err := BindDateFormat(ts, value.StringValue(c.format))
		if err != nil {
			t.Fatalf("%s: %v", c.format, err)
		}
		if mustString(t, got) != c.want {
			t.Fatalf("%s: got %q want %q", c.format, mustString(t, got), c.want)
		}
	}
}

// TestBindQuoteEscapes exercises every escape branch in BindQuote.
func TestBindQuoteEscapes(t *testing.T) {
	t.Parallel()

	// Each branch: ' \ \0 \n \r + default rune.
	got, err := BindQuote(value.StringValue("a'b\\c\x00d\ne\rf"))
	if err != nil {
		t.Fatal(err)
	}
	want := `'a\'b\\c\0d\ne\rf'`
	if mustString(t, got) != want {
		t.Fatalf("got %q want %q", mustString(t, got), want)
	}
}

func TestBindStrcmpNull(t *testing.T) {
	t.Parallel()

	if v, _ := BindStrcmp(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindStrcmp(value.StringValue("x"), nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindInetNtoaNull(t *testing.T) {
	t.Parallel()

	if v, _ := BindInetNtoa(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindInetNtoa(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindIsIPv4MappedNull(t *testing.T) {
	t.Parallel()

	if v, _ := BindIsIPv4Mapped(nil); v != nil {
		t.Fatal("expected null")
	}

	// Non-mapped 16-byte input returns false.
	nonMapped := value.BytesValue{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	got, err := BindIsIPv4Mapped(nonMapped)
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false")
	}
}

func TestBindIsUUIDNull(t *testing.T) {
	t.Parallel()
	if v, _ := BindIsUUID(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindFromDaysWrongArgCount(t *testing.T) {
	t.Parallel()
	if _, err := BindFromDays(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindToDaysWrongArgCount(t *testing.T) {
	t.Parallel()
	if _, err := BindToDays(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindIsIPv4Null(t *testing.T) {
	t.Parallel()
	if v, _ := BindIsIPv4(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindIsIPv6(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindIsIPv4CompatNull(t *testing.T) {
	t.Parallel()
	if v, _ := BindIsIPv4Compat(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindInet6AtonNull(t *testing.T) {
	t.Parallel()
	if v, _ := BindInet6Aton(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindInet6Ntoa(nil); v != nil {
		t.Fatal("expected null")
	}

	// Invalid IP string -> null.
	v, err := BindInet6Aton(value.StringValue("not-an-ip"))
	if err != nil || v != nil {
		t.Fatal("expected null")
	}
}

func TestBindUUIDToBinNull(t *testing.T) {
	t.Parallel()
	if v, _ := BindUUIDToBin(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindBinToUUID(nil); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindMakeDateWrongArgCount(t *testing.T) {
	t.Parallel()
	if v, _ := BindMakeDate(nil, value.IntValue(1)); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindLocateOutOfRange(t *testing.T) {
	t.Parallel()

	// pos > len(haystack) -> 0.
	got, err := BindLocate(value.StringValue("a"), value.StringValue("abc"), value.IntValue(99))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 0 {
		t.Fatalf("got %d", mustInt64(t, got))
	}

	// pos = 0 (1-based -> 0) clamps to 0.
	got, err = BindLocate(value.StringValue("b"), value.StringValue("abc"), value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	if mustInt64(t, got) != 2 {
		t.Fatalf("got %d", mustInt64(t, got))
	}
}

func TestBindSubstringIndexEmptyDelim(t *testing.T) {
	t.Parallel()
	got, err := BindSubstringIndex(value.StringValue("hello"), value.StringValue(""), value.IntValue(1))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty")
	}

	// Negative beyond range returns full string.
	got, err = BindSubstringIndex(value.StringValue("a.b"), value.StringValue("."), value.IntValue(-99))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "a.b" {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindFloat64ArrayFromJsonEmpty(t *testing.T) {
	t.Parallel()
	got, err := BindFloat64ArrayFromJson(value.StringValue("[]"))
	if err != nil {
		t.Fatal(err)
	}
	arr, ok := got.(*value.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %T", got)
	}
	if len(arr.Values) != 0 {
		t.Fatal("expected empty")
	}

	// Array with a null element.
	got, err = BindFloat64ArrayFromJson(value.StringValue("[1, null, 3]"))
	if err != nil {
		t.Fatal(err)
	}
	arr, _ = got.(*value.ArrayValue)
	if len(arr.Values) != 3 || arr.Values[1] != nil {
		t.Fatal("expected null middle element")
	}
}

func TestBindBindJsonUnquoteWithUnicodeEscape(t *testing.T) {
	t.Parallel()
	// jsonDecodeString \u branch.
	got, err := BindJsonUnquote(value.StringValue(`"A"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "A" {
		t.Fatalf("got %q", mustString(t, got))
	}

	// \u with non-hex content -> pass-through.
	got, err = BindJsonUnquote(value.StringValue(`"\uZZZZ"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `"\uZZZZ"` {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Trailing backslash -> pass-through.
	got, err = BindJsonUnquote(value.StringValue(`"a\"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `"a\"` {
		t.Fatalf("got %q", mustString(t, got))
	}
}

func TestBindJsonQuoteUnquote(t *testing.T) {
	t.Parallel()

	got, err := BindJsonQuote(value.StringValue(`he said "hi"`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != `"he said \"hi\""` {
		t.Fatalf("got %q", mustString(t, got))
	}

	back, err := BindJsonUnquote(value.StringValue(`"he said \"hi\""`))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, back) != `he said "hi"` {
		t.Fatalf("got %q", mustString(t, back))
	}

	// Unquoted input passes through.
	back, err = BindJsonUnquote(value.StringValue("plain"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, back) != "plain" {
		t.Fatal("expected pass-through")
	}

	// Newline encoding.
	got, err = BindJsonQuote(value.StringValue("a\nb"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), `\n`) {
		t.Fatal("expected newline encoded")
	}

	if v, _ := BindJsonQuote(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindJsonQuote(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindJsonUnquote(); err == nil {
		t.Fatal("expected error")
	}
}
