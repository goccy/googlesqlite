package googlesqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestQuotedDottedNamePath resolves a project- and dataset-qualified
// table and TVF through every spelling BigQuery accepts for the same
// path: one quoted identifier per component, the whole path in one
// quoted identifier, and a quoted `project.dataset` prefix followed by
// the object name.
func TestQuotedDottedNamePath(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=quoted_dotted_name_path")
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

	for _, stmt := range []string{
		"CREATE TABLE `my-proj`.`my_ds`.`one_row` AS SELECT 1 AS k",
		"CREATE TABLE FUNCTION `my-proj`.`my_ds`.`plus_one`(k INT64) AS (SELECT k + 1 AS n)",
	} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	for _, path := range []string{
		"`my-proj`.`my_ds`.`%s`",
		"`my-proj`.my_ds.%s",
		"`my-proj.my_ds.%s`",
		"`my-proj.my_ds`.%s",
	} {
		t.Run("table "+fmt.Sprintf(path, "one_row"), func(t *testing.T) {
			var k int64
			if err := conn.QueryRowContext(ctx, "SELECT k FROM "+fmt.Sprintf(path, "one_row")).Scan(&k); err != nil {
				t.Fatalf("query: %v", err)
			}
			if k != 1 {
				t.Errorf("k = %d, want 1", k)
			}
		})
		t.Run("TVF "+fmt.Sprintf(path, "plus_one"), func(t *testing.T) {
			var n int64
			if err := conn.QueryRowContext(ctx, "SELECT n FROM "+fmt.Sprintf(path, "plus_one")+"(1)").Scan(&n); err != nil {
				t.Fatalf("call: %v", err)
			}
			if n != 2 {
				t.Errorf("n = %d, want 2", n)
			}
		})
	}
}
