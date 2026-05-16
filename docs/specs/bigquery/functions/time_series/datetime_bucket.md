---
name: DATETIME_BUCKET
dialect: bigquery
category: functions/time_series
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#datetime_bucket
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#datetime_bucket
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/time_series/datetime_bucket.yaml
---

# DATETIME_BUCKET

## Summary
Returns the inclusive lower bound (a `DATETIME`) of the fixed-width datetime
bucket that contains the supplied datetime, given a bucket width and an
optional bucket origin.

## Signatures
- `DATETIME_BUCKET(datetime_in_bucket, bucket_width)`
- `DATETIME_BUCKET(datetime_in_bucket, bucket_width, bucket_origin_datetime)`

## Behavior
- `datetime_in_bucket` is a `DATETIME` value used to locate the containing
  bucket.
- `bucket_width` is an `INTERVAL` value that defines the bucket size; only a
  single interval with date and time parts is supported.
- `bucket_origin_datetime` is a `DATETIME` anchoring the bucket grid; buckets
  extend in both directions from this point.
- When `bucket_origin_datetime` is omitted, the default origin is
  `1950-01-01 00:00:00`.
- Buckets are half-open intervals `[lower_bound, upper_bound)`, and the
  function returns the inclusive `lower_bound` of the bucket containing
  `datetime_in_bucket`.
- The return type is `DATETIME`.

## Examples
```sql
-- Default origin (1950-01-01 00:00:00), 12-hour buckets
WITH some_datetimes AS (
  SELECT DATETIME '1949-12-30 13:00:00' AS my_datetime UNION ALL
  SELECT DATETIME '1949-12-31 00:00:00' UNION ALL
  SELECT DATETIME '1949-12-31 13:00:00' UNION ALL
  SELECT DATETIME '1950-01-01 00:00:00' UNION ALL
  SELECT DATETIME '1950-01-01 13:00:00' UNION ALL
  SELECT DATETIME '1950-01-02 00:00:00'
)
SELECT DATETIME_BUCKET(my_datetime, INTERVAL 12 HOUR) AS bucket_lower_bound
FROM some_datetimes;
-- expected: 1949-12-30T12:00:00, 1949-12-31T00:00:00, 1949-12-31T12:00:00,
--           1950-01-01T00:00:00, 1950-01-01T12:00:00, 1950-01-02T00:00:00
```

```sql
-- Custom origin 2000-12-22 12:00:00, 7-day buckets
WITH some_datetimes AS (
  SELECT DATETIME '2000-12-20 00:00:00' AS my_datetime UNION ALL
  SELECT DATETIME '2000-12-21 00:00:00' UNION ALL
  SELECT DATETIME '2000-12-22 00:00:00' UNION ALL
  SELECT DATETIME '2000-12-23 00:00:00' UNION ALL
  SELECT DATETIME '2000-12-24 00:00:00' UNION ALL
  SELECT DATETIME '2000-12-25 00:00:00'
)
SELECT DATETIME_BUCKET(
  my_datetime,
  INTERVAL 7 DAY,
  DATETIME '2000-12-22 12:00:00') AS bucket_lower_bound
FROM some_datetimes;
-- expected: 2000-12-15T12:00:00, 2000-12-15T12:00:00, 2000-12-15T12:00:00,
--           2000-12-22T12:00:00, 2000-12-22T12:00:00, 2000-12-22T12:00:00
```

## Edge cases
- Only single-interval `INTERVAL` values composed of date and time parts are
  supported for `bucket_width`; multi-part intervals are not allowed.
- Datetimes earlier than the origin still resolve to a bucket because the
  bucket grid extends in both directions from `bucket_origin_datetime`.

## Reference (upstream)

See the upstream BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#datetime_bucket>
