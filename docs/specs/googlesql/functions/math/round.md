---
name: ROUND
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#round
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/round.yaml
---

# ROUND

## Summary

Rounds a numeric value `X` to the nearest integer, or to `N` decimal places when `N` is supplied. For `NUMERIC` and `BIGNUMERIC` inputs, an optional `rounding_mode` selects the half-case tie-breaking rule.

## Signatures

- `ROUND(X)`
- `ROUND(X, N)`
- `ROUND(X, N, rounding_mode)` — only when `X` is `NUMERIC` or `BIGNUMERIC`

## Behavior

- Return type is `NUMERIC` for `NUMERIC` input, `BIGNUMERIC` for `BIGNUMERIC` input, and `DOUBLE` for all other numeric input types (`INT32`, `INT64`, `UINT32`, `UINT64`, `FLOAT`, `DOUBLE`).
- With only `X`, rounds to the nearest integer.
- With `N`, rounds `X` to `N` decimal places after the decimal point.
- A negative `N` rounds digits to the left of the decimal point.
- The default tie-breaking rule rounds halfway cases away from zero.
- For `NUMERIC` or `BIGNUMERIC` inputs, `rounding_mode` may be `"ROUND_HALF_AWAY_FROM_ZERO"` (default) or `"ROUND_HALF_EVEN"` (rounds halfway cases toward the nearest even digit).
- Generates an error if the result overflows.

## Examples

```sql
SELECT ROUND(2.5);
-- expected 3.0
```

```sql
SELECT ROUND(123.7, -1);
-- expected 120.0
```

```sql
SELECT ROUND(NUMERIC "2.25", 1, "ROUND_HALF_EVEN");
-- expected 2.2
```

## Edge cases

- `ROUND(+inf)` returns `+inf`, `ROUND(-inf)` returns `-inf`, and `ROUND(NaN)` returns `NaN`.
- Halfway cases such as `ROUND(2.5)` and `ROUND(-2.5)` round away from zero by default (`3.0` and `-3.0`).
- Supplying `rounding_mode` for any input type other than `NUMERIC` or `BIGNUMERIC` raises an error.
- An overflow in the rounded result raises an error rather than wrapping or saturating.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/mathematical_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ROUND`

```
ROUND(X [, N [, rounding_mode]])
```

**Description**

If only X is present, rounds X to the nearest integer. If N is present,
rounds X to N decimal places after the decimal point. If N is negative,
rounds off digits to the left of the decimal point. Rounds halfway cases
away from zero. Generates an error if overflow occurs.

If X is a `NUMERIC` or `BIGNUMERIC` type, then you can
explicitly set `rounding_mode`
to one of the following:

+   [`"ROUND_HALF_AWAY_FROM_ZERO"`][round-half-away-from-zero]: (Default) Rounds
    halfway cases away from zero.
+   [`"ROUND_HALF_EVEN"`][round-half-even]: Rounds halfway cases
    towards the nearest even digit.

If you set the `rounding_mode` and X isn't a `NUMERIC` or `BIGNUMERIC` type,
then the function generates an error.

<table>
  <thead>
    <tr>
      <th>Expression</th>
      <th>Return Value</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>ROUND(2.0)</code></td>
      <td>2.0</td>
    </tr>
    <tr>
      <td><code>ROUND(2.3)</code></td>
      <td>2.0</td>
    </tr>
    <tr>
      <td><code>ROUND(2.8)</code></td>
      <td>3.0</td>
    </tr>
    <tr>
      <td><code>ROUND(2.5)</code></td>
      <td>3.0</td>
    </tr>
    <tr>
      <td><code>ROUND(-2.3)</code></td>
      <td>-2.0</td>
    </tr>
    <tr>
      <td><code>ROUND(-2.8)</code></td>
      <td>-3.0</td>
    </tr>
    <tr>
      <td><code>ROUND(-2.5)</code></td>
      <td>-3.0</td>
    </tr>
    <tr>
      <td><code>ROUND(0)</code></td>
      <td>0</td>
    </tr>
    <tr>
      <td><code>ROUND(+inf)</code></td>
      <td><code>+inf</code></td>
    </tr>
    <tr>
      <td><code>ROUND(-inf)</code></td>
      <td><code>-inf</code></td>
    </tr>
    <tr>
      <td><code>ROUND(NaN)</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>ROUND(123.7, -1)</code></td>
      <td>120.0</td>
    </tr>
    <tr>
      <td><code>ROUND(1.235, 2)</code></td>
      <td>1.24</td>
    </tr>
    <tr>
      <td><code>ROUND(NUMERIC "2.25", 1, "ROUND_HALF_EVEN")</code></td>
      <td>2.2</td>
    </tr>
    <tr>
      <td><code>ROUND(NUMERIC "2.35", 1, "ROUND_HALF_EVEN")</code></td>
      <td>2.4</td>
    </tr>
    <tr>
      <td><code>ROUND(NUMERIC "2.251", 1, "ROUND_HALF_EVEN")</code></td>
      <td>2.3</td>
    </tr>
    <tr>
      <td><code>ROUND(NUMERIC "-2.5", 0, "ROUND_HALF_EVEN")</code></td>
      <td>-2</td>
    </tr>
    <tr>
      <td><code>ROUND(NUMERIC "2.5", 0, "ROUND_HALF_AWAY_FROM_ZERO")</code></td>
      <td>3</td>
    </tr>
    <tr>
      <td><code>ROUND(NUMERIC "-2.5", 0, "ROUND_HALF_AWAY_FROM_ZERO")</code></td>
      <td>-3</td>
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

[round-half-away-from-zero]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_away_from_zero

[round-half-even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
