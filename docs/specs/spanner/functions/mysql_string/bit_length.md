---
name: BIT_LENGTH
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Returns the bit length of a STRING (= bytes * 8). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_bit_length in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#bit_length
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#bit_length
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/bit_length.yaml
---

# BIT_LENGTH

## Summary

Returns the length of `str` measured in bits. `BIT_LENGTH(s) = 8 * OCTET_LENGTH(s)` for binary-safe inputs.

## Signatures

- `BIT_LENGTH(str)`

## Arguments

- `str`: `STRING` or `BYTES`. Multi-byte characters are counted by their byte representation.

## Return type

`INT64`.

## Behavior

- For `BYTES`, the result is `8 * BYTE_LENGTH(str)`.
- For `STRING`, the result is `8 * BYTE_LENGTH(CAST(str AS BYTES))`, i.e. eight times the UTF-8 byte length.
- Returns `NULL` if `str` is `NULL`.

## Examples

```sql
SELECT BIT_LENGTH("abc");      -- 24
SELECT BIT_LENGTH(b"\\x00\\xff"); -- 16
SELECT BIT_LENGTH("ä");        -- 16  (two-byte UTF-8 sequence)
```

## Edge cases

- An empty string returns `0`.
- Counting characters (not bytes) requires `CHAR_LENGTH` instead.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#bit_length>.
