---
name: GAP_FILL
dialect: bigquery
category: functions/time_series
status: implemented
notes: |
  Implemented as a SQL pre-rewrite in internal/gap_fill.go,
  invoked from analyzer.go before the upstream parser sees the
  query. The lowered SQL builds bucket boundaries via
  GENERATE_TIMESTAMP_ARRAY against per-(partition?) min/max
  bounds, LEFT JOINs the input source on (ts, partitioning
  columns), then applies the value-column fill methods:
  - null: passes the joined value through unchanged.
  - locf: COUNT(value) OVER (ORDER BY ts) tags every row with
    its last-non-null group id; MAX(value) within that group
    surfaces the carried value. (The classic IGNORE-NULLS
    idiom without engine support for the modifier.)
  - linear: same group-id trick forward and backward, then a
    fractional interpolation between the previous and next
    non-null using TIMESTAMP_DIFF over the ratio.
  partitioning_columns, origin, value_columns, ignore_null_values
  are all parsed; TABLE name and (subquery) input shapes both
  accepted.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#gap_fill
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#gap_fill
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/time_series/gap_fill.yaml
---

# GAP_FILL

## Summary
Table-valued function that finds and fills gaps in a time series, producing
one row per fixed-width time bucket and using a configurable gap-filling
method to populate value columns where source rows are missing.

## Signatures
- `GAP_FILL(TABLE time_series_table, time_series_column, bucket_width [, partitioning_columns => value] [, value_columns => value] [, origin => value] [, ignore_null_values => { TRUE | FALSE }])`
- `GAP_FILL((time_series_subquery), time_series_column, bucket_width [, partitioning_columns => values] [, value_columns => value] [, origin => value] [, ignore_null_values => { TRUE | FALSE }])`

## Behavior
- `time_series_column` names the column carrying time points, and must be
  of type `DATE`, `DATETIME`, or `TIMESTAMP`; `origin` (when supplied) must
  be the same type as `time_series_column`.
- `bucket_width` is an `INTERVAL` value that defines the fixed width of the
  output time buckets, matching the type domain of `time_series_column`.
- `partitioning_columns` is a named argument of type `ARRAY<STRING>` listing
  zero or more column names that identify independent time series; gap
  filling is applied per partition, with the same column-type rules as
  `PARTITION BY`.
- `value_columns` is a named argument of type `ARRAY<STRUCT<STRING, STRING>>`
  whose elements are `(column_name, gap_filling_method)` pairs; each
  `column_name` must come from the input and may appear at most once. The
  gap-filling method is one of: `null` (default; fill missing values with
  `NULL`), `linear` (linear interpolation; column must be numeric), or
  `locf` (last-observation-carried-forward).
- When `value_columns` is omitted, the `null` gap-filling method is applied
  to every non-time, non-partitioning column.
- `origin` is an optional named `DATE` / `DATETIME` / `TIMESTAMP` argument
  that anchors the bucket grid; buckets expand in both directions from it.
  When omitted, the default origin matches the time column type:
  `DATE '1950-01-01'`, `DATETIME '1950-01-01 00:00:00'`, or
  `TIMESTAMP '1950-01-01 00:00:00'`.
- `ignore_null_values` is a `BOOL` named argument that defaults to `TRUE`,
  causing `NULL` values in the input to be skipped during gap filling; set
  it to `FALSE` to include them.
- When multiple input rows fall into the same bucket and carry equal
  values, the function condenses them into a single output row.
- The function returns a `TABLE`.

## Examples
```sql
-- locf gap-filling
CREATE TEMP TABLE device_data AS
SELECT * FROM UNNEST(
  ARRAY<STRUCT<device_id INT64, time DATETIME, signal INT64, state STRING>>[
    STRUCT(1, DATETIME '2023-11-01 09:34:01', 74, 'INACTIVE'),
    STRUCT(2, DATETIME '2023-11-01 09:36:00', 77, 'ACTIVE'),
    STRUCT(3, DATETIME '2023-11-01 09:37:00', 78, 'ACTIVE'),
    STRUCT(4, DATETIME '2023-11-01 09:38:01', 80, 'ACTIVE')
  ]);

SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'locf')]
)
ORDER BY time;
-- expected: (09:35,74) (09:36,77) (09:37,78) (09:38,78)
```

```sql
-- linear gap-filling
SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'linear')]
)
ORDER BY time;
-- expected: (09:35,75) (09:36,77) (09:37,78) (09:38,80)
```

