---
name: EXTRACT
dialect: googlesql
category: functions/interval
status: implemented
source_url: docs/third_party/googlesql-docs/interval_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/interval_functions.md#extract
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/interval/extract.yaml
---

# EXTRACT

## Summary

(TBD — refine from the upstream reference below.)

## Signatures

(TBD)

## Behavior

(TBD)

## Examples

(TBD)

## Edge cases

(TBD)

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/interval_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `EXTRACT`

```googlesql
EXTRACT(part FROM interval_expression)
```

**Description**

Returns the value corresponding to the specified date part. The `part` must be
one of `YEAR`, `MONTH`, `DAY`, `HOUR`, `MINUTE`, `SECOND`, `MILLISECOND` or
`MICROSECOND`.

**Return Data Type**

`INTERVAL`

**Examples**

In the following example, different parts of two intervals are extracted.

```googlesql
SELECT
  EXTRACT(YEAR FROM i) AS year,
  EXTRACT(MONTH FROM i) AS month,
  EXTRACT(DAY FROM i) AS day,
  EXTRACT(HOUR FROM i) AS hour,
  EXTRACT(MINUTE FROM i) AS minute,
  EXTRACT(SECOND FROM i) AS second,
  EXTRACT(MILLISECOND FROM i) AS milli,
  EXTRACT(MICROSECOND FROM i) AS micro
FROM
  UNNEST([INTERVAL '1-2 3 4:5:6.789999' YEAR TO SECOND,
          INTERVAL '0-13 370 48:61:61' YEAR TO SECOND]) AS i

/*------+-------+-----+------+--------+--------+-------+--------+
 | year | month | day | hour | minute | second | milli | micro  |
 +------+-------+-----+------+--------+--------+-------+--------+
 | 1    | 2     | 3   | 4    | 5      | 6      | 789   | 789999 |
 | 1    | 1     | 370 | 49   | 2      | 1      | 0     | 0      |
 +------+-------+-----+------+--------+--------+-------+--------*/
```

When a negative sign precedes the time part in an interval, the negative sign
distributes over the hours, minutes, and seconds. For example:

```googlesql
SELECT
  EXTRACT(HOUR FROM i) AS hour,
  EXTRACT(MINUTE FROM i) AS minute
FROM
  UNNEST([INTERVAL '10 -12:30' DAY TO MINUTE]) AS i

/*------+--------+
 | hour | minute |
 +------+--------+
 | -12  | -30    |
 +------+--------*/
```

When a negative sign precedes the year and month part in an interval, the
negative sign distributes over the years and months. For example:

```googlesql
SELECT
  EXTRACT(YEAR FROM i) AS year,
  EXTRACT(MONTH FROM i) AS month
FROM
  UNNEST([INTERVAL '-22-6 10 -12:30' YEAR TO MINUTE]) AS i

/*------+--------+
 | year | month  |
 +------+--------+
 | -22  | -6     |
 +------+--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/interval_functions.md`.
