package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	specRoot         = "docs/specs"
	googlesqlBase    = "googlesql"
	googlesqlBlobURL = "https://github.com/google/googlesql/blob/master/docs"
)

// normalizeFile maps one upstream markdown file to a normalized output
// destination. Files not listed here are intentionally skipped: concept
// pages, indexes, syntax guides, and graph extensions don't fit the
// "one entry per H2" model.
type normalizeFile struct {
	upstream string // basename inside docs/third_party/googlesql-docs/
	category string // logical category, e.g. "functions/string", "types"
	kind     specKind
}

type specKind int

const (
	specKindFunction specKind = iota
	specKindType
)

var normalizeFiles = []normalizeFile{
	// Scalar function categories.
	{"string_functions.md", "functions/string", specKindFunction},
	{"mathematical_functions.md", "functions/math", specKindFunction},
	{"array_functions.md", "functions/array", specKindFunction},
	{"date_functions.md", "functions/date", specKindFunction},
	{"datetime_functions.md", "functions/datetime", specKindFunction},
	{"time_functions.md", "functions/time", specKindFunction},
	{"timestamp_functions.md", "functions/timestamp", specKindFunction},
	{"interval_functions.md", "functions/interval", specKindFunction},
	{"json_functions.md", "functions/json", specKindFunction},
	{"hash_functions.md", "functions/hash", specKindFunction},
	{"bit_functions.md", "functions/bit", specKindFunction},
	{"net_functions.md", "functions/net", specKindFunction},
	{"geography_functions.md", "functions/geography", specKindFunction},
	{"range-functions.md", "functions/range", specKindFunction},
	{"hll_functions.md", "functions/hll", specKindFunction},
	{"conversion_functions.md", "functions/conversion", specKindFunction},
	{"security_functions.md", "functions/security", specKindFunction},
	{"debugging_functions.md", "functions/debug", specKindFunction},
	{"protocol_buffer_functions.md", "functions/proto", specKindFunction},

	// Aggregates and window functions. Approximate, statistical and
	// differential-privacy variants live under their own subdirectories
	// because they share names with the standard aggregates (e.g. SUM,
	// AVG) and would otherwise overwrite each other.
	{"aggregate_functions.md", "functions/aggregate", specKindFunction},
	{"approximate_aggregate_functions.md", "functions/aggregate/approximate", specKindFunction},
	{"statistical_aggregate_functions.md", "functions/aggregate/statistical", specKindFunction},
	{"aggregate-dp-functions.md", "functions/aggregate/differential_privacy", specKindFunction},
	{"numbering_functions.md", "functions/window", specKindFunction},
	{"navigation_functions.md", "functions/window", specKindFunction},

	// Data types.
	{"data-types.md", "types", specKindType},
}

// nonEntryHeadings are H2s in the upstream files that are not individual
// function/type entries (TOC, summary tables, general sections).
var nonEntryHeadings = map[string]bool{
	"function list":           true,
	"data type list":          true,
	"data type properties":    true,
	"deterministic functions": true,
	"deprecated functions":    true,
}

func init() {
	register(command{
		name:    "normalize",
		summary: "generate spec skeletons for upstream functions that lack one (never modifies existing specs)",
		run:     runNormalize,
	})
}

// runNormalize generates spec skeletons only for upstream functions and
// types that do not yet have a spec file under docs/specs/. It never
// modifies existing specs: any upstream entry whose destination spec
// already exists is left byte-for-byte untouched (no status reset, no
// last_synced bump, no body re-rendering). Running normalize on a
// fully-covered repository therefore produces a zero git diff.
func runNormalize(_ context.Context, args []string) error {
	flags := flag.NewFlagSet("normalize", flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: specctl normalize [--dry-run]")
		fmt.Fprintln(os.Stderr, "generates spec skeletons only for upstream functions that do not")
		fmt.Fprintln(os.Stderr, "yet have a spec; never modifies existing specs.")
		flags.PrintDefaults()
	}
	dryRun := flags.Bool("dry-run", false, "print new skeletons that would be written, without writing them")
	if err := flags.Parse(args); err != nil {
		return err
	}

	now := time.Now().UTC().Format("2006-01-02")

	report := &normalizeReport{}
	for _, nf := range normalizeFiles {
		path := filepath.Join("docs/third_party", "googlesql-docs", nf.upstream)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := normalizeOne(nf, string(data), now, *dryRun, report); err != nil {
			return fmt.Errorf("normalize %s: %w", nf.upstream, err)
		}
	}

	report.print()
	return nil
}

