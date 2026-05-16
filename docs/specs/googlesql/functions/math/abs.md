---
name: ABS
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#abs
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/abs.yaml
---

# ABS

## Summary

Computes the absolute value of `X`.

## Signatures

- `ABS(X)`

## Behavior

- Return type matches the input type: `INT32`, `INT64`, `UINT32`, `UINT64`, `NUMERIC`, `BIGNUMERIC`, `FLOAT`, or `DOUBLE`.
- Returns the magnitude of `X`: positive inputs are returned unchanged and negative inputs are negated.
- For unsigned integer inputs, the value is already non-negative and is returned unchanged.
- `ABS(+inf)` and `ABS(-inf)` both return `+inf`.
- Raises an error when the argument is an integer whose absolute value cannot be represented as the same type (i.e. the largest negative value of a signed integer type).

## Examples

```sql
SELECT ABS(25);
-- expected: 25
```

```sql
SELECT ABS(-25);
-- expected: 25
```

```sql
SELECT ABS(CAST('-inf' AS DOUBLE));
-- expected: +inf
```

## Edge cases

- For signed integer types, applying `ABS` to the most negative representable value raises an error because the result has no positive representation in the same type.
- `ABS(-inf)` evaluates to `+inf` rather than raising.
- `NULL` input propagates to a `NULL` result (standard SQL function behaviour).

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/mathematical_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ABS`

```
ABS(X)
```

**Description**

Computes absolute value. Returns an error if the argument is an integer and the
output value can't be represented as the same type; this happens only for the
largest negative input value, which has no positive representation.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>ABS(X)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>25</td>
      <td>25</td>
    </tr>
    <tr>
      <td>-25</td>
      <td>25</td>
    </tr>
    <tr>
      <td><code>+inf</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>+inf</code></td>
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
<tr><th>OUTPUT</th><td style="vertical-align:middle"><code>INT32</code></td><td style="vertical-align:middle"><code>INT64</code></td><td style="vertical-align:middle"><code>UINT32</code></td><td style="vertical-align:middle"><code>UINT64</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>FLOAT</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
</tbody>

</table>

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
