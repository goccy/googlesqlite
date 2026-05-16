---
name: LOCATE
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindLocate in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#locate
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#locate
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/locate.yaml
---

# LOCATE

## Summary

Returns the 1-based position of the first occurrence of `substr` within `str`, optionally starting the search at position `pos`.

## Signatures

- `LOCATE(substr, str)`
- `LOCATE(substr, str, pos)`

## Arguments

- `substr`: `STRING` pattern to find.
- `str`: `STRING` to search.
- `pos`: optional `INT64` 1-based starting position.

## Return type

`INT64`. `0` indicates no match (per MySQL convention).

## Behavior

- Comparison is byte-exact for `BYTES` and Unicode-codepoint exact for `STRING`.
- A `NULL` `substr` or `str` returns `NULL`; a `NULL` `pos` returns `NULL`.
- If `pos` exceeds the length of `str`, the result is `0`.

## Examples

```sql
SELECT LOCATE("bar", "foobarbar");      -- 4
SELECT LOCATE("bar", "foobarbar", 5);   -- 7
SELECT LOCATE("xyz", "foobarbar");      -- 0
```

## Edge cases

- `LOCATE` is the symmetric counterpart of `INSTR`/`POSITION`; argument order differs from `STRPOS`.
- An empty `substr` returns `1` for any non-`NULL` `str`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#locate>.
