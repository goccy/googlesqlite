---
name: LN
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#ln
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/ln.yaml
---

# LN

## Summary

Computes the natural logarithm (base e) of `X`.

## Signatures

- `LN(X)`

## Behavior

- Return type is `DOUBLE` for `INT32`, `INT64`, `UINT32`, `UINT64`, `FLOAT`, and `DOUBLE` inputs; `NUMERIC` for `NUMERIC`; `BIGNUMERIC` for `BIGNUMERIC`.
- Computes the natural logarithm of `X`.
- `LN(1.0)` returns `0.0`.
- `LN(+inf)` returns `+inf`.
- Generates an error when `X <= 0`.

## Examples

```sql
SELECT LN(1.0);
-- expected 0.0
```

```sql
SELECT LN(CAST('inf' AS FLOAT64));
-- expected +inf
```

## Edge cases

- Raises an error when `X` is zero or negative.
- `LN(+inf)` propagates positive infinity.
- `NULL` input yields `NULL` (standard SQL function behavior).

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/mathematical_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `LN`

```
LN(X)
```

**Description**

Computes the natural logarithm of X. Generates an error if X is less than or
equal to zero.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>LN(X)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>1.0</td>
      <td>0.0</td>
    </tr>
    <tr>
      <td><code>+inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td><code>X &lt;= 0</code></td>
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
