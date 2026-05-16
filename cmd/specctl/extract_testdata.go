package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/googlesqlite/internal/specmeta"
)

func init() {
	register(command{
		name:    "extract-testdata",
		summary: "auto-generate testdata YAML from each spec's upstream Examples section",
		run: func(_ context.Context, args []string) error {
			return runExtractTestdata(args)
		},
	})
}

// runExtractTestdata scans every spec markdown under docs/specs/ for
// its upstream Examples section (the verbatim Apache-2.0 derivative
// pulled in by `specctl normalize`) and writes a testdata YAML file
// at the spec's declared testdata path. Each (SQL, ASCII-table)
// example pair becomes one case.
//
// The Examples section is the **only** authoritative source for
// expected values. The tool will NOT write a testdata case when the
// spec has no upstream Examples — those specs must be downgraded
// from `implemented` because we have no source of truth for what
// their output should be.
//
// Existing testdata YAML files are NOT overwritten unless --force is
// supplied. --only-missing skips specs that already have a testdata
// file on disk. --dry-run prints what would be written without
// changing files.
func runExtractTestdata(args []string) error {
	force := false
	dryRun := false
	onlyMissing := false
	for _, a := range args {
		switch a {
		case "--force":
			force = true
		case "--dry-run":
			dryRun = true
		case "--only-missing":
			onlyMissing = true
		}
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	specs, err := specmeta.LoadAllSpecs(root)
	if err != nil {
		return err
	}
	written := 0
	skipped := 0
	noExamples := 0
	overwritten := 0
	for _, sp := range specs {
		td := strings.TrimSpace(sp.Frontmatter.Testdata)
		if td == "" {
			continue
		}
		full := filepath.Join(root, td)
		exists := false
		if _, err := os.Stat(full); err == nil {
			exists = true
		}
		if onlyMissing && exists {
			continue
		}
		if exists && !force {
			skipped++
			continue
		}
		cases, err := extractExamplesFromSpec(filepath.Join(root, sp.Path))
		if err != nil {
			return fmt.Errorf("%s: %w", sp.Path, err)
		}
		if len(cases) == 0 {
			noExamples++
			fmt.Fprintf(os.Stderr, "no-examples: %s\n", sp.Path)
			continue
		}
		yaml := renderTestdataYAML(sp, cases)
		if dryRun {
			fmt.Fprintf(os.Stderr, "would-write: %s (%d case(s))\n", td, len(cases))
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(yaml), 0o644); err != nil {
			return err
		}
		if exists {
			overwritten++
		} else {
			written++
		}
	}
	fmt.Fprintf(os.Stderr, "extract-testdata: %d written, %d overwritten, %d skipped (exists), %d no-examples\n",
		written, overwritten, skipped, noExamples)
	return nil
}

// extractExamplesFromSpec reads the spec md file, isolates the
// `## Reference (upstream)` section (which holds the Apache-2.0
// derivative of the upstream docs), finds each `**Examples**`
// subsection, and pulls out each (SQL, ASCII-table) example pair.
func extractExamplesFromSpec(path string) ([]testdataCase, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	upstream := isolateUpstreamSection(string(body))
	if upstream == "" {
		return nil, nil
	}
	var all []testdataCase
	// A single spec md can carry several `## <FUNCTION_NAME>` sections
	// (e.g., compound types/conditional helpers). Iterate every
	// Examples subsection.
	for _, examples := range allExampleBodies(upstream) {
		all = append(all, parseExampleBlocks(examples)...)
	}
	return all, nil
}

// isolateUpstreamSection returns the substring after the
// `## Reference (upstream)` header.
func isolateUpstreamSection(body string) string {
	idx := strings.Index(body, "## Reference (upstream)")
	if idx < 0 {
		return ""
	}
	return body[idx:]
}

// allExampleBodies pulls every `**Example**` / `**Examples**`
// subsection out of the upstream block. Each subsection runs from
// the marker to the next bold marker or top-level heading.
//
// Upstream docs are inconsistent — string_functions.md uses
// `**Examples**` while debugging_functions.md uses `**Example**`
// (singular), and a single spec may carry multiple of either.
func allExampleBodies(s string) []string {
	var out []string
	exampleHeading := exampleHeadingRe
	rest := s
	for {
		loc := exampleHeading.FindStringIndex(rest)
		if loc == nil {
			return out
		}
		tail := rest[loc[1]:]
		stopMarkers := []string{
			"\n**Description**",
			"\n**Return type**",
			"\n**Stipulations**",
			"\n**Definitions**",
			"\n**Constraints**",
			"\n**Arguments**",
			"\n**Example**",
			"\n**Examples**",
			"\n## ",
		}
		endIdx := len(tail)
		for _, m := range stopMarkers {
			if i := strings.Index(tail, m); i >= 0 && i < endIdx {
				endIdx = i
			}
		}
		out = append(out, tail[:endIdx])
		rest = tail[endIdx:]
	}
}

// exampleHeadingRe matches both **Example** and **Examples** as
// markdown bold headings used by upstream googlesql docs.
var exampleHeadingRe = regexp.MustCompile(`\*\*Examples?\*\*`)

// parseExampleBlocks walks one Examples body and pulls out each
// ```googlesql / ```sql code block. Each block typically holds the
// SQL followed by an ASCII-table comment with the expected output.
func parseExampleBlocks(s string) []testdataCase {
	var cases []testdataCase
	blocks := codeBlockRe.FindAllStringSubmatch(s, -1)
	for _, m := range blocks {
		if len(m) < 2 {
			continue
		}
		block := m[1]
		sql, rows, ok := splitSQLAndExpected(block)
		if !ok {
			continue
		}
		cases = append(cases, testdataCase{SQL: sql, Rows: rows})
	}
	return cases
}

var codeBlockRe = regexp.MustCompile("(?s)```(?:googlesql|sql)\\s*\\n(.*?)```")

// splitSQLAndExpected breaks a single code block into (SQL,
// expectedRows). The SQL is everything up to the first ASCII-table
// comment (`/*--`); rows are parsed from the comment.
func splitSQLAndExpected(block string) (string, [][]string, bool) {
	commentStart := strings.Index(block, "/*--")
	if commentStart < 0 {
		// Some upstream tables begin with `/*-+` (column-border-only).
		commentStart = strings.Index(block, "/*-+")
		if commentStart < 0 {
			return "", nil, false
		}
	}
	sql := strings.TrimSpace(block[:commentStart])
	sql = strings.TrimSuffix(sql, ";")
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "", nil, false
	}
	commentEnd := strings.Index(block[commentStart:], "*/")
	if commentEnd < 0 {
		return "", nil, false
	}
	table := block[commentStart : commentStart+commentEnd]
	rows := parseAsciiTable(table)
	if len(rows) == 0 {
		return "", nil, false
	}
	return sql, rows, true
}

