---
name: SAFE_NEGATE
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#safe_negate
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/safe_negate.yaml
---

# SAFE_NEGATE

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

## `SAFE_NEGATE`

```
SAFE_NEGATE(X)
```

**Description**

Equivalent to the unary minus operator (`-`), but returns
`NULL` if overflow occurs.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>SAFE_NEGATE(X)</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>+1</td>
      <td>-1</td>
    </tr>
    <tr>
      <td>-1</td>
      <td>+1</td>
    </tr>
    <tr>
      <td>0</td>
      <td>0</td>
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
<tr><th>OUTPUT</th><td style="vertical-align:middle"><code>INT32</code></td><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle">ERROR</td><td style="vertical-align:middle">ERROR</td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>FLOAT</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
</tbody>

</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
