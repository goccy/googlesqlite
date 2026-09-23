package googlesqlite_test

import (
	"database/sql"
	"testing"
)

// TestSQLUDFBodyNestedSubqueryColumnIDs covers a SQL UDF body with a
// scalar subquery nested inside an EXISTS subquery. Both subqueries
// produce a column the analyzer names $col1; the body must still be
// formatted with distinct column references.
func TestSQLUDFBodyNestedSubqueryColumnIDs(t *testing.T) {
	const body = `EXISTS (SELECT 1 FROM UNNEST([1]) AS _f WHERE COALESCE((SELECT c = 'x'), FALSE))`
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{
			name:  "temp function",
			query: `CREATE TEMP FUNCTION w3(c STRING) RETURNS BOOL AS (` + body + `); SELECT w3('x')`,
			want:  true,
		},
		{
			name:  "temp function false",
			query: `CREATE TEMP FUNCTION w3(c STRING) RETURNS BOOL AS (` + body + `); SELECT w3('y')`,
			want:  false,
		},
		{
			name:  "persistent function",
			query: `CREATE FUNCTION w3(c STRING) RETURNS BOOL AS (` + body + `); SELECT w3('x')`,
			want:  true,
		},
		{
			name:  "no parameter reference",
			query: `CREATE TEMP FUNCTION w3(c STRING) RETURNS BOOL AS (EXISTS (SELECT 1 FROM UNNEST([1]) AS _f WHERE (SELECT TRUE))); SELECT w3('x')`,
			want:  true,
		},
		{
			name: "called from a table function",
			query: `CREATE TEMP FUNCTION w3(c STRING) RETURNS BOOL AS (` + body + `);
CREATE TEMP TABLE FUNCTION t(c STRING) AS (SELECT w3(c) AS v);
SELECT v FROM t('x')`,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("googlesqlite", ":memory:?_test=udf_subquery_column_ids")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			var got bool
			if err := db.QueryRow(tt.query).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
