---
name: EXTERNAL_QUERY
dialect: bigquery
category: functions/federated
status: unsupported
notes: |
  Executes SQL against an external RDBMS via BigQuery's connection-resource plumbing. googlesqlite is a pure-Go local engine and has no foreign-connection facility; revisit only if the consumer ships a federated stub.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/federated_query_functions
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/federated_query_functions
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/federated/external_query.yaml
---

# EXTERNAL_QUERY

## Summary

Executes a query on an external database through a BigQuery connection
resource and returns the results as a temporary table that can be used
inside a GoogleSQL query (typically in a `FROM` clause).

## Signatures

- `EXTERNAL_QUERY('connection_id', '''external_database_query''' [, 'options'])`

## Arguments

- `connection_id`: STRING. The ID of the connection resource that
  configures the link between BigQuery and the external database.
  When no default project is configured, use the fully qualified form
  `projects/PROJECT_ID/locations/LOCATION/connections/CONNECTION_ID`
  (for example, `projects/example-project/locations/us/connections/sql-bq`).
- `external_database_query`: STRING. The SQL statement to run against
  the external database, written in that database's dialect.
- `options`: STRING (optional). A JSON map of case-sensitive option
  names and values, e.g. `'{"default_type_for_decimal_columns":"numeric"}'`.
  Supported keys:
  - `default_type_for_decimal_columns`: one of `"float64"`, `"numeric"`,
    `"bignumeric"`, or `"string"`. Controls how MySQL `DECIMAL` and
    PostgreSQL `NUMERIC` columns are mapped into BigQuery. Defaults to
    BigQuery `NUMERIC` when omitted.
  - `query_execution_priority`: one of `"low"`, `"medium"`, or `"high"`.
    Spanner-only; sets the execution priority of the external query.
    Defaults to `"medium"`.

## Return type

A BigQuery table whose columns are produced by mapping the external
database's column types to GoogleSQL types.

## Behavior

- The external query runs on the remote database and its result set is
  materialised as a temporary table that the surrounding BigQuery
  statement can consume.
- Column types from the external database are converted to GoogleSQL
  types according to the documented MySQL-to-BigQuery and
  PostgreSQL-to-BigQuery mappings (see the upstream "Data type
  mappings" tables for the full list).
- The function is normally used inside a `FROM` clause and may be
  joined against native BigQuery tables.
- It can also be used to read external metadata, e.g. by querying
  `information_schema.tables` or `information_schema.columns` on the
  external database.
- Row ordering from the external query is not preserved: even if the
  external statement contains `ORDER BY`, BigQuery does not guarantee
  that the federated result rows are returned in that order.
- When a view that uses `EXTERNAL_QUERY` is shared across projects, the
  fully qualified `projects/.../connections/...` form must be used or
  the wrong project may be selected.

## Examples

```sql
-- Join external Postgres results with a native BigQuery table.
SELECT
  c.customer_id,
  c.name,
  SUM(t.amount) AS total_revenue,
  rq.first_order_date
FROM customers AS c
INNER JOIN transaction_fact AS t
  ON c.customer_id = t.customer_id
LEFT OUTER JOIN EXTERNAL_QUERY(
  'connection_id',
  '''SELECT customer_id, MIN(order_date) AS first_order_date
     FROM orders GROUP BY customer_id'''
) AS rq
  ON rq.customer_id = c.customer_id
GROUP BY c.customer_id, c.name, rq.first_order_date;

-- Read metadata from the external database.
SELECT *
FROM EXTERNAL_QUERY(
  'connection_id',
  '''SELECT * FROM information_schema.tables'''
);

SELECT *
FROM EXTERNAL_QUERY(
  'connection_id',
  '''SELECT * FROM information_schema.columns WHERE table_name='x';'''
);

-- ORDER BY in the external query does NOT order BigQuery output.
SELECT *
FROM EXTERNAL_QUERY(
  'connection_id',
  '''SELECT * FROM customers AS c ORDER BY c.customer_id'''
);
```

## Edge cases

- If the external query produces a column whose type has no supported
  BigQuery mapping, the federated query fails immediately. Cast the
  offending column to a supported MySQL/PostgreSQL type inside the
  external query to recover.
- BigQuery `NUMERIC` has a smaller value range than MySQL `DECIMAL` and
  PostgreSQL `NUMERIC`; use `default_type_for_decimal_columns` to map
  to `BIGNUMERIC`, `FLOAT64`, or `STRING` when the default range is
  insufficient.
- PostgreSQL types without a BigQuery counterpart (for example
  `money`, `path`, `uuid`, `box`) are unsupported.
- MySQL `TIMESTAMP` values are always retrieved in UTC, regardless of
  the caller's session timezone.
- BigQuery `TIME` covers `00:00:00`-`23:59:59` only, which is narrower
  than MySQL `TIME`'s `-838:59:59`-`838:59:59` range.
- The omitted-default-project form of `connection_id` (a bare ID) can
  resolve to the wrong project when used inside a shared view; prefer
  the fully qualified path in that case.
- `query_execution_priority` is only honoured for Spanner connections.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/federated_query_functions#external_query>
for the canonical description, full data-type mapping tables, and any
later additions. Upstream prose is intentionally not redistributed
here.
