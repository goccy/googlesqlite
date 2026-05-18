package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	googlesql "github.com/goccy/go-googlesql"
)

// nullText is the rendered form of a SQL NULL value.
const nullText = "NULL"

// formatValue renders a scanned cell value as its plain display
// string, without any ANSI colour codes. Every display-width
// calculation in the renderer must run on this plain form so that
// colour escape bytes never inflate a column width — the root cause
// of the CJK column-misalignment bug in the predecessor CLI.
func formatValue(v any) string {
	switch x := v.(type) {
	case nil:
		return nullText
	case bool:
		if x {
			return "true"
		}
		return "false"
	case string:
		return x
	case []byte:
		return string(x)
	case []any:
		// ARRAY and STRUCT values both scan to []any; both render with
		// bracket notation, which is unambiguous enough for a console.
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = formatValue(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprint(v)
	}
}

// nullColor is the colour applied to NULL cells.
var nullColor = color.New(color.FgRed)

// colorForKind maps a GoogleSQL column type kind to its display
// colour. The mapping mirrors common SQL-console conventions: numbers
// magenta, floats yellow, strings green, temporal types blue.
func colorForKind(kind googlesql.TypeKind) *color.Color {
	switch kind {
	case googlesql.TypeKindTypeInt32,
		googlesql.TypeKindTypeInt64,
		googlesql.TypeKindTypeUint32,
		googlesql.TypeKindTypeUint64,
		googlesql.TypeKindTypeNumeric,
		googlesql.TypeKindTypeBignumeric:
		return color.New(color.FgMagenta)
	case googlesql.TypeKindTypeBool:
		return color.New(color.FgCyan)
	case googlesql.TypeKindTypeFloat,
		googlesql.TypeKindTypeDouble:
		return color.New(color.FgYellow)
	case googlesql.TypeKindTypeString,
		googlesql.TypeKindTypeBytes:
		return color.New(color.FgGreen)
	case googlesql.TypeKindTypeDate,
		googlesql.TypeKindTypeDatetime,
		googlesql.TypeKindTypeTime,
		googlesql.TypeKindTypeTimestamp,
		googlesql.TypeKindTypeInterval:
		return color.New(color.FgHiBlue)
	default:
		return color.New(color.FgWhite)
	}
}

// colorizeCell returns the coloured form of a cell's plain text. raw
// is the original scanned value (used to detect NULL); kind is the
// column's type kind. The returned string has the same display width
// as plain — only ANSI escapes are added.
func colorizeCell(raw any, plain string, kind googlesql.TypeKind) string {
	if raw == nil {
		return nullColor.Sprint(plain)
	}
	return colorForKind(kind).Sprint(plain)
}
