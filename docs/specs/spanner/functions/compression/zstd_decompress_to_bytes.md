---
name: ZSTD_DECOMPRESS_TO_BYTES
dialect: spanner
category: functions/compression
status: implemented
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_decompress_to_bytes
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_decompress_to_bytes
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/compression/zstd_decompress_to_bytes.yaml
---

# ZSTD_DECOMPRESS_TO_BYTES

## Summary

Decompresses a zstd frame and returns the result as `BYTES`.

## Signatures

- `ZSTD_DECOMPRESS_TO_BYTES(compressed)`

## Arguments

- `compressed`: `BYTES` produced by `ZSTD_COMPRESS` (or any zstd-compatible encoder).

## Return type

`BYTES`.

## Behavior

- Returns `NULL` if `compressed` is `NULL`.
- Malformed input or truncated frames raise an error.

## Examples

```sql
SELECT ZSTD_DECOMPRESS_TO_BYTES(ZSTD_COMPRESS(b"hello"));   -- b"hello"
```

## Edge cases

- For `STRING` payloads, prefer `ZSTD_DECOMPRESS_TO_STRING` to avoid an explicit byte→string cast.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_decompress_to_bytes>.
