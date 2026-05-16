---
name: ST_EQUALS
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_equals
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_equals.yaml
---

# ST_EQUALS

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

Verbatim copy from `docs/third_party/googlesql-docs/geography_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ST_EQUALS`

```googlesql
ST_EQUALS(geography_1, geography_2)
```

**Description**

Checks if two `GEOGRAPHY` values represent the same `GEOGRAPHY` value. Returns
`TRUE` if the values are the same, otherwise returns `FALSE`.

**Definitions**

+ `geography_1`: The first `GEOGRAPHY` value to compare.
+ `geography_2`: The second `GEOGRAPHY` value to compare.

**Details**

As long as they still represent the same geometric structure, two
`GEOGRAPHY` values can be equal even if the ordering of points or vertices
differ. This means that one of the following conditions must be true for this
function to return `TRUE`:

+   Both `ST_COVERS(geography_1, geography_2)` and
    `ST_COVERS(geography_2, geography_1)` are `TRUE`.
+   Both `geography_1` and `geography_2` are empty.

`ST_EQUALS` isn't guaranteed to be a transitive function.

**Return type**

`BOOL`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
