// Package specmeta defines the shared frontmatter and on-disk layout
// used by spec markdown files and their accompanying testdata YAML.
//
// Both the spec-test runner (the root TestSpec suite) and the specctl
// CLI (cmd/specctl) parse spec metadata; this package keeps the schema
// in one place.
package specmeta

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

// Status is the lifecycle marker stored in a spec's frontmatter.
type Status string

const (
	StatusDrafted     Status = "drafted"
	StatusReviewed    Status = "reviewed"
	StatusTested      Status = "tested"
	StatusImplemented Status = "implemented"
	StatusPartial     Status = "partial"
	StatusUnsupported Status = "unsupported"
)

// AllStatuses lists statuses in display order (left-to-right in the
// support-matrix narrative — "more done" rightward).
var AllStatuses = []Status{
	StatusDrafted,
	StatusReviewed,
	StatusTested,
	StatusImplemented,
	StatusPartial,
	StatusUnsupported,
}

// Frontmatter is the YAML header block that opens every spec.
type Frontmatter struct {
	Name        string `yaml:"name"`
	Dialect     string `yaml:"dialect"`
	Category    string `yaml:"category"`
	Status      Status `yaml:"status"`
	SourceURL   string `yaml:"source_url"`
	UpstreamURL string `yaml:"upstream_url"`
	LastSynced  string `yaml:"last_synced"`
	Testdata    string `yaml:"testdata"`
	Notes       string `yaml:"notes,omitempty"`
}

// Spec couples the parsed frontmatter with the file path it came from.
type Spec struct {
	Path        string // path relative to project root, e.g. docs/specs/googlesql/functions/string/upper.md
	Frontmatter Frontmatter
}

// SpecRoot is the project-root-relative directory holding spec files.
const SpecRoot = "docs/specs"

// LoadAllSpecs walks SpecRoot under root and returns every spec found.
// The "_template.md" file is intentionally excluded.
func LoadAllSpecs(root string) ([]Spec, error) {
	dir := filepath.Join(root, SpecRoot)
	var out []Spec
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == "_template.md" || base == "INDEX.md" {
			return nil
		}
		if !strings.HasSuffix(base, ".md") {
			return nil
		}
		spec, err := loadSpec(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		spec.Path = filepath.ToSlash(rel)
		out = append(out, spec)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func loadSpec(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	fm, err := ParseFrontmatter(data)
	if err != nil {
		return Spec{}, err
	}
	return Spec{Frontmatter: fm}, nil
}

// ParseFrontmatter extracts the YAML frontmatter block from raw markdown
// content. The block is the first contiguous run of lines bracketed by
// `---` markers at the start of the file.
func ParseFrontmatter(data []byte) (Frontmatter, error) {
	const sep = "---"
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) == 0 || string(bytes.TrimSpace(lines[0])) != sep {
		return Frontmatter{}, fmt.Errorf("frontmatter marker missing (first line must be %q)", sep)
	}
	var inside [][]byte
	closed := false
	for _, line := range lines[1:] {
		if string(bytes.TrimSpace(line)) == sep {
			closed = true
			break
		}
		inside = append(inside, line)
	}
	if !closed {
		return Frontmatter{}, fmt.Errorf("frontmatter not closed by %q", sep)
	}
	var fm Frontmatter
	if err := yaml.Unmarshal(bytes.Join(inside, []byte("\n")), &fm); err != nil {
		return Frontmatter{}, fmt.Errorf("yaml: %w", err)
	}
	return fm, nil
}

// Testdata is the on-disk shape of a testdata YAML file.
type Testdata struct {
	Spec    string `yaml:"spec"`
	Dialect string `yaml:"dialect"`
	Cases   []Case `yaml:"cases"`
}

// Case is a single declarative test case.
type Case struct {
	Desc      string         `yaml:"desc"`
	Setup     []string       `yaml:"setup,omitempty"`
	SQL       string         `yaml:"sql"`
	Args      []any          `yaml:"args,omitempty"`
	NamedArgs map[string]any `yaml:"named_args,omitempty"`
	Expected  Expected       `yaml:"expected"`
	// RegisterProtos names test-side proto bundles (see
	// `internal/spectest/protos.go`) that the runner installs via
	// Conn.RegisterProto before the setup statements run. Use this
	// for spec cases whose upstream Example references a user-defined
	// proto message / enum (e.g. `Item`, `googlesql.examples.music.Album`)
	// that the analyzer would otherwise reject as Type-not-found.
	RegisterProtos []string `yaml:"register_protos,omitempty"`
	// Nondeterministic, when true, runs the SQL but only asserts the
	// case produced the right number of result rows — the values
	// themselves are not compared. Use this exclusively for upstream
	// Examples whose docs explicitly disclaim deterministic output
	// (e.g. DP-with-noise queries where "These results will change
	// each time you run the query"). The `expected:` rows field
	// stays in the YAML as documentation of the docs' specific
	// realization.
	Nondeterministic bool `yaml:"nondeterministic,omitempty"`
}

// Expected captures the expected outcome of a case. Exactly one of
// Rows / Error must be populated.
type Expected struct {
	Rows  [][]any `yaml:"rows,omitempty"`
	Error *Error  `yaml:"error,omitempty"`
	// Unordered, when true, treats Rows as a multiset: a case passes
	// if `got` is a permutation of `Rows`. Use this only when the
	// SQL's ORDER BY (or absence thereof) leaves the row order
	// implementation-defined — e.g. GROUP BY GROUPING SETS where the
	// outer ORDER BY does not fully break the per-group tie.
	Unordered bool `yaml:"unordered,omitempty"`
}

// Error specifies an expected error from the engine.
type Error struct {
	Contains string `yaml:"contains"`
	Phase    string `yaml:"phase,omitempty"` // "analysis" | "execution"
}

// LoadTestdata reads a testdata YAML at path (project-root-relative).
func LoadTestdata(root, relpath string) (Testdata, error) {
	p := filepath.Join(root, relpath)
	data, err := os.ReadFile(p)
	if err != nil {
		return Testdata{}, err
	}
	var td Testdata
	if err := yaml.Unmarshal(data, &td); err != nil {
		return Testdata{}, fmt.Errorf("%s: yaml: %w", relpath, err)
	}
	return td, nil
}
