---
name: CBRT
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#cbrt
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/cbrt.yaml
---

# CBRT

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

## `CBRT`

```
CBRT(X)
```

**Description**

Computes the cube root of `X`. `X` can be any data type
that [coerces to `DOUBLE`][conversion-rules].
Supports the `SAFE.` prefix.

<table>
  <thead>
    <tr>
      <th>X</th>
      <th>CBRT(X)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td><code>+inf</code></td>
      <td><code>inf</code></td>
    </tr>
    <tr>
      <td><code>-inf</code></td>
      <td><code>-inf</code></td>
    </tr>
    <tr>
      <td><code>NaN</code></td>
      <td><code>NaN</code></td>
    </tr>
    <tr>
      <td><code>0</code></td>
      <td><code>0</code></td>
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
SELECT CBRT(27) AS cube_root;

/*--------------------+
 | cube_root          |
 +--------------------+
 | 3.0000000000000004 |
 +--------------------*/
```

[conversion-rules]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
