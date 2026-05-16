---
name: LOG10
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#log10
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/log10.yaml
---

# LOG10

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

## `LOG10`

```
LOG10(X)
```

**Description**

Similar to `LOG`, but computes logarithm to base 10.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>LOG10(X)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>100.0</td>
      <td>2.0</td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>+inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td>X &lt;= 0</td>
      <td>Error</td>
    </tr>
  </tbody>
</table>

**Return Data Type**

<table>

<thead>
<tr>
<th>INPUT</th><th><code>INT32</code></th><th><code>INT64</code></th><th><code>UINT32</code></th><th><code>UINT64</code></th><th><code>NUMERIC</code></th><th><code>BIGNUMERIC</code></th><th><code>FLOAT</code></th><th><code>DOUBLE</code></th>
</tr>
</thead>
<tbody>
<tr><th>OUTPUT</th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
</tbody>

</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
