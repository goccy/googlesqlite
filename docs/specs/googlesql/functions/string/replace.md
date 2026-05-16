---
name: REPLACE
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#replace
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/replace.yaml
---

# REPLACE

## Summary

Replaces all occurrences of `from_pattern` with `to_pattern` in `original_value`.

## Signatures

- `REPLACE(original_value, from_pattern, to_pattern)`

## Behavior

- Returns `STRING` or `BYTES`, matching the input types.
- Replaces every occurrence of `from_pattern` in `original_value` with `to_pattern`.
- If `from_pattern` is empty, no replacement is made and `original_value` is returned unchanged.
- Supports specifying collation for string comparison.

## Examples

```googlesql
WITH desserts AS
  (SELECT 'apple pie' as dessert
  UNION ALL
  SELECT 'blackberry pie' as dessert
  UNION ALL
  SELECT 'cherry pie' as dessert)

SELECT
  REPLACE (dessert, 'pie', 'cobbler') as example
FROM desserts;
-- expected: 'apple cobbler', 'blackberry cobbler', 'cherry cobbler'
```

## Edge cases

- An empty `from_pattern` results in no replacement; `original_value` is returned as-is.
- Collation, when specified, governs how `from_pattern` is matched against `original_value`.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `REPLACE`

```googlesql
REPLACE(original_value, from_pattern, to_pattern)
```

**Description**

Replaces all occurrences of `from_pattern` with `to_pattern` in
`original_value`. If `from_pattern` is empty, no replacement is made.

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
WITH desserts AS
  (SELECT 'apple pie' as dessert
  UNION ALL
  SELECT 'blackberry pie' as dessert
  UNION ALL
  SELECT 'cherry pie' as dessert)

SELECT
  REPLACE (dessert, 'pie', 'cobbler') as example
FROM desserts;

/*--------------------+
 | example            |
 +--------------------+
 | apple cobbler      |
 | blackberry cobbler |
 | cherry cobbler     |
 +--------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
