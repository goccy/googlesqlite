package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/mattn/go-runewidth"
)

// displayWidth returns the terminal display width of s, counting
// full-width (CJK) runes as 2 columns. All column-alignment maths in
// this file goes through it.
func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

// RenderOptions controls how a Result is rendered.
type RenderOptions struct {
	// Color enables ANSI colourisation of values.
	Color bool
	// Debug, when true, prints the translated SQLite query above the
	// result.
	Debug bool
}

// RenderResult writes a Result to w: an error line, an affected-rows
// line for non-queries, or a bordered table / \G group view for
// queries, followed by a row-count summary.
func RenderResult(w io.Writer, res Result, opt RenderOptions) {
	if opt.Debug && res.SQLiteQuery != "" {
		for _, line := range strings.Split(res.SQLiteQuery, "\n") {
			fmt.Fprintf(w, "-- sqlite: %s\n", line)
		}
	}
	if res.Err != nil {
		fmt.Fprintf(w, "ERROR: %v\n", res.Err)
		return
	}
	if !res.IsQuery {
		fmt.Fprintf(w, "Query OK, %d row(s) affected (%s)\n",
			res.RowsAffected, formatDuration(res.Elapsed))
		return
	}
	if res.GroupMode {
		renderGroup(w, res, opt.Color)
	} else {
		renderTable(w, res, opt.Color)
	}
	fmt.Fprintf(w, "%d row(s) in set (%s)\n",
		len(res.Rows), formatDuration(res.Elapsed))
}

// renderTable writes res as a bordered, display-width-aligned table.
// Column widths are computed on the plain (un-coloured) cell text via
// runewidth so that full-width CJK characters and ANSI colour codes
// never skew the alignment.
func renderTable(w io.Writer, res Result, useColor bool) {
	ncol := len(res.Columns)
	if ncol == 0 {
		fmt.Fprintln(w, "Empty set")
		return
	}

	plain := make([][]string, len(res.Rows))
	for r, row := range res.Rows {
		plain[r] = make([]string, ncol)
		for c := 0; c < ncol; c++ {
			if c < len(row) {
				plain[r][c] = formatValue(row[c])
			}
		}
	}

	widths := make([]int, ncol)
	for c, name := range res.Columns {
		widths[c] = displayWidth(name)
	}
	for r := range plain {
		for c := 0; c < ncol; c++ {
			if cw := displayWidth(plain[r][c]); cw > widths[c] {
				widths[c] = cw
			}
		}
	}

	border := tableBorder(widths)
	fmt.Fprintln(w, border)
	fmt.Fprintln(w, tableLine(res.Columns, widths, nil, nil, false))
	fmt.Fprintln(w, border)
	for r := range plain {
		fmt.Fprintln(w, tableLine(plain[r], widths, res.Rows[r], res.ColumnKinds, useColor))
	}
	fmt.Fprintln(w, border)
}

// tableLine builds one "| a | b |" row. Padding is computed from the
// plain cell width; colour, when enabled, is applied to the cell text
// only after the gap is measured.
func tableLine(cells []string, widths []int, raw []any, kinds []googlesql.TypeKind, useColor bool) string {
	var b strings.Builder
	b.WriteString("|")
	for c, cell := range cells {
		gap := widths[c] - displayWidth(cell)
		if gap < 0 {
			gap = 0
		}
		shown := cell
		if useColor {
			var rawVal any
			if raw != nil && c < len(raw) {
				rawVal = raw[c]
			}
			kind := googlesql.TypeKindTypeUnknown
			if kinds != nil && c < len(kinds) {
				kind = kinds[c]
			}
			shown = colorizeCell(rawVal, cell, kind)
		}
		b.WriteString(" ")
		b.WriteString(shown)
		b.WriteString(strings.Repeat(" ", gap))
		b.WriteString(" |")
	}
	return b.String()
}

// tableBorder builds a "+----+----+" separator line for the given
// column widths.
func tableBorder(widths []int) string {
	var b strings.Builder
	b.WriteString("+")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w+2))
		b.WriteString("+")
	}
	return b.String()
}

// renderGroup writes res in vertical \G form: one stanza per row, the
// column names right-aligned to the widest name.
func renderGroup(w io.Writer, res Result, useColor bool) {
	nameWidth := 0
	for _, name := range res.Columns {
		if cw := displayWidth(name); cw > nameWidth {
			nameWidth = cw
		}
	}
	for r, row := range res.Rows {
		fmt.Fprintf(w, "*************************** %d. row ***************************\n", r+1)
		for c, name := range res.Columns {
			var rawVal any
			if c < len(row) {
				rawVal = row[c]
			}
			plain := formatValue(rawVal)
			shown := plain
			if useColor {
				kind := googlesql.TypeKindTypeUnknown
				if c < len(res.ColumnKinds) {
					kind = res.ColumnKinds[c]
				}
				shown = colorizeCell(rawVal, plain, kind)
			}
			pad := strings.Repeat(" ", nameWidth-displayWidth(name))
			fmt.Fprintf(w, "%s%s: %s\n", pad, name, shown)
		}
	}
}

// formatDuration renders an execution duration compactly.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return d.String()
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	default:
		return d.Round(time.Microsecond).String()
	}
}
