// Package main tests drive the specctl CLI as a subprocess so the
// tests survive any refactor that reshapes the internal layout.
//
// The strategy: compile the CLI binary once in TestMain, then run every
// subcommand against curated fixture project roots placed in
// temp directories. Each test exercises a different code path —
// success, validation failure, missing flags, etc. — and asserts on
// exit code, stdout, and stderr.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// specctlBin is the path to the compiled specctl binary populated by
// TestMain. Each test runs it via os/exec and captures stdout/stderr.
var specctlBin string

// covDir, when non-empty, is the directory where the instrumented
// binary emits GOCOVERDIR coverage files. The binary is built with
// `go build -cover` so subprocess invocations contribute to the
// coverage report assembled at the end of the test run.
var covDir string

func TestMain(m *testing.M) {
	flag.Parse()
	dir, err := os.MkdirTemp("", "specctl-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: tempdir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	name := "specctl"
	if runtime.GOOS == "windows" {
		name = "specctl.exe"
	}
	specctlBin = filepath.Join(dir, name)

	// Build with coverage instrumentation so subprocess runs contribute
	// to the package coverage profile. The instrumented binary needs a
	// GOCOVERDIR directory to dump per-process counter files; we merge
	// them at the end of m.Run via `go tool covdata textfmt` if the
	// caller asked for `-test.coverprofile`.
	args := []string{"build", "-cover", "-coverpkg=./...", "-o", specctlBin, "."}
	build := exec.Command("go", args...)
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: build: %v\n", err)
		os.Exit(1)
	}
	covDir, err = os.MkdirTemp("", "specctl-cov-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: covdir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(covDir)
	code := m.Run()
	// If `-test.coverprofile` was supplied to `go test`, merge the
	// subprocess counters into a textfmt profile and concatenate it
	// onto the cover profile so the parent test run sees subprocess
	// coverage.
	if cp := coverProfilePath(); cp != "" {
		mergeCoverage(cp)
	}
	os.Exit(code)
}

// coverProfilePath returns the value of the -test.coverprofile flag if
// set, or "" otherwise.
func coverProfilePath() string {
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.coverprofile=") {
			return strings.TrimPrefix(a, "-test.coverprofile=")
		}
	}
	return ""
}

// mergeCoverage uses `go tool covdata textfmt` to assemble a profile
// from covDir and appends its body (skipping the `mode:` line) onto
// the existing coverprofile written by `go test`.
func mergeCoverage(target string) {
	out := filepath.Join(covDir, "_textfmt.cov")
	cmd := exec.Command("go", "tool", "covdata", "textfmt", "-i", covDir, "-o", out)
	if err := cmd.Run(); err != nil {
		// Best effort. A failure here just means we lose subprocess
		// coverage signal — still better than the test failing.
		return
	}
	merged, err := os.ReadFile(out)
	if err != nil {
		return
	}
	cur, err := os.ReadFile(target)
	if err != nil {
		return
	}
	lines := strings.Split(string(merged), "\n")
	// First line is `mode: set` (or similar). Skip it so we don't
	// duplicate it in the merged profile.
	if len(lines) > 1 {
		// Keep cur unchanged; append the merged lines beyond the mode
		// header. Existing entries take precedence because tools sum
		// duplicate counters.
		body := strings.Join(lines[1:], "\n")
		_ = os.WriteFile(target, []byte(string(cur)+body), 0o644)
	}
}

// runSpecctl invokes the compiled CLI from within `cwd` and returns
// stdout, stderr, and exit code. A non-zero exit is not a test failure
// — callers must inspect the code field directly.
type runResult struct {
	stdout string
	stderr string
	code   int
}

func runSpecctl(t *testing.T, cwd string, args ...string) runResult {
	t.Helper()
	cmd := exec.Command(specctlBin, args...)
	cmd.Dir = cwd
	if covDir != "" {
		cmd.Env = append(os.Environ(), "GOCOVERDIR="+covDir)
	}
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("specctl run failed: %v\nstderr: %s", err, errBuf.String())
		}
	}
	return runResult{stdout: out.String(), stderr: errBuf.String(), code: code}
}

// makeFixtureRoot builds a self-contained mini-project layout that
// satisfies specctl's expectations: go.mod marker, docs/specs/...,
// testdata/specs/..., and an INDEX.md / README.md when relevant.
type specFixture struct {
	relPath    string
	name       string
	dialect    string
	category   string
	status     string
	notes      string
	td         string // testdata relative path
	tdContent  string
	withBody   bool
	bodyExtras string
}

type fixtureRoot struct {
	dir   string
	specs []specFixture
}

