---
name: ST_ISEMPTY
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Whether the geography contains zero coordinate positions. Runtime entry: BindStIsEmpty in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_isempty
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_isempty.yaml
---

# ST_ISEMPTY

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

## `ST_ISEMPTY`

```googlesql
ST_ISEMPTY(geography_expression)
```

**Description**

Returns `TRUE` if the given `GEOGRAPHY` is empty; that is, the `GEOGRAPHY`
doesn't contain any points, lines, or polygons.

NOTE: An empty `GEOGRAPHY` isn't associated with a particular geometry shape.
For example, the results of expressions `ST_GEOGFROMTEXT('POINT EMPTY')` and
`ST_GEOGFROMTEXT('GEOMETRYCOLLECTION EMPTY')` are identical.

**Return type**

`BOOL`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
