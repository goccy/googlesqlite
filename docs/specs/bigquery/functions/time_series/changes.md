---
name: CHANGES
dialect: bigquery
category: functions/time_series
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#changes
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#changes
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/time_series/changes.yaml
---

# CHANGES

## Summary
Table-valued function that returns the rows changed in a base table during a
half-open time range. Used to incrementally consume change-data-capture (CDC)
output for tables that have change history enabled.

## Signatures
- `CHANGES(TABLE table_name, start_timestamp, end_timestamp)`

## Behavior
- The function is a table-valued function; it appears in the `FROM` clause and
  returns a relation whose user columns mirror the base table's schema, plus
  CDC pseudo-columns describing each change.
- The base table must have its `enable_change_history` table option set to
  `TRUE`; otherwise the query returns an error.
- `start_timestamp` is exclusive and `end_timestamp` is inclusive; the
  function reports rows changed in the half-open interval
  `(start_timestamp, end_timestamp]`.
- Both timestamp arguments must be `TIMESTAMP` values and `start_timestamp`
  must be less than or equal to `end_timestamp`.
- `end_timestamp` cannot be in the future; the typical use is
  `CURRENT_TIMESTAMP() - INTERVAL N MINUTE` to allow time for change metadata
  to materialize.
- `start_timestamp` cannot be earlier than the table's time-travel retention
  window (default 7 days); the table option `max_staleness` and CDC retention
  control how far back history is available.
- Each output row carries pseudo-columns: `_CHANGE_TYPE` (`STRING`, one of
  `INSERT`, `UPDATE`, or `DELETE`) and `_CHANGE_TIMESTAMP` (`TIMESTAMP`)
  identifying when the change was applied.
- The captured operations include `INSERT`, `UPDATE`, `DELETE`, `MERGE`,
  `CREATE TABLE` writes, batch loads, streaming ingestion, `TRUNCATE TABLE`,
  and jobs with `writeDisposition = WRITE_TRUNCATE`, as well as individual
  partition deletions.
- Querying the function requires `bigquery.tables.getData` on the base table;
  row-level access policies further restrict which historical rows the caller
  can observe.

## Examples
```sql
-- Read all changes that occurred in the last 10 minutes, ending one minute ago
SELECT *
FROM CHANGES(
  TABLE mydataset.mytable,
  TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 11 MINUTE),
  TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL  1 MINUTE)
);
-- expected: one row per change with user columns plus _CHANGE_TYPE and _CHANGE_TIMESTAMP
```

```sql
-- Count inserts vs. deletes since a known checkpoint
SELECT _CHANGE_TYPE, COUNT(*) AS n
FROM CHANGES(
  TABLE mydataset.orders,
  TIMESTAMP '2026-04-01 00:00:00 UTC',
  TIMESTAMP '2026-05-01 00:00:00 UTC'
)
GROUP BY _CHANGE_TYPE;
-- expected: rows like ('INSERT', 1234), ('UPDATE', 56), ('DELETE', 7)
```

## Edge cases
- Returns an error if the base table does not have `enable_change_history =
  TRUE` enabled.
- Returns an error if `start_timestamp` falls outside the table's available
  change-history / time-travel window.
- Returns an error if `end_timestamp` is in the future or precedes
  `start_timestamp`.
- `NULL` for either timestamp argument produces an error; both bounds must be
  concrete `TIMESTAMP` values.
- Rows fully written and then deleted within the same window may surface
  multiple change events with distinct `_CHANGE_TIMESTAMP` values.
- Row-level security policies filter the visible change history; users without
  override permission only see changes affecting rows they can normally read.
- Enabling change history adds storage and compute cost proportional to the
  volume of mutations on the table.

## Reference (upstream)

See the upstream BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#changes>
