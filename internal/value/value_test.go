package value

import (
	"testing"
	"time"
)

func formatTimestamp(s string) (string, error) {
	loc, err := time.LoadLocation("")
	if err != nil {
		return "", err
	}
	t, err := parseTimestamp(s, loc)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC3339Nano), nil
}

func TestTimestampValue(t *testing.T) {
	if !datetimeRe.MatchString("2022-01-01 00:00:00") {
		t.Fatalf("mismatch timestamp value")
	}
	formatted, err := formatTimestamp("2022-01-01 00:00:00")
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "2022-01-01T00:00:00Z" {
		t.Fatalf("failed to format timestamp")
	}
}

// The canonical literal parsers back the STRING -> DATE / DATETIME / TIME
// cast, so each one must accept its own literal -- including the
// one-digit component forms -- and reject every other temporal shape.
func TestParseLiterals(t *testing.T) {
	tests := []struct {
		name  string
		parse func(string) (time.Time, error)
		in    string
		want  string // RFC3339Nano, or "" when the parse must fail
	}{
		{"date", parseDateLiteral, "2024-01-15", "2024-01-15T00:00:00Z"},
		{"date one-digit month and day", parseDateLiteral, "2024-1-5", "2024-01-05T00:00:00Z"},
		{"date with time", parseDateLiteral, "2024-01-15 00:00:00", ""},
		{"date with zone", parseDateLiteral, "2024-01-15T00:00:00Z", ""},
		{"datetime space separator", parseDatetimeLiteral, "2024-01-15 10:30:00", "2024-01-15T10:30:00Z"},
		{"datetime T separator", parseDatetimeLiteral, "2024-01-15T10:30:00", "2024-01-15T10:30:00Z"},
		{"datetime t separator", parseDatetimeLiteral, "2024-01-15t10:30:00", "2024-01-15T10:30:00Z"},
		{"datetime fractional seconds", parseDatetimeLiteral, "2024-01-15 10:30:00.123", "2024-01-15T10:30:00.123Z"},
		{"datetime without time part", parseDatetimeLiteral, "2024-01-15", "2024-01-15T00:00:00Z"},
		{"datetime with zone", parseDatetimeLiteral, "2024-01-15 10:30:00+00", ""},
		{"time", parseTimeLiteral, "10:30:00", "0000-01-01T10:30:00Z"},
		{"time one-digit components", parseTimeLiteral, "1:2:3", "0000-01-01T01:02:03Z"},
		{"time fractional seconds", parseTimeLiteral, "10:30:00.123456", "0000-01-01T10:30:00.123456Z"},
		{"time with date", parseTimeLiteral, "2024-01-15 10:30:00", ""},
		{"time with zone", parseTimeLiteral, "10:30:00Z", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parse(tt.in)
			if tt.want == "" {
				if err == nil {
					t.Fatalf("parsed %q as %s, want a parse error", tt.in, got.Format(time.RFC3339Nano))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if formatted := got.Format(time.RFC3339Nano); formatted != tt.want {
				t.Fatalf("parsed %q as %s, want %s", tt.in, formatted, tt.want)
			}
		})
	}
}
