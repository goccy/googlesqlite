---
name: DATE_BUCKET
dialect: bigquery
category: functions/time_series
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#date_bucket
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#date_bucket
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/time_series/date_bucket.yaml
---

# DATE_BUCKET

## Summary
Returns the inclusive lower bound (a `DATE`) of the fixed-width date bucket
that contains the supplied date, given a bucket width and an optional bucket
origin.

## Signatures
- `DATE_BUCKET(date_in_bucket, bucket_width)`
- `DATE_BUCKET(date_in_bucket, bucket_width, bucket_origin_date)`

## Behavior
- `date_in_bucket` is a `DATE` value used to locate the containing bucket.
- `bucket_width` is an `INTERVAL` value that defines the bucket size; only a
  single interval with date parts is supported.
- `bucket_origin_date` is a `DATE` anchoring the bucket grid; buckets extend
  in both directions from this point.
- When `bucket_origin_date` is omitted, the default origin is `1950-01-01`.
- Buckets are half-open intervals `[lower_bound, upper_bound)`, and the
  function returns the inclusive `lower_bound` of the bucket containing
  `date_in_bucket`.
- The return type is `DATE`.

## Examples
```sql
-- Default origin (1950-01-01), 2-day buckets
WITH some_dates AS (
  SELECT DATE '1949-12-29' AS my_date UNION ALL
  SELECT DATE '1949-12-30' UNION ALL
  SELECT DATE '1949-12-31' UNION ALL
  SELECT DATE '1950-01-01' UNION ALL
  SELECT DATE '1950-01-02' UNION ALL
  SELECT DATE '1950-01-03'
)
SELECT DATE_BUCKET(my_date, INTERVAL 2 DAY) AS bucket_lower_bound
FROM some_dates;
-- expected: 1949-12-28, 1949-12-30, 1949-12-30, 1950-01-01, 1950-01-01, 1950-01-03
```

```sql
-- Custom origin 2000-12-24, 7-day buckets
WITH some_dates AS (
  SELECT DATE '2000-12-20' AS my_date UNION ALL
  SELECT DATE '2000-12-21' UNION ALL
  SELECT DATE '2000-12-22' UNION ALL
  SELECT DATE '2000-12-23' UNION ALL
  SELECT DATE '2000-12-24' UNION ALL
  SELECT DATE '2000-12-25'
)
SELECT DATE_BUCKET(
  my_date,
  INTERVAL 7 DAY,
  DATE '2000-12-24') AS bucket_lower_bound
FROM some_dates;
-- expected: 2000-12-17, 2000-12-17, 2000-12-17, 2000-12-17, 2000-12-24, 2000-12-24
```

## Edge cases
- Only `INTERVAL` values consisting of date parts are supported for
  `bucket_width`; sub-day parts are not valid.
- Dates earlier than the origin still resolve to a bucket because the bucket
  grid extends in both directions from `bucket_origin_date`.

## Reference (upstream)

See the upstream BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#date_bucket>
