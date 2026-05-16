---
name: SEC
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#sec
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/sec.yaml
---

# SEC

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

## `SEC`

```
SEC(X)
```

**Description**

Computes the secant for the angle of `X`, where `X` is specified in radians.
`X` can be any data type
that [coerces to `DOUBLE`][conversion-rules].

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>SEC(X)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td><code>+inf</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>NaN</code></td>
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
SELECT SEC(100) AS a, SEC(-1) AS b;

/*----------------+---------------+
 | a              | b             |
 +----------------+---------------+
 | 1.159663822905 | 1.85081571768 |
 +----------------+---------------*/
```

[conversion-rules]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
