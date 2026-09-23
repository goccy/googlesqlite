package googlesqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestQuotedDottedUDFNamePath resolves a project- and dataset-qualified
// scalar UDF through every spelling BigQuery accepts for the same path,
// both when it is created and when it is called, including from inside
// a TVF body. `project.dataset`.f used to fail with "Function not found".
func TestQuotedDottedUDFNamePath(t *testing.T) {
	t.Parallel()
	paths := []string{
		"`my-proj`.`my_ds`.`%s`",
		"`my-proj`.my_ds.%s",
		"`my-proj.my_ds.%s`",
		"`my-proj.my_ds`.%s",
	}
	for ci, createPath := range paths {
		t.Run("created as "+fmt.Sprintf(createPath, "f"), func(t *testing.T) {
			t.Parallel()
			db, err := sql.Open("googlesqlite", fmt.Sprintf(":memory:?_test=quoted_udf_name_path_%d", ci))
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

			if _, err := conn.ExecContext(ctx, "CREATE FUNCTION "+fmt.Sprintf(createPath, "plus_one")+"(x INT64) RETURNS INT64 AS (x + 1)"); err != nil {
				t.Fatalf("CREATE FUNCTION: %v", err)
			}
			for _, callPath := range paths {
				call := fmt.Sprintf(callPath, "plus_one")
				var n int64
				if err := conn.QueryRowContext(ctx, "SELECT "+call+"(1)").Scan(&n); err != nil {
					t.Errorf("SELECT %s(1): %v", call, err)
				} else if n != 2 {
					t.Errorf("SELECT %s(1) = %d, want 2", call, n)
				}
			}
			tvf := fmt.Sprintf("`my-proj`.`my_ds`.`tf_%d`", ci)
			if _, err := conn.ExecContext(ctx, "CREATE TABLE FUNCTION "+tvf+"(k INT64) AS (SELECT `my-proj.my_ds`.plus_one(k) AS n)"); err != nil {
				t.Fatalf("CREATE TABLE FUNCTION calling the UDF: %v", err)
			}
			var n int64
			if err := conn.QueryRowContext(ctx, "SELECT n FROM "+tvf+"(1)").Scan(&n); err != nil {
				t.Errorf("TVF call: %v", err)
			} else if n != 2 {
				t.Errorf("TVF call = %d, want 2", n)
			}
		})
	}
}
