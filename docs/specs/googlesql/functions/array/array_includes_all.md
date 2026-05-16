---
name: ARRAY_INCLUDES_ALL
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_includes_all
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_includes_all.yaml
---

# ARRAY_INCLUDES_ALL

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

## `ARRAY_INCLUDES_ALL`

```googlesql
ARRAY_INCLUDES_ALL(array_to_search, search_values)
```

**Description**

Takes an array to search and an array of search values. Returns `TRUE` if all
search values are in the array to search, otherwise returns `FALSE`.

+   `array_to_search`: The array to search.
+   `search_values`: The array that contains the elements to search for.

Returns `NULL` if `array_to_search` or `search_values` is
`NULL`.

**Return type**

`BOOL`

**Example**

In the following example, the query first checks to see if `3`, `4`, and `5`
exists in an array. Then the query checks to see if `4`, `5`, and `6` exists in
an array.

```googlesql
SELECT
  ARRAY_INCLUDES_ALL([1,2,3,4,5], [3,4,5]) AS a1,
  ARRAY_INCLUDES_ALL([1,2,3,4,5], [4,5,6]) AS a2;

/*------+-------+
 | a1   | a2    |
 +------+-------+
 | true | false |
 +------+-------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
