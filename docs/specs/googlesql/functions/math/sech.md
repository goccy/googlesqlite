---
name: SECH
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#sech
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/sech.yaml
---

# SECH

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

## `SECH`

```
SECH(X)
```

**Description**

Computes the hyperbolic secant for the angle of `X`, where `X` is specified
in radians. `X` can be any data type
that [coerces to `DOUBLE`][conversion-rules].
Never produces an error.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>SECH(X)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td><code>+inf</code></td>
      <td><code>0</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>0</code></td>
    </tr>
    <tr>
      <td><code>NaN</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>NULL</code></td>
      <td><code>NULL</code></td>
    </tr>
  </tbody>
</table>

**Return Data Type**

`DOUBLE`

**Example**

```googlesql
SELECT SECH(0.5) AS a, SECH(-2) AS b, SECH(100) AS c;

/*----------------+----------------+---------------------+
 | a              | b              | c                   |
 +----------------+----------------+---------------------+
 | 0.88681888397  | 0.265802228834 | 7.4401519520417E-44 |
 +----------------+----------------+---------------------*/
```

[conversion-rules]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
