package googlesqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// MERGE with a subquery in USING. BigQuery accepts any query as the
// source; dbt-bigquery's incremental materialization always emits one
// (`USING (SELECT * FROM <tmp>) AS DBT_INTERNAL_SOURCE ON FALSE ...`) and
// the snowplow_unified manifest update uses a cross join of two
// subqueries. Before the fix every such MERGE failed with
// "MERGE: source must be a single-table reference".
func TestMergeSubquerySource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		setup []string
		merge string
		check string
		want  []string
	}{
		{
			// dbt insert_overwrite-style incremental: replace the
			// partitions the source covers, keep the rest.
			name: "dbt_on_false_delete_by_source_then_insert",
			setup: []string{
				"CREATE TABLE dest (d INT64, v STRING)",
				"CREATE TABLE tmp (d INT64, v STRING)",
				"INSERT INTO dest VALUES (1, 'old1'), (2, 'old2'), (3, 'keep')",
				"INSERT INTO tmp VALUES (1, 'new1'), (2, 'new2')",
			},
			merge: `MERGE dest AS DBT_INTERNAL_DEST
USING (SELECT * FROM tmp) AS DBT_INTERNAL_SOURCE
ON FALSE
WHEN NOT MATCHED BY SOURCE AND DBT_INTERNAL_DEST.d IN (1, 2) THEN DELETE
WHEN NOT MATCHED THEN INSERT (d, v) VALUES (d, v)`,
			check: "SELECT FORMAT('%d=%s', d, v) FROM dest ORDER BY d",
			want:  []string{"1=new1", "2=new2", "3=keep"},
		},
		{
			// snowplow_unified incremental manifest update.
			name: "snowplow_manifest_cross_join_union_all",
			setup: []string{
				"CREATE TABLE manifest (model STRING, last_success INT64)",
				"CREATE TABLE events (ts INT64)",
				"INSERT INTO manifest VALUES ('a', 5), ('b', 50)",
				"INSERT INTO events VALUES (10), (20)",
			},
			merge: `MERGE manifest m
USING (
  SELECT b.model, a.last_success
  FROM (SELECT MAX(ts) AS last_success FROM events) a,
       (SELECT 'a' AS model UNION ALL SELECT 'b' AS model UNION ALL SELECT 'c' AS model) b
  WHERE a.last_success IS NOT NULL
) s
ON m.model = s.model
WHEN MATCHED THEN UPDATE SET last_success = GREATEST(m.last_success, s.last_success)
WHEN NOT MATCHED THEN INSERT (model, last_success) VALUES (model, last_success)`,
			check: "SELECT FORMAT('%s=%d', model, last_success) FROM manifest ORDER BY model",
			want:  []string{"a=20", "b=50", "c=20"},
		},
		{
			// Filtered and computed source columns, matched DELETE.
			name: "filtered_source_matched_delete_and_update",
			setup: []string{
				"CREATE TABLE t (k INT64, v INT64)",
				"CREATE TABLE s (k INT64, v INT64, drop_it BOOL)",
				"INSERT INTO t VALUES (1, 1), (2, 2), (3, 3)",
				"INSERT INTO s VALUES (1, 10, FALSE), (2, 20, TRUE), (3, 30, FALSE), (4, 40, FALSE)",
			},
			merge: `MERGE t
USING (SELECT k, v * 2 AS v, drop_it FROM s WHERE k != 3) src
ON t.k = src.k
WHEN MATCHED AND src.drop_it THEN DELETE
WHEN MATCHED THEN UPDATE SET v = src.v
WHEN NOT MATCHED BY TARGET THEN INSERT (k, v) VALUES (src.k, src.v)`,
			check: "SELECT FORMAT('%d=%d', k, v) FROM t ORDER BY k",
			want:  []string{"1=20", "3=3", "4=80"},
		},
		{
			// An empty source: nothing matched, BY SOURCE fires for all.
			name: "empty_source_not_matched_by_source_update",
			setup: []string{
				"CREATE TABLE t (k INT64, v INT64)",
				"CREATE TABLE s (k INT64, v INT64)",
				"INSERT INTO t VALUES (1, 1), (2, 2)",
			},
			merge: `MERGE t USING (SELECT * FROM s WHERE FALSE) src ON t.k = src.k
WHEN NOT MATCHED BY SOURCE THEN UPDATE SET v = v + 100`,
			check: "SELECT FORMAT('%d=%d', k, v) FROM t ORDER BY k",
			want:  []string{"1=101", "2=102"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db, err := sql.Open("googlesqlite", fmt.Sprintf(":memory:?_test=merge_subquery_%s", tc.name))
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
			for _, q := range tc.setup {
				if _, err := conn.ExecContext(ctx, q); err != nil {
					t.Fatalf("exec %q: %v", q, err)
				}
			}
			if _, err := conn.ExecContext(ctx, tc.merge); err != nil {
				t.Fatalf("MERGE: %v", err)
			}
			rows, err := conn.QueryContext(ctx, tc.check)
			if err != nil {
				t.Fatalf("check query: %v", err)
			}
			defer rows.Close()
			var got []string
			for rows.Next() {
				var s string
				if err := rows.Scan(&s); err != nil {
					t.Fatalf("Scan: %v", err)
				}
				got = append(got, s)
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("got %v; want %v", got, tc.want)
			}
		})
	}
}
