---
name: APPENDS
dialect: bigquery
category: functions/time_series
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#appends
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#appends
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/time_series/appends.yaml
---

# APPENDS

## Summary
Table-valued function (Preview) that returns every row appended to a regular
BigQuery table within a given time range, along with `_CHANGE_TYPE` and
`_CHANGE_TIMESTAMP` pseudo-columns describing each insertion.

## Signatures
- `APPENDS(TABLE table, start_timestamp DEFAULT NULL, end_timestamp DEFAULT NULL)`

## Behavior
- The function appears in the `FROM` clause and returns a relation whose user
  columns mirror the input table's current schema, plus pseudo-columns
  `_CHANGE_TYPE` (`STRING`, always `'INSERT'`) and `_CHANGE_TIMESTAMP`
  (`TIMESTAMP`, the commit time of the appending transaction).
- `table` must be a regular BigQuery table and must be preceded by the keyword
  `TABLE`; clones, snapshots, views, materialized views, external tables, and
  wildcard tables are not supported.
- `start_timestamp` is a `TIMESTAMP` lower bound (inclusive); `NULL` means
  start from table creation. If the table was created later than
  `start_timestamp`, the actual creation time is used instead.
- `end_timestamp` is a `TIMESTAMP` upper bound and is exclusive; `NULL` means
  up to the query's start time. A future `end_timestamp` makes the query fail.
- Tracked operations are `CREATE TABLE`, `INSERT`, the appending side of
  `MERGE`, batch load jobs, and streaming ingestion. `UPDATE` and `DELETE`
  have no effect on the output, and a row's values reflect what was originally
  inserted, not later updates.
- Output is bounded by the table's time-travel window (default seven days,
  configurable to less); requesting `start_timestamp` earlier than that — or
  passing `NULL` when the table itself was created earlier than the window —
  returns an error.
- Output uses the table's current schema: a column added after the recorded
  inserts shows up as `NULL` for rows inserted before the column existed.

## Examples
```sql
-- Set up a table and append a few rows.
CREATE TABLE mydataset.Produce (product STRING, inventory INT64) AS (
  SELECT 'apples' AS product, 10 AS inventory);

INSERT INTO mydataset.Produce VALUES ('bananas', 20), ('carrots', 30);

-- Read the full append history within the time-travel window.
SELECT
  product,
  inventory,
  _CHANGE_TYPE  AS change_type,
  _CHANGE_TIMESTAMP AS change_time
FROM
  APPENDS(TABLE mydataset.Produce, NULL, NULL);
-- expected: three rows ('apples',10,'INSERT',...), ('bananas',20,'INSERT',...),
-- ('carrots',30,'INSERT',...)
```

```sql
-- After ALTER TABLE ... ADD COLUMN color, INSERT ('grapes',40,'purple'),
-- UPDATE inventory += 5, and DELETE WHERE product='bananas',
-- APPENDS still shows the original inserts for the apples/bananas/carrots rows
-- with color = NULL and inventory at insert time, plus the new grapes row.
SELECT product, inventory, color, _CHANGE_TYPE, _CHANGE_TIMESTAMP
FROM APPENDS(TABLE mydataset.Produce, NULL, NULL);
-- expected: bananas row still present (deletes ignored); inventory unchanged
-- by the UPDATE; color is NULL for rows inserted before the ADD COLUMN.
```

## Edge cases
- An error is returned if `start_timestamp` is earlier than allowed by the
  table's time-travel window, or if `start_timestamp` is `NULL` and the table
  was created earlier than the window.
- An error is returned if `end_timestamp` is in the future.
- Deletions are not reflected; a deleted row remains in the `APPENDS` output
  for as long as the time-travel window covers its insertion.
- Updates to a row produce no change event; only the original inserted values
  are returned.
- Copying a table resets the append history: calling `APPENDS` on the copy
  reports every row as inserted at the copy's creation time.
- The output schema follows the current table schema; columns that exist in
  the schema but did not exist when the row was inserted appear as `NULL`.
- Cannot be invoked inside a multi-statement transaction; rows that are
  appended and then updated or deleted within the same multi-statement
  transaction may not be captured.
- Partition pseudo-columns for ingestion-time partitioned tables
  (`_PARTITIONTIME`, `_PARTITIONDATE`) are not included in the output.
- Feature is in Preview and subject to the Pre-GA Offerings Terms; behaviour
  and availability may change.

## Reference (upstream)

See the upstream BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#appends>
