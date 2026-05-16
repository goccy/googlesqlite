// Declarative spec-test runtime: walks the YAML cases under
// testdata/specs/ (alongside the spec markdown in docs/specs/) and
// runs each as a Go subtest against a fresh in-memory googlesqlite
// connection. Entry point: TestSpec.

package googlesqlite_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/specmeta"
)

// TestSpec runs every testdata YAML under testdata/specs/ as a Go
// subtest. Use `go test -run 'TestSpec/<path>'` to scope to a
// category or a single spec.
func TestSpec(t *testing.T) {
	runSpecSuite(t)
}

// runSpecSuite discovers every testdata YAML file rooted at the
// project's testdata/specs/ directory and runs each case as a Go
// subtest. It fails the parent test only if a discovery / I/O error
// prevents running any cases at all; individual case failures are
// reported through t.Errorf.
func runSpecSuite(t *testing.T) {
	t.Helper()
	root, err := projectRoot()
	if err != nil {
		t.Fatalf("locate project root: %v", err)
	}
	testdataRoot := filepath.Join(root, "testdata", "specs")

	count := 0
	walkErr := filepath.WalkDir(testdataRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		rel, err := filepath.Rel(testdataRoot, path)
		if err != nil {
			return err
		}
		count++
		runFile(t, root, rel, path)
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk testdata: %v", walkErr)
	}
	if count == 0 {
		t.Skip("no testdata YAML files under testdata/specs/")
	}
}

func runFile(t *testing.T, root, rel, path string) {
	t.Helper()
	td, err := specmeta.LoadTestdata(root, filepath.Join("testdata", "specs", rel))
	if err != nil {
		t.Errorf("load %s: %v", rel, err)
		return
	}
	if td.Dialect == "" {
		t.Errorf("%s: dialect field missing", rel)
		return
	}

	// The spec status acts as the test gate. Spec authors keep testdata
	// even for specs that don't fully match upstream yet (so the gap is
	// documented in YAML rather than hidden); we only assert the
	// expected rows when the spec actually claims `implemented` /
	// `tested`. Lower statuses run the SQL but do not fail on
	// divergence — failures there are tracked through `specctl
	// coverage` instead.
	gateAsserts := true
	if td.Spec != "" {
		if fm, err := loadSpecFrontmatter(root, td.Spec); err == nil {
			switch fm.Status {
			case specmeta.StatusImplemented, specmeta.StatusTested:
				gateAsserts = true
			default:
				gateAsserts = false
			}
		}
	}

	subName := strings.TrimSuffix(filepath.ToSlash(rel), ".yaml")
	t.Run(subName, func(t *testing.T) {
		t.Parallel()
		for i, c := range td.Cases {
			name := c.Desc
			if name == "" {
				name = fmt.Sprintf("case_%d", i)
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				if !gateAsserts {
					t.Skip("spec status is not implemented/tested; case retained as documentation only")
				}
				runCase(t, td.Dialect, c)
			})
		}
	})
}