func normalizeOne(nf normalizeFile, src, syncDate string, dryRun bool, r *normalizeReport) error {
	_, sections := splitH2(src)
	outDir := filepath.Join(specRoot, googlesqlBase, nf.category)
	if !dryRun {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
	}

	for _, sec := range sections {
		entry, ok := classifyHeading(sec.heading, nf.kind)
		if !ok {
			r.skippedHeadings++
			continue
		}

		slug := slugify(entry.name)
		if slug == "" {
			r.skippedHeadings++
			continue
		}
		filename := slug + ".md"
		out := filepath.Join(outDir, filename)

		// Additive-only: never touch a spec that already exists. The
		// 700+ spec files have been heavily hand-refined since bootstrap
		// (status promoted, notes written, testdata wired, body
		// sections filled in). normalize only generates a skeleton for a
		// brand-new upstream function/type that has no spec yet.
		if _, err := os.Stat(out); err == nil {
			r.unchanged++
			continue
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", out, err)
		}

		fm := frontmatter{
			Name:        canonicalName(entry.name, nf.kind),
			Dialect:     googlesqlBase,
			Category:    nf.category,
			Status:      "drafted",
			SourceURL:   "docs/third_party/googlesql-docs/" + nf.upstream,
			UpstreamURL: upstreamAnchor(googlesqlBlobURL, nf.upstream, sec.heading),
			LastSynced:  syncDate,
			Testdata:    filepath.Join("testdata", "specs", googlesqlBase, nf.category, slug+".yaml"),
		}

		body := renderSpecBody(fm, sec, nf)
		if dryRun {
			r.wouldWrite = append(r.wouldWrite, out)
			continue
		}

		if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
		r.created = append(r.created, out)
	}
	return nil
}

type entryHeading struct {
	name string // raw extracted name (function or type)
}

func classifyHeading(heading string, kind specKind) (entryHeading, bool) {
	stripped := strings.TrimSpace(strings.TrimPrefix(heading, "##"))
	if nonEntryHeadings[strings.ToLower(strings.TrimSpace(strings.ReplaceAll(stripped, "`", "")))] {
		return entryHeading{}, false
	}
	switch kind {
	case specKindFunction:
		name := extractFunctionName(heading)
		if name == "" {
			return entryHeading{}, false
		}
		return entryHeading{name: name}, true
	case specKindType:
		name := extractTypeName(heading)
		if name == "" {
			return entryHeading{}, false
		}
		return entryHeading{name: name}, true
	}
	return entryHeading{}, false
}

func canonicalName(raw string, kind specKind) string {
	switch kind {
	case specKindFunction:
		return strings.ToUpper(raw)
	case specKindType:
		return strings.ToUpper(raw)
	}
	return raw
}

// renderSpecBody composes the spec markdown body for a brand-new spec
// skeleton. The per-section content is intentionally a thin scaffold:
// the upstream excerpt is copied verbatim into a fenced "Reference
// (upstream)" block, while the other sections are placeholders for
// human / agent refinement. Once the skeleton exists, normalize never
// re-renders it — subsequent refinements are preserved verbatim.
func renderSpecBody(fm frontmatter, sec section, nf normalizeFile) string {
	var b strings.Builder
	// strings.Builder never errors, so the writeTo error is dropped.
	_ = fm.writeTo(&b)
	fmt.Fprintf(&b, "\n# %s\n\n", fm.Name)

	b.WriteString("## Summary\n\n")
	b.WriteString("(TBD — refine from the upstream reference below.)\n\n")

	b.WriteString("## Signatures\n\n")
	b.WriteString("(TBD)\n\n")

	b.WriteString("## Behavior\n\n")
	b.WriteString("(TBD)\n\n")

	b.WriteString("## Examples\n\n")
	b.WriteString("(TBD)\n\n")

	b.WriteString("## Edge cases\n\n")
	b.WriteString("(TBD)\n\n")

	b.WriteString("## Reference (upstream)\n\n")
	b.WriteString("Verbatim copy from `" + fm.SourceURL + "`. Auto-managed by\n")
	b.WriteString("`specctl normalize`; do not edit by hand.\n\n")
	b.WriteString(sec.heading)
	b.WriteString("\n")
	body := strings.TrimRight(sec.body, "\n")
	b.WriteString(body)
	b.WriteString("\n\n")

	b.WriteString("## References\n\n")
	fmt.Fprintf(&b, "- Apache 2.0 derivative of `%s`.\n", fm.SourceURL)

	return b.String()
}

// normalizeReport accumulates the outcome of an additive-only normalize
// run. unchanged counts upstream entries whose spec already exists and
// was therefore left untouched; created lists newly generated skeletons.
type normalizeReport struct {
	created         []string
	unchanged       int
	skippedHeadings int
	wouldWrite      []string
}

func (r *normalizeReport) print() {
	sort.Strings(r.created)
	sort.Strings(r.wouldWrite)

	if len(r.wouldWrite) > 0 {
		fmt.Printf("normalize (dry-run): %d new skeleton(s) would be written\n", len(r.wouldWrite))
		for _, p := range r.wouldWrite {
			fmt.Printf("  + %s\n", p)
		}
		return
	}

	fmt.Printf("normalize: created=%d unchanged=%d skipped_headings=%d\n",
		len(r.created), r.unchanged, r.skippedHeadings)
	for _, p := range r.created {
		fmt.Printf("  + %s\n", p)
	}
}
