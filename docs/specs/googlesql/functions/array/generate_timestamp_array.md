---
name: GENERATE_TIMESTAMP_ARRAY
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#generate_timestamp_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/generate_timestamp_array.yaml
---

# GENERATE_TIMESTAMP_ARRAY

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

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `GENERATE_TIMESTAMP_ARRAY`

```googlesql
GENERATE_TIMESTAMP_ARRAY(start_timestamp, end_timestamp,
                         INTERVAL step_expression date_part)
```

**Description**

Returns an `ARRAY` of `TIMESTAMPS` separated by a given interval. The
`start_timestamp` and `end_timestamp` parameters determine the inclusive
lower and upper bounds of the `ARRAY`.

The `GENERATE_TIMESTAMP_ARRAY` function accepts the following data types as
inputs:

+ `start_timestamp`: `TIMESTAMP`
+ `end_timestamp`: `TIMESTAMP`
+ `step_expression`: `INT64`
+ Allowed `date_part` values are:
  `PICOSECOND`,
  `NANOSECOND`,
  `MICROSECOND`, `MILLISECOND`, `SECOND`, `MINUTE`, `HOUR`, or `DAY`.

The `step_expression` parameter determines the increment used to generate
timestamps.

**Return Data Type**

An `ARRAY` containing 0 or more `TIMESTAMP` values.

**Examples**

The following example returns an `ARRAY` of `TIMESTAMP`s at intervals of 1 day.

```googlesql
SELECT GENERATE_TIMESTAMP_ARRAY('2016-10-05 00:00:00', '2016-10-07 00:00:00',
                                INTERVAL 1 DAY) AS timestamp_array;

/*--------------------------------------------------------------------------+
 | timestamp_array                                                          |
 +--------------------------------------------------------------------------+
 | [2016-10-05 00:00:00+00, 2016-10-06 00:00:00+00, 2016-10-07 00:00:00+00] |
 +--------------------------------------------------------------------------*/
```

The following example returns an `ARRAY` of `TIMESTAMP`s at intervals of 1
second.

```googlesql
SELECT GENERATE_TIMESTAMP_ARRAY('2016-10-05 00:00:00', '2016-10-05 00:00:02',
                                INTERVAL 1 SECOND) AS timestamp_array;

/*--------------------------------------------------------------------------+
 | timestamp_array                                                          |
 +--------------------------------------------------------------------------+
 | [2016-10-05 00:00:00+00, 2016-10-05 00:00:01+00, 2016-10-05 00:00:02+00] |
 +--------------------------------------------------------------------------*/
```

The following example returns an `ARRAY` of `TIMESTAMPS` with a negative
interval.

```googlesql
SELECT GENERATE_TIMESTAMP_ARRAY('2016-10-06 00:00:00', '2016-10-01 00:00:00',
                                INTERVAL -2 DAY) AS timestamp_array;

/*--------------------------------------------------------------------------+
 | timestamp_array                                                          |
 +--------------------------------------------------------------------------+
 | [2016-10-06 00:00:00+00, 2016-10-04 00:00:00+00, 2016-10-02 00:00:00+00] |
 +--------------------------------------------------------------------------*/
```

The following example returns an `ARRAY` with a single element, because
`start_timestamp` and `end_timestamp` have the same value.

```googlesql
SELECT GENERATE_TIMESTAMP_ARRAY('2016-10-05 00:00:00', '2016-10-05 00:00:00',
                                INTERVAL 1 HOUR) AS timestamp_array;

/*--------------------------+
 | timestamp_array          |
 +--------------------------+
 | [2016-10-05 00:00:00+00] |
 +--------------------------*/
```

The following example returns an empty `ARRAY`, because `start_timestamp` is
later than `end_timestamp`.

```googlesql
SELECT GENERATE_TIMESTAMP_ARRAY('2016-10-06 00:00:00', '2016-10-05 00:00:00',
                                INTERVAL 1 HOUR) AS timestamp_array;

/*-----------------+
 | timestamp_array |
 +-----------------+
 | []              |
 +-----------------*/
```

The following example returns a null `ARRAY`, because one of the inputs is
`NULL`.

```googlesql
SELECT GENERATE_TIMESTAMP_ARRAY('2016-10-05 00:00:00', NULL, INTERVAL 1 HOUR)
  AS timestamp_array;

/*-----------------+
 | timestamp_array |
 +-----------------+
 | NULL            |
 +-----------------*/
```

The following example generates `ARRAY`s of `TIMESTAMP`s from columns containing
values for `start_timestamp` and `end_timestamp`.

```googlesql
SELECT GENERATE_TIMESTAMP_ARRAY(start_timestamp, end_timestamp, INTERVAL 1 HOUR)
  AS timestamp_array
FROM
  (SELECT
    TIMESTAMP '2016-10-05 00:00:00' AS start_timestamp,
    TIMESTAMP '2016-10-05 02:00:00' AS end_timestamp
   UNION ALL
   SELECT
    TIMESTAMP '2016-10-05 12:00:00' AS start_timestamp,
    TIMESTAMP '2016-10-05 14:00:00' AS end_timestamp
   UNION ALL
   SELECT
    TIMESTAMP '2016-10-05 23:59:00' AS start_timestamp,
    TIMESTAMP '2016-10-06 01:59:00' AS end_timestamp);

/*--------------------------------------------------------------------------+
 | timestamp_array                                                          |
 +--------------------------------------------------------------------------+
 | [2016-10-05 00:00:00+00, 2016-10-05 01:00:00+00, 2016-10-05 02:00:00+00] |
 | [2016-10-05 12:00:00+00, 2016-10-05 13:00:00+00, 2016-10-05 14:00:00+00] |
 | [2016-10-05 23:59:00+00, 2016-10-06 00:59:00+00, 2016-10-06 01:59:00+00] |
 +--------------------------------------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
