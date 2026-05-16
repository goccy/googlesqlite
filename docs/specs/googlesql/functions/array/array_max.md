---
name: ARRAY_MAX
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_max
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_max.yaml
---

# ARRAY_MAX

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

## `ARRAY_MAX`

```googlesql
ARRAY_MAX(input_array)
```

**Description**

Returns the maximum non-`NULL` value in an array.

Caveats:

+ If the array is `NULL`, empty, or contains only `NULL`s, returns
  `NULL`.
+ If the array contains `NaN`, returns `NaN`.

**Supported Argument Types**

In the input array, `ARRAY<T>`, `T` can be an
[orderable data type][data-type-properties].

**Return type**

The same data type as `T` in the input array.

**Examples**

```googlesql
SELECT ARRAY_MAX([8, 37, NULL, 55, 4]) as max

/*-----+
 | max |
 +-----+
 | 55  |
 +-----*/
```

[data-type-properties]: https://github.com/google/googlesql/blob/master/docs/data-types.md#data_type_properties

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
