---
name: REGEXP_LIKE
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindRegexpLike in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#regexp_like
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#regexp_like
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/regexp_like.yaml
---

# REGEXP_LIKE

## Summary

Returns `TRUE` if `str` matches the regular-expression `pattern`, `FALSE` otherwise. Equivalent to `REGEXP_CONTAINS` but follows MySQLs ordering and semantics.

## Signatures

- `REGEXP_LIKE(str, pattern)`
- `REGEXP_LIKE(str, pattern, match_type)`

## Arguments

- `str`: `STRING` to test.
- `pattern`: `STRING` regular expression in RE2 syntax.
- `match_type`: optional `STRING` containing flag characters: `c` (case-sensitive, default), `i` (case-insensitive), `m` (multi-line `^`/`$`), `n` (`.` matches newline), `u` (Unicode mode).

## Return type

`BOOL`.

## Behavior

- The match is unanchored; use explicit `^` and `$` for full-string matching.
- An invalid `pattern` raises an error at evaluation time.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT REGEXP_LIKE("hello", "l+");                -- TRUE
SELECT REGEXP_LIKE("hello", "^world$");           -- FALSE
SELECT REGEXP_LIKE("Hello", "hello", "i");        -- TRUE
```

## Edge cases

- Conflicting flags in `match_type` (e.g. `c` and `i` together) follow MySQLs "later wins" convention.
- For full anchoring semantics matching `LIKE` `^...$`, use `REGEXP_CONTAINS(str, "^pattern$")` instead.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#regexp_like>.