func loadSpecFrontmatter(root, specPath string) (specmeta.Frontmatter, error) {
	full := filepath.Join(root, filepath.FromSlash(specPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return specmeta.Frontmatter{}, err
	}
	return specmeta.ParseFrontmatter(data)
}

func runCase(t *testing.T, dialect string, c specmeta.Case) {
	t.Helper()
	if c.SQL == "" {
		t.Fatalf("case has empty sql")
	}

	dsn := ":memory:"
	if dialect != "" && dialect != "googlesql" {
		// Reserved for future per-dialect DSN routing. Today only the
		// googlesql baseline runs; bigquery/spanner specs will start
		// returning a Skip until their analyzer wiring lands.
		dsn = fmt.Sprintf(":memory:?dialect=%s", dialect)
	}
	db, err := sql.Open("googlesqlite", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	// Pin every statement of a single case to ONE driver connection
	// so setup-side catalog mutations (CREATE TABLE, etc.) are
	// visible to the query that follows. Without this, sql.DB's
	// pool may hand the query a fresh connection whose catalog
	// never saw the setup.
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	for _, name := range c.RegisterProtos {
		bytes, ok := testProtoBundles[name]
		if !ok {
			t.Fatalf("register_protos: unknown bundle %q (defined in internal/spectest/protos.go)", name)
		}
		if err := registerProtoOnConn(conn, bytes); err != nil {
			t.Fatalf("RegisterProto(%q): %v", name, err)
		}
	}

	for _, stmt := range c.Setup {
		if _, err := conn.ExecContext(context.Background(), stmt); err != nil {
			t.Fatalf("setup failed: %v\nsetup: %s", err, stmt)
		}
	}

	bindArgs := buildBindArgs(c)
	rows, queryErr := conn.QueryContext(context.Background(), c.SQL, bindArgs...)
	if c.Expected.Error != nil {
		// Some runtime errors only surface once the driver starts
		// streaming rows (e.g. ERROR(msg), divide-by-zero from a
		// non-folded expression). Drain to elicit rows.Err() before
		// reporting.
		if queryErr == nil && rows != nil {
			for rows.Next() {
				// drain
			}
			queryErr = rows.Err()
		}
		assertExpectedError(t, c.Expected.Error, queryErr)
		if rows != nil {
			rows.Close()
		}
		return
	}
	if queryErr != nil {
		t.Fatalf("query failed: %v\nsql: %s", queryErr, c.SQL)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("columns: %v", err)
	}

	var got [][]any
	for rows.Next() {
		buf := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range buf {
			ptrs[i] = &buf[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, buf)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v\nsql: %s", err, c.SQL)
	}

	want := c.Expected.Rows
	if c.Nondeterministic {
		// Inherently non-deterministic case (e.g. DP-with-noise):
		// only verify the query executed without error. Neither
		// values nor row count are asserted, because DP thresholding
		// can drop entire groups in addition to the per-row noise
		// budget. The `expected:` block stays as documentation of
		// the docs' specific realization.
		_ = want
		return
	}
	if c.Expected.Unordered {
		if !rowsEqualUnordered(got, want) {
			t.Errorf("rows mismatch (unordered)\n  got:  %s\n  want: %s\n  sql:  %s",
				formatRows(got), formatRows(want), c.SQL)
		}
		return
	}
	if !rowsEqual(got, want) {
		t.Errorf("rows mismatch\n  got:  %s\n  want: %s\n  sql:  %s",
			formatRows(got), formatRows(want), c.SQL)
	}
}

// rowsEqualUnordered compares two row sets as multisets — every row in
// `got` must match exactly one row in `want`, and vice versa, but the
// per-row order is irrelevant. Use only for cases where the SQL's
// ORDER BY leaves the order implementation-defined.
func rowsEqualUnordered(got, want [][]any) bool {
	if len(got) != len(want) {
		return false
	}
	matched := make([]bool, len(want))
	for _, g := range got {
		found := false
		for j, w := range want {
			if matched[j] {
				continue
			}
			if len(g) != len(w) {
				continue
			}
			eq := true
			for k := range g {
				if !valueEqual(g[k], w[k]) {
					eq = false
					break
				}
			}
			if eq {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func assertExpectedError(t *testing.T, want *specmeta.Error, got error) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected error containing %q, got nil", want.Contains)
	}
	if want.Contains != "" && !strings.Contains(got.Error(), want.Contains) {
		t.Errorf("error %q does not contain %q", got.Error(), want.Contains)
	}
}

// buildBindArgs materialises the case's NamedArgs / Args fields into
// a single args slice suitable for db.QueryContext. NamedArgs become
// sql.Named entries (insertion-order is undefined for maps, but the
// driver routes them by name so position is irrelevant). Args become
// positional values in their declared order.
func buildBindArgs(c specmeta.Case) []any {
	var out []any
	for name, val := range c.NamedArgs {
		out = append(out, sql.Named(name, val))
	}
	out = append(out, c.Args...)
	return out
}

// rowsEqual compares two row sets after normalising types that
// database/sql, YAML, and SQLite represent differently. It treats
// integers, floats, byte-slices as a STRING, and nils with care.
func rowsEqual(got, want [][]any) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if len(got[i]) != len(want[i]) {
			return false
		}
		for j := range got[i] {
			if !valueEqual(got[i][j], want[i][j]) {
				return false
			}
		}
	}
	return true
}

func valueEqual(got, want any) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	g := normalize(got)
	w := normalize(want)
	switch gv := g.(type) {
	case int64:
		if wv, ok := w.(int64); ok {
			return gv == wv
		}
		if wv, ok := w.(float64); ok {
			return float64(gv) == wv
		}
		if wv, ok := w.(bool); ok {
			// SQLite returns BOOL as int64(0/1); YAML parses
			// `true`/`false` as bool. Bridge the two so testdata
			// can use natural booleans.
			return (gv != 0) == wv
		}
		if wv, ok := w.(string); ok {
			// Auto-extracted testdata renders INT64 ASCII-cell
			// values as quoted strings ("3", "1234", ...). Parse
			// and compare numerically when the string is an integer
			// literal; otherwise compare as the formatted string.
			if i, err := strconv.ParseInt(wv, 10, 64); err == nil {
				return gv == i
			}
			return strconv.FormatInt(gv, 10) == wv
		}
	case float64:
		if wv, ok := w.(float64); ok {
			return floatNear(gv, wv)
		}
		if wv, ok := w.(int64); ok {
			return floatNear(gv, float64(wv))
		}
		if wv, ok := w.(string); ok {
			// Auto-extracted testdata serialises numeric ASCII-table
			// cells as quoted strings ('4.71238898038469', '0', ...).
			// Treat them as numeric when the string parses as a float.
			if f, err := strconv.ParseFloat(wv, 64); err == nil {
				return floatNear(gv, f)
			}
		}
	case string:
		if wv, ok := w.(string); ok {
			if gv == wv {
				return true
			}
			// When both sides parse as a WKT geography literal, compare
			// after collapsing the optional space between the type
			// keyword and the open paren. OGC SFS uses a space
			// (`POINT (0 0)`), BigQuery's display strips it
			// (`POINT(0 0)`); both encode the same geometry.
			if gn, wn := normaliseWKTLiteral(gv), normaliseWKTLiteral(wv); gn != "" && wn != "" && gn == wn {
				return true
			}
			return false
		}
	case bool:
		if wv, ok := w.(bool); ok {
			return gv == wv
		}
		if wv, ok := w.(int64); ok {
			return gv == (wv != 0)
		}
	case []byte:
		if wv, ok := w.([]byte); ok {
			return string(gv) == string(wv)
		}
		if wv, ok := w.(string); ok {
			return string(gv) == wv
		}
	case []any:
		// Array comparison — element-wise via valueEqual.
		if wv, ok := w.([]any); ok {
			if len(gv) != len(wv) {
				return false
			}
			for i := range gv {
				if !valueEqual(gv[i], wv[i]) {
					return false
				}
			}
			return true
		}
		// Auto-extracted testdata renders ARRAY ASCII-cell values as a
		// single quoted string like `[1, 2, 3]` or `[(1, 'a'), (2, 'b')]`.
		// Format the Go slice as the canonical BigQuery display form
		// and compare against the string verbatim. Tolerates either
		// `'x'` (single-quoted strings) or `x` (raw) for STRING members.
		if wv, ok := w.(string); ok {
			// Normalise both sides: ignore quoting and STRUCT delimiter
			// style. See normaliseArrayString for the rationale.
			return normaliseArrayString(formatArrayBQ(gv)) == normaliseArrayString(wv)
		}
		return false
	case map[string]any:
		// Struct / object comparison — key-wise.
		wv, ok := w.(map[string]any)
		if !ok {
			return false
		}
		if len(gv) != len(wv) {
			return false
		}
		for k, gx := range gv {
			wx, ok := wv[k]
			if !ok {
				return false
			}
			if !valueEqual(gx, wx) {
				return false
			}
		}
		return true
	}
	return false
}

// formatArrayBQ renders a Go slice as BigQuery's ARRAY display form
// (`[v1, v2, v3]`). Primitive elements render via fmt.Sprint;
// nested arrays and structs recurse through scalarRenderBQ.
func formatArrayBQ(s []any) string {
	parts := make([]string, 0, len(s))
	for _, v := range s {
		parts = append(parts, scalarRenderBQ(v))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// scalarRenderBQ renders one element of an ARRAY (or a sub-element
// of a nested ARRAY) the way BigQuery's ASCII-cell display would.
// Strings are written verbatim (no quoting); the comparison side
// strips quote characters before comparing.
func scalarRenderBQ(v any) string {
	if v == nil {
		return "NULL"
	}
	switch x := v.(type) {
	case []any:
		return formatArrayBQ(x)
	case string:
		return x
	case []byte:
		return string(x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	}
	return fmt.Sprint(v)
}

// normaliseArrayString reduces a BigQuery ARRAY / STRUCT display
// string to a canonical form that ignores quoting and delimiter
// style. Strip single quotes (`'a'` → `a`), normalise `(` / `{` → `[`
// and `)` / `}` → `]`, drop `<identifier>:` field-name prefixes
// (`{xmin:1, ymin:2}` → `[1, 2]`), and collapse any whitespace
// introduced by the upstream rendering. This lets the runner compare
// a Go slice rendered as `[v1, v2]` against a yaml string written in
// either `[(1, 'a'), (2, 'b')]` or `[{1, 'a'}, {2, 'b'}]` form, or
// against a STRUCT cell rendered as `{xmin:1, ymin:2, ...}`.
func normaliseArrayString(s string) string {
	// Collapse `POINT (` -> `POINT(` and similar WKT type-spacing
	// variants before bracket folding, so a WKT literal nested
	// inside an array compares equal regardless of which display
	// convention either side chose.
	s = collapseWKTSpaces(s)
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "(", "[")
	s = strings.ReplaceAll(s, ")", "]")
	s = strings.ReplaceAll(s, "{", "[")
	s = strings.ReplaceAll(s, "}", "]")
	s = stripStructFieldLabels(s)
	s = stripStructTrailingLabels(s)
	// Drop element-separator commas — BigQuery's CLI display of a
	// repeated proto field omits them (entries are space-separated),
	// whereas a Go slice render uses comma + space. Treat both as
	// equivalent.
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.Join(strings.Fields(s), " ")
	// Drop whitespace adjacent to brackets so `[ x ]` and `[x]`
	// compare equal — different docs sources use both conventions for
	// the same proto map / struct payloads.
	s = strings.ReplaceAll(s, "[ ", "[")
	s = strings.ReplaceAll(s, " ]", "]")
	// Strip the redundant outer `[` … `]` wrapper if and only if the
	// opening bracket truly pairs with the closing one (depth returns
	// to zero only at the end). This matches the case where one side
	// renders an ARRAY<X> as `[<elem1>, <elem2>]` while the other
	// simply concatenates the element renderings (e.g. proto map
	// fields in BigQuery's CLI: `{ ... } { ... }`).
	if isWrappedInBrackets(s) {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

// isWrappedInBrackets reports whether s starts with `[`, ends with
// `]`, and the opening bracket's closing match is exactly the final
// character (no premature depth-zero in between). Used so that
// strings like `[a]` or `[[a] [b]]` are recognised as wrapped but
// `[a] [b]` is left alone.
func isWrappedInBrackets(s string) bool {
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return false
	}
	depth := 0
	for i, c := range s {
		switch c {
		case '[':
			depth++
		case ']':
			depth--
		}
		if depth == 0 && i < len(s)-1 {
			return false
		}
	}
	return depth == 0
}

// stripStructTrailingLabels rewrites `[1 A1, a alias_inferred]` to
// `[1, a]` by removing any whitespace+identifier suffix that
// immediately follows a value (not a separator) and does not start a
// new bracket group. This matches BigQuery's ARRAY_ZIP cell display
// where each struct member is rendered as `<value> <field_name>` —
// the inverse of the compliance suite's value-only convention — so
// both forms compare equal after normalization. The "preceded by a
// value char" rule keeps element separators like `, a` from being
// mistaken for trailing field-name annotations.
func stripStructTrailingLabels(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		c := s[i]
		if c == ' ' && i > 0 && isValueChar(s[i-1]) && i+1 < len(s) && isIdentStart(s[i+1]) {
			j := i + 1
			for j < len(s) && isIdentCont(s[j]) {
				j++
			}
			if j < len(s) && (s[j] == ',' || s[j] == ']') {
				ident := s[i+1 : j]
				upper := strings.ToUpper(ident)
				if upper != "NULL" && upper != "TRUE" && upper != "FALSE" {
					i = j
					continue
				}
			}
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// isValueChar reports whether c can plausibly end a value token —
// digit, letter, single-character quote / unquote, or closing
// bracket of a nested value. Used by stripStructTrailingLabels to
// distinguish ` <ident>` after a value (a trailing field annotation
// to strip) from ` <ident>` after a separator (the start of the
// next array element to preserve).
func isValueChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		c == '_' || c == ']' || c == '.' || c == '-' || c == '+'
}

// stripStructFieldLabels rewrites `[xmin:1, ymin:2]` to `[1, 2]`
// by removing any `<identifier>:` prefix that immediately precedes a
// value. Recognises identifiers containing letters, digits, and `_`.
func stripStructFieldLabels(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		c := s[i]
		if isIdentStart(c) {
			j := i + 1
			for j < len(s) && isIdentCont(s[j]) {
				j++
			}
			if j < len(s) && s[j] == ':' && (j+1 >= len(s) || s[j+1] != ':') {
				i = j + 1
				continue
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isIdentCont(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

// wktTypeNames enumerates the OGC Simple Features Specification
// geometry type keywords that the driver's WKT renderer can emit.
// Ordered long-to-short so a prefix scan picks `MULTIPOINT` before
// `POINT`.
var wktTypeNames = []string{
	"GEOMETRYCOLLECTION",
	"MULTIPOLYGON",
	"MULTILINESTRING",
	"MULTIPOINT",
	"POLYGON",
	"LINESTRING",
	"POINT",
}

// normaliseWKTLiteral returns a canonical form of `s` when it
// contains one or more WKT geometry literals, with any optional
// space between a type keyword and the following `(` collapsed.
// Returns the empty string when `s` does not begin with a WKT type
// keyword — callers fall back to strict string comparison in that
// case. Recurses through nested constructs so
// `GEOMETRYCOLLECTION (POINT (0 0), LINESTRING (1 2, 2 1))`
// normalises to
// `GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1))`.
//
// Also folds the empty-geography display variants — `<TYPE> EMPTY`
// for any WKT type — to the canonical `GEOMETRYCOLLECTION EMPTY`
// BigQuery prints, and reorders MULTIPOINT vertices into a lex-
// sorted canonical set so the per-engine S2 ordering does not
// break the comparison.
func normaliseWKTLiteral(s string) string {
	t := strings.TrimSpace(s)
	if !looksLikeWKT(t) {
		return ""
	}
	t = collapseWKTSpaces(t)
	t = canonicaliseEmptyWKT(t)
	t = canonicaliseMultiPointWKT(t)
	return t
}

// canonicaliseEmptyWKT rewrites any `<TYPE> EMPTY` to the canonical
// `GEOMETRYCOLLECTION EMPTY` form BigQuery's CLI uses regardless of
// the empty geometry's declared type.
func canonicaliseEmptyWKT(s string) string {
	for _, name := range wktTypeNames {
		if strings.EqualFold(s, name+" EMPTY") {
			return "GEOMETRYCOLLECTION EMPTY"
		}
	}
	return s
}

// canonicaliseMultiPointWKT sorts the vertices of a MULTIPOINT
// literal lexicographically so the two engines' internal vertex
// orderings (S2's cell-ID order vs the input order) compare equal.
// Untouched for other geometry types where order is semantically
// meaningful.
func canonicaliseMultiPointWKT(s string) string {
	const prefix = "MULTIPOINT("
	if !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, ")") {
		return s
	}
	inner := s[len(prefix) : len(s)-1]
	parts := strings.Split(inner, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	sort.Strings(parts)
	return prefix + strings.Join(parts, ", ") + ")"
}

func looksLikeWKT(s string) bool {
	for _, name := range wktTypeNames {
		if strings.HasPrefix(s, name) {
			return true
		}
	}
	return false
}

// collapseWKTSpaces scans `s` and removes any single ASCII space
// that appears between a WKT type keyword and the following `(`.
// Other characters and whitespace are passed through unchanged.
func collapseWKTSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		matched := false
		for _, name := range wktTypeNames {
			if strings.HasPrefix(s[i:], name) {
				j := i + len(name)
				// Skip a single space, if present, before `(`.
				if j < len(s) && s[j] == ' ' && j+1 < len(s) && s[j+1] == '(' {
					b.WriteString(name)
					b.WriteByte('(')
					i = j + 2
					matched = true
					break
				}
				b.WriteString(name)
				i = j
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// floatNear compares two float64 values with a 1e-9 tolerance
// (absolute or relative). NaNs compare equal.
func floatNear(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	diff := math.Abs(a - b)
	if diff <= 1e-9 {
		return true
	}
	scale := math.Max(math.Abs(a), math.Abs(b))
	return diff/scale <= 1e-9
}

// normalize converts the result of yaml.Unmarshal-into-any (which uses
// int / float64 / string / bool / nil) and database/sql Scan-into-any
// (int64 / float64 / string / []byte / bool / time.Time / nil) into a
// common shape.
func normalize(v any) any {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint64:
		return int64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	case bool, string, nil:
		return x
	case []byte:
		return x
	default:
		return v
	}
}

func formatRows(r [][]any) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, row := range r {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('[')
		for j, v := range row {
			if j > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%v(%T)", v, v)
		}
		b.WriteByte(']')
	}
	b.WriteByte(']')
	return b.String()
}

// projectRoot walks up from this file's location to find go.mod.
//
// Tests run with cwd set to the package directory, but the testdata
// tree is at the repository root. Caller info gives a stable anchor
// regardless of where `go test` is invoked from.
func projectRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for range 8 {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("go.mod not found above runner.go")
}
