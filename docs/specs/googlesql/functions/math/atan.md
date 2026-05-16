---
name: ATAN
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#atan
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/atan.yaml
---

# ATAN

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

## `ATAN`

```
ATAN(X)
```

**Description**

Computes the principal value of the inverse tangent of X. The return value is
in the range [-&pi;/2,&pi;/2]. Doesn't fail.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>ATAN(X)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>+inf</code></td>
      <td>&pi;/2</td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td>-&pi;/2</td>
    </tr>
    <tr>
      <td><code>NaN</code></td>
      <td><code>NaN</code></td>
    </tr>
  </tbody>
</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
