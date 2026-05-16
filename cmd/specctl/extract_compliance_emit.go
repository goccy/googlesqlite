package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/googlesqlite/cmd/specctl/compliancetest"
	"github.com/goccy/googlesqlite/internal/specmeta"
)

func init() {
	register(command{
		name:    "extract-compliance",
		summary: "render candidate testdata YAML for one spec from compliance .test cases",
		run: func(_ context.Context, args []string) error {
			return runExtractCompliance(args)
		},
	})
}

// runExtractCompliance prints a candidate testdata YAML body for a
// single spec, sourced from one or more compliance .test files. It
// does NOT write to disk on its own — the operator pipes the output
// into the spec's testdata path after a human review pass. That keeps
// the "authoritative source only" guarantee: we never overwrite an
// existing testdata YAML with cases we have not visually confirmed
// against the upstream fixture.
//
// Usage:
//
//	specctl extract-compliance --root <path> --spec <docs/specs/.../foo.md> [--limit N]
//
// The function name is read from the spec's frontmatter and used to
// pick matching cases. Only cases that ConvertibleToYAML returns true
// for are emitted; everything else is summarised on stderr so the
// operator can fall back to manual extraction.
func runExtractCompliance(args []string) error {
	root := ""
	specPath := ""
	limit := 25
	selfContained := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--root":
			if i+1 < len(args) {
				root = args[i+1]
				i++
			}
		case "--spec":
			if i+1 < len(args) {
				specPath = args[i+1]
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				if n, err := atoi(args[i+1]); err == nil && n > 0 {
					limit = n
				}
				i++
			}
		case "--self-contained":
			// Only emit cases that don't reference external tables
			// (`FROM <name>` other than `FROM UNNEST(...)` or
			// `FROM (SELECT ...)`). Useful for sourcing testdata that
			// will run against a fresh :memory: connection with no
			// schema setup.
			selfContained = true
		}
	}
	if root == "" {
		return fmt.Errorf("--root is required (path to compliance/testdata directory)")
	}
	if specPath == "" {
		return fmt.Errorf("--spec is required (path to docs/specs/.../<name>.md)")
	}

	projRoot, err := projectRoot()
	if err != nil {
		return err
	}
	specAbs := specPath
	if !filepath.IsAbs(specAbs) {
		specAbs = filepath.Join(projRoot, specPath)
	}
	data, err := os.ReadFile(specAbs)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}
	fm, err := specmeta.ParseFrontmatter(data)
	if err != nil {
		return fmt.Errorf("parse frontmatter: %w", err)
	}
	fnName := strings.ToUpper(strings.TrimSpace(fm.Name))
	if fnName == "" {
		return fmt.Errorf("spec frontmatter has no `name:`")
	}

	// Walk the compliance directory and collect matching cases.
	var matches []compliancetest.Case
	skipped := 0
	unsafe := 0
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".test" {
			return nil
		}
		cases, perr := compliancetest.ParseFile(path)
		if perr != nil {
			return nil
		}
		for _, c := range cases {
			if !mentionsFunction(c.SQL, fnName) {
				continue
			}
			if len(c.Features) > 0 {
				skipped++
				continue
			}
			if !isSafeForConversion(c) {
				unsafe++
				continue
			}
			if selfContained && !isSelfContainedSQL(c.SQL) {
				skipped++
				continue
			}
			matches = append(matches, c)
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	// Stable order by name so reruns are diff-friendly.
	sort.Slice(matches, func(i, j int) bool { return matches[i].Name < matches[j].Name })
	if len(matches) > limit {
		matches = matches[:limit]
	}

	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "no safely-convertible cases for %s (skipped %d feature-gated, %d non-scalar)\n",
			fnName, skipped, unsafe)
		return nil
	}

	// Render YAML to stdout.
	specRel, _ := filepath.Rel(projRoot, specAbs)
	specRel = filepath.ToSlash(specRel)
	dialect := fm.Dialect
	if dialect == "" {
		dialect = "googlesql"
	}
	fmt.Printf("spec: %s\n", specRel)
	fmt.Printf("dialect: %s\n", dialect)
	fmt.Printf("# Candidate testdata sourced from googlesql compliance fixtures.\n")
	fmt.Printf("# Review each case before committing as authoritative.\n")
	fmt.Printf("cases:\n")
	for _, c := range matches {
		rs, err := compliancetest.ParseExpectedRows(c.Expected)
		if err != nil {
			continue
		}
		desc := c.Name
		if desc == "" {
			desc = "compliance case"
		}
		fmt.Printf("  - desc: %s\n", desc)
		fmt.Printf("    sql: |-\n")
		for _, l := range strings.Split(strings.TrimRight(c.SQL, ";"), "\n") {
			fmt.Printf("      %s\n", l)
		}
		fmt.Printf("    expected:\n")
		fmt.Printf("      rows:\n")
		for _, row := range rs.Rows {
			fmt.Printf("        - [")
			for j, cell := range row {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Print(formatYAMLCell(cell))
			}
			fmt.Print("]\n")
		}
	}
	fmt.Fprintf(os.Stderr, "extract-compliance: %d cases written (skipped %d feature-gated, %d non-scalar)\n",
		len(matches), skipped, unsafe)
	return nil
}

func atoi(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("bad int %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// isSelfContainedSQL is true when the SQL has no `FROM <identifier>`
// referencing an external table -- only inline forms like
// `FROM UNNEST([...])`, `FROM (SELECT ...)`, or no FROM at all. The
// match is intentionally simple: locate every uppercase " FROM "
// boundary and check the next token. Anything else makes the case
// dependent on test-fixture setup we don't carry.
func isSelfContainedSQL(sql string) bool {
	upper := strings.ToUpper(sql)
	rest := upper
	for {
		idx := strings.Index(rest, " FROM ")
		if idx < 0 {
			return true
		}
		// Move just past " FROM ".
		tail := rest[idx+6:]
		tail = strings.TrimLeft(tail, " \t\n")
		if tail == "" {
			return true
		}
		switch tail[0] {
		case '(':
			// `FROM (SELECT ...)` — inline subquery, fine.
		case 'U':
			if !strings.HasPrefix(tail, "UNNEST") {
				return false
			}
		default:
			// Anything else is a table reference we can't satisfy.
			return false
		}
		rest = tail
	}
}

func formatYAMLCell(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", x)
	case float64:
		return fmt.Sprintf("%g", x)
	case string:
		return "'" + strings.ReplaceAll(x, "'", "''") + "'"
	}
	return "null"
}
