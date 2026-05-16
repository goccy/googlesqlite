---
name: SUBSTRING_INDEX
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindSubstringIndex in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#substring_index
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#substring_index
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/substring_index.yaml
---

# SUBSTRING_INDEX

## Summary

Returns the substring of `str` before the `count`-th occurrence of `delim`. A negative `count` selects from the end.

## Signatures

- `SUBSTRING_INDEX(str, delim, count)`

## Arguments

- `str`: `STRING` to slice.
- `delim`: `STRING` separator. Matched literally (not as a regex).
- `count`: `INT64`. Positive `count` keeps text **before** the `count`-th delimiter from the left; negative `count` keeps text **after** the `(-count)`-th delimiter from the right; zero returns the empty string.

## Return type

`STRING`.

## Behavior

- If the delimiter occurs fewer than `|count|` times, the entire string is returned.
- Counts are by occurrence, not by characters.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT SUBSTRING_INDEX("a.b.c.d", ".", 2);    -- "a.b"
SELECT SUBSTRING_INDEX("a.b.c.d", ".", -2);   -- "c.d"
SELECT SUBSTRING_INDEX("a.b.c.d", ".", 0);    -- ""
SELECT SUBSTRING_INDEX("abc", ".", 1);        -- "abc"
```

## Edge cases

- An empty `delim` returns the entire `str` regardless of `count`.
- Used heavily in MySQL log/URL parsing; consider `SPLIT` for array-returning equivalents.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#substring_index>.
