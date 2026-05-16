---
name: TIMESTAMP_BUCKET
dialect: bigquery
category: functions/time_series
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#timestamp_bucket
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#timestamp_bucket
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/time_series/timestamp_bucket.yaml
---

# TIMESTAMP_BUCKET

## Summary

Returns the lower bound of the timestamp bucket that contains a
given `TIMESTAMP`, given a bucket width and optional origin.

## Signatures

- `TIMESTAMP_BUCKET(timestamp_in_bucket, bucket_width)`
- `TIMESTAMP_BUCKET(timestamp_in_bucket, bucket_width, bucket_origin_timestamp)`

## Behavior

- `bucket_width` is an `INTERVAL` (a single interval combining
  date and time parts).
- `bucket_origin_timestamp` defaults to `1950-01-01 00:00:00`
  (UTC). Buckets expand left and right from the origin.
- Each bucket is a half-open interval `[lower, upper)` of width
  `bucket_width`.
- Returns `TIMESTAMP`.

## Examples

```sql
SELECT TIMESTAMP_BUCKET(my_timestamp, INTERVAL 12 HOUR) AS bucket
FROM UNNEST([
  TIMESTAMP '1949-12-31 13:00:00.00',
  TIMESTAMP '1950-01-01 00:00:00.00',
  TIMESTAMP '1950-01-01 13:00:00.00'
]) AS my_timestamp;
-- buckets are anchored on the 1950-01-01 default origin
```

## Edge cases

- Default origin is `1950-01-01 00:00:00`. Pass an explicit
  `bucket_origin_timestamp` to anchor elsewhere.
- The interval must combine into a meaningful single duration; the
  page documents only single-part intervals as supported.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/time-series-functions#timestamp_bucket>.
