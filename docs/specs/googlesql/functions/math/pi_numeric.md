---
name: PI_NUMERIC
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#pi_numeric
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/pi_numeric.yaml
---

# PI_NUMERIC

## Summary

(TBD — refine from the upstream reference below.)

## Signatures

(TBD)

## Behavior

(TBD)

## Examples

(TBD)

## Edge cases

(TBD)

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/mathematical_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PI_NUMERIC`

```googlesql
PI_NUMERIC()
```

**Description**

Returns the mathematical constant `π` as a `NUMERIC` value.

**Return type**

`NUMERIC`

**Example**

```googlesql
SELECT PI_NUMERIC() AS pi

/*-------------+
 | pi          |
 +-------------+
 | 3.141592654 |
 +-------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