// parseAsciiTable extracts row values from an upstream ASCII-table
// comment. Each data row is a line wrapped in pipes between border
// lines; the first non-border row is the header; the rest are
// values.
func parseAsciiTable(s string) [][]string {
	lines := strings.Split(s, "\n")
	var dataRows [][]string
	for _, line := range lines {
		l := strings.TrimSpace(line)
		l = strings.TrimPrefix(l, "/*")
		l = strings.TrimSuffix(l, "*/")
		l = strings.TrimSpace(l)
		if !strings.HasPrefix(l, "|") || !strings.HasSuffix(l, "|") {
			continue
		}
		inner := l[1 : len(l)-1]
		// Skip border-only rows like "----+----".
		hasAlnum := false
		for _, r := range inner {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				hasAlnum = true
				break
			}
		}
		if !hasAlnum {
			continue
		}
		cells := strings.Split(inner, "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		for len(cells) > 0 && cells[len(cells)-1] == "" {
			cells = cells[:len(cells)-1]
		}
		if len(cells) == 0 {
			continue
		}
		dataRows = append(dataRows, cells)
	}
	if len(dataRows) <= 1 {
		return nil
	}
	return dataRows[1:]
}

// renderTestdataYAML produces the testdata YAML body for a spec's
// extracted examples.
func renderTestdataYAML(sp specmeta.Spec, cases []testdataCase) string {
	dialect := sp.Frontmatter.Dialect
	if dialect == "" {
		dialect = "googlesql"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "spec: %s\n", sp.Path)
	fmt.Fprintf(&b, "dialect: %s\n", dialect)
	fmt.Fprintf(&b, "# Auto-generated from the spec's upstream Examples section.\n")
	fmt.Fprintf(&b, "# Source: %s\n", sp.Frontmatter.SourceURL)
	fmt.Fprintf(&b, "cases:\n")
	for i, c := range cases {
		fmt.Fprintf(&b, "  - desc: upstream example %d\n", i+1)
		b.WriteString("    sql: |-\n")
		for _, line := range strings.Split(c.SQL, "\n") {
			b.WriteString("      ")
			b.WriteString(line)
			b.WriteByte('\n')
		}
		b.WriteString("    expected:\n")
		b.WriteString("      rows:\n")
		for _, row := range c.Rows {
			b.WriteString("        - [")
			for j, cell := range row {
				if j > 0 {
					b.WriteString(", ")
				}
				b.WriteString(quoteCell(cell))
			}
			b.WriteString("]\n")
		}
	}
	return b.String()
}

func quoteCell(s string) string {
	if s == "NULL" || s == "null" {
		return "null"
	}
	if s == "true" || s == "TRUE" {
		return "true"
	}
	if s == "false" || s == "FALSE" {
		return "false"
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

type testdataCase struct {
	SQL  string
	Rows [][]string
}
