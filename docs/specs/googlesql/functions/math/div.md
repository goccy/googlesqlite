---
name: DIV
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#div
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/div.yaml
---

# DIV

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

## `DIV`

```
DIV(X, Y)
```

**Description**

Returns the result of integer division of X by Y. Division by zero returns
an error. Division by -1 may overflow.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>Y</th>
      <th>DIV(X, Y)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>20</td>
      <td>4</td>
      <td>5</td>
    </tr>
    <tr>
      <td>12</td>
      <td>-7</td>
      <td>-1</td>
    </tr>
    <tr>
      <td>20</td>
      <td>3</td>
      <td>6</td>
    </tr>
    <tr>
      <td>0</td>
      <td>20</td>
      <td>0</td>
    </tr>
    <tr>
      <td>20</td>
      <td>0</td>
      <td>Error</td>
    </tr>
  </tbody>
</table>

**Return Data Type**

The return data type is determined by the argument types with the following
table.
<table>

<thead>
<tr>
<th>INPUT</th><th><code>INT32</code></th><th><code>INT64</code></th><th><code>UINT32</code></th><th><code>UINT64</code></th><th><code>NUMERIC</code></th><th><code>BIGNUMERIC</code></th>
</tr>
</thead>
<tbody>
<tr><th><code>INT32</code></th><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle">ERROR</td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td></tr>
<tr><th><code>INT64</code></th><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle">ERROR</td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td></tr>
<tr><th><code>UINT32</code></th><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle"><code>UINT64</code></td><td style="vertical-align:middle"><code>UINT64</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td></tr>
<tr><th><code>UINT64</code></th><td style="vertical-align:middle">ERROR</td><td style="vertical-align:middle">ERROR</td><td style="vertical-align:middle"><code>UINT64</code></td><td style="vertical-align:middle"><code>UINT64</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td></tr>
<tr><th><code>NUMERIC</code></th><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td></tr>
<tr><th><code>BIGNUMERIC</code></th><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td></tr>
</tbody>

</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
