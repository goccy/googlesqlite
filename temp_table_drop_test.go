package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestTempTableDroppedInScript covers a script that drops its own temp
// table. BigQuery drops a script's remaining temp tables when the script
// ends; one the script already dropped is simply gone. The end-of-script
// cleanup used to delete the spec a second time and fail the whole
// script with "failed to find table spec from map", after every
// statement had run.
func TestTempTableDroppedInScript(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=temp_table_dropped_in_script")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for _, tc := range []struct {
		name   string
		script string
		gone   []string // temp tables that must not outlive the script
	}{
		{"CTAS then DROP", "CREATE TEMP TABLE a AS SELECT 1 AS x; DROP TABLE a", []string{"a"}},
		{"column list then DROP", "CREATE TEMP TABLE b (x INT64); DROP TABLE b", []string{"b"}},
		{
			"DROP then re-create under the same name",
			"CREATE TEMP TABLE c AS SELECT 1 AS x; DROP TABLE c; CREATE TEMP TABLE c AS SELECT 2 AS x",
			[]string{"c"},
		},
		{
			"one dropped, one left for the cleanup",
			"CREATE TEMP TABLE d AS SELECT 1 AS x; CREATE TEMP TABLE e AS SELECT 1 AS x; DROP TABLE d",
			[]string{"d", "e"},
		},
	} {
		if _, err := conn.ExecContext(ctx, tc.script); err != nil {
			t.Errorf("%s: %s: %v", tc.name, tc.script, err)
			continue
		}
		for _, name := range tc.gone {
			if _, err := conn.ExecContext(ctx, "SELECT * FROM "+name); err == nil {
				t.Errorf("%s: temp table %s outlived the script", tc.name, name)
			}
		}
	}
}
