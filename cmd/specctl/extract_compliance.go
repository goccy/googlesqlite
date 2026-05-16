package main

import (
	"bufio"
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
		name:    "scan-compliance",
		summary: "scan an external compliance .test directory and report cases applicable to each spec",
		run: func(_ context.Context, args []string) error {
			return runScanCompliance(args)
		},
	})
}

// runScanCompliance walks a compliance testdata directory (the
// GoogleSQL "*.test" fixtures hosted under e.g.
// googlesql-wasm/googlesql/googlesql/compliance/testdata/) and reports
// which function names have at least one safely-translatable case
// available there. Used as a feasibility scan before doing a manual
// re-extraction pass into our testdata YAML.
//
// Usage:
//
//	specctl scan-compliance --root <path/to/compliance/testdata> [--filter <substring>]
//
// Output format (TSV on stdout): function-name, case-count,
// safe-case-count, file-list.
func runScanCompliance(args []string) error {
	root := ""
	filter := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--root":
			if i+1 < len(args) {
				root = args[i+1]
				i++
			}
		case "--filter":
			if i+1 < len(args) {
				filter = args[i+1]
				i++
			}
		}
	}
	if root == "" {
		return fmt.Errorf("--root is required (path to compliance/testdata directory)")
	}
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("--root: %w", err)
	}

	specs, err := loadSpecsForCoverage()
	if err != nil {
		return err
	}

	// Build a function-name → spec map (uppercased). The compliance
	// SQL refers to functions by name, so we use a name match.
	specByFn := map[string]specmeta.Spec{}
	for _, sp := range specs {
		key := strings.ToUpper(strings.TrimSpace(sp.Frontmatter.Name))
		if key == "" {
			continue
		}
		// First-write wins; multi-binding specs (e.g. JSON_VALUE in
		// both bigquery/ and googlesql/) get one entry and we surface
		// the conflict in a later pass if needed.
		if _, ok := specByFn[key]; !ok {
			specByFn[key] = sp
		}
	}

	// Aggregate per function name.
	type stat struct {
		name     string
		caseCnt  int
		safeCnt  int
		filenms  map[string]struct{}
		linkedTo string
	}
	stats := map[string]*stat{}
	get := func(name string) *stat {
		s, ok := stats[name]
		if !ok {
			s = &stat{name: name, filenms: map[string]struct{}{}}
			stats[name] = s
		}
		return s
	}

	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".test" {
			return nil
		}
		cases, perr := compliancetest.ParseFile(path)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "warn parse %s: %v\n", path, perr)
			return nil
		}
		fileName := filepath.Base(path)
		for _, c := range cases {
			// Heuristic: scan SQL for tokens that match registered
			// function names. We accumulate per-function regardless
			// of whether the function call is bare (`X(...)`) or
			// dotted (`X.MERGE(...)`). The cost of imperfect matches
			// is acceptable for a scan tool.
			seen := map[string]bool{}
			for fn := range specByFn {
				if seen[fn] {
					continue
				}
				if !mentionsFunction(c.SQL, fn) {
					continue
				}
				seen[fn] = true
				s := get(fn)
				s.caseCnt++
				s.filenms[fileName] = struct{}{}
				s.linkedTo = specByFn[fn].Path
				// "Safe" cases are those whose expected block is
				// ExpectedRows AND the rows convert to YAML-friendly
				// scalars AND the SQL has no required_features.
				if isSafeForConversion(c) {
					s.safeCnt++
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	// Sort by function name and emit TSV.
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintln(out, "function\tcases\tsafe_cases\tfiles\tspec")
	for _, k := range keys {
		s := stats[k]
		if filter != "" && !strings.Contains(strings.ToLower(k), strings.ToLower(filter)) {
			continue
		}
		files := make([]string, 0, len(s.filenms))
		for f := range s.filenms {
			files = append(files, f)
		}
		sort.Strings(files)
		fmt.Fprintf(out, "%s\t%d\t%d\t%s\t%s\n",
			k, s.caseCnt, s.safeCnt, strings.Join(files, ","), s.linkedTo)
	}
	return nil
}

func loadSpecsForCoverage() ([]specmeta.Spec, error) {
	root, err := projectRoot()
	if err != nil {
		return nil, err
	}
	return specmeta.LoadAllSpecs(root)
}

// mentionsFunction reports whether sql contains a token call to fn.
// fn is the uppercase name; we match case-insensitively against an
// identifier boundary (so SUBSTR matches "SUBSTR(", "substr(" but not
// "MYSUBSTR(").
func mentionsFunction(sql, fn string) bool {
	upper := strings.ToUpper(sql)
	target := fn
	for {
		idx := strings.Index(upper, target)
		if idx < 0 {
			return false
		}
		// Left boundary: start of string or non-identifier char.
		if idx > 0 {
			c := upper[idx-1]
			if isIdentByte(c) {
				upper = upper[idx+len(target):]
				continue
			}
		}
		// Right boundary: must be followed by `(` (optionally after
		// whitespace) — function call syntax.
		j := idx + len(target)
		for j < len(upper) && (upper[j] == ' ' || upper[j] == '\t' || upper[j] == '\n') {
			j++
		}
		if j < len(upper) && upper[j] == '(' {
			return true
		}
		upper = upper[idx+len(target):]
	}
}

func isIdentByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// isSafeForConversion is true when the case has no required_features
// and the expected block is a row-equality block whose cells are all
// YAML-friendly scalars.
func isSafeForConversion(c compliancetest.Case) bool {
	if len(c.Features) > 0 {
		return false
	}
	if compliancetest.ClassifyExpected(c.Expected) != compliancetest.ExpectedRows {
		return false
	}
	rs, err := compliancetest.ParseExpectedRows(c.Expected)
	if err != nil {
		return false
	}
	return compliancetest.ConvertibleToYAML(rs)
}
