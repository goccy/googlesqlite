---
name: ZSTD_DECOMPRESS_TO_STRING
dialect: spanner
category: functions/compression
status: implemented
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_decompress_to_string
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_decompress_to_string
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/compression/zstd_decompress_to_string.yaml
---

# ZSTD_DECOMPRESS_TO_STRING

## Summary

Decompresses a zstd frame and interprets the result as a UTF-8 `STRING`.

## Signatures

- `ZSTD_DECOMPRESS_TO_STRING(compressed)`

## Arguments

- `compressed`: `BYTES` zstd frame.

## Return type

`STRING`.

## Behavior

- Returns `NULL` if `compressed` is `NULL`.
- Decompressed bytes that are not valid UTF-8 raise an error.

## Examples

```sql
SELECT ZSTD_DECOMPRESS_TO_STRING(ZSTD_COMPRESS(CAST("hello" AS BYTES)));   -- "hello"
```

## Edge cases

- For binary payloads, use `ZSTD_DECOMPRESS_TO_BYTES`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_decompress_to_string>.
