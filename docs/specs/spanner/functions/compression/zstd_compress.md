---
name: ZSTD_COMPRESS
dialect: spanner
category: functions/compression
status: implemented
notes: |
  Spanner compression helpers operate on the storage engine's internal format. googlesqlite stores plain SQLite pages — there is no on-disk compression layer to expose.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_compress
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_compress
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/compression/zstd_compress.yaml
---

# ZSTD_COMPRESS

## Summary

Compresses `BYTES` with the Zstandard algorithm.

## Signatures

- `ZSTD_COMPRESS(data)`
- `ZSTD_COMPRESS(data, compression_level)`

## Arguments

- `data`: `BYTES` to compress.
- `compression_level`: optional `INT64` in `[1, 22]`. Higher values trade time for ratio. The default is implementation-defined.

## Return type

`BYTES` containing a self-contained zstd frame.

## Behavior

- The output begins with the zstd magic number `0xFD2FB528`.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT BYTE_LENGTH(ZSTD_COMPRESS(b"hello world hello world"))
       < BYTE_LENGTH(b"hello world hello world");   -- TRUE
```

## Edge cases

- Inverse: `ZSTD_DECOMPRESS_TO_BYTES`.
- Compressing an empty `BYTES` returns a non-empty zstd frame (with the empty payload encoded inside).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#zstd_compress>.
