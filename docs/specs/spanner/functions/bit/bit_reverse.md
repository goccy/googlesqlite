---
name: BIT_REVERSE
dialect: spanner
category: functions/bit
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/bit_functions#bit_reverse
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/bit_functions#bit_reverse
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/bit/bit_reverse.yaml
---

# BIT_REVERSE

## Summary

Returns `value` with its bits reversed (most significant bit becomes least significant, etc.). The width of the reversal can optionally be restricted to a smaller bit range.

## Signatures

- `BIT_REVERSE(value, preserve_width)`

## Arguments

- `value`: `INT64`. Reversed as if it were a 64-bit twos-complement integer.
- `preserve_width`: `BOOL`. When `TRUE`, the reversal is performed on the minimal number of bits needed to represent `value` (so leading zero bits are preserved on the right). When `FALSE` (the typical form), reversal is over a full 64-bit width.

## Return type

`INT64`.

## Behavior

- For `preserve_width = FALSE`: `BIT_REVERSE(BIT_REVERSE(x, FALSE), FALSE) = x`.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT BIT_REVERSE(1, FALSE);   -- the bit pattern 0x8000000000000000
SELECT BIT_REVERSE(0, FALSE);   -- 0
```

## Edge cases

- Useful for hash partition shuffling. Treat the result as opaque rather than ordered.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/bit_functions#bit_reverse>.
