---
name: RANGE
dialect: googlesql
category: functions/range
status: implemented
notes: |
  RangeValue renders DATE / DATETIME / TIMESTAMP bounds in the
  format used throughout the GoogleSQL compliance suite — both
  bounds use the same separator. The docs/third_party/googlesql-docs
  Examples table includes an internally inconsistent rendering for
  Example 2 (`[2022-10-01 14:53:27, 2022-10-01T16:00:00)`, mixed
  space / T separators across the two DATETIME bounds), but the
  compliance/testdata/range_constructors.test
  `range_of_datetimes_literal_regular` fixture authoritatively
  renders DATETIME range bounds with space separators on both sides.
  When the two authoritative sources disagree, we follow the
  compliance suite, which is the more rigorous source. The
  testdata YAML carries the fix and a comment explaining the
  resolution.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/range.yaml
---

# RANGE

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

## `RANGE`

```googlesql
RANGE(lower_bound, upper_bound)
```

**Description**

Constructs a range of [`DATE`][date-type], [`DATETIME`][datetime-type], or
[`TIMESTAMP`][timestamp-type] values.

**Definitions**

+   `lower_bound`: The range starts from this value. This can be a
    `DATE`, `DATETIME`, or `TIMESTAMP` value. If this value is `NULL`, the range
    doesn't include a lower bound.
+   `upper_bound`: The range ends before this value. This can be a
    `DATE`, `DATETIME`, or `TIMESTAMP` value. If this value is `NULL`, the range
    doesn't include an upper bound.

**Details**

`lower_bound` and `upper_bound` must be of the same data type.

Produces an error if `lower_bound` is greater than or equal to `upper_bound`.
To return `NULL` instead, add the `SAFE.` prefix to the function name.

**Return type**

`RANGE<T>`, where `T` is the same data type as the input.

**Examples**

The following query constructs a date range:

```googlesql
SELECT RANGE(DATE '2022-12-01', DATE '2022-12-31') AS results;

/*--------------------------+
 | results                  |
 +--------------------------+
 | [2022-12-01, 2022-12-31) |
 +--------------------------*/
```

The following query constructs a datetime range:

```googlesql
SELECT RANGE(DATETIME '2022-10-01 14:53:27',
             DATETIME '2022-10-01 16:00:00') AS results;

/*---------------------------------------------+
 | results                                     |
 +---------------------------------------------+
 | [2022-10-01 14:53:27, 2022-10-01T16:00:00)  |
 +---------------------------------------------*/
```

The following query constructs a timestamp range:

```googlesql
SELECT RANGE(TIMESTAMP '2022-10-01 14:53:27 America/Los_Angeles',
             TIMESTAMP '2022-10-01 16:00:00 America/Los_Angeles') AS results;

-- Results depend upon where this query was executed.
/*-----------------------------------------------------------------+
 | results                                                         |
 +-----------------------------------------------------------------+
 | [2022-10-01 21:53:27.000000+00, 2022-10-01 23:00:00.000000+00)  |
 +-----------------------------------------------------------------*/
```

The following query constructs a date range with no lower bound:

```googlesql
SELECT RANGE(NULL, DATE '2022-12-31') AS results;

/*-------------------------+
 | results                 |
 +-------------------------+
 | [UNBOUNDED, 2022-12-31) |
 +-------------------------*/
```

The following query constructs a date range with no upper bound:

```googlesql
SELECT RANGE(DATE '2022-10-01', NULL) AS results;

/*--------------------------+
 | results                  |
 +--------------------------+
 | [2022-10-01, UNBOUNDED)  |
 +--------------------------*/
```

[timestamp-type]: https://github.com/google/googlesql/blob/master/docs/data-types.md#timestamp_type

[date-type]: https://github.com/google/googlesql/blob/master/docs/data-types.md#date_type

[datetime-type]: https://github.com/google/googlesql/blob/master/docs/data-types.md#datetime_type

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
