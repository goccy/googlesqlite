---
name: MAKE_INTERVAL
dialect: googlesql
category: functions/interval
status: implemented
source_url: docs/third_party/googlesql-docs/interval_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/interval_functions.md#make_interval
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/interval/make_interval.yaml
---

# MAKE_INTERVAL

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

## `MAKE_INTERVAL`

```googlesql
MAKE_INTERVAL(
  [ [ year => ] value ]
  [, [ month => ] value ]
  [, [ day => ] value ]
  [, [ hour => ] value ]
  [, [ minute => ] value ]
  [, [ second => ] value ]
)
```

**Description**

Constructs an [`INTERVAL`][interval-type] object using `INT64` values
representing the year, month, day, hour, minute, and second. All arguments are
optional, `0` by default, and can be [named arguments][named-arguments].

**Return Data Type**

`INTERVAL`

**Example**

```googlesql
SELECT
  MAKE_INTERVAL(1, 6, 15) AS i1,
  MAKE_INTERVAL(hour => 10, second => 20) AS i2,
  MAKE_INTERVAL(1, minute => 5, day => 2) AS i3

/*--------------+---------------+-------------+
 | i1           | i2            | i3          |
 +--------------+---------------+-------------+
 | 1-6 15 0:0:0 | 0-0 0 10:0:20 | 1-0 2 0:5:0 |
 +--------------+---------------+-------------*/
```

[interval-type]: https://github.com/google/googlesql/blob/master/docs/data-types.md#interval_type

[named-arguments]: https://github.com/google/googlesql/blob/master/docs/functions-reference.md#named_arguments

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/interval_functions.md`.
