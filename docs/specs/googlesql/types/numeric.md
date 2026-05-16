---
name: NUMERIC
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#numeric-types
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/numeric.yaml
---

# NUMERIC

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

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## Numeric types 
<a id="numeric_types"></a>

Numeric types include the following types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `NUMERIC` with alias `DECIMAL`
+ `BIGNUMERIC` with alias `BIGDECIMAL`
+ `FLOAT`
+ `DOUBLE`

### Integer types 
<a id="integer_types"></a>

Integers are numeric values that don't have fractional components.

<table>
<thead>
<tr>
<th>Name</th>
<th>Range</th>
</tr>
</thead>
<tbody>

<tr>
<td><code>INT32</code></td>
<td>-2,147,483,648 to 2,147,483,647</td>
</tr>

<tr>
<td><code>UINT32</code></td>
<td>0 to 4,294,967,295</td>
</tr>

<tr>
<td><code>INT64</code>
</td>
<td>-9,223,372,036,854,775,808 to 9,223,372,036,854,775,807</td>
</tr>

<tr>
<td><code>UINT64</code></td>
<td>0 to 18,446,744,073,709,551,615</td>
</tr>

</tbody>
</table>

To learn more about the literal representation of an integer type,
see [Integer literals][integer-literals].

### Decimal types 
<a id="decimal_types"></a>

Decimal type values are numeric values with fixed decimal precision and scale.
Precision is the number of digits that the number contains. Scale is
how many of these digits appear after the decimal point.

This type can represent decimal fractions exactly, and is suitable for financial
calculations.

<table>
<thead>
<tr>
  <th>Name</th>
  <th>Precision, Scale, and Range</th>
</tr>
</thead>
<tbody>

<tr id="numeric_type">
  <td id="numeric-type" style="vertical-align:middle"><code>NUMERIC</code>
    <br><code>DECIMAL</code></td>
  <td style="vertical-align:middle">
    Precision: 38<br>
    Scale: 9<br>
    Minimum value greater than 0 that can be handled: 1e-9<br>
    Min: -9.9999999999999999999999999999999999999E+28<br>
    Max: 9.9999999999999999999999999999999999999E+28<br>
  </td>
</tr>

<tr id="bignumeric_type">
  <td id="bignumeric-type" style="vertical-align:middle"><code>BIGNUMERIC</code>
    <br><code>BIGDECIMAL</code></td>
  <td style="vertical-align:middle">
    Precision: approximately 76.8 digits (the 77th digit is partial)<br>
    Scale: 38<br>
    Minimum value greater than 0 that can be handled: 1e-38<br>
    Min: <small>-5.7896044618658097711785492504343953926634992332820282019728792003956564819968E+38</small><br>
    Max: <small>5.7896044618658097711785492504343953926634992332820282019728792003956564819967E+38</small><br>
  </td>
</tr>

</tbody>
</table>

`DECIMAL` is an alias for `NUMERIC`.

`BIGDECIMAL` is an alias for `BIGNUMERIC`.

To learn more about the literal representation of a `NUMERIC` type,
see [`NUMERIC` literals][numeric-literals].

To learn more about the literal representation of a `BIGNUMERIC` type,
see [`BIGNUMERIC` literals][bignumeric-literals].

### Floating point types 
<a id="floating_point_types"></a>

Floating point values are approximate numeric values with fractional components.

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr id="float_type">
  <td id="float-type" style="vertical-align:middle"><code>FLOAT</code>
    <br><code>FLOAT32</code></td>
<td>
  Single precision (approximate) numeric values.
</td>
</tr>

<tr id="double_type">
  <td id="double-type" style="vertical-align:middle"><code>DOUBLE</code>
    <br><code>FLOAT64</code></td>
<td>Double precision (approximate) numeric values.</td>
</tr>
</tbody>
</table>

`FLOAT32` is an alias for `FLOAT`.

`FLOAT64` is an alias for `DOUBLE`.

To learn more about the literal representation of a floating point type,
see [Floating point literals][floating-point-literals].

#### Floating point semantics

When working with floating point numbers, there are special non-numeric values
that need to be considered: `NaN` and `+/-inf`

Arithmetic operators provide standard IEEE-754 behavior for all finite input
values that produce finite output and for all operations for which at least one
input is non-finite.

