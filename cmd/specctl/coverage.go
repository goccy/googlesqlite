package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/googlesqlite/internal/specmeta"
)

func init() {
	register(command{
		name:    "coverage",
		summary: "report gaps between specs, testdata, and implementation",
		run:     runCoverage,
	})
}

func runCoverage(_ context.Context, args []string) error {
	flags := flag.NewFlagSet("coverage", flag.ContinueOnError)
	limit := flags.Int("limit", 20, "max gap entries to print per category")
	verbose := flags.Bool("verbose", false, "print every gap instead of summarising")
	jsonOut := flags.Bool("json", false, "emit a machine-readable JSON report instead of text")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root, err := projectRoot()
	if err != nil {
		return err
	}
	specs, err := specmeta.LoadAllSpecs(root)
	if err != nil {
		return fmt.Errorf("load specs: %w", err)
	}
	report := buildCoverageReport(root, specs)

	if *jsonOut {
		return report.printJSON(os.Stdout)
	}
	cap := *limit
	if *verbose {
		cap = -1
	}
	report.printText(os.Stdout, cap)

	if report.exitCode() != 0 {
		os.Exit(report.exitCode())
	}
	return nil
}

// coverageReport is the data extracted from the spec / testdata pair.
//
// Fields:
//   - byStatus: histogram of status frontmatter values
//   - byCategory: per top-level category, total and ready counts
//   - missingTestdata: spec drafted without a testdata file on disk
//   - testdataNoSpec: testdata present but its `spec:` field points to
//     a missing markdown
//   - statusMismatch: status in {tested,implemented} but testdata
//     missing or empty
type coverageReport struct {
	total           int
	byStatus        map[specmeta.Status]int
	byCategory      map[string]*categoryStat
	missingTestdata []string
	testdataNoSpec  []string
	statusMismatch  []string
}

type categoryStat struct {
	dialect string
	bucket  string // "types" or "functions/<sub>"
	total   int
	ready   int
}

func buildCoverageReport(root string, specs []specmeta.Spec) *coverageReport {
	r := &coverageReport{
		total:      len(specs),
		byStatus:   map[specmeta.Status]int{},
		byCategory: map[string]*categoryStat{},
	}
	for _, sp := range specs {
		fm := sp.Frontmatter
		r.byStatus[fm.Status]++

		key := fmt.Sprintf("%s/%s", fm.Dialect, fm.Category)
		stat, ok := r.byCategory[key]
		if !ok {
			stat = &categoryStat{dialect: fm.Dialect, bucket: fm.Category}
			r.byCategory[key] = stat
		}
		stat.total++
		if isReadyStatus(fm.Status) {
			stat.ready++
		}

		// File-system audits.
		td := strings.TrimSpace(fm.Testdata)
		readyStatus := isReadyStatus(fm.Status)
		if td == "" {
			r.missingTestdata = append(r.missingTestdata, sp.Path)
			if readyStatus {
				// A spec claiming tested / implemented status MUST
				// have a testdata pointer. Surface as a hard mismatch
				// so `specctl check` fails on the regression.
				r.statusMismatch = append(r.statusMismatch, sp.Path)
			}
		} else if _, err := os.Stat(filepath.Join(root, td)); err != nil {
			if os.IsNotExist(err) {
				r.missingTestdata = append(r.missingTestdata, sp.Path)
				if readyStatus {
					// Pointer points at a missing file; same story —
					// implemented status without a runnable testdata
					// file is a coverage hole.
					r.statusMismatch = append(r.statusMismatch, sp.Path)
				}
			}
		} else if readyStatus {
			// status claims tested/implemented; sanity-check that the
			// testdata file isn't empty and parses as YAML with at
			// least one case.
			tdRel := filepath.ToSlash(td)
			parsed, err := specmeta.LoadTestdata(root, tdRel)
			if err != nil || len(parsed.Cases) == 0 {
				r.statusMismatch = append(r.statusMismatch, sp.Path)
			}
		}
	}

	// Cross-check: testdata files whose `spec:` field points at a
	// spec that doesn't exist (drift after a rename).
	r.testdataNoSpec = findOrphanTestdata(root, specs)
	return r
}