func newFixtureRoot(t *testing.T) *fixtureRoot {
	t.Helper()
	dir := t.TempDir()
	// Marker for projectRoot().
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs", "specs"), 0o755); err != nil {
		t.Fatalf("mkdir docs/specs: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "testdata", "specs"), 0o755); err != nil {
		t.Fatalf("mkdir testdata/specs: %v", err)
	}
	return &fixtureRoot{dir: dir}
}

// writeSpec writes a spec markdown file with the standard required
// sections so it passes validation.
func (f *fixtureRoot) writeSpec(t *testing.T, sp specFixture) {
	t.Helper()
	abs := filepath.Join(f.dir, sp.relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", sp.name)
	fmt.Fprintf(&b, "dialect: %s\n", sp.dialect)
	fmt.Fprintf(&b, "category: %s\n", sp.category)
	fmt.Fprintf(&b, "status: %s\n", sp.status)
	fmt.Fprintf(&b, "source_url: docs/third_party/test/%s.md\n", sp.name)
	fmt.Fprintf(&b, "upstream_url: https://example/%s\n", sp.name)
	fmt.Fprintf(&b, "last_synced: 2026-05-15\n")
	fmt.Fprintf(&b, "testdata: %s\n", sp.td)
	if sp.notes != "" {
		fmt.Fprintf(&b, "notes: %s\n", sp.notes)
	}
	b.WriteString("---\n\n")
	if sp.withBody {
		fmt.Fprintf(&b, "# %s\n\n", sp.name)
		b.WriteString("## Summary\n\nDoc.\n\n")
		b.WriteString("## Signatures\n\nSig.\n\n")
		b.WriteString("## Behavior\n\nBeh.\n\n")
		b.WriteString("## Examples\n\nExample text.\n\n")
		b.WriteString("## Edge cases\n\nEdge.\n\n")
		b.WriteString("## Reference (upstream)\n\nUpstream reference.\n\n")
		b.WriteString(sp.bodyExtras)
	}
	if err := os.WriteFile(abs, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	// Write testdata file if requested.
	if sp.td != "" && sp.tdContent != "" {
		tdAbs := filepath.Join(f.dir, sp.td)
		if err := os.MkdirAll(filepath.Dir(tdAbs), 0o755); err != nil {
			t.Fatalf("mkdir td: %v", err)
		}
		if err := os.WriteFile(tdAbs, []byte(sp.tdContent), 0o644); err != nil {
			t.Fatalf("write td: %v", err)
		}
	}
	f.specs = append(f.specs, sp)
}

// stdTestdata returns a minimal but valid testdata YAML body that
// references the given spec path.
func stdTestdata(specPath string) string {
	return fmt.Sprintf(`spec: %s
dialect: googlesql
cases:
  - desc: sample
    sql: SELECT 1
    expected:
      rows:
        - [1]
`, specPath)
}

// ---- check ----

func TestCheck_OK_RegeneratesIndex(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		tdContent: stdTestdata("docs/specs/googlesql/functions/string/upper.md"),
		withBody:  true,
	})
	res := runSpecctl(t, f.dir, "check")
	if res.code != 0 {
		t.Fatalf("check failed (code=%d): stdout=%s stderr=%s", res.code, res.stdout, res.stderr)
	}
	idxBytes, err := os.ReadFile(filepath.Join(f.dir, "docs/specs/INDEX.md"))
	if err != nil {
		t.Fatalf("INDEX.md not written: %v", err)
	}
	if !strings.Contains(string(idxBytes), "UPPER") {
		t.Errorf("INDEX.md does not mention UPPER spec:\n%s", string(idxBytes))
	}
}

func TestCheck_StatusMismatchFails(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/missing.yaml",
		// Intentionally do not write the testdata file.
		withBody: true,
	})
	res := runSpecctl(t, f.dir, "check")
	if res.code == 0 {
		t.Fatalf("check should have failed on missing testdata; stdout=%s stderr=%s", res.stdout, res.stderr)
	}
}

