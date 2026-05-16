---
name: IS_UUID
dialect: spanner
category: functions/mysql_utility
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindIsUUID in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_uuid
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_uuid
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_utility/is_uuid.yaml
---

# IS_UUID

## Summary

Returns `TRUE` if `s` is a textually valid UUID (canonical, hex, or curly-brace form).

## Signatures

- `IS_UUID(s)`

## Arguments

- `s`: `STRING`.

## Return type

`BOOL`.

## Behavior

- Accepts the canonical `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` form and MySQLs curly-brace and 32-hex compact forms.
- Returns `NULL` only if `s` is `NULL`.

## Examples

```sql
SELECT IS_UUID("00000000-0000-0000-0000-000000000000");   -- TRUE
SELECT IS_UUID("not-a-uuid");                              -- FALSE
```

## Edge cases

- Case (upper vs lower hex) does not affect validity.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#is_uuid>.
