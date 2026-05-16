---
name: LAST_VALUE
dialect: googlesql
category: functions/window
status: implemented
source_url: docs/third_party/googlesql-docs/navigation_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/navigation_functions.md#last_value
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/window/last_value.yaml
---

# LAST_VALUE

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

Verbatim copy from `docs/third_party/googlesql-docs/navigation_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `LAST_VALUE`

```googlesql
LAST_VALUE (value_expression [{RESPECT | IGNORE} NULLS])
OVER over_clause

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]
  ORDER BY expression [ { ASC | DESC }  ] [, ...]
  [ window_frame_clause ]

```

**Description**

Returns the value of the `value_expression` for the last row in the current
window frame.

This function includes `NULL` values in the calculation unless `IGNORE NULLS` is
present. If `IGNORE NULLS` is present, the function excludes `NULL` values from
the calculation.

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

`value_expression` can be any data type that an expression can return.

**Return Data Type**

Same type as `value_expression`.

**Examples**

The following example computes the slowest time for each division.

```googlesql
WITH finishers AS
 (SELECT 'Sophia Liu' as name,
  TIMESTAMP '2016-10-18 2:51:45' as finish_time,
  'F30-34' as division
  UNION ALL SELECT 'Lisa Stelzner', TIMESTAMP '2016-10-18 2:54:11', 'F35-39'
  UNION ALL SELECT 'Nikki Leith', TIMESTAMP '2016-10-18 2:59:01', 'F30-34'
  UNION ALL SELECT 'Lauren Matthews', TIMESTAMP '2016-10-18 3:01:17', 'F35-39'
  UNION ALL SELECT 'Desiree Berry', TIMESTAMP '2016-10-18 3:05:42', 'F35-39'
  UNION ALL SELECT 'Suzy Slane', TIMESTAMP '2016-10-18 3:06:24', 'F35-39'
  UNION ALL SELECT 'Jen Edwards', TIMESTAMP '2016-10-18 3:06:36', 'F30-34'
  UNION ALL SELECT 'Meghan Lederer', TIMESTAMP '2016-10-18 3:07:41', 'F30-34'
  UNION ALL SELECT 'Carly Forte', TIMESTAMP '2016-10-18 3:08:58', 'F25-29'
  UNION ALL SELECT 'Lauren Reasoner', TIMESTAMP '2016-10-18 3:10:14', 'F30-34')
SELECT name,
  FORMAT_TIMESTAMP('%X', finish_time) AS finish_time,
  division,
  FORMAT_TIMESTAMP('%X', slowest_time) AS slowest_time,
  TIMESTAMP_DIFF(slowest_time, finish_time, SECOND) AS delta_in_seconds
FROM (
  SELECT name,
  finish_time,
  division,
  LAST_VALUE(finish_time)
    OVER (PARTITION BY division ORDER BY finish_time ASC
    ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) AS slowest_time
  FROM finishers);

/*-----------------+-------------+----------+--------------+------------------+
 | name            | finish_time | division | slowest_time | delta_in_seconds |
 +-----------------+-------------+----------+--------------+------------------+
 | Carly Forte     | 03:08:58    | F25-29   | 03:08:58     | 0                |
 | Sophia Liu      | 02:51:45    | F30-34   | 03:10:14     | 1109             |
 | Nikki Leith     | 02:59:01    | F30-34   | 03:10:14     | 673              |
 | Jen Edwards     | 03:06:36    | F30-34   | 03:10:14     | 218              |
 | Meghan Lederer  | 03:07:41    | F30-34   | 03:10:14     | 153              |
 | Lauren Reasoner | 03:10:14    | F30-34   | 03:10:14     | 0                |
 | Lisa Stelzner   | 02:54:11    | F35-39   | 03:06:24     | 733              |
 | Lauren Matthews | 03:01:17    | F35-39   | 03:06:24     | 307              |
 | Desiree Berry   | 03:05:42    | F35-39   | 03:06:24     | 42               |
 | Suzy Slane      | 03:06:24    | F35-39   | 03:06:24     | 0                |
 +-----------------+-------------+----------+--------------+------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/navigation_functions.md`.