func findOrphanTestdata(root string, specs []specmeta.Spec) []string {
	specPaths := map[string]struct{}{}
	for _, sp := range specs {
		specPaths[filepath.ToSlash(sp.Path)] = struct{}{}
	}
	var orphans []string
	tdRoot := filepath.Join(root, "testdata", "specs")
	_ = filepath.Walk(tdRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		td, err := specmeta.LoadTestdata(root, filepath.ToSlash(rel))
		if err != nil {
			orphans = append(orphans, filepath.ToSlash(rel))
			return nil
		}
		if td.Spec == "" {
			orphans = append(orphans, filepath.ToSlash(rel))
			return nil
		}
		if _, ok := specPaths[filepath.ToSlash(td.Spec)]; !ok {
			orphans = append(orphans, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(orphans)
	return orphans
}

func isReadyStatus(s specmeta.Status) bool {
	return s == specmeta.StatusImplemented || s == specmeta.StatusTested
}

// printText writes a human-readable report. capHidden=-1 means print
// every entry; capHidden=N truncates each gap list to N rows with a
// "... and X more" trailer.
func (r *coverageReport) printText(w *os.File, capHidden int) {
	fmt.Fprintf(w, "Specs:                  %d\n", r.total)

	statuses := make([]specmeta.Status, 0, len(r.byStatus))
	for s := range r.byStatus {
		statuses = append(statuses, s)
	}
	sort.Slice(statuses, func(i, j int) bool { return string(statuses[i]) < string(statuses[j]) })
	fmt.Fprintln(w, "By status:")
	for _, s := range statuses {
		fmt.Fprintf(w, "  %-14s %d\n", string(s), r.byStatus[s])
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "By category (ready / total):")
	keys := make([]string, 0, len(r.byCategory))
	for k := range r.byCategory {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s := r.byCategory[k]
		fmt.Fprintf(w, "  %-50s %4d / %4d  (%s)\n",
			k, s.ready, s.total, percentText(s.ready, s.total))
	}

	r.printGapSection(w, "Specs missing testdata", r.missingTestdata, capHidden)
	r.printGapSection(w, "Specs with status=tested|implemented but no usable testdata", r.statusMismatch, capHidden)
	r.printGapSection(w, "Orphan testdata (spec field points to missing spec)", r.testdataNoSpec, capHidden)
}

func (r *coverageReport) printGapSection(w *os.File, title string, entries []string, capHidden int) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s: %d\n", title, len(entries))
	if len(entries) == 0 {
		return
	}
	limit := len(entries)
	if capHidden >= 0 && limit > capHidden {
		limit = capHidden
	}
	for _, e := range entries[:limit] {
		fmt.Fprintf(w, "  - %s\n", e)
	}
	if limit < len(entries) {
		fmt.Fprintf(w, "  ... and %d more (use --verbose to print all)\n",
			len(entries)-limit)
	}
}

func (r *coverageReport) printJSON(w *os.File) error {
	type catEntry struct {
		Dialect string `json:"dialect"`
		Bucket  string `json:"bucket"`
		Total   int    `json:"total"`
		Ready   int    `json:"ready"`
	}
	type body struct {
		Total           int                 `json:"total"`
		ByStatus        map[string]int      `json:"by_status"`
		ByCategory      map[string]catEntry `json:"by_category"`
		MissingTestdata []string            `json:"missing_testdata"`
		StatusMismatch  []string            `json:"status_mismatch"`
		OrphanTestdata  []string            `json:"orphan_testdata"`
	}
	b := body{
		Total:           r.total,
		ByStatus:        map[string]int{},
		ByCategory:      map[string]catEntry{},
		MissingTestdata: r.missingTestdata,
		StatusMismatch:  r.statusMismatch,
		OrphanTestdata:  r.testdataNoSpec,
	}
	for s, n := range r.byStatus {
		b.ByStatus[string(s)] = n
	}
	for k, s := range r.byCategory {
		b.ByCategory[k] = catEntry{Dialect: s.dialect, Bucket: s.bucket, Total: s.total, Ready: s.ready}
	}
	enc := jsonEncoder(w)
	return enc.Encode(b)
}

// percentText formats `num/denom` as a percent, gracefully handling
// the divide-by-zero case.
func percentText(num, denom int) string {
	if denom == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d%%", num*100/denom)
}

// exitCode returns non-zero when there are integrity issues that
// should fail CI. Pure status gaps (drafted entries) are NOT
// considered failures here; they're tracked as work-to-do, not as
// repo invariants.
func (r *coverageReport) exitCode() int {
	if len(r.statusMismatch) > 0 || len(r.testdataNoSpec) > 0 {
		return 1
	}
	return 0
}
