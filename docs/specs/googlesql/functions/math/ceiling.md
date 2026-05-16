---
name: CEILING
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#ceiling
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/ceiling.yaml
---

# CEILING

## Summary

`CEILING(X)` is a synonym of `CEIL(X)`; it returns the smallest integral value not less than `X`.

## Signatures

- `CEILING(X)`

## Behavior

- Behaves identically to `CEIL(X)` in every respect; `CEILING` is provided as an alias.
- Returns the smallest integral value that is greater than or equal to the input `X`.
- Result type and numeric semantics match those of `CEIL(X)`; refer to the `CEIL` spec for full details.

## Examples

The upstream reference does not include dedicated examples for `CEILING`; see `CEIL` for representative usage. Equivalent calls:

```
SELECT CEILING(2.3);   -- same as CEIL(2.3)
SELECT CEILING(-1.5);  -- same as CEIL(-1.5)
```

## Edge cases

- All edge cases (NULL input, infinities, NaN, type promotion, errors) are inherited from `CEIL(X)`; the upstream reference defines no additional edge behaviour for `CEILING`.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/mathematical_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `CEILING`

```
CEILING(X)
```

**Description**

Synonym of CEIL(X)

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
