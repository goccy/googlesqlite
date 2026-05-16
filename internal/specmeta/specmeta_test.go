package specmeta_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/googlesqlite/internal/specmeta"
)

// validFrontmatter is the minimal frontmatter shape every parser test
// uses as its starting point. The fields mirror the project's
// `_template.md` so the assertions stay in sync with the canonical
// spec layout.
const validFrontmatter = `---
name: TEST_FUNC
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://example.com/x
last_synced: 2026-05-15
testdata: testdata/specs/googlesql/functions/string/test_func.yaml
notes: |
  example notes
---

# Body
`

func TestParseFrontmatterRoundtripsAllFields(t *testing.T) {
	t.Parallel()

	fm, err := specmeta.ParseFrontmatter([]byte(validFrontmatter))
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if fm.Name != "TEST_FUNC" {
		t.Errorf("Name = %q", fm.Name)
	}
	if fm.Dialect != "googlesql" {
		t.Errorf("Dialect = %q", fm.Dialect)
	}
	if fm.Category != "functions/string" {
		t.Errorf("Category = %q", fm.Category)
	}
	if fm.Status != specmeta.StatusImplemented {
		t.Errorf("Status = %q", fm.Status)
	}
	if fm.SourceURL == "" {
		t.Error("SourceURL empty")
	}
	if fm.UpstreamURL == "" {
		t.Error("UpstreamURL empty")
	}
	if fm.LastSynced != "2026-05-15" {
		t.Errorf("LastSynced = %q", fm.LastSynced)
	}
	if fm.Testdata == "" {
		t.Error("Testdata empty")
	}
}

func TestParseFrontmatterMissingMarker(t *testing.T) {
	t.Parallel()

	if _, err := specmeta.ParseFrontmatter([]byte("# no marker\n")); err == nil {
		t.Fatal("expected error when first line is not `---`")
	}
}

func TestParseFrontmatterUnterminatedMarker(t *testing.T) {
	t.Parallel()

	body := "---\nname: X\n# never closed\n"
	if _, err := specmeta.ParseFrontmatter([]byte(body)); err == nil {
		t.Fatal("expected error when frontmatter is not closed")
	}
}

func TestParseFrontmatterInvalidYAML(t *testing.T) {
	t.Parallel()

	// `: : :` is not a valid YAML key/value sequence.
	body := "---\n: : :\n---\n"
	if _, err := specmeta.ParseFrontmatter([]byte(body)); err == nil {
		t.Fatal("expected YAML parse error")
	}
}