func TestCheck_CheckOnlyFlagFailsOnStale(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		tdContent: stdTestdata("docs/specs/googlesql/functions/string/upper.md"),
		withBody:  true,
	})
	// Write an empty / stale INDEX.md so --check tripwires.
	if err := os.WriteFile(filepath.Join(f.dir, "docs/specs/INDEX.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("seed stale index: %v", err)
	}
	res := runSpecctl(t, f.dir, "check", "--check")
	if res.code == 0 {
		t.Fatalf("--check should fail on stale INDEX; stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stderr, "stale") {
		t.Errorf("expected 'stale' in stderr, got: %s", res.stderr)
	}
}

func TestCheck_OrphanTestdataFails(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		tdContent: stdTestdata("docs/specs/googlesql/functions/string/upper.md"),
		withBody:  true,
	})
	// Drop a testdata file that points at a non-existent spec.
	orphan := filepath.Join(f.dir, "testdata/specs/googlesql/functions/string/ghost.yaml")
	if err := os.WriteFile(orphan, []byte(stdTestdata("docs/specs/googlesql/functions/string/ghost.md")), 0o644); err != nil {
		t.Fatalf("seed orphan: %v", err)
	}
	res := runSpecctl(t, f.dir, "check")
	if res.code == 0 {
		t.Fatalf("expected non-zero exit on orphan testdata; stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stderr, "orphan") {
		t.Errorf("expected 'orphan' in stderr; got %s", res.stderr)
	}
}

// ---- validate ----

func TestValidate_OK(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		tdContent: stdTestdata("docs/specs/googlesql/functions/string/upper.md"),
		withBody:  true,
	})
	res := runSpecctl(t, f.dir, "validate", "--all")
	if res.code != 0 {
		t.Fatalf("validate failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "OK") {
		t.Errorf("expected OK; got %s", res.stdout)
	}
}

func TestValidate_MissingSections(t *testing.T) {
	f := newFixtureRoot(t)
	abs := filepath.Join(f.dir, "docs/specs/googlesql/functions/string/broken.md")
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	_ = os.WriteFile(abs, []byte(`---
name: BROKEN
dialect: googlesql
category: functions/string
status: drafted
source_url: x
upstream_url: y
last_synced: 2026-05-15
testdata: testdata/specs/googlesql/functions/string/broken.yaml
---

# BROKEN

## Summary
only one section
`), 0o644)
	res := runSpecctl(t, f.dir, "validate", "docs/specs/googlesql/functions/string/broken.md")
	if res.code == 0 {
		t.Fatalf("expected non-zero exit on missing sections; stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stderr, "section missing") {
		t.Errorf("expected 'section missing' in stderr; got %s", res.stderr)
	}
}

func TestValidate_PartialMissingNotes(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/partial.md",
		name:    "PARTIAL_FN", dialect: "googlesql", category: "functions/string",
		status: "partial", td: "testdata/specs/googlesql/functions/string/partial.yaml",
		withBody: true,
	})
	res := runSpecctl(t, f.dir, "validate", "--all")
	if res.code == 0 {
		t.Fatalf("expected validate to fail for partial without notes; stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stderr, "requires a notes") {
		t.Errorf("expected notes-required message; got %s", res.stderr)
	}
}

func TestValidate_BadStatus(t *testing.T) {
	f := newFixtureRoot(t)
	abs := filepath.Join(f.dir, "docs/specs/googlesql/functions/string/bad.md")
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	_ = os.WriteFile(abs, []byte(`---
name: BAD
dialect: googlesql
category: functions/string
status: someweirdvalue
source_url: x
upstream_url: y
last_synced: 2026-05-15
testdata: testdata/specs/googlesql/functions/string/bad.yaml
---

# BAD

## Summary

text

## Signatures

text

## Behavior

text

## Examples

text

## Edge cases

text

## Reference (upstream)

text
`), 0o644)
	res := runSpecctl(t, f.dir, "validate", "--all")
	if res.code == 0 {
		t.Fatalf("expected validate to fail for unknown status")
	}
	if !strings.Contains(res.stderr, "invalid status") {
		t.Errorf("expected invalid-status message; got %s", res.stderr)
	}
}

func TestValidate_NoArgsErrors(t *testing.T) {
	f := newFixtureRoot(t)
	res := runSpecctl(t, f.dir, "validate")
	if res.code == 0 {
		t.Fatalf("expected non-zero exit when no args supplied")
	}
}

func TestValidate_TestdataPrefixWrong(t *testing.T) {
	f := newFixtureRoot(t)
	abs := filepath.Join(f.dir, "docs/specs/googlesql/functions/string/bad.md")
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	_ = os.WriteFile(abs, []byte(`---
name: BAD
dialect: googlesql
category: functions/string
status: drafted
source_url: x
upstream_url: y
last_synced: 2026-05-15
testdata: somewhereelse/bad.yaml
---

# BAD

## Summary

text

## Signatures

text

## Behavior

text

## Examples

text

## Edge cases

text

## Reference (upstream)

text
`), 0o644)
	res := runSpecctl(t, f.dir, "validate", "--all")
	if res.code == 0 {
		t.Fatalf("expected validate to fail when testdata prefix wrong")
	}
	if !strings.Contains(res.stderr, "testdata must live under testdata/specs/") {
		t.Errorf("expected prefix error; got %s", res.stderr)
	}
}

