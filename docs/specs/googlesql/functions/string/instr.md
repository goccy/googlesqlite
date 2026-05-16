---
name: INSTR
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#instr
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/instr.yaml
---

# INSTR

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

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `INSTR`

```googlesql
INSTR(value, subvalue[, position[, occurrence]])
```

**Description**

Returns the lowest 1-based position of `subvalue` in `value`.
`value` and `subvalue` must be the same type, either
`STRING` or `BYTES`.

If `position` is specified, the search starts at this position in
`value`, otherwise it starts at `1`, which is the beginning of
`value`. If `position` is negative, the function searches backwards
from the end of `value`, with `-1` indicating the last character.
`position` is of type `INT64` and can't be `0`.

If `occurrence` is specified, the search returns the position of a specific
instance of `subvalue` in `value`. If not specified, `occurrence`
defaults to `1` and returns the position of the first occurrence.
For `occurrence` > `1`, the function includes overlapping occurrences.
`occurrence` is of type `INT64` and must be positive.

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

Returns `0` if:

+ No match is found.
+ If `occurrence` is greater than the number of matches found.
+ If `position` is greater than the length of `value`.

Returns `NULL` if:

+ Any input argument is `NULL`.

Returns an error if:

+ `position` is `0`.
+ `occurrence` is `0` or negative.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT
  'banana' AS value, 'an' AS subvalue, 1 AS position, 1 AS occurrence,
  INSTR('banana', 'an', 1, 1) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | banana       | an           | 1        | 1          | 2     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'banana' AS value, 'an' AS subvalue, 1 AS position, 2 AS occurrence,
  INSTR('banana', 'an', 1, 2) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | banana       | an           | 1        | 2          | 4     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'banana' AS value, 'an' AS subvalue, 1 AS position, 3 AS occurrence,
  INSTR('banana', 'an', 1, 3) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | banana       | an           | 1        | 3          | 0     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'banana' AS value, 'an' AS subvalue, 3 AS position, 1 AS occurrence,
  INSTR('banana', 'an', 3, 1) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | banana       | an           | 3        | 1          | 4     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'banana' AS value, 'an' AS subvalue, -1 AS position, 1 AS occurrence,
  INSTR('banana', 'an', -1, 1) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | banana       | an           | -1       | 1          | 4     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'banana' AS value, 'an' AS subvalue, -3 AS position, 1 AS occurrence,
  INSTR('banana', 'an', -3, 1) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | banana       | an           | -3       | 1          | 4     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'banana' AS value, 'ann' AS subvalue, 1 AS position, 1 AS occurrence,
  INSTR('banana', 'ann', 1, 1) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | banana       | ann          | 1        | 1          | 0     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'helloooo' AS value, 'oo' AS subvalue, 1 AS position, 1 AS occurrence,
  INSTR('helloooo', 'oo', 1, 1) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | helloooo     | oo           | 1        | 1          | 5     |
 +--------------+--------------+----------+------------+-------*/
```

```googlesql
SELECT
  'helloooo' AS value, 'oo' AS subvalue, 1 AS position, 2 AS occurrence,
  INSTR('helloooo', 'oo', 1, 2) AS instr;

/*--------------+--------------+----------+------------+-------+
 | value        | subvalue     | position | occurrence | instr |
 +--------------+--------------+----------+------------+-------+
 | helloooo     | oo           | 1        | 2          | 6     |
 +--------------+--------------+----------+------------+-------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
