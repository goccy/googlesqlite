package compliancetest

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ExpectedKind classifies the top-level shape of the expected block.
type ExpectedKind int

const (
	// ExpectedRows is the standard ARRAY<STRUCT<...>>[...] form.
	ExpectedRows ExpectedKind = iota
	// ExpectedError is `ERROR: ...`.
	ExpectedError
	// ExpectedUnknown is anything else (e.g., a scalar value notation
	// that does not start with ARRAY<STRUCT>; we don't try to parse).
	ExpectedUnknown
)

// ExpectedRowSet is the parsed form of an `ARRAY<STRUCT<...>>[...]`
// block. Rows[i][j] holds one cell value; the cell is one of:
//   - nil (NULL)
//   - bool
//   - int64
//   - float64
//   - string (already-unquoted contents of a "..." literal)
//   - rawValue (kept verbatim for non-scalar shapes — caller decides
//     whether the case is convertible to rows-equality assertion)
type ExpectedRowSet struct {
	Rows [][]any
}

// rawValue holds the original text for cell shapes the simple
// renderer leaves unhandled (arrays, structs, ranges, byte literals
// with hex escapes, etc.). Tests that involve a rawValue cell are not
// safely translatable to our YAML rows form and should be skipped.
type rawValue string

// ClassifyExpected returns the top-level kind of the expected block.
func ClassifyExpected(s string) ExpectedKind {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(s, "ERROR:"):
		return ExpectedError
	case strings.HasPrefix(s, "ARRAY<STRUCT"):
		return ExpectedRows
	}
	return ExpectedUnknown
}

// ParseExpectedRows extracts the row list from an
// `ARRAY<STRUCT<...>>[{...}, {...}]` block. Any cell whose value
// notation is more complex than a top-level scalar / NULL / bool
// becomes a rawValue, signalling to the caller that this case is not
// safe to convert to a YAML rows-equality assertion.
func ParseExpectedRows(s string) (ExpectedRowSet, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "ARRAY<STRUCT") {
		return ExpectedRowSet{}, errors.New("not an ARRAY<STRUCT<...>> block")
	}
	// Skip the type prefix, locate the `[` that opens the row list.
	bracket := strings.Index(s, "[")
	if bracket < 0 {
		return ExpectedRowSet{}, errors.New("missing [ in row list")
	}
	tail := s[bracket+1:]
	tail = strings.TrimSpace(tail)
	if strings.HasSuffix(tail, "]") {
		tail = strings.TrimSuffix(tail, "]")
	} else {
		return ExpectedRowSet{}, errors.New("missing ] at end of row list")
	}

	rows, err := splitTopLevel(tail, '{', '}')
	if err != nil {
		return ExpectedRowSet{}, err
	}
	var out ExpectedRowSet
	for _, rowBody := range rows {
		rowBody = strings.TrimSpace(rowBody)
		if !strings.HasPrefix(rowBody, "{") || !strings.HasSuffix(rowBody, "}") {
			return ExpectedRowSet{}, fmt.Errorf("malformed row: %q", rowBody)
		}
		inner := strings.TrimSpace(rowBody[1 : len(rowBody)-1])
		cells, err := splitCells(inner)
		if err != nil {
			return ExpectedRowSet{}, err
		}
		parsed := make([]any, len(cells))
		for i, raw := range cells {
			parsed[i] = parseScalarCell(raw)
		}
		out.Rows = append(out.Rows, parsed)
	}
	return out, nil
}

// splitTopLevel walks s and emits the top-level groups delimited by
// `open`...`close` braces. Anything between groups (commas, whitespace)
// is dropped. Strings ("...") are skipped to avoid being misled by
// brace characters embedded in literals.
func splitTopLevel(s string, open, close rune) ([]string, error) {
	var out []string
	depth := 0
	start := -1
	inStr := false
	var quote rune
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inStr {
			if r == '\\' && i+1 < len(runes) {
				i++
				continue
			}
			if r == quote {
				inStr = false
			}
			continue
		}
		if r == '"' || r == '\'' {
			inStr = true
			quote = r
			continue
		}
		switch r {
		case open:
			if depth == 0 {
				start = i
			}
			depth++
		case close:
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("unbalanced %c at offset %d", close, i)
			}
			if depth == 0 && start >= 0 {
				out = append(out, string(runes[start:i+1]))
				start = -1
			}
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("unbalanced %c%c", open, close)
	}
	return out, nil
}

// splitCells splits the content of a `{...}` row into cell strings on
// top-level commas. Strings, arrays, and nested structs are honoured.
func splitCells(s string) ([]string, error) {
	var out []string
	depth := 0
	start := 0
	inStr := false
	var quote rune
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inStr {
			if r == '\\' && i+1 < len(runes) {
				i++
				continue
			}
			if r == quote {
				inStr = false
			}
			continue
		}
		if r == '"' || r == '\'' {
			inStr = true
			quote = r
			continue
		}
		switch r {
		case '[', '{', '<':
			depth++
		case ']', '}', '>':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(string(runes[start:i])))
				start = i + 1
			}
		}
	}
	tail := strings.TrimSpace(string(runes[start:]))
	if tail != "" {
		out = append(out, tail)
	}
	if depth != 0 {
		return nil, errors.New("unbalanced delimiter in cell list")
	}
	return out, nil
}

// parseScalarCell turns one cell text into a typed value, or
// rawValue when the shape isn't a scalar.
func parseScalarCell(s string) any {
	s = strings.TrimSpace(s)
	switch s {
	case "NULL", "null":
		return nil
	case "true", "TRUE":
		return true
	case "false", "FALSE":
		return false
	case "nan", "NaN", "NAN":
		// keep as raw so callers know to skip
		return rawValue(s)
	case "inf", "Inf", "INF", "-inf", "-Inf", "-INF":
		return rawValue(s)
	}
	// Quoted string literal.
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		raw := s[1 : len(s)-1]
		unq, err := unquoteString(raw)
		if err == nil {
			return unq
		}
		return rawValue(s)
	}
	// Integer.
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	// Float.
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// Hex byte literal b"\x00..." etc., array `[..]`, struct `{..}`,
	// range, etc. — keep raw.
	return rawValue(s)
}

// unquoteString decodes the common GoogleSQL escape sequences inside
// a quoted string literal.
func unquoteString(s string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		if i+1 >= len(s) {
			return "", errors.New("dangling backslash")
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '"':
			b.WriteByte('"')
		case '\'':
			b.WriteByte('\'')
		case '\\':
			b.WriteByte('\\')
		case '0':
			b.WriteByte(0)
		case 'x':
			if i+2 >= len(s) {
				return "", errors.New("\\x needs two hex digits")
			}
			n, err := strconv.ParseUint(s[i+1:i+3], 16, 8)
			if err != nil {
				return "", err
			}
			b.WriteByte(byte(n))
			i += 2
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String(), nil
}

// ConvertibleToYAML returns true when every cell of every row is a
// scalar that the testdata YAML round-trips cleanly: int64, float64,
// bool, string, nil. rawValue cells (arrays, structs, NaN/Inf, byte
// blobs, etc.) make the case unsafe to assert with simple row
// equality.
func ConvertibleToYAML(rs ExpectedRowSet) bool {
	for _, row := range rs.Rows {
		for _, cell := range row {
			switch cell.(type) {
			case nil, bool, int64, float64, string:
			default:
				return false
			}
		}
	}
	return true
}
