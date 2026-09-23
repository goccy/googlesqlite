package googlesqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	_ "github.com/goccy/googlesqlite"
)

// TestTVFAnyTableArgument covers table-valued functions with a
// templated ANY TABLE parameter (and ANY TYPE scalars alongside it).
// The first two functions are the ANY TABLE Examples from
// docs/third_party/googlesql-docs/table-functions.md, created without
// TEMP so they outlive the CREATE statement.
func TestTVFAnyTableArgument(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=tvf_any_table")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	for _, stmt := range []string{
		"CREATE TABLE Customers (id INT64, creation_time INT64)",
		"INSERT INTO Customers VALUES (1, 10), (2, 20), (3, 30)",
		"CREATE TABLE Other (name STRING)",
		"INSERT INTO Other VALUES ('a'), ('b')",
		`CREATE TABLE FUNCTION CustomerCreationTimeRange(
		    min_creation_time INT64,
		    max_creation_time INT64,
		    SelectedCustomers ANY TABLE)
		  AS
		    SELECT *
		    FROM SelectedCustomers
		    WHERE creation_time >= min_creation_time
		    AND creation_time <= max_creation_time`,
		`CREATE TABLE FUNCTION MyFunction(
		     first_value ANY TYPE,
		     second_value ANY TYPE,
		     MyInputTable ANY TABLE)
		   AS
		     SELECT *
		     FROM MyInputTable
		     WHERE first_value > second_value`,
		"CREATE TABLE FUNCTION CountRows(t ANY TABLE) AS (SELECT COUNT(*) AS n FROM t)",
		// A fixed-schema TABLE<...> parameter goes through the same
		// relation-argument path without being templated.
		"CREATE TABLE FUNCTION IdsAtLeast(lo INT64, t TABLE<id INT64>) AS (SELECT id FROM t WHERE id >= lo)",
	} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}

	for _, tc := range []struct {
		name  string
		query string
		want  [][]any
	}{
		{
			name:  "ANY TABLE with scalar arguments filters the input table",
			query: "SELECT id, creation_time FROM CustomerCreationTimeRange(15, 30, TABLE Customers) ORDER BY id",
			want:  [][]any{{int64(2), int64(20)}, {int64(3), int64(30)}},
		},
		{
			name:  "ANY TYPE predicate true returns every input row",
			query: "SELECT id FROM MyFunction(2, 1, TABLE Customers) ORDER BY id",
			want:  [][]any{{int64(1)}, {int64(2)}, {int64(3)}},
		},
		{
			name:  "ANY TYPE predicate false returns no rows",
			query: "SELECT id FROM MyFunction('a', 'b', TABLE Customers)",
			want:  nil,
		},
		{
			name:  "the same function accepts a table with a different schema",
			query: "SELECT name FROM MyFunction(2, 1, TABLE Other) ORDER BY name",
			want:  [][]any{{"a"}, {"b"}},
		},
		{
			name:  "ANY TABLE given a table",
			query: "SELECT n FROM CountRows(TABLE Customers)",
			want:  [][]any{{int64(3)}},
		},
		{
			name:  "TABLE<...> parameter",
			query: "SELECT id FROM IdsAtLeast(2, TABLE Customers) ORDER BY id",
			want:  [][]any{{int64(2)}, {int64(3)}},
		},
		{
			name:  "templated TVF call joined with another table",
			query: "SELECT c.id, r.n FROM Customers AS c CROSS JOIN CountRows(TABLE Other) AS r ORDER BY c.id",
			want:  [][]any{{int64(1), int64(2)}, {int64(2), int64(2)}, {int64(3), int64(2)}},
		},
		{
			name:  "ANY TABLE given a subquery",
			query: "SELECT n FROM CountRows((SELECT x FROM UNNEST([1, 2, 3, 4]) AS x WHERE x > 1))",
			want:  [][]any{{int64(3)}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := conn.QueryContext(ctx, tc.query)
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			defer rows.Close()
			cols, err := rows.Columns()
			if err != nil {
				t.Fatal(err)
			}
			var got [][]any
			for rows.Next() {
				vals := make([]any, len(cols))
				ptrs := make([]any, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					t.Fatal(err)
				}
				got = append(got, vals)
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("rows: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestTVFAnyTableArgumentAcrossConnections creates a templated TVF on
// one connection and calls it on another, which reloads the TVF from
// the persisted catalog rather than the in-memory one.
func TestTVFAnyTableArgumentAcrossConnections(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", "file:"+filepath.Join(t.TempDir(), "tvf.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	creator, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := creator.ExecContext(ctx,
		"CREATE TABLE FUNCTION CountRows(t ANY TABLE) AS (SELECT COUNT(*) AS n FROM t)"); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION: %v", err)
	}

	caller, err := db.Conn(ctx) // creator is still checked out, so this is a new connection
	if err != nil {
		t.Fatal(err)
	}
	runtime.GC() // the templated TVF handle must survive collection too
	time.Sleep(10 * time.Millisecond)
	defer caller.Close()
	defer creator.Close()
	var n int64
	if err := caller.QueryRowContext(ctx,
		"SELECT n FROM CountRows((SELECT x FROM UNNEST([1, 2, 3]) AS x))").Scan(&n); err != nil {
		t.Fatalf("call on a second connection: %v", err)
	}
	if n != 3 {
		t.Errorf("n = %d, want 3", n)
	}
}
