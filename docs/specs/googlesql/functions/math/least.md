---
name: LEAST
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#least
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/least.yaml
---

# LEAST

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

## `LEAST`

```
LEAST(X1,...,XN)
```

**Description**

Returns the least value among `X1,...,XN`. If any argument is `NULL`, returns
`NULL`. Otherwise, in the case of floating-point arguments, if any argument is
`NaN`, returns `NaN`. In all other cases, returns the value among `X1,...,XN`
that has the least value according to the ordering used by the `ORDER BY`
clause. The arguments `X1, ..., XN` must be coercible to a common supertype, and
the supertype must support ordering.

<table>
  <thead>
    <tr>
      <th>X1,...,XN</th>
      <th>LEAST(X1,...,XN)</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>3,5,1</td>
      <td>1</td>
    </tr>
  </tbody>
</table>

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Return Data Types**

Data type of the input values.

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
