---
name: COTH
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#coth
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/coth.yaml
---

# COTH

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

## `COTH`

```
COTH(X)
```

**Description**

Computes the hyperbolic cotangent for the angle of `X`, where `X` is specified
in radians. `X` can be any data type
that [coerces to `DOUBLE`][conversion-rules].
Supports the `SAFE.` prefix.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>COTH(X)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td><code>+inf</code></td>
      <td><code>1</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>-1</code></td>
    </tr>
    <tr>
      <td><code>NaN</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>0</code></td>
      <td><code>Error</code></td>
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
SELECT COTH(1) AS a, SAFE.COTH(0) AS b;

/*----------------+------+
 | a              | b    |
 +----------------+------+
 | 1.313035285499 | NULL |
 +----------------+------*/
```

[conversion-rules]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
