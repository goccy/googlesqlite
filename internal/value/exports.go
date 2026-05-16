package value

import "time"

// Exported wrappers for the package-internal parsers / classifiers /
// helpers that the rest of internal/ used to call as un-prefixed
// identifiers.

func ParseDate(s string) (time.Time, error)     { return parseDate(s) }
func ParseDatetime(s string) (time.Time, error) { return parseDatetime(s) }
func ParseTime(s string) (time.Time, error)     { return parseTime(s) }
func ParseTimestamp(s string, loc *time.Location) (time.Time, error) {
	return parseTimestamp(s, loc)
}
func ParseInterval(s string) (*IntervalValue, error) { return parseInterval(s) }
func IsDate(s string) bool                           { return isDate(s) }
func IsDatetime(s string) bool                       { return isDatetime(s) }
func IsTime(s string) bool                           { return isTime(s) }
func IsTimestamp(s string) bool                      { return isTimestamp(s) }
func IsNullValue(v any) bool                         { return isNullValue(v) }

func ToLocation(timeZone string) (*time.Location, error) { return toLocation(timeZone) }
func ModifyTimeZone(t time.Time, loc *time.Location) (time.Time, error) {
	return modifyTimeZone(t, loc)
}

func TimeFromUnixNano(n int64) time.Time { return timeFromUnixNano(n) }

// Helpers used by encoder/decoder when constructing literals.
func CoerceIntsToFloats(v any) any { return coerceIntsToFloats(v) }
func PrintableChar(b byte) bool    { return printableChar(b) }
