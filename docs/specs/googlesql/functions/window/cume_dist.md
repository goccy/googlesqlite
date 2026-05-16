---
name: CUME_DIST
dialect: googlesql
category: functions/window
status: implemented
source_url: docs/third_party/googlesql-docs/numbering_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/numbering_functions.md#cume_dist
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/window/cume_dist.yaml
---

# CUME_DIST

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

Verbatim copy from `docs/third_party/googlesql-docs/numbering_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `CUME_DIST`

```googlesql
CUME_DIST()
OVER over_clause

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]
  ORDER BY expression [ { ASC | DESC }  ] [, ...]

```

**Description**

Return the relative rank of a row defined as NP/NR. NP is defined to be the
number of rows that either precede or are peers with the current row. NR is the
number of rows in the partition.

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Return Type**

`DOUBLE`

**Example**

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
  UNION ALL SELECT 'Meghan Lederer', TIMESTAMP '2016-10-18 2:59:01', 'F30-34')
SELECT name,
  finish_time,
  division,
  CUME_DIST() OVER (PARTITION BY division ORDER BY finish_time ASC) AS finish_rank
FROM finishers;

/*-----------------+------------------------+----------+-------------+
 | name            | finish_time            | division | finish_rank |
 +-----------------+------------------------+----------+-------------+
 | Sophia Liu      | 2016-10-18 09:51:45+00 | F30-34   | 0.25        |
 | Meghan Lederer  | 2016-10-18 09:59:01+00 | F30-34   | 0.75        |
 | Nikki Leith     | 2016-10-18 09:59:01+00 | F30-34   | 0.75        |
 | Jen Edwards     | 2016-10-18 10:06:36+00 | F30-34   | 1           |
 | Lisa Stelzner   | 2016-10-18 09:54:11+00 | F35-39   | 0.25        |
 | Lauren Matthews | 2016-10-18 10:01:17+00 | F35-39   | 0.5         |
 | Desiree Berry   | 2016-10-18 10:05:42+00 | F35-39   | 0.75        |
 | Suzy Slane      | 2016-10-18 10:06:24+00 | F35-39   | 1           |
 +-----------------+------------------------+----------+-------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/numbering_functions.md`.
