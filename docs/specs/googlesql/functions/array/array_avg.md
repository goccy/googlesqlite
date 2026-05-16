---
name: ARRAY_AVG
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_avg
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_avg.yaml
---

# ARRAY_AVG

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

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY_AVG`

```googlesql
ARRAY_AVG(input_array)
```

**Description**

Returns the average of non-`NULL` values in an array.

Caveats:

+ If the array is `NULL`, empty, or contains only `NULL`s, returns
  `NULL`.
+ If the array contains `NaN`, returns `NaN`.
+ If the array contains `[+|-]Infinity`, returns either `[+|-]Infinity`
  or `NaN`.
+ If there is numeric overflow, produces an error.
+ If a [floating-point type][floating-point-types] is returned, the result is
  [non-deterministic][non-deterministic], which means you might receive a
  different result each time you use this function.

[floating-point-types]: https://github.com/google/googlesql/blob/master/docs/data-types.md#floating_point_types

[non-deterministic]: https://github.com/google/googlesql/blob/master/docs/data-types.md#floating_point_semantics

**Supported Argument Types**

In the input array, `ARRAY<T>`, `T` can represent one of the following
data types:

+ Any numeric input type
+ `INTERVAL`

**Return type**

The return type depends upon `T` in the input array:

<table>

<thead>
<tr>
<th>INPUT</th><th><code>INT32</code></th><th><code>INT64</code></th><th><code>UINT32</code></th><th><code>UINT64</code></th><th><code>NUMERIC</code></th><th><code>BIGNUMERIC</code></th><th><code>FLOAT</code></th><th><code>DOUBLE</code></th><th><code>INTERVAL</code></th>
</tr>
</thead>
<tbody>
<tr><th>OUTPUT</th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>INTERVAL</code></td></tr>
</tbody>

</table>

**Examples**

```googlesql
SELECT ARRAY_AVG([0, 2, NULL, 4, 4, 5]) as avg

/*-----+
 | avg |
 +-----+
 | 3   |
 +-----*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
