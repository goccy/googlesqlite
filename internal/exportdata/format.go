package exportdata

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format is the EXPORT DATA output format selected by the `format` option.
type Format string

const (
	FormatCSV    Format = "CSV"
	FormatNDJSON Format = "NEWLINE_DELIMITED_JSON"
)

// ParseFormat normalizes the `format` option value. An empty string maps to
// CSV (matching real BigQuery's default). Formats that BigQuery accepts but
// the emulator does not yet implement (AVRO, PARQUET) return a descriptive
// error rather than silently falling back, so callers see the gap instead
// of a corrupt file.
func ParseFormat(s string) (Format, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "", "CSV":
		return FormatCSV, nil
	case "JSON", "NEWLINE_DELIMITED_JSON":
		return FormatNDJSON, nil
	case "AVRO", "PARQUET":
		return "", fmt.Errorf("EXPORT DATA: format %q is not yet supported by googlesqlite", s)
	default:
		return "", fmt.Errorf("EXPORT DATA: unknown format %q", s)
	}
}

// RowSource yields successive rows of an EXPORT DATA output. Each call must
// return the next row's values in the order matching the column list passed
// to EncodeRows. The bool result is true while more rows remain and false
// once iteration is exhausted; an error short-circuits encoding.
type RowSource func() (values []any, hasMore bool, err error)

// EncodeRows streams rows from src into w using the chosen format. CSV
// writes a header row of column names; NDJSON writes one JSON object per
// row keyed by column name.
func EncodeRows(w io.Writer, format Format, columns []string, src RowSource) error {
	switch format {
	case FormatCSV:
		return encodeCSV(w, columns, src)
	case FormatNDJSON:
		return encodeNDJSON(w, columns, src)
	}
	return fmt.Errorf("EXPORT DATA: unsupported format %q", format)
}

func encodeCSV(w io.Writer, columns []string, src RowSource) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(columns); err != nil {
		return fmt.Errorf("EXPORT DATA: write CSV header: %w", err)
	}
	for {
		values, hasMore, err := src()
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
		record := make([]string, len(values))
		for i, v := range values {
			record[i] = csvString(v)
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("EXPORT DATA: write CSV row: %w", err)
		}
	}
	cw.Flush()
	return cw.Error()
}

func encodeNDJSON(w io.Writer, columns []string, src RowSource) error {
	enc := json.NewEncoder(w)
	for {
		values, hasMore, err := src()
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
		obj := make(map[string]any, len(columns))
		for i, name := range columns {
			if i < len(values) {
				obj[name] = values[i]
			}
		}
		if err := enc.Encode(obj); err != nil {
			return fmt.Errorf("EXPORT DATA: write JSON row: %w", err)
		}
	}
	return nil
}

// csvString renders a Go value to a CSV field. Strings pass through as-is,
// nil becomes an empty field, and everything else is JSON-marshaled to a
// stable textual form (matching how the BigQuery emulator's existing
// extract-job CSV encoder renders non-scalar values).
func csvString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	s := string(b)
	// json.Marshal wraps a string in quotes; for scalars rendered as JSON
	// strings (rare here since we already special-cased string above), drop
	// the surrounding quotes so the CSV cell carries the value, not a JSON
	// literal.
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		var unquoted string
		if json.Unmarshal(b, &unquoted) == nil {
			return unquoted
		}
	}
	return s
}
