---
name: OCT
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindOct in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#oct
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#oct
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/oct.yaml
---

# OCT

## Summary

Returns the octal (base-8) string representation of an integer.

## Signatures

- `OCT(n)`

## Arguments

- `n`: `INT64`.

## Return type

`STRING`. No leading zero is emitted (matching MySQL); the literal `0` returns `"0"`.

## Behavior

- Negative integers are formatted as the twos-complement of their 64-bit representation, producing up to 22 octal digits.
- `NULL` input returns `NULL`.

## Examples

```sql
SELECT OCT(8);     -- "10"
SELECT OCT(255);   -- "377"
SELECT OCT(0);     -- "0"
SELECT OCT(-1);    -- "1777777777777777777777"
```

## Edge cases

- The result is right-aligned by value, not zero-padded. Use `LPAD` if a fixed width is required.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#oct>.
