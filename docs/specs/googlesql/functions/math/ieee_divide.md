---
name: IEEE_DIVIDE
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#ieee_divide
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/ieee_divide.yaml
---

# IEEE_DIVIDE

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

## `IEEE_DIVIDE`

```
IEEE_DIVIDE(X, Y)
```

**Description**

Divides X by Y; this function never fails. Returns
`DOUBLE` unless
both X and Y are `FLOAT`, in which case it returns
`FLOAT`. Unlike the division operator (/),
this function doesn't generate errors for division by zero or overflow.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>Y</th>
      <th>IEEE_DIVIDE(X, Y)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>20.0</td>
      <td>4.0</td>
      <td>5.0</td>
    </tr>
    <tr>
      <td>20.0</td>
      <td>6.0</td>
      <td>3.3333333333333335</td>
    </tr>
    <tr>
      <td>0.0</td>
      <td>25.0</td>
      <td>0.0</td>
    </tr>
    <tr>
      <td>25.0</td>
      <td>0.0</td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td>-25.0</td>
      <td>0.0</td>
      <td><code>-inf</code></td>
    </tr>
    <tr>
      <td>25.0</td>
      <td>-0.0</td>
      <td><code>-inf</code></td>
    </tr>
    <tr>
      <td>0.0</td>
      <td>0.0</td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td>0.0</td>
      <td><code>NaN</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>NaN</code></td>
      <td>0.0</td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>+inf</code></td>
      <td><code>+inf</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>-inf</code></td>
      <td><code>NaN</code></td>
    </tr>
  </tbody>
</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
