---
name: GENERATE_RANGE_ARRAY
dialect: googlesql
category: functions/range
status: implemented
notes: |
  ARRAY<RANGE<DATE>> rendering is now stable end-to-end. The runtime
  fix in BindGenerateRangeArray re-coerces each computed bucket
  boundary back to the source RANGE's element type — without that
  coercion, DateValue.Add(IntervalValue) returns a DatetimeValue and
  downstream buckets render as `[2020-01-01, 2020-01-02 00:00:00)`
  instead of `[2020-01-01, 2020-01-02)`.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#generate_range_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/generate_range_array.yaml
---

# GENERATE_RANGE_ARRAY

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

Verbatim copy from `docs/third_party/googlesql-docs/range-functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `GENERATE_RANGE_ARRAY`

```googlesql
GENERATE_RANGE_ARRAY(range_to_split, step_interval)
```

```googlesql
GENERATE_RANGE_ARRAY(range_to_split, step_interval, include_last_partial_range)
```

**Description**

Splits a range into an array of subranges.

**Definitions**

+   `range_to_split`: The `RANGE<T>` value to split.
+   `step_interval`: The `INTERVAL` value, which determines the maximum size of
    each subrange in the resulting array. An
    [interval single date and time part][interval-single]
    is supported, but an interval range of date and time parts isn't.

    +   If `range_to_split` is `RANGE<DATE>`, these interval
        date parts are supported: `YEAR` to `DAY`.

    +   If `range_to_split` is `RANGE<DATETIME>`, these interval
        date and time parts are supported: `YEAR` to `SECOND`.

    +   If `range_to_split` is `RANGE<TIMESTAMP>`, these interval
        date and time parts are supported: `DAY` to `SECOND`.
+   `include_last_partial_range`: A `BOOL` value, which determines whether or
    not to include the last subrange if it's a partial subrange.
    If this argument isn't specified, the default value is `TRUE`.

    +   `TRUE` (default): The last subrange is included, even if it's
        smaller than `step_interval`.

    +   `FALSE`: Exclude the last subrange if it's smaller than
        `step_interval`.

**Details**

Returns `NULL` if any input is` NULL`.

**Return type**

`ARRAY<RANGE<T>>`

**Examples**

In the following example, a date range between `2020-01-01` and `2020-01-06`
is split into an array of subranges that are one day long. There are
no partial ranges.

```googlesql
SELECT GENERATE_RANGE_ARRAY(
  RANGE(DATE '2020-01-01', DATE '2020-01-06'),
  INTERVAL 1 DAY) AS results;

/*----------------------------+
 | results                    |
 +----------------------------+
 | [                          |
 |  [2020-01-01, 2020-01-02), |
 |  [2020-01-02, 2020-01-03), |
 |  [2020-01-03, 2020-01-04), |
 |  [2020-01-04, 2020-01-05), |
 |  [2020-01-05, 2020-01-06), |
 | ]                          |
 +----------------------------*/
```

In the following examples, a date range between `2020-01-01` and `2020-01-06`
is split into an array of subranges that are two days long. The final subrange
is smaller than two days:

```googlesql
SELECT GENERATE_RANGE_ARRAY(
  RANGE(DATE '2020-01-01', DATE '2020-01-06'),
  INTERVAL 2 DAY) AS results;

/*----------------------------+
 | results                    |
 +----------------------------+
 | [                          |
 |  [2020-01-01, 2020-01-03), |
 |  [2020-01-03, 2020-01-05), |
 |  [2020-01-05, 2020-01-06)  |
 | ]                          |
 +----------------------------*/
```

```googlesql
SELECT GENERATE_RANGE_ARRAY(
  RANGE(DATE '2020-01-01', DATE '2020-01-06'),
  INTERVAL 2 DAY,
  TRUE) AS results;

/*----------------------------+
 | results                    |
 +----------------------------+
 | [                          |
 |  [2020-01-01, 2020-01-03), |
 |  [2020-01-03, 2020-01-05), |
 |  [2020-01-05, 2020-01-06)  |
 | ]                          |
 +----------------------------*/
```

In the following example, a date range between `2020-01-01` and `2020-01-06`
is split into an array of subranges that are two days long, but the final
subrange is excluded because it's smaller than two days:

```googlesql
SELECT GENERATE_RANGE_ARRAY(
  RANGE(DATE '2020-01-01', DATE '2020-01-06'),
  INTERVAL 2 DAY,
  FALSE) AS results;

/*----------------------------+
 | results                    |
 +----------------------------+
 | [                          |
 |  [2020-01-01, 2020-01-03), |
 |  [2020-01-03, 2020-01-05)  |
 | ]                          |
 +----------------------------*/
```

[interval-single]: https://github.com/google/googlesql/blob/master/docs/data-types.md#single_datetime_part_interval

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
