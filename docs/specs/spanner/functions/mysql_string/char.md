---
name: CHAR
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindChar in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#char
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#char
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/char.yaml
---

# CHAR

## Summary

Returns the string formed by interpreting each integer argument as a Unicode code point.

## Signatures

- `CHAR(n1[, n2, ...])`

## Arguments

- `n1, n2, ...`: `INT64` Unicode code points in the range `[0, 1114111]`.

## Return type

`STRING`.

## Behavior

- The result is the concatenation of the characters whose code points are the given integers, in order.
- Out-of-range code points raise an error.
- A `NULL` argument propagates: the entire result is `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT CHAR(72, 105);       -- "Hi"
SELECT CHAR(9731);          -- "☃"
```

## Edge cases

- Surrogate-half code points (`0xD800`–`0xDFFF`) are rejected as invalid Unicode scalars.
- Code points above `0x10FFFF` return an error.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#char>.
