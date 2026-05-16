---
name: PI_BIGNUMERIC
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#pi_bignumeric
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/pi_bignumeric.yaml
---

# PI_BIGNUMERIC

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

## `PI_BIGNUMERIC`

```googlesql
PI_BIGNUMERIC()
```

**Description**

Returns the mathematical constant `π` as a `BIGNUMERIC` value.

**Return type**

`BIGNUMERIC`

**Example**

```googlesql
SELECT PI_BIGNUMERIC() AS pi

/*-----------------------------------------+
 | pi                                      |
 +-----------------------------------------+
 | 3.1415926535897932384626433832795028842 |
 +-----------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
