---
name: EDIT_DISTANCE
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#edit_distance
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/edit_distance.yaml
---

# EDIT_DISTANCE

## Summary

Computes the Levenshtein distance between two `STRING` or `BYTES` values.

## Signatures

- ```googlesql
  EDIT_DISTANCE(
    value1,
    value2,
    [ max_distance => max_distance_value ]
  )
  ```

## Behavior

- Return type is `INT64`.
- Computes the Levenshtein distance between `value1` and `value2`.
- Both `value1` and `value2` must be of the same type (`STRING` or `BYTES`); otherwise an error is produced.
- The optional named argument `max_distance` is an `INT64` greater than or equal to zero that caps the maximum distance to compute.
- If the computed distance exceeds `max_distance`, the function exits early and returns `max_distance`.
- The default value of `max_distance` is the maximum size of `value1` and `value2`.

## Examples

```googlesql
SELECT EDIT_DISTANCE('a', 'b') AS results;
-- expected results: 1
```

```googlesql
SELECT EDIT_DISTANCE('aa', 'b') AS results;
-- expected results: 2
```

```googlesql
SELECT EDIT_DISTANCE('abcdefg', 'a', max_distance => 2) AS results;
-- expected results: 2 (early-exit at max_distance)
```

## Edge cases

- Returns `NULL` if either `value1` or `value2` is `NULL`.
- Raises an error when `value1` and `value2` are not of the same type (e.g. mixing `STRING` and `BYTES`).
- `max_distance` must be `>= 0`; when the true Levenshtein distance exceeds it, the function returns `max_distance` rather than the actual distance.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `EDIT_DISTANCE`

```googlesql
EDIT_DISTANCE(
  value1,
  value2,
  [ max_distance => max_distance_value ]
)
```

**Description**

Computes the [Levenshtein distance][l-distance] between two `STRING` or
`BYTES` values.

**Definitions**

+   `value1`: The first `STRING` or `BYTES` value to compare.
+   `value2`: The second `STRING` or `BYTES` value to compare.
+   `max_distance`: A named argument with a `INT64` value that's greater than
    or equal to zero. Represents the maximum distance between the two values
    to compute.

    If this distance is exceeded, the function returns this value.
    The default value for this argument is the maximum size of
    `value1` and `value2`.

**Details**

If `value1` or `value2` is `NULL`, `NULL` is returned.

You can only compare values of the same type. Otherwise, an error is produced.

**Return type**

`INT64`

**Examples**

In the following example, the first character in both strings is different:

```googlesql
SELECT EDIT_DISTANCE('a', 'b') AS results;

/*---------+
 | results |
 +---------+
 | 1       |
 +---------*/
```

In the following example, the first and second characters in both strings are
different:

```googlesql
SELECT EDIT_DISTANCE('aa', 'b') AS results;

/*---------+
 | results |
 +---------+
 | 2       |
 +---------*/
```

In the following example, only the first character in both strings is
different:

```googlesql
SELECT EDIT_DISTANCE('aa', 'ba') AS results;

/*---------+
 | results |
 +---------+
 | 1       |
 +---------*/
```

In the following example, the last six characters are different, but because
the maximum distance is `2`, this function exits early and returns `2`, the
maximum distance:

```googlesql
SELECT EDIT_DISTANCE('abcdefg', 'a', max_distance => 2) AS results;

/*---------+
 | results |
 +---------+
 | 2       |
 +---------*/
```

[l-distance]: https://en.wikipedia.org/wiki/Levenshtein_distance

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