// TestLoadAllSpecs builds a small spec tree under a tempdir and
// asserts the walker finds the canonical spec, sorts the output, and
// excludes the template / INDEX files.
func TestLoadAllSpecs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specDir := filepath.Join(root, specmeta.SpecRoot, "googlesql", "functions", "string")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Two valid specs (sorted alphabetically: lower.md, upper.md).
	if err := os.WriteFile(filepath.Join(specDir, "upper.md"), []byte(validFrontmatter), 0o644); err != nil {
		t.Fatal(err)
	}
	lowerFrontmatter := []byte(`---
name: LOWER
dialect: googlesql
category: functions/string
status: drafted
source_url: x
upstream_url: y
last_synced: 2026-05-15
testdata: t
---
`)
	if err := os.WriteFile(filepath.Join(specDir, "lower.md"), lowerFrontmatter, 0o644); err != nil {
		t.Fatal(err)
	}

	// Excluded files: _template.md and INDEX.md.
	if err := os.WriteFile(filepath.Join(specDir, "_template.md"), []byte("---\nname: T\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, specmeta.SpecRoot, "INDEX.md"), []byte("---\nname: I\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Non-markdown files are silently skipped.
	if err := os.WriteFile(filepath.Join(specDir, "stray.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	specs, err := specmeta.LoadAllSpecs(root)
	if err != nil {
		t.Fatalf("LoadAllSpecs: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs, got %d: %+v", len(specs), specs)
	}
	// Sorted by path.
	if specs[0].Frontmatter.Name != "LOWER" {
		t.Errorf("first spec Name = %q", specs[0].Frontmatter.Name)
	}
	if specs[1].Frontmatter.Name != "TEST_FUNC" {
		t.Errorf("second spec Name = %q", specs[1].Frontmatter.Name)
	}
	// Path is project-root-relative and uses forward slashes.
	if specs[0].Path != "docs/specs/googlesql/functions/string/lower.md" {
		t.Errorf("first spec Path = %q", specs[0].Path)
	}
}

// TestLoadAllSpecsMissingDir is a no-op when the spec root doesn't
// exist (e.g. running the test outside the repo).
func TestLoadAllSpecsMissingDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specs, err := specmeta.LoadAllSpecs(root)
	if err != nil {
		t.Fatalf("LoadAllSpecs: %v", err)
	}
	if len(specs) != 0 {
		t.Fatalf("expected empty result, got %d", len(specs))
	}
}

// TestLoadAllSpecsPropagatesParseError surfaces a spec with an
// invalid frontmatter as a wrapped error.
func TestLoadAllSpecsPropagatesParseError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specDir := filepath.Join(root, specmeta.SpecRoot)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "bad.md"), []byte("not frontmatter"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := specmeta.LoadAllSpecs(root); err == nil {
		t.Fatal("expected error from bad spec")
	}
}

// TestLoadTestdata round-trips the canonical testdata YAML shape.
func TestLoadTestdata(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rel := "testdata/specs/x/y.yaml"
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte(`spec: docs/specs/x/y.md
dialect: googlesql
cases:
  - desc: simple
    sql: SELECT 1
    expected:
      rows:
        - [1]
  - desc: with error
    sql: SELECT bad
    expected:
      error:
        contains: bad
        phase: analysis
  - desc: with named args
    sql: SELECT @a
    named_args:
      a: 1
    expected:
      rows:
        - [1]
    register_protos: [foo]
    nondeterministic: true
  - desc: unordered
    sql: SELECT * FROM T
    expected:
      unordered: true
      rows:
        - [1]
        - [2]
`)
	if err := os.WriteFile(full, body, 0o644); err != nil {
		t.Fatal(err)
	}
	td, err := specmeta.LoadTestdata(root, rel)
	if err != nil {
		t.Fatalf("LoadTestdata: %v", err)
	}
	if td.Spec != "docs/specs/x/y.md" {
		t.Errorf("Spec = %q", td.Spec)
	}
	if td.Dialect != "googlesql" {
		t.Errorf("Dialect = %q", td.Dialect)
	}
	if len(td.Cases) != 4 {
		t.Fatalf("expected 4 cases, got %d", len(td.Cases))
	}
	if td.Cases[1].Expected.Error == nil || td.Cases[1].Expected.Error.Contains != "bad" {
		t.Errorf("case 2 error: %+v", td.Cases[1].Expected.Error)
	}
	if !td.Cases[2].Nondeterministic {
		t.Error("case 3 should be nondeterministic")
	}
	if len(td.Cases[2].RegisterProtos) == 0 {
		t.Error("case 3 should register protos")
	}
	if !td.Cases[3].Expected.Unordered {
		t.Error("case 4 should be unordered")
	}
}

func TestLoadTestdataMissingFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if _, err := specmeta.LoadTestdata(root, "does/not/exist.yaml"); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadTestdataMalformedYAML(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	rel := "x.yaml"
	full := filepath.Join(root, rel)
	if err := os.WriteFile(full, []byte("not:\n  - valid: : :"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := specmeta.LoadTestdata(root, rel); err == nil {
		t.Fatal("expected error")
	}
}

func TestAllStatusesContainsCanonicalSet(t *testing.T) {
	t.Parallel()

	want := []specmeta.Status{
		specmeta.StatusDrafted,
		specmeta.StatusReviewed,
		specmeta.StatusTested,
		specmeta.StatusImplemented,
		specmeta.StatusPartial,
		specmeta.StatusUnsupported,
	}
	if len(specmeta.AllStatuses) != len(want) {
		t.Fatalf("AllStatuses length: %d", len(specmeta.AllStatuses))
	}
	for i, s := range want {
		if specmeta.AllStatuses[i] != s {
			t.Errorf("AllStatuses[%d] = %q want %q", i, specmeta.AllStatuses[i], s)
		}
	}
}
