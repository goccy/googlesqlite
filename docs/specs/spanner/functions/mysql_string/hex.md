---
name: HEX
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Hex-encodes BYTES as upper-case STRING. Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_hex in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#hex
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#hex
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/hex.yaml
---

# HEX

## Summary

Returns the hexadecimal representation of an integer or the bytes of a string/bytes value.

## Signatures

- `HEX(int_val)`
- `HEX(str_or_bytes)`

## Arguments

- `int_val`: `INT64`. Negative values are formatted as the twos-complement of their 64-bit representation.
- `str_or_bytes`: `STRING` or `BYTES`. For `STRING`, the UTF-8 byte representation is hex-encoded.

## Return type

`STRING`. Letters in the result are upper-case (`0-9A-F`).

## Behavior

- For integers, the result has no leading zeros (except for the literal value `0`, which returns `"0"`).
- For `BYTES`, every byte becomes exactly two hex characters, so the result has even length.
- `NULL` input returns `NULL`.

## Examples

```sql
SELECT HEX(255);              -- "FF"
SELECT HEX(0);                -- "0"
SELECT HEX(-1);               -- "FFFFFFFFFFFFFFFF"
SELECT HEX("abc");            -- "616263"
SELECT HEX(b"\\x00\\x01");      -- "0001"
```

## Edge cases

- The companion `UNHEX` is the inverse only for byte/string inputs; `HEX(int)` reverses with `CAST(CONCAT("0x", HEX(x)) AS INT64)` style conversions.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#hex>.
