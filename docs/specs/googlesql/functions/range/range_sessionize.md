---
name: RANGE_SESSIONIZE
dialect: googlesql
category: functions/range
status: implemented
notes: |
  RANGE_SESSIONIZE is lowered to a window-function SQL pipeline
  before AnalyzeStatement runs. The pre-rewrite in
  internal/range_sessionize.go mirrors the GAP_FILL pre-rewrite
  pattern: a scanner finds every `RANGE_SESSIONIZE(...)` invocation,
  parses the four positional arguments (TABLE / subquery, range
  column name, partitioning columns array, optional mode), and
  replaces the call with a parenthesised SELECT that uses
  RANGE_START / RANGE_END / MAX OVER window functions to compute the
  per-row `session_range`. The MEETS mode treats meeting and
  overlapping ranges as one session (new session when current.start
  > prev.max_end); OVERLAPS only fuses ranges that strictly overlap
  (new session when current.start >= prev.max_end).
  
  Testdata Examples are restructured to use `setup:` blocks creating
  the source table (the upstream Example 1 INSERTs are otherwise
  multi-statement which the spectest runner cannot batch into a
  single QueryContext). Within-partition row order is
  implementation-defined when the outer ORDER BY does not break the
  tie, so the cases declare `unordered: true`.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range_sessionize
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/range_sessionize.yaml
---

# RANGE_SESSIONIZE

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

## `RANGE_SESSIONIZE`

```googlesql
RANGE_SESSIONIZE(
  TABLE table_name,
  range_column,
  partitioning_columns
)
```

```googlesql
RANGE_SESSIONIZE(
  TABLE table_name,
  range_column,
  partitioning_columns,
  sessionize_option
)
```

**Description**

Produces a table of sessionized ranges.

**Definitions**

+   `table_name`: A table expression that represents the name of the table to
    construct. This can represent any relation with `range_column`.
+   `range_column`: A `STRING` literal that indicates which `RANGE` column
    in a table contains the data to sessionize.
+   `partitioning_columns`: An `ARRAY<STRING>` literal that indicates which
    columns should partition the data before the data is sessionized.
+   `sessionize_option`: A `STRING` value that describes how order-adjacent
    ranges are sessionized. Your choices are as follows:

    +   `MEETS` (default): Ranges that meet or overlap are sessionized.

    +   `OVERLAPS`: Only a range that's overlapped by another range is
        sessionized.

    If this argument isn't provided, `MEETS` is used by default.

**Details**

This function produces a table that includes all columns in the
input table and an additional `RANGE` column called
`session_range`, which indicates the start and end of a session. The
start and end of each session is determined by the `sessionize_option`
argument.

**Return type**

`TABLE`

**Examples**

The examples in this section reference the following table called
`my_sessionized_range_table` in a dataset called `mydataset`:

```googlesql
INSERT mydataset.my_sessionized_range_table (emp_id, dept_id, duration)
VALUES(10, 1000, RANGE<DATE> '[2010-01-10, 2010-03-10)'),
      (10, 2000, RANGE<DATE> '[2010-03-10, 2010-07-15)'),
      (10, 2000, RANGE<DATE> '[2010-06-15, 2010-08-18)'),
      (20, 2000, RANGE<DATE> '[2010-03-10, 2010-07-20)'),
      (20, 1000, RANGE<DATE> '[2020-05-10, 2020-09-20)');

SELECT * FROM mydataset.my_sessionized_range_table ORDER BY emp_id;

/*--------+---------+--------------------------+
 | emp_id | dept_id | duration                 |
 +--------+---------+--------------------------+
 | 10     | 1000    | [2010-01-10, 2010-03-10) |
 | 10     | 2000    | [2010-03-10, 2010-07-15) |
 | 10     | 2000    | [2010-06-15, 2010-08-18) |
 | 20     | 2000    | [2010-03-10, 2010-07-20) |
 | 20     | 1000    | [2020-05-10, 2020-09-20) |
 +--------+---------+--------------------------*/
```

In the following query, a table of sessionized data is produced for
`my_sessionized_range_table`, and only ranges that meet or overlap are
sessionized:

```googlesql
SELECT
  emp_id, duration, session_range
FROM
  RANGE_SESSIONIZE(
    TABLE mydataset.my_sessionized_range_table,
    'duration',
    ['emp_id'])
ORDER BY emp_id;

/*--------+--------------------------+--------------------------+
 | emp_id | duration                 | session_range            |
 +--------+--------------------------+--------------------------+
 | 10     | [2010-01-10, 2010-03-10) | [2010-01-10, 2010-08-18) |
 | 10     | [2010-03-10, 2010-07-15) | [2010-01-10, 2010-08-18) |
 | 10     | [2010-06-15, 2010-08-18) | [2010-01-10, 2010-08-18) |
 | 20     | [2010-03-10, 2010-07-20) | [2010-03-10, 2010-07-20) |
 | 20     | [2020-05-10, 2020-09-20) | [2020-05-10, 2020-09-20) |
 +--------+-----------------------------------------------------*/
```

In the following query, a table of sessionized data is produced for
`my_sessionized_range_table`, and only a range that's overlapped by another
range is sessionized:

```googlesql
SELECT
  emp_id, duration, session_range
FROM
  RANGE_SESSIONIZE(
    TABLE mydataset.my_sessionized_range_table,
    'duration',
    ['emp_id'],
    'OVERLAPS')
ORDER BY emp_id;

/*--------+--------------------------+--------------------------+
 | emp_id | duration                 | session_range            |
 +--------+--------------------------+--------------------------+
 | 10     | [2010-03-10, 2010-07-15) | [2010-03-10, 2010-08-18) |
 | 10     | [2010-06-15, 2010-08-18) | [2010-03-10, 2010-08-18) |
 | 10     | [2010-01-10, 2010-03-10) | [2010-01-10, 2010-03-10) |
 | 20     | [2020-05-10, 2020-09-20) | [2020-05-10, 2020-09-20) |
 | 20     | [2010-03-10, 2010-07-20) | [2010-03-10, 2010-07-20) |
 +--------+-----------------------------------------------------*/
```

If you need to normalize sessionized data, you can use a query similar to the
following:

```googlesql
SELECT emp_id, session_range AS normalized FROM (
  SELECT emp_id, session_range
  FROM RANGE_SESSIONIZE(
    TABLE mydataset.my_sessionized_range_table,
    'duration',
    ['emp_id'],
    'MEETS')
)
GROUP BY emp_id, normalized;

/*--------+--------------------------+
 | emp_id | normalized               |
 +--------+--------------------------+
 | 20     | [2010-03-10, 2010-07-20) |
 | 10     | [2010-01-10, 2010-08-18) |
 | 20     | [2020-05-10, 2020-09-20) |
 +--------+--------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