Function calls and operators return an overflow error if the input is finite
but the output would be non-finite. If the input contains non-finite values, the
output can be non-finite. In general functions don't introduce `NaN`s or
`+/-inf`. However, specific functions like `IEEE_DIVIDE` can return non-finite
values on finite input. All such cases are noted explicitly in
[Mathematical functions][mathematical-functions].

Floating point values are approximations.

+ The binary format used to represent floating point values can only represent
  a subset of the numbers between the most positive number and most
  negative number in the value range. This enables efficient handling of a
  much larger range than would be possible otherwise.
  Numbers that aren't exactly representable are approximated by utilizing a
  close value instead. For example, `0.1` can't be represented as an integer
  scaled by a power of `2`. When this value is displayed as a string, it's
  rounded to a limited number of digits, and the value approximating `0.1`
  might appear as `"0.1"`, hiding the fact that the value isn't precise.
  In other situations, the approximation can be visible.
+ Summation of floating point values might produce surprising results because
  of [limited precision][floating-point-accuracy]. For example,
  `(1e30 + 1) - 1e30 = 0`, while `(1e30 - 1e30) + 1 = 1.0`. This is
  because the floating point value doesn't have enough precision to
  represent `(1e30 + 1)`, and the result is rounded to `1e30`.
  This example also shows that the result of the `SUM` aggregate function of
  floating points values depends on the order in which the values are
  accumulated. In general, this order isn't deterministic and therefore the
  result isn't deterministic. Thus, the resulting `SUM` of
  floating point values might not be deterministic and two executions of the
  same query on the same tables might produce different results.
+ If the above points are concerning, use a
  [decimal type][decimal-types] instead.

##### Mathematical function examples

<table>
<thead>
<tr>
<th>Left Term</th>
<th>Operator</th>
<th>Right Term</th>
<th>Returns</th>
</tr>
</thead>
<tbody>
<tr>
<td>Any value</td>
<td><code>+</code></td>
<td><code>NaN</code></td>
<td><code>NaN</code></td>
</tr>
<tr>
<td>1.0</td>
<td><code>+</code></td>
<td><code>+inf</code></td>
<td><code>+inf</code></td>
</tr>
<tr>
<td>1.0</td>
<td><code>+</code></td>
<td><code>-inf</code></td>
<td><code>-inf</code></td>
</tr>
<tr>
<td><code>-inf</code></td>
<td><code>+</code></td>
<td><code>+inf</code></td>
<td><code>NaN</code></td>
</tr>
<tr>
<td>Maximum <code>DOUBLE</code> value</td>
<td><code>+</code></td>
<td>Maximum <code>DOUBLE</code> value</td>
<td>Overflow error</td>
</tr>
<tr>
<td>Minimum <code>DOUBLE</code> value</td>
<td><code>/</code></td>
<td>2.0</td>
<td>0.0</td>
</tr>
<tr>
<td>1.0</td>
<td><code>/</code></td>
<td><code>0.0</code></td>
<td>"Divide by zero" error</td>
</tr>
</tbody>
</table>

Comparison operators provide standard IEEE-754 behavior for floating point
input.

##### Comparison operator examples

<table>
<thead>
<tr>
<th>Left Term</th>
<th>Operator</th>
<th>Right Term</th>
<th>Returns</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>NaN</code></td>
<td><code>=</code></td>
<td>Any value</td>
<td><code>FALSE</code></td>
</tr>
<tr>
<td><code>NaN</code></td>
<td><code>&lt;</code></td>
<td>Any value</td>
<td><code>FALSE</code></td>
</tr>
<tr>
<td>Any value</td>
<td><code>&lt;</code></td>
<td><code>NaN</code></td>
<td><code>FALSE</code></td>
</tr>
<tr>
<td>-0.0</td>
<td><code>=</code></td>
<td>0.0</td>
<td><code>TRUE</code></td>
</tr>
<tr>
<td>-0.0</td>
<td><code>&lt;</code></td>
<td>0.0</td>
<td><code>FALSE</code></td>
</tr>
</tbody>
</table>

For more information on how these values are ordered and grouped so they
can be compared,
see [Ordering floating point values][orderable-floating-points].

[floating-point-accuracy]: https://en.wikipedia.org/wiki/Floating-point_arithmetic#Accuracy_problems

[decimal-types]: #decimal_types

[orderable-floating-points]: #orderable_floating_points

[integer-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#integer_literals

[floating-point-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#floating_point_literals

[mathematical-functions]: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md

[numeric-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#numeric_literals

[bignumeric-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#bignumeric_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
