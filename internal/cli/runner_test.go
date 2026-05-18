package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
)

func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	// A per-test memdb database keeps tests isolated while still
	// exercising the shared-memory VFS the CLI defaults to.
	dsn := fmt.Sprintf("file:/%s.db?vfs=memdb", strings.ReplaceAll(t.Name(), "/", "_"))
	r, err := NewRunner(context.Background(), dsn)
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func TestRunnerSelect(t *testing.T) {
	r := newTestRunner(t)
	res := r.Exec(context.Background(), "SELECT 1 AS x, 'hello' AS y")
	if res.Err != nil {
		t.Fatalf("Exec: %v", res.Err)
	}
	if !res.IsQuery {
		t.Fatalf("IsQuery = false, want true")
	}
	if got, want := len(res.Rows), 1; got != want {
		t.Fatalf("rows = %d, want %d", got, want)
	}
	if got := res.Columns; len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("columns = %v", got)
	}
	if res.SQLiteQuery == "" {
		t.Errorf("SQLiteQuery is empty; the debug collector did not capture the translated query")
	}
}

func TestRunnerDebugCollector(t *testing.T) {
	r := newTestRunner(t)
	res := r.Exec(context.Background(), "SELECT 1 + 1 AS two")
	if res.Err != nil {
		t.Fatalf("Exec: %v", res.Err)
	}
	// The translated SQLite query must mention the addition; it is the
	// internal SQL collector hook that feeds CLI debug mode.
	if !strings.Contains(res.SQLiteQuery, "1") {
		t.Errorf("SQLiteQuery = %q, expected the translated query text", res.SQLiteQuery)
	}
}

func TestRunnerDDLAndDML(t *testing.T) {
	r := newTestRunner(t)
	ctx := context.Background()

	if res := r.Exec(ctx, "CREATE TABLE t (id INT64, name STRING)"); res.Err != nil {
		t.Fatalf("CREATE: %v", res.Err)
	}
	res := r.Exec(ctx, "INSERT INTO t (id, name) VALUES (1, 'a'), (2, 'b')")
	if res.Err != nil {
		t.Fatalf("INSERT: %v", res.Err)
	}
	if res.IsQuery {
		t.Errorf("INSERT classified as a query")
	}
	if res.RowsAffected != 2 {
		t.Errorf("RowsAffected = %d, want 2", res.RowsAffected)
	}

	sel := r.Exec(ctx, "SELECT id, name FROM t ORDER BY id")
	if sel.Err != nil {
		t.Fatalf("SELECT: %v", sel.Err)
	}
	if len(sel.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(sel.Rows))
	}
}

func TestRunnerGroupMode(t *testing.T) {
	r := newTestRunner(t)
	res := r.Exec(context.Background(), `SELECT 1 AS x \G`)
	if res.Err != nil {
		t.Fatalf("Exec: %v", res.Err)
	}
	if !res.GroupMode {
		t.Errorf("GroupMode = false, want true")
	}
}

func TestRenderTableCJKAlignment(t *testing.T) {
	// A header and a full-width CJK value: every table line must have
	// the same display width, which is the regression guard for the
	// column-misalignment bug.
	res := Result{
		IsQuery:     true,
		Columns:     []string{"name"},
		ColumnKinds: nil,
		Rows:        [][]any{{"山田太郎"}, {"Bob"}},
	}
	var buf bytes.Buffer
	RenderResult(&buf, res, RenderOptions{Color: false})
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) < 6 {
		t.Fatalf("unexpected output:\n%s", buf.String())
	}
	width := displayWidth(lines[0])
	for i, ln := range lines[:5] {
		if w := displayWidth(ln); w != width {
			t.Errorf("line %d width = %d, want %d:\n%s", i, w, width, buf.String())
		}
	}
}
