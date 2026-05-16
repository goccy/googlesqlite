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
		name:    "validate",
		summary: "validate the frontmatter and section structure of spec files",
		run:     runValidate,
	})
}

// validStatuses lists every value `status:` is allowed to take.
var validStatuses = map[specmeta.Status]bool{
	specmeta.StatusDrafted:     true,
	specmeta.StatusReviewed:    true,
	specmeta.StatusTested:      true,
	specmeta.StatusImplemented: true,
	specmeta.StatusPartial:     true,
	specmeta.StatusUnsupported: true,
}

// requiredFunctionSections are the H2 headers every function spec
// must carry. Type specs share the same set today; if that diverges
// we'll add a per-kind table.
var requiredSections = []string{
	"## Summary",
	"## Signatures",
	"## Behavior",
	"## Examples",
	"## Edge cases",
	"## Reference (upstream)",
}

func runValidate(_ context.Context, args []string) error {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	all := flags.Bool("all", false, "validate every spec under docs/specs/")
	if err := flags.Parse(args); err != nil {
		return err
	}

	root, err := projectRoot()
	if err != nil {
		return err
	}

	var paths []string
	if *all {
		specs, err := specmeta.LoadAllSpecs(root)
		if err != nil {
			return err
		}
		for _, sp := range specs {
			paths = append(paths, filepath.Join(root, sp.Path))
		}
	} else {
		if flags.NArg() == 0 {
			return fmt.Errorf("usage: validate [--all] <path> [<path>...]")
		}
		for _, p := range flags.Args() {
			abs, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			paths = append(paths, abs)
		}
	}

	var failures []validateFailure
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			failures = append(failures, validateFailure{path: p, errs: []string{err.Error()}})
			continue
		}
		errs := validateSpecBytes(data)
		if len(errs) > 0 {
			rel, _ := filepath.Rel(root, p)
			failures = append(failures, validateFailure{path: filepath.ToSlash(rel), errs: errs})
		}
	}

	sort.Slice(failures, func(i, j int) bool { return failures[i].path < failures[j].path })
	for _, f := range failures {
		for _, msg := range f.errs {
			fmt.Fprintf(os.Stderr, "%s: %s\n", f.path, msg)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%d spec file(s) failed validation", len(failures))
	}
	fmt.Printf("validate: %d spec file(s) OK\n", len(paths))
	return nil
}

type validateFailure struct {
	path string
	errs []string
}

// validateSpecBytes returns a list of issues found in the given spec
// markdown. An empty slice means the file is well-formed.
func validateSpecBytes(data []byte) []string {
	var errs []string
	fm, err := specmeta.ParseFrontmatter(data)
	if err != nil {
		errs = append(errs, fmt.Sprintf("frontmatter: %v", err))
		return errs
	}
	if fm.Name == "" {
		errs = append(errs, "frontmatter: name is empty")
	}
	if fm.Dialect == "" {
		errs = append(errs, "frontmatter: dialect is empty")
	}
	if fm.Category == "" {
		errs = append(errs, "frontmatter: category is empty")
	}
	if !validStatuses[fm.Status] {
		errs = append(errs, fmt.Sprintf("frontmatter: invalid status %q", fm.Status))
	}
	if fm.SourceURL == "" {
		errs = append(errs, "frontmatter: source_url is empty")
	}
	if fm.Testdata == "" {
		errs = append(errs, "frontmatter: testdata is empty")
	} else if !strings.HasPrefix(fm.Testdata, "testdata/specs/") {
		errs = append(errs, fmt.Sprintf("frontmatter: testdata must live under testdata/specs/, got %q", fm.Testdata))
	}
	// `partial` and `unsupported` require a notes line explaining the
	// gap; everything else allows it but doesn't require it.
	if (fm.Status == specmeta.StatusPartial || fm.Status == specmeta.StatusUnsupported) && strings.TrimSpace(fm.Notes) == "" {
		errs = append(errs, fmt.Sprintf("frontmatter: status %q requires a notes line", fm.Status))
	}

	body := string(data)
	for _, header := range requiredSections {
		if !strings.Contains(body, "\n"+header+"\n") {
			errs = append(errs, fmt.Sprintf("section missing: %s", header))
		}
	}
	return errs
}