```sql
-- null gap-filling
SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'null')]
)
ORDER BY time;
-- expected: (09:35,NULL) (09:36,77) (09:37,78) (09:38,NULL)
```

```sql
-- ignore_null_values defaults to TRUE: NULL inputs are skipped before fill
CREATE TEMP TABLE device_data AS
SELECT * FROM UNNEST(
  ARRAY<STRUCT<device_id INT64, time DATETIME, signal INT64, state STRING>>[
    STRUCT(1, DATETIME '2023-11-01 09:34:01', 74, 'INACTIVE'),
    STRUCT(2, DATETIME '2023-11-01 09:36:00', 77, 'ACTIVE'),
    STRUCT(3, DATETIME '2023-11-01 09:37:00', NULL, 'ACTIVE'),
    STRUCT(4, DATETIME '2023-11-01 09:38:01', 80, 'ACTIVE')
  ]);

SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'linear')]
)
ORDER BY time;
-- expected: (09:35,75) (09:36,77) (09:37,78) (09:38,80)
```

```sql
-- ignore_null_values => FALSE: NULL inputs participate in fill
SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'linear')],
  ignore_null_values => FALSE
)
ORDER BY time;
-- expected: (09:35,75) (09:36,77) (09:37,NULL) (09:38,NULL)
```

```sql
-- value_columns omitted: 'null' method applied to every non-time column
SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE
)
ORDER BY time;
-- expected: (09:35,NULL,NULL,NULL) (09:36,2,77,'ACTIVE')
--           (09:37,3,79,'ACTIVE') (09:38,NULL,NULL,NULL)
```

```sql
-- partitioning_columns: gap fill applied per partition
SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  partitioning_columns => ['device_id'],
  value_columns => [('signal', 'locf')]
)
ORDER BY device_id;
-- expected: one row per (bucket, device_id) with locf fill within partition
```

```sql
-- Multiple value columns each with their own method
SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'linear'), ('state', 'locf')]
)
ORDER BY time;
-- expected: (09:35,75,'ACTIVE') (09:36,77,'INACTIVE')
--           (09:37,78,'INACTIVE') (09:38,78,'ACTIVE') (09:39,80,'ACTIVE')
```

```sql
-- Custom origin shifts the bucket grid
SELECT *
FROM GAP_FILL(
  TABLE device_data,
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'null')],
  origin => DATETIME '2023-11-01 09:30:01'
)
ORDER BY time;
-- expected: (09:34:01,74) (09:35:01,NULL) (09:36:01,NULL)
--           (09:37:01,NULL) (09:38:01,NULL) (09:39:01,80)
```

```sql
-- Subquery form
SELECT *
FROM GAP_FILL(
  (SELECT * FROM UNNEST(
    ARRAY<STRUCT<device_id INT64, time DATETIME, signal INT64, state STRING>>[
      STRUCT(1, DATETIME '2023-11-01 09:34:01', 74, 'INACTIVE'),
      STRUCT(2, DATETIME '2023-11-01 09:36:00', 77, 'ACTIVE'),
      STRUCT(3, DATETIME '2023-11-01 09:37:00', 78, 'ACTIVE'),
      STRUCT(4, DATETIME '2023-11-01 09:38:01', 80, 'ACTIVE')
    ])),
  ts_column => 'time',
  bucket_width => INTERVAL 1 MINUTE,
  value_columns => [('signal', 'linear')]
)
ORDER BY time;
-- expected: (09:35,75) (09:36,77) (09:37,78) (09:38,80)
```

## Edge cases
- The `linear` method requires the target column to be a numeric data
  type; non-numeric columns must use `null` or `locf`.
- A column may appear at most once in `value_columns`.
- With `ignore_null_values => TRUE` (default), `locf` and `linear` carry
  the previous or next non-`NULL` value, so output value columns are never
  `NULL` except at the edges; with `ignore_null_values => FALSE`, `locf`
  yields `NULL` when the previous value is `NULL`, and `linear` yields
  `NULL` when either endpoint of the interpolated segment is `NULL`.
- For the `null` method, an input `NULL` always emits `NULL` regardless of
  `ignore_null_values`.
- `time_series_column` and `origin` must share the same data type
  (`DATE`, `DATETIME`, or `TIMESTAMP`).
- Multiple input rows that fall into the same bucket and carry the same
  values are condensed into a single output row.

## Reference (upstream)

See the upstream BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#gap_fill>
