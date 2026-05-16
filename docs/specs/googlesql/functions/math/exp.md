---
name: EXP
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#exp
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/exp.yaml
---

# EXP

## Summary

Computes *e* raised to the power of `X` (the natural exponential function).

## Signatures

- `EXP(X)`

## Behavior

- Return type is `DOUBLE` for integer and floating-point inputs, `NUMERIC` for `NUMERIC` input, and `BIGNUMERIC` for `BIGNUMERIC` input.
- `EXP(0.0)` returns `1.0`.
- `EXP(+inf)` returns `+inf`.
- `EXP(-inf)` returns `0.0`.
- If the result underflows, returns zero.
- If the result overflows, raises an error.

## Examples

```sql
SELECT EXP(0.0);
-- expected 1.0
```

```sql
SELECT EXP(CAST('+inf' AS FLOAT64));
-- expected +inf
```

```sql
SELECT EXP(CAST('-inf' AS FLOAT64));
-- expected 0.0
```

## Edge cases

- Underflowing results are flushed to zero rather than producing a denormal/error.
- Overflowing results raise an error instead of returning `+inf`.
- `EXP(-inf)` is defined as `0.0`, and `EXP(+inf)` is defined as `+inf`.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/mathematical_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `EXP`

```
EXP(X)
```

**Description**

Computes *e* to the power of X, also called the natural exponential function. If
the result underflows, this function returns a zero. Generates an error if the
result overflows.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>EXP(X)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>0.0</td>
      <td>1.0</td>
    </tr>
    <tr>
      <td><code>+inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td>0.0</td>
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
