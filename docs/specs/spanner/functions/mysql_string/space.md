---
name: SPACE
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Returns a STRING of `n` spaces. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_space in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#space
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#space
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/space.yaml
---

# SPACE

## Summary

Returns a string of `n` ASCII space characters.

## Signatures

- `SPACE(n)`

## Arguments

- `n`: `INT64` non-negative count of space characters.

## Return type

`STRING`.

## Behavior

- `n <= 0` returns the empty string.
- Returns `NULL` if `n` is `NULL`.

## Examples

```sql
SELECT SPACE(3);       -- "   "
SELECT SPACE(0);       -- ""
SELECT LENGTH(SPACE(5)); -- 5
```

## Edge cases

- For padding, `LPAD`/`RPAD` are usually a clearer fit; `SPACE` exists primarily for MySQL portability.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#space>.
