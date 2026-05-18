package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// HistoryEntry is one executed statement and its outcome, in a form
// that is convenient to persist (browser storage) and to export.
type HistoryEntry struct {
	Statement    string     `json:"statement"`
	SQLiteQuery  string     `json:"sqliteQuery,omitempty"`
	IsQuery      bool       `json:"isQuery"`
	Columns      []string   `json:"columns,omitempty"`
	Rows         [][]string `json:"rows,omitempty"`
	RowsAffected int64      `json:"rowsAffected,omitempty"`
	Error        string     `json:"error,omitempty"`
	ElapsedMs    float64    `json:"elapsedMs"`
	Timestamp    time.Time  `json:"timestamp"`
}

// NewHistoryEntry converts an executed Result into a HistoryEntry,
// flattening every cell to its plain display string.
func NewHistoryEntry(res Result) HistoryEntry {
	e := HistoryEntry{
		Statement:    res.Statement,
		SQLiteQuery:  res.SQLiteQuery,
		IsQuery:      res.IsQuery,
		Columns:      res.Columns,
		RowsAffected: res.RowsAffected,
		ElapsedMs:    float64(res.Elapsed.Microseconds()) / 1000.0,
		Timestamp:    time.Now(),
	}
	if res.Err != nil {
		e.Error = res.Err.Error()
	}
	if len(res.Rows) > 0 {
		e.Rows = make([][]string, len(res.Rows))
		for r, row := range res.Rows {
			cells := make([]string, len(row))
			for c, v := range row {
				cells[c] = formatValue(v)
			}
			e.Rows[r] = cells
		}
	}
	return e
}

// History is an append-only log of executed statements. It is the
// model the wasm Playground persists to the browser and exports.
type History struct {
	Entries []HistoryEntry `json:"entries"`
}

// Add appends res to the history.
func (h *History) Add(res Result) {
	h.Entries = append(h.Entries, NewHistoryEntry(res))
}

// Clear drops every entry.
func (h *History) Clear() { h.Entries = nil }

// Len returns the number of entries.
func (h *History) Len() int { return len(h.Entries) }

// ExportJSON renders the whole history as indented JSON.
func (h *History) ExportJSON() (string, error) {
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ExportSQL renders the history as a runnable SQL script: every
// statement, terminated by a semicolon, with a comment for failures.
func (h *History) ExportSQL() string {
	var b strings.Builder
	for _, e := range h.Entries {
		if e.Error != "" {
			fmt.Fprintf(&b, "-- error: %s\n", e.Error)
		}
		b.WriteString(e.Statement)
		if !strings.HasSuffix(strings.TrimSpace(e.Statement), ";") {
			b.WriteString(";")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ExportMarkdown renders the history as a Markdown document: each
// statement in a fenced block followed by its result table.
func (h *History) ExportMarkdown() string {
	var b strings.Builder
	for i, e := range h.Entries {
		fmt.Fprintf(&b, "## Query %d\n\n", i+1)
		b.WriteString("```sql\n")
		b.WriteString(strings.TrimSpace(e.Statement))
		b.WriteString("\n```\n\n")
		switch {
		case e.Error != "":
			fmt.Fprintf(&b, "Error: `%s`\n\n", e.Error)
		case e.IsQuery:
			b.WriteString(markdownTable(e.Columns, e.Rows))
			b.WriteString("\n")
		default:
			fmt.Fprintf(&b, "%d row(s) affected.\n\n", e.RowsAffected)
		}
	}
	return b.String()
}

// markdownTable renders columns and rows as a GitHub-flavoured
// Markdown table.
func markdownTable(columns []string, rows [][]string) string {
	if len(columns) == 0 {
		return "_no columns_\n"
	}
	var b strings.Builder
	b.WriteString("| " + strings.Join(escapeMarkdownCells(columns), " | ") + " |\n")
	b.WriteString("|" + strings.Repeat(" --- |", len(columns)) + "\n")
	for _, row := range rows {
		b.WriteString("| " + strings.Join(escapeMarkdownCells(row), " | ") + " |\n")
	}
	return b.String()
}

func escapeMarkdownCells(cells []string) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = strings.ReplaceAll(strings.ReplaceAll(c, "|", "\\|"), "\n", " ")
	}
	return out
}

// ExportCSV renders the result sets in the history as CSV. Each query
// entry is preceded by a `# <statement>` comment line and separated
// from the next by a blank line. Non-query entries are skipped.
func (h *History) ExportCSV() (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)
	first := true
	for _, e := range h.Entries {
		if !e.IsQuery || e.Error != "" {
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false
		fmt.Fprintf(&b, "# %s\n", strings.ReplaceAll(e.Statement, "\n", " "))
		if len(e.Columns) > 0 {
			if err := w.Write(e.Columns); err != nil {
				return "", err
			}
		}
		for _, row := range e.Rows {
			if err := w.Write(row); err != nil {
				return "", err
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

// Export renders the history in the named format: "json", "sql",
// "markdown" / "md", or "csv".
func (h *History) Export(format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return h.ExportJSON()
	case "sql":
		return h.ExportSQL(), nil
	case "markdown", "md":
		return h.ExportMarkdown(), nil
	case "csv":
		return h.ExportCSV()
	default:
		return "", fmt.Errorf("unknown export format %q (want json, sql, markdown or csv)", format)
	}
}