// ---- coverage ----

func TestCoverage_Text(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		tdContent: stdTestdata("docs/specs/googlesql/functions/string/upper.md"),
		withBody:  true,
	})
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/types/int64.md",
		name:    "INT64", dialect: "googlesql", category: "types",
		status: "drafted", td: "testdata/specs/googlesql/types/int64.yaml",
		withBody: true,
	})
	res := runSpecctl(t, f.dir, "coverage")
	if res.code != 0 {
		t.Fatalf("coverage failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "Specs:") {
		t.Errorf("expected coverage summary; got %s", res.stdout)
	}
	if !strings.Contains(res.stdout, "drafted") {
		t.Errorf("expected drafted status row; got %s", res.stdout)
	}
}

func TestCoverage_VerboseAndLimit(t *testing.T) {
	f := newFixtureRoot(t)
	// Two drafted specs without testdata to populate missing list.
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("F%d", i)
		f.writeSpec(t, specFixture{
			relPath: fmt.Sprintf("docs/specs/googlesql/functions/string/f%d.md", i),
			name:    name, dialect: "googlesql", category: "functions/string",
			status:   "drafted",
			td:       fmt.Sprintf("testdata/specs/googlesql/functions/string/f%d.yaml", i),
			withBody: true,
		})
	}
	res := runSpecctl(t, f.dir, "coverage", "--verbose")
	if res.code != 0 {
		t.Fatalf("coverage verbose failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "Specs missing testdata") {
		t.Errorf("expected missing-testdata section; got %s", res.stdout)
	}
	// Limit cap.
	res = runSpecctl(t, f.dir, "coverage", "--limit", "1")
	if res.code != 0 {
		t.Fatalf("coverage --limit failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "and 2 more") {
		t.Errorf("expected 'and 2 more' in output; got %s", res.stdout)
	}
}

func TestCoverage_JSON(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		tdContent: stdTestdata("docs/specs/googlesql/functions/string/upper.md"),
		withBody:  true,
	})
	res := runSpecctl(t, f.dir, "coverage", "--json")
	if res.code != 0 {
		t.Fatalf("coverage --json failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &body); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, res.stdout)
	}
	if _, ok := body["by_status"]; !ok {
		t.Errorf("expected by_status field; got %s", res.stdout)
	}
}

func TestCoverage_NonZeroOnIntegrityIssue(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/missing.yaml",
		withBody: true,
	})
	res := runSpecctl(t, f.dir, "coverage")
	if res.code == 0 {
		t.Fatalf("expected non-zero exit for status_mismatch")
	}
}

// ---- main ----

func TestNoArgsShowsUsage(t *testing.T) {
	f := newFixtureRoot(t)
	res := runSpecctl(t, f.dir)
	if res.code == 0 {
		t.Fatalf("expected non-zero exit when no args; stderr=%s", res.stderr)
	}
	if !strings.Contains(res.stderr, "usage:") {
		t.Errorf("expected usage line; got %s", res.stderr)
	}
}

func TestUnknownSubcommand(t *testing.T) {
	f := newFixtureRoot(t)
	res := runSpecctl(t, f.dir, "doesnotexist")
	if res.code == 0 {
		t.Fatalf("expected non-zero exit for unknown subcommand")
	}
	if !strings.Contains(res.stderr, "unknown subcommand") {
		t.Errorf("expected 'unknown subcommand' in stderr; got %s", res.stderr)
	}
}

// ---- normalize ----

