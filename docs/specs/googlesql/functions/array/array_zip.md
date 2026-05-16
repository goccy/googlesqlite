---
name: ARRAY_ZIP
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_zip
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_zip.yaml
---

# ARRAY_ZIP

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

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY_ZIP`

```googlesql
ARRAY_ZIP(
  array_input [ AS alias ],
  array_input [ AS alias ][, ... ]
  [, [ transformation => ] value ]
  [, mode => { 'STRICT' | 'TRUNCATE' | 'PAD' } ]
)
```

**Description**

Combines the elements from two to four arrays into one array.

**Definitions**

+   `array_input`: An input `ARRAY` value to be zipped with the other array
    inputs. `ARRAY_ZIP` supports two to four input arrays.
+   `alias`: An alias optionally supplied for an `array_input`. In the results,
    the alias is the name of the associated `STRUCT` field.
+   `transformation`: A named argument with a lambda expression.
    The lambda expression specifies how elements are combined as they are
    zipped. This overrides the default `STRUCT` creation behavior.
+   `mode`: A named argument with a `STRING` value. Determines how arrays of
    differing lengths are zipped. If this argument isn't supplied, the
    function uses `STRICT` mode. This argument can be one of the
    following values:

    +   `STRICT` (default): If the length of any array is different from the
        others, produce an error.

    +   `TRUNCATE`: Truncate longer arrays to match the length of the shortest
        array.

    +   `PAD`: Pad shorter arrays with `NULL` values to match the length of the
        longest array.

**Details**

+   If an `array_input` or `mode` is `NULL`, this function returns `NULL`, even when
    `mode` is `STRICT`.
+   Argument aliases can't be used with the `transformation` argument.

**Return type**

+   If `transformation` is used and returns type `T`, the
    return type is `ARRAY<T>`.
+   Otherwise, the return type is `ARRAY<STRUCT>`, with the `STRUCT` having a
    number of fields equal to the number of input arrays. Each field's name is
    either the user-provided `alias` for the corresponding `array_input`, or a
    default alias assigned by the compiler, following the same logic used for
    [naming columns in a SELECT list][implicit-aliases].

**Examples**

The following query zips two arrays into one:

```googlesql
SELECT ARRAY_ZIP([1, 2], ['a', 'b']) AS results

/*----------------------+
 | results              |
 +----------------------+
 | [(1, 'a'), (2, 'b')] |
 +----------------------*/
```

You can give an array an alias. For example, in the following
query, the returned array is of type `ARRAY<STRUCT<A1, alias_inferred>>`,
where:

+   `A1` is the alias provided for array `[1, 2]`.
+   `alias_inferred` is the inferred alias provided for array `['a', 'b']`.

```googlesql
WITH T AS (
  SELECT ['a', 'b'] AS alias_inferred
)
SELECT ARRAY_ZIP([1, 2] AS A1, alias_inferred) AS results
FROM T

/*----------------------------------------------------------+
 | results                                                  |
 +----------------------------------------------------------+
 | [{1 A1, 'a' alias_inferred}, {2 A1, 'b' alias_inferred}] |
 +----------------------------------------------------------*/
```

To provide a custom transformation of the input arrays, use the `transformation`
argument:

```googlesql
SELECT ARRAY_ZIP([1, 2], [3, 4], transformation => (e1, e2) -> (e1 + e2))

/*---------+
 | results |
 +---------+
 | [4, 6]  |
 +---------*/
```

The argument name `transformation` isn't required. For example:

```googlesql
SELECT ARRAY_ZIP([1, 2], [3, 4], (e1, e2) -> (e1 + e2))

/*---------+
 | results |
 +---------+
 | [4, 6]  |
 +---------*/
```

When `transformation` is provided, the input arrays aren't allowed to have
aliases. For example, the following query is invalid:

```googlesql {.bad}
-- Error: ARRAY_ZIP function with lambda argument doesn't allow providing
-- argument aliases
SELECT ARRAY_ZIP([1, 2], [3, 4] AS alias_not_allowed, (e1, e2) -> (e1 + e2))
```

To produce an error when arrays with different lengths are zipped, don't
add `mode`, or if you do, set it as `STRICT`. For example:

```googlesql {.bad}
-- Error: Unequal array length
SELECT ARRAY_ZIP([1, 2], ['a', 'b', 'c', 'd']) AS results
```

```googlesql {.bad}
-- Error: Unequal array length
SELECT ARRAY_ZIP([1, 2], ['a', 'b', 'c', 'd'], mode => 'STRICT') AS results
```

Use the `PAD` mode to pad missing values with `NULL` when input arrays have
different lengths. For example:

```googlesql
SELECT ARRAY_ZIP([1, 2], ['a', 'b', 'c', 'd'], [], mode => 'PAD') AS results

/*------------------------------------------------------------------------+
 | results                                                                |
 +------------------------------------------------------------------------+
 | [{1, 'a', NULL}, {2, 'b', NULL}, {NULL, 'c', NULL}, {NULL, 'd', NULL}] |
 +------------------------------------------------------------------------*/
```

Use the `TRUNCATE` mode to truncate all arrays that are longer than the shortest
array. For example:

```googlesql
SELECT ARRAY_ZIP([1, 2], ['a', 'b', 'c', 'd'], mode => 'TRUNCATE') AS results

/*----------------------+
 | results              |
 +----------------------+
 | [(1, 'a'), (2, 'b')] |
 +----------------------*/
```

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[implicit-aliases]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#implicit_aliases

<!-- mdlint on -->

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
