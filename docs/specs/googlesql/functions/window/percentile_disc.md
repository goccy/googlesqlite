---
name: PERCENTILE_DISC
dialect: googlesql
category: functions/window
status: implemented
source_url: docs/third_party/googlesql-docs/navigation_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/navigation_functions.md#percentile_disc
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/window/percentile_disc.yaml
---

# PERCENTILE_DISC

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

## `PERCENTILE_DISC`

```googlesql
PERCENTILE_DISC (value_expression, percentile [{RESPECT | IGNORE} NULLS])
OVER over_clause

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]

```

**Description**

Computes the specified percentile value for a discrete `value_expression`. The
returned value is the first sorted value of `value_expression` with cumulative
distribution greater than or equal to the given `percentile` value.

This function ignores `NULL`
values unless
`RESPECT NULLS` is present.

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

+ `value_expression` can be any orderable type.
+ `percentile` must be a literal in the range `[0, 1]`, with one of the
  following types:
   + `NUMERIC`
   + `BIGNUMERIC`
   + `DOUBLE`

**Return Data Type**

Same type as `value_expression`.

**Examples**

The following example computes the value for some percentiles from a column of
values while ignoring nulls.

```googlesql
SELECT
  x,
  PERCENTILE_DISC(x, 0) OVER() AS min,
  PERCENTILE_DISC(x, 0.5) OVER() AS median,
  PERCENTILE_DISC(x, 1) OVER() AS max
FROM UNNEST(['c', NULL, 'b', 'a']) AS x;

/*------+-----+--------+-----+
 | x    | min | median | max |
 +------+-----+--------+-----+
 | c    | a   | b      | c   |
 | NULL | a   | b      | c   |
 | b    | a   | b      | c   |
 | a    | a   | b      | c   |
 +------+-----+--------+-----*/
```

The following example computes the value for some percentiles from a column of
values while respecting nulls.

```googlesql
SELECT
  x,
  PERCENTILE_DISC(x, 0 RESPECT NULLS) OVER() AS min,
  PERCENTILE_DISC(x, 0.5 RESPECT NULLS) OVER() AS median,
  PERCENTILE_DISC(x, 1 RESPECT NULLS) OVER() AS max
FROM UNNEST(['c', NULL, 'b', 'a']) AS x;

/*------+------+--------+-----+
 | x    | min  | median | max |
 +------+------+--------+-----+
 | c    | NULL | a      | c   |
 | NULL | NULL | a      | c   |
 | b    | NULL | a      | c   |
 | a    | NULL | a      | c   |
 +------+------+--------+-----*/

```

[sketches]: https://github.com/google/googlesql/blob/master/docs/sketches.md

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/navigation_functions.md`.