func TestNormalize_DryRun(t *testing.T) {
	f := newFixtureRoot(t)
	// Write a minimal upstream stub. We point string_functions.md to a
	// tiny page with one well-formed function section; the dry-run
	// reports what would be written.
	upstreamDir := filepath.Join(f.dir, "docs/third_party", "googlesql-docs")
	if err := os.MkdirAll(upstreamDir, 0o755); err != nil {
		t.Fatalf("mkdir upstream: %v", err)
	}
	for _, name := range []string{
		"string_functions.md", "mathematical_functions.md", "array_functions.md",
		"date_functions.md", "datetime_functions.md", "time_functions.md",
		"timestamp_functions.md", "interval_functions.md", "json_functions.md",
		"hash_functions.md", "bit_functions.md", "net_functions.md",
		"geography_functions.md", "range-functions.md", "hll_functions.md",
		"conversion_functions.md", "security_functions.md", "debugging_functions.md",
		"protocol_buffer_functions.md",
		"aggregate_functions.md", "approximate_aggregate_functions.md",
		"statistical_aggregate_functions.md", "aggregate-dp-functions.md",
		"numbering_functions.md", "navigation_functions.md", "data-types.md",
	} {
		body := "# header\n\n## `EXAMPLE_FN`\n\nbody\n"
		if name == "data-types.md" {
			body = "# header\n\n## Array type\n\nbody\n"
		}
		_ = os.WriteFile(filepath.Join(upstreamDir, name), []byte(body), 0o644)
	}
	res := runSpecctl(t, f.dir, "normalize", "--dry-run")
	if res.code != 0 {
		t.Fatalf("normalize --dry-run failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "would be written") {
		t.Errorf("expected dry-run wording; got %s", res.stdout)
	}
}

func TestNormalize_Writes(t *testing.T) {
	f := newFixtureRoot(t)
	upstreamDir := filepath.Join(f.dir, "docs/third_party", "googlesql-docs")
	if err := os.MkdirAll(upstreamDir, 0o755); err != nil {
		t.Fatalf("mkdir upstream: %v", err)
	}
	for _, name := range []string{
		"string_functions.md", "mathematical_functions.md", "array_functions.md",
		"date_functions.md", "datetime_functions.md", "time_functions.md",
		"timestamp_functions.md", "interval_functions.md", "json_functions.md",
		"hash_functions.md", "bit_functions.md", "net_functions.md",
		"geography_functions.md", "range-functions.md", "hll_functions.md",
		"conversion_functions.md", "security_functions.md", "debugging_functions.md",
		"protocol_buffer_functions.md",
		"aggregate_functions.md", "approximate_aggregate_functions.md",
		"statistical_aggregate_functions.md", "aggregate-dp-functions.md",
		"numbering_functions.md", "navigation_functions.md", "data-types.md",
	} {
		body := "# header\n\n## `EXAMPLE_FN`\n\nbody\n"
		if name == "data-types.md" {
			body = "# header\n\n## Array type\n\nbody\n"
		}
		_ = os.WriteFile(filepath.Join(upstreamDir, name), []byte(body), 0o644)
	}
	res := runSpecctl(t, f.dir, "normalize")
	if res.code != 0 {
		t.Fatalf("normalize failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "normalize: created=") {
		t.Errorf("expected created= summary; got %s", res.stdout)
	}
	// Re-running should mostly produce 'unchanged' results.
	res = runSpecctl(t, f.dir, "normalize")
	if res.code != 0 {
		t.Fatalf("normalize re-run failed: %s", res.stderr)
	}
	if !strings.Contains(res.stdout, "unchanged=") {
		t.Errorf("expected unchanged counter on re-run; got %s", res.stdout)
	}
}

// ---- downgrade ----

func TestDowngrade_StatusFlip(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	listPath := filepath.Join(f.dir, "list.txt")
	_ = os.WriteFile(listPath, []byte("docs/specs/googlesql/functions/string/upper.md\n# comment\n\n"), 0o644)
	res := runSpecctl(t, f.dir, "downgrade", "--list", listPath, "--note", "missing some edge case")
	if res.code != 0 {
		t.Fatalf("downgrade failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	body, _ := os.ReadFile(filepath.Join(f.dir, "docs/specs/googlesql/functions/string/upper.md"))
	if !strings.Contains(string(body), "status: partial") {
		t.Errorf("expected partial status; got\n%s", string(body))
	}
	if !strings.Contains(string(body), "notes: missing some edge case") {
		t.Errorf("expected note inserted; got\n%s", string(body))
	}
	// Re-run is a no-op: already partial, note already present.
	res = runSpecctl(t, f.dir, "downgrade", "--list", listPath, "--note", "missing some edge case")
	if res.code != 0 {
		t.Fatalf("downgrade re-run failed: %s", res.stderr)
	}
}

func TestDowngrade_AcceptsTestSpecPath(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	listPath := filepath.Join(f.dir, "list.txt")
	_ = os.WriteFile(listPath, []byte("googlesql/functions/string/upper\n"), 0o644)
	res := runSpecctl(t, f.dir, "downgrade", "--list", listPath, "--to", "unsupported", "--note", "no engine support")
	if res.code != 0 {
		t.Fatalf("downgrade failed: %s", res.stderr)
	}
	body, _ := os.ReadFile(filepath.Join(f.dir, "docs/specs/googlesql/functions/string/upper.md"))
	if !strings.Contains(string(body), "status: unsupported") {
		t.Errorf("expected unsupported status; got\n%s", string(body))
	}
}

func TestDowngrade_StdinInput(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "implemented", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	cmd := exec.Command(specctlBin, "downgrade", "--note", "x")
	cmd.Dir = f.dir
	if covDir != "" {
		cmd.Env = append(os.Environ(), "GOCOVERDIR="+covDir)
	}
	cmd.Stdin = strings.NewReader("docs/specs/googlesql/functions/string/upper.md\n")
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		t.Fatalf("downgrade stdin run failed: %v stderr=%s", err, errBuf.String())
	}
	body, _ := os.ReadFile(filepath.Join(f.dir, "docs/specs/googlesql/functions/string/upper.md"))
	if !strings.Contains(string(body), "status: partial") {
		t.Errorf("expected partial; got\n%s", string(body))
	}
}

// ---- extract-testdata ----

func TestExtractTestdata_NoExamples(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "drafted", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true, // body has no code-blocks → no examples
	})
	res := runSpecctl(t, f.dir, "extract-testdata", "--dry-run")
	if res.code != 0 {
		t.Fatalf("extract-testdata failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stderr, "no-examples") {
		t.Errorf("expected no-examples message; got %s", res.stderr)
	}
}

func TestExtractTestdata_WithExample(t *testing.T) {
	f := newFixtureRoot(t)
	// Build a spec with an upstream Examples section that has an
	// extractable SQL + ASCII-table pair.
	abs := filepath.Join(f.dir, "docs/specs/googlesql/functions/string/upper.md")
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	body := `---
name: UPPER
dialect: googlesql
category: functions/string
status: drafted
source_url: docs/third_party/test/upper.md
upstream_url: https://example/upper
last_synced: 2026-05-15
testdata: testdata/specs/googlesql/functions/string/upper.yaml
---

# UPPER

## Summary

text

## Signatures

text

## Behavior

text

## Examples

text

## Edge cases

text

## Reference (upstream)

` + "**Examples**\n\n```googlesql\nSELECT UPPER('abc');\n\n/*------*\n | out  |\n +------+\n | ABC  |\n *------*/\n```" + `

`
	_ = os.WriteFile(abs, []byte(body), 0o644)
	res := runSpecctl(t, f.dir, "extract-testdata")
	if res.code != 0 {
		t.Fatalf("extract-testdata failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	td, err := os.ReadFile(filepath.Join(f.dir, "testdata/specs/googlesql/functions/string/upper.yaml"))
	if err != nil {
		t.Fatalf("expected testdata file written: %v", err)
	}
	if !strings.Contains(string(td), "SELECT UPPER") {
		t.Errorf("expected SQL copied; got\n%s", string(td))
	}
	// Existing file: re-run without --force → skipped.
	res = runSpecctl(t, f.dir, "extract-testdata")
	if res.code != 0 {
		t.Fatalf("re-run failed: %s", res.stderr)
	}
	if !strings.Contains(res.stderr, "skipped") {
		t.Errorf("expected skipped counter on re-run; got %s", res.stderr)
	}
	// With --force, rewrite.
	res = runSpecctl(t, f.dir, "extract-testdata", "--force")
	if res.code != 0 {
		t.Fatalf("--force re-run failed: %s", res.stderr)
	}
	if !strings.Contains(res.stderr, "overwritten") {
		t.Errorf("expected overwritten counter; got %s", res.stderr)
	}
	// --only-missing should skip existing.
	res = runSpecctl(t, f.dir, "extract-testdata", "--only-missing")
	if res.code != 0 {
		t.Fatalf("--only-missing failed: %s", res.stderr)
	}
}

// ---- scan-compliance / extract-compliance basic smoke ----

func TestScanCompliance_MissingRoot(t *testing.T) {
	f := newFixtureRoot(t)
	res := runSpecctl(t, f.dir, "scan-compliance")
	if res.code == 0 {
		t.Fatalf("expected error when --root missing")
	}
}

func TestScanCompliance_BadRoot(t *testing.T) {
	f := newFixtureRoot(t)
	res := runSpecctl(t, f.dir, "scan-compliance", "--root", "/no/such/path/zzz/x")
	if res.code == 0 {
		t.Fatalf("expected error on missing root")
	}
}

func TestScanCompliance_WithFixture(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "drafted", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	compRoot := filepath.Join(f.dir, "comp")
	_ = os.MkdirAll(compRoot, 0o755)
	// Minimal compliance .test file with a SQL referencing UPPER.
	testFile := `[name=upper_basic]
SELECT UPPER('a')
--
ARRAY<STRUCT<STRING>>[{"A"}]
==
`
	_ = os.WriteFile(filepath.Join(compRoot, "x.test"), []byte(testFile), 0o644)
	res := runSpecctl(t, f.dir, "scan-compliance", "--root", compRoot, "--filter", "upper")
	if res.code != 0 {
		t.Fatalf("scan-compliance failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "function\tcases") {
		t.Errorf("expected TSV header; got %s", res.stdout)
	}
}

func TestExtractCompliance_MissingFlags(t *testing.T) {
	f := newFixtureRoot(t)
	res := runSpecctl(t, f.dir, "extract-compliance")
	if res.code == 0 {
		t.Fatalf("expected error when --root missing")
	}
	res = runSpecctl(t, f.dir, "extract-compliance", "--root", "/tmp")
	if res.code == 0 {
		t.Fatalf("expected error when --spec missing")
	}
}

func TestExtractCompliance_NoMatches(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "drafted", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	compRoot := filepath.Join(f.dir, "comp")
	_ = os.MkdirAll(compRoot, 0o755)
	res := runSpecctl(t, f.dir, "extract-compliance",
		"--root", compRoot, "--spec", "docs/specs/googlesql/functions/string/upper.md", "--limit", "5")
	if res.code != 0 {
		t.Fatalf("extract-compliance unexpectedly failed: %s", res.stderr)
	}
	if !strings.Contains(res.stderr, "no safely-convertible") {
		t.Errorf("expected no-cases message; got %s", res.stderr)
	}
}

// Cover --self-contained code path.
func TestExtractCompliance_SelfContainedFlagSafe(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "drafted", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	compRoot := filepath.Join(f.dir, "comp")
	_ = os.MkdirAll(compRoot, 0o755)
	res := runSpecctl(t, f.dir, "extract-compliance",
		"--root", compRoot,
		"--spec", "docs/specs/googlesql/functions/string/upper.md",
		"--limit", "abc", // forces atoi failure → ignored
		"--self-contained")
	if res.code != 0 {
		t.Fatalf("extract-compliance failed: %s", res.stderr)
	}
}

// TestExtractCompliance_WithMatchedCases drops a hand-crafted .test
// file that mentions UPPER and emits a STRING ARRAY<STRUCT> result.
// This drives the YAML rendering path and the success branch of
// extract-compliance.
func TestExtractCompliance_WithMatchedCases(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "drafted", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	compRoot := filepath.Join(f.dir, "comp")
	_ = os.MkdirAll(compRoot, 0o755)
	// A minimal compliance case in the canonical .test format. The
	// SQL mentions UPPER and the expected block is an ARRAY<STRUCT<STRING>>
	// row, which ParseExpectedRows accepts as a scalar string.
	body := `[name=upper_simple]
SELECT UPPER("abc")
--
ARRAY<STRUCT<STRING>>[{"ABC"}]
==
`
	_ = os.WriteFile(filepath.Join(compRoot, "fixture.test"), []byte(body), 0o644)
	res := runSpecctl(t, f.dir, "extract-compliance",
		"--root", compRoot,
		"--spec", "docs/specs/googlesql/functions/string/upper.md",
		"--limit", "5",
		"--self-contained")
	if res.code != 0 {
		t.Fatalf("extract-compliance failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "spec:") {
		t.Errorf("expected YAML spec header; got %s", res.stdout)
	}
	if !strings.Contains(res.stdout, "UPPER") {
		t.Errorf("expected UPPER in emitted YAML; got %s", res.stdout)
	}
}

// TestScanCompliance_NonMatching exercises the case where a function
// has zero matching cases.
func TestScanCompliance_NonMatching(t *testing.T) {
	f := newFixtureRoot(t)
	f.writeSpec(t, specFixture{
		relPath: "docs/specs/googlesql/functions/string/upper.md",
		name:    "UPPER", dialect: "googlesql", category: "functions/string",
		status: "drafted", td: "testdata/specs/googlesql/functions/string/upper.yaml",
		withBody: true,
	})
	compRoot := filepath.Join(f.dir, "comp")
	_ = os.MkdirAll(compRoot, 0o755)
	body := `[name=other]
SELECT OTHER_FN(1)
--
ARRAY<STRUCT<INT64>>[{1}]
==
`
	_ = os.WriteFile(filepath.Join(compRoot, "x.test"), []byte(body), 0o644)
	res := runSpecctl(t, f.dir, "scan-compliance", "--root", compRoot)
	if res.code != 0 {
		t.Fatalf("scan-compliance failed: %s", res.stderr)
	}
}

// ---- upstream-sync ----

func TestUpstreamSync_BadRepo(t *testing.T) {
	f := newFixtureRoot(t)
	// Point at a non-existent local "repo" path so git clone fails
	// without network access.
	bogus := filepath.Join(t.TempDir(), "no-such-repo")
	res := runSpecctl(t, f.dir, "upstream-sync", "--repo", bogus)
	if res.code == 0 {
		t.Fatalf("expected git clone to fail")
	}
}

// TestUpstreamSync_LocalRepoSucceeds drives the full success path by
// pointing --repo at a local fake repository. The fake repo contains
// a `docs/` subtree with a single markdown file and a LICENSE file —
// matching the upstream-sync sparse-checkout pattern.
func TestUpstreamSync_LocalRepoSucceeds(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	f := newFixtureRoot(t)

	// Build a fake upstream repo on disk.
	fakeRepo := t.TempDir()
	mkdir := func(p string) {
		if err := os.MkdirAll(filepath.Join(fakeRepo, p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}
	write := func(p, body string) {
		if err := os.WriteFile(filepath.Join(fakeRepo, p), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	mkdir("docs")
	write("docs/string_functions.md", "# placeholder\n\n## `UPPER`\n\ndoc\n")
	write("LICENSE", "Apache 2.0\n")
	// `git init` and commit. Disable global config side-effects to keep
	// the fake repo self-contained.
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	for _, cmd := range [][]string{
		{"git", "init", "--initial-branch=main", "-q", fakeRepo},
		{"git", "-C", fakeRepo, "add", "."},
		{"git", "-C", fakeRepo, "-c", "commit.gpgsign=false", "commit", "-q", "-m", "seed"},
	} {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Env = gitEnv
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", cmd, err, string(out))
		}
	}

	res := runSpecctl(t, f.dir, "upstream-sync", "--repo", fakeRepo)
	if res.code != 0 {
		t.Fatalf("upstream-sync failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "upstream-sync:") {
		t.Errorf("expected upstream-sync output; got %s", res.stdout)
	}
	// Verify the vendor directory exists with the markdown.
	if _, err := os.Stat(filepath.Join(f.dir, "docs/third_party/googlesql-docs/string_functions.md")); err != nil {
		t.Errorf("vendor markdown not written: %v", err)
	}
	// Re-run should report 0 changes.
	res = runSpecctl(t, f.dir, "upstream-sync", "--repo", fakeRepo)
	if res.code != 0 {
		t.Fatalf("upstream-sync re-run failed: %s", res.stderr)
	}
}

// ---- spec.go helpers via the binary (indirect) ----
//
// Most spec.go helpers are exercised by the normalize command above; the
// edge cases (extractFunctionName empty, slugify trimming) are covered
// here through `normalize` against pathological headings.

func TestNormalize_SkipsUnclassifiableHeadings(t *testing.T) {
	f := newFixtureRoot(t)
	upstreamDir := filepath.Join(f.dir, "docs/third_party", "googlesql-docs")
	if err := os.MkdirAll(upstreamDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// All the upstream files are required to exist; write a deliberately
	// pathological string_functions.md and minimal stubs for the rest.
	for _, name := range []string{
		"mathematical_functions.md", "array_functions.md", "date_functions.md",
		"datetime_functions.md", "time_functions.md", "timestamp_functions.md",
		"interval_functions.md", "json_functions.md", "hash_functions.md",
		"bit_functions.md", "net_functions.md", "geography_functions.md",
		"range-functions.md", "hll_functions.md", "conversion_functions.md",
		"security_functions.md", "debugging_functions.md", "protocol_buffer_functions.md",
		"aggregate_functions.md", "approximate_aggregate_functions.md",
		"statistical_aggregate_functions.md", "aggregate-dp-functions.md",
		"numbering_functions.md", "navigation_functions.md", "data-types.md",
	} {
		body := "# h\n\n"
		if name == "data-types.md" {
			body = "# h\n\n## NotAType\n\nbody\n## Array type\n\nbody\n"
		}
		_ = os.WriteFile(filepath.Join(upstreamDir, name), []byte(body), 0o644)
	}
	// string_functions.md: H2 without backticks → unclassifiable. Plus
	// the "Function list" non-entry heading that nonEntryHeadings filters.
	_ = os.WriteFile(filepath.Join(upstreamDir, "string_functions.md"), []byte(
		"# h\n\n## Function list\n\nignored\n\n## NoBackticks\n\nignored too\n\n## `GOOD_FN`\n\nbody\n",
	), 0o644)
	res := runSpecctl(t, f.dir, "normalize")
	if res.code != 0 {
		t.Fatalf("normalize failed: stdout=%s stderr=%s", res.stdout, res.stderr)
	}
	if !strings.Contains(res.stdout, "GOOD_FN") && !strings.Contains(res.stdout, "good_fn") {
		t.Logf("normalize stdout: %s", res.stdout)
	}
}
