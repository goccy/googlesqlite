---
name: ST_ISCLOSED
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  True when every constituent LINESTRING starts and ends at the same vertex. Runtime entry: BindStIsClosed in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_isclosed
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_isclosed.yaml
---

# ST_ISCLOSED

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

## `ST_ISCLOSED`

```googlesql
ST_ISCLOSED(geography_expression)
```

**Description**

Returns `TRUE` for a non-empty Geography, where each element in the Geography
has an empty boundary. The boundary for each element can be defined with
[`ST_BOUNDARY`][st-boundary].

+   A point is closed.
+   A linestring is closed if the start and end points of the linestring are
    the same.
+   A polygon is closed only if it's a full polygon.
+   A collection is closed if and only if every element in the collection is
    closed.

An empty `GEOGRAPHY` isn't closed.

**Return type**

`BOOL`

**Example**

```googlesql
WITH example AS(
  SELECT ST_GEOGFROMTEXT('POINT(5 0)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('LINESTRING(0 1, 4 3, 2 6, 0 1)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('LINESTRING(2 6, 1 3, 3 9)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1))') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION EMPTY'))
SELECT
  geography,
  ST_ISCLOSED(geography) AS is_closed,
FROM example;

/*------------------------------------------------------+-----------+
 | geography                                            | is_closed |
 +------------------------------------------------------+-----------+
 | POINT(5 0)                                           | TRUE      |
 | LINESTRING(0 1, 4 3, 2 6, 0 1)                       | TRUE      |
 | LINESTRING(2 6, 1 3, 3 9)                            | FALSE     |
 | GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1)) | FALSE     |
 | GEOMETRYCOLLECTION EMPTY                             | FALSE     |
 +------------------------------------------------------+-----------*/
```

[st-boundary]: #st_boundary

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
