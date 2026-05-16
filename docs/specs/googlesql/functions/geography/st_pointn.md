---
name: ST_POINTN
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Nth (1-based) vertex of a LINESTRING; negative N counts from end. Runtime entry: BindStPointN in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_pointn
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_pointn.yaml
---

# ST_POINTN

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

## `ST_POINTN`

```googlesql
ST_POINTN(linestring_geography, index)
```

**Description**

Returns the Nth point of a linestring geography as a point geography, where N is
the index. The index is 1-based. Negative values are counted backwards from the
end of the linestring, so that -1 is the last point. Returns an error if the
input isn't a linestring, if the input is empty, or if there is no vertex at
the given index. Use the `SAFE` prefix to obtain `NULL` for invalid input
instead of an error.

**Return Type**

Point `GEOGRAPHY`

**Example**

The following example uses `ST_POINTN`, [`ST_STARTPOINT`][st-startpoint] and
[`ST_ENDPOINT`][st-endpoint] to extract points from a linestring.

```googlesql
WITH linestring AS (
    SELECT ST_GEOGFROMTEXT('LINESTRING(1 1, 2 1, 3 2, 3 3)') g
)
SELECT ST_POINTN(g, 1) AS first, ST_POINTN(g, -1) AS last,
    ST_POINTN(g, 2) AS second, ST_POINTN(g, -2) AS second_to_last
FROM linestring;

/*--------------+--------------+--------------+----------------+
 | first        | last         | second       | second_to_last |
 +--------------+--------------+--------------+----------------+
 | POINT(1 1)   | POINT(3 3)   | POINT(2 1)   | POINT(3 2)     |
 +--------------+--------------+--------------+----------------*/
```

[st-startpoint]: #st_startpoint

[st-endpoint]: #st_endpoint

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
