---
name: IS_FIRST
dialect: googlesql
category: functions/window
status: implemented
source_url: docs/third_party/googlesql-docs/numbering_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/numbering_functions.md#is_first
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/window/is_first.yaml
---

# IS_FIRST

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

## `IS_FIRST`

```googlesql
IS_FIRST(k)
OVER over_clause

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]
  [ ORDER BY expression [ { ASC | DESC }  ] [, ...] ]

```

**Description**

Returns `true` if the current row is in the first `k` rows (1-based) in the
window; otherwise, returns `false`. This function doesn't require the `ORDER BY`
clause.

**Details**

* The `k` value must be positive; otherwise, a runtime error is raised.
* If `k` is 0, the scenario is considered a degenerate case where the result is always `false`.
* If `k` is `NULL`, the result is `NULL`.
* Disallows the window framing clause, similar to the `ROW_NUMBER` function.
* If any rows are tied or if `ORDER BY` is omitted, the result is non-deterministic.
  If the `ORDER BY` clause is unspecified or if all rows are tied, the
  result is equivalent to `ANY-k`.

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Return Type**

`BOOL`

**Examples**

```googlesql
WITH Numbers AS
 (SELECT 1 as x
  UNION ALL SELECT 2
  UNION ALL SELECT 2
  UNION ALL SELECT 5
  UNION ALL SELECT 8
  UNION ALL SELECT 10
  UNION ALL SELECT 10
)
SELECT x,
  IS_FIRST(2) OVER (ORDER BY x) AS is_first
FROM Numbers

/*-------------------------+
 | x          | is_first   |
 +-------------------------+
 | 1          | true       |
 | 2          | true       |
 | 2          | false      |
 | 5          | false      |
 | 8          | false      |
 | 10         | false      |
 | 10         | false      |
 +-------------------------*/
```

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
  IS_FIRST(2) OVER (PARTITION BY division ORDER BY finish_time ASC) AS is_first
FROM finishers;

/*-----------------+------------------------+----------+-------------+
 | name            | finish_time            | division | finish_rank |
 +-----------------+------------------------+----------+-------------+
 | Sophia Liu      | 2016-10-18 09:51:45+00 | F30-34   | true        |
 | Meghan Lederer  | 2016-10-18 09:59:01+00 | F30-34   | true        |
 | Nikki Leith     | 2016-10-18 09:59:01+00 | F30-34   | false       |
 | Jen Edwards     | 2016-10-18 10:06:36+00 | F30-34   | false       |
 | Lisa Stelzner   | 2016-10-18 09:54:11+00 | F35-39   | true        |
 | Lauren Matthews | 2016-10-18 10:01:17+00 | F35-39   | true        |
 | Desiree Berry   | 2016-10-18 10:05:42+00 | F35-39   | false       |
 | Suzy Slane      | 2016-10-18 10:06:24+00 | F35-39   | false       |
 +-----------------+------------------------+----------+-------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/numbering_functions.md`.
