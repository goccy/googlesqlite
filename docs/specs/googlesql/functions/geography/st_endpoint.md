---
name: ST_ENDPOINT
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Last vertex of a LINESTRING as a POINT. Runtime entry: BindStEndPoint in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_endpoint
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_endpoint.yaml
---

# ST_ENDPOINT

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

## `ST_ENDPOINT`

```googlesql
ST_ENDPOINT(linestring_geography)
```

**Description**

Returns the last point of a linestring geography as a point geography. Returns
an error if the input isn't a linestring or if the input is empty. Use the
`SAFE` prefix to obtain `NULL` for invalid input instead of an error.

**Return Type**

Point `GEOGRAPHY`

**Example**

```googlesql
SELECT ST_ENDPOINT(ST_GEOGFROMTEXT('LINESTRING(1 1, 2 1, 3 2, 3 3)')) last

/*--------------+
 | last         |
 +--------------+
 | POINT(3 3)   |
 +--------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
