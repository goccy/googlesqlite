---
name: ARRAY_MIN
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_min
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_min.yaml
---

# ARRAY_MIN

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

## `ARRAY_MIN`

```googlesql
ARRAY_MIN(input_array)
```

**Description**

Returns the minimum non-`NULL` value in an array.

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
SELECT ARRAY_MIN([8, 37, NULL, 4, 55]) as min

/*-----+
 | min |
 +-----+
 | 4   |
 +-----*/
```

[data-type-properties]: https://github.com/google/googlesql/blob/master/docs/data-types.md#data_type_properties

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
