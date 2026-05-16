---
name: POW
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#pow
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/pow.yaml
---

# POW

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

## `POW`

```
POW(X, Y)
```

**Description**

Returns the value of X raised to the power of Y. If the result underflows and
isn't representable, then the function returns a value of zero.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>Y</th>
      <th>POW(X, Y)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>2.0</td>
      <td>3.0</td>
      <td>8.0</td>
    </tr>
    <tr>
      <td>1.0</td>
      <td>Any value including <code>NaN</code></td>
      <td>1.0</td>
    </tr>
    <tr>
      <td>Any value including <code>NaN</code></td>
      <td>0</td>
      <td>1.0</td>
    </tr>
    <tr>
      <td>-1.0</td>
      <td><code>+inf</code></td>
      <td>1.0</td>
    </tr>
    <tr>
      <td>-1.0</td>
      <td><code>-inf</code></td>
      <td>1.0</td>
    </tr>
    <tr>
      <td>ABS(X) &lt; 1</td>
      <td><code>-inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td>ABS(X) &gt; 1</td>
      <td><code>-inf</code></td>
      <td>0.0</td>
    </tr>
    <tr>
      <td>ABS(X) &lt; 1</td>
      <td><code>+inf</code></td>
      <td>0.0</td>
    </tr>
    <tr>
      <td>ABS(X) &gt; 1</td>
      <td><code>+inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td>Y &lt; 0</td>
      <td>0.0</td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td>Y &gt; 0</td>
      <td><code>-inf</code> if Y is an odd integer, <code>+inf</code> otherwise</td>
    </tr>
    <tr>
      <td><code>+inf</code></td>
      <td>Y &lt; 0</td>
      <td>0</td>
    </tr>
    <tr>
      <td><code>+inf</code></td>
      <td>Y &gt; 0</td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td>Finite value &lt; 0</td>
      <td>Non-integer</td>
      <td>Error</td>
    </tr>
    <tr>
      <td>0</td>
      <td>Finite value &lt; 0</td>
      <td>Error</td>
    </tr>
  </tbody>
</table>

**Return Data Type**

The return data type is determined by the argument types with the following
table.

<table style="font-size:small">

<thead>
<tr>
<th>INPUT</th><th><code>INT32</code></th><th><code>INT64</code></th><th><code>UINT32</code></th><th><code>UINT64</code></th><th><code>NUMERIC</code></th><th><code>BIGNUMERIC</code></th><th><code>FLOAT</code></th><th><code>DOUBLE</code></th>
</tr>
</thead>
<tbody>
<tr><th><code>INT32</code></th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>INT64</code></th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>UINT32</code></th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>UINT64</code></th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>NUMERIC</code></th><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>BIGNUMERIC</code></th><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>FLOAT</code></th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>DOUBLE</code></th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
</tbody>

</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
