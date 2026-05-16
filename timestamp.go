package googlesqlite

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeFromTimestampValue converts a googlesqlite TIMESTAMP cell value back
// to time.Time. The driver's canonical scan form for TIMESTAMP is
// "YYYY-MM-DD HH:MM:SS[.ffffff]+00" — those land in the canonical branch
// below. The function also accepts the predecessor's "<unix-secs>[.frac]"
// epoch-float spelling so callers carrying old testdata or external
// integrations relying on that shape keep working.
func TimeFromTimestampValue(v string) (time.Time, error) {
	if v == "" {
		return time.Time{}, fmt.Errorf("empty timestamp string")
	}
	// Canonical form has a colon (time separator) or 'T' (RFC3339).
	// Epoch-float form is digits, optional leading '-' and a single '.',
	// so it can never contain those.
	if strings.ContainsAny(v, ":T") {
		return parseCanonicalTimestamp(v)
	}
	return parseEpochFloatTimestamp(v)
}

// parseCanonicalTimestamp accepts the formats emitted by
// internal.formatTimestampCanonical and value_range.go: an ISO-like form
// in UTC with optional fractional seconds and an optional `+00` /
// numeric / "Z" offset.
func parseCanonicalTimestamp(v string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02 15:04:05.999999999+00",
		"2006-01-02 15:04:05+00",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid timestamp string %q", v)
}

// parseEpochFloatTimestamp keeps the predecessor's encoding alive for
// callers that still receive it (e.g. older serialised testdata).
func parseEpochFloatTimestamp(v string) (time.Time, error) {
	// ParseFloat is too imprecise to use, instead split into seconds and fractional seconds
	parts := strings.Split(v, ".")
	if len(parts) > 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp string (multiple delimiters) %s", v)
	}
	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	micros := int64(0)
	if len(parts) == 2 {
		// Pad fractional places to microseconds (".1" → "100000" micros).
		// The pre-fix version constructed `microsString` but parsed
		// `parts[1]`, so ".1" became 1 microsecond instead of 0.1
		// second.
		microsString := parts[1]
		for len(microsString) < 6 {
			microsString += "0"
		}
		m, err := strconv.ParseInt(microsString, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		micros = m
	}
	nanos := micros * int64(time.Microsecond)
	return time.Unix(seconds, nanos), err
}
