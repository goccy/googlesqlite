---
name: COSH
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#cosh
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/cosh.yaml
---

# COSH

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

## `COSH`

```
COSH(X)
```

**Description**

Computes the hyperbolic cosine of X where X is specified in radians.
Generates an error if overflow occurs.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>COSH(X)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>+inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td><code>NaN</code></td>
      <td><code>NaN</code></td>
    </tr>
  </tbody>
</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
