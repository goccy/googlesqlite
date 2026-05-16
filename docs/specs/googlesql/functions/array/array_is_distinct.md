---
name: ARRAY_IS_DISTINCT
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_is_distinct
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_is_distinct.yaml
---

# ARRAY_IS_DISTINCT

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

## `ARRAY_IS_DISTINCT`

```googlesql
ARRAY_IS_DISTINCT(value)
```

**Description**

Returns `TRUE` if the array contains no repeated elements, using the same
equality comparison logic as `SELECT DISTINCT`.

**Return type**

`BOOL`

**Examples**

```googlesql
SELECT ARRAY_IS_DISTINCT([1, 2, 3]) AS is_distinct

/*-------------+
 | is_distinct |
 +-------------+
 | true        |
 +-------------*/
```

```googlesql
SELECT ARRAY_IS_DISTINCT([1, 1, 1]) AS is_distinct

/*-------------+
 | is_distinct |
 +-------------+
 | false       |
 +-------------*/
```

```googlesql
SELECT ARRAY_IS_DISTINCT([1, 2, NULL]) AS is_distinct

/*-------------+
 | is_distinct |
 +-------------+
 | true        |
 +-------------*/
```

```googlesql
SELECT ARRAY_IS_DISTINCT([1, 1, NULL]) AS is_distinct

/*-------------+
 | is_distinct |
 +-------------+
 | false       |
 +-------------*/
```

```googlesql
SELECT ARRAY_IS_DISTINCT([1, NULL, NULL]) AS is_distinct

/*-------------+
 | is_distinct |
 +-------------+
 | false       |
 +-------------*/
```
```googlesql
SELECT ARRAY_IS_DISTINCT([]) AS is_distinct

/*-------------+
 | is_distinct |
 +-------------+
 | true        |
 +-------------*/
```

```googlesql
SELECT ARRAY_IS_DISTINCT(NULL) AS is_distinct

/*-------------+
 | is_distinct |
 +-------------+
 | NULL        |
 +-------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
