---
name: SHA2
dialect: spanner
category: functions/mysql_encryption
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindSHA2 in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/encryption_functions#sha2
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/encryption_functions#sha2
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_encryption/sha2.yaml
---

# SHA2

## Summary

Returns the hex-encoded hash of `str` using a SHA-2 family digest selected by `length`.

## Signatures

- `SHA2(str, length)`

## Arguments

- `str`: `STRING` or `BYTES`.
- `length`: `INT64` selecting the SHA-2 variant. Allowed values: `224`, `256`, `384`, `512`. `0` is treated as `256`.

## Return type

`STRING` of lowercase hex characters (length `2*length/8`).

## Behavior

- Returns `NULL` if `str` is `NULL` or `length` is `NULL`.
- Unsupported `length` values return `NULL`.

## Examples

```sql
SELECT SHA2("", 256);    -- "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
SELECT SHA2("abc", 512); -- 128 hex chars
```

## Edge cases

- For raw `BYTES` output, use `SHA256`/`SHA512` (return type `BYTES`).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/encryption_functions#sha2>.
