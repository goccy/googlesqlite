---
name: ST_LINESUBSTRING
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStLineSubstring walks every vertex of the polyline whose
  arc-length fraction falls strictly between `start` and `end`,
  with new endpoint vertices interpolated at the exact cut
  positions via `s2.Polyline.Interpolate`. When `start == end`,
  returns the single interpolated POINT (matches Example 2).
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_linesubstring
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_linesubstring.yaml
---

# ST_LINESUBSTRING

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

## `ST_LINESUBSTRING`

```googlesql
ST_LINESUBSTRING(linestring_geography, start_fraction, end_fraction);
```

**Description**

Gets a segment of a linestring at a specific starting and ending fraction.

**Definitions**

+   `linestring_geography`: The LineString `GEOGRAPHY` value that represents the
    linestring from which to extract a segment.
+   `start_fraction`: `DOUBLE` value that represents
    the starting fraction of the total length of `linestring_geography`.
    This must be an inclusive value between 0 and 1 (0-100%).
+   `end_fraction`: `DOUBLE` value that represents
    the ending fraction of the total length of `linestring_geography`.
    This must be an inclusive value between 0 and 1 (0-100%).

**Details**

`end_fraction` must be greater than or equal to `start_fraction`.

If `start_fraction` and `end_fraction` are equal, a linestring with only
one point is produced.

**Return type**

+   LineString `GEOGRAPHY` if the resulting geography has more than one point.
+   Point `GEOGRAPHY` if the resulting geography has only one point.

**Example**

The following query returns the second half of the linestring:

```googlesql
WITH data AS (
  SELECT ST_GEOGFROMTEXT('LINESTRING(20 70, 70 60, 10 70, 70 70)') AS geo1
)
SELECT ST_LINESUBSTRING(geo1, 0.5, 1) AS segment
FROM data;

/*-------------------------------------------------------------+
 | segment                                                     |
 +-------------------------------------------------------------+
 | LINESTRING(49.4760661523471 67.2419539103851, 10 70, 70 70) |
 +-------------------------------------------------------------*/
```

The following query returns a linestring that only contains one point:

```googlesql
WITH data AS (
  SELECT ST_GEOGFROMTEXT('LINESTRING(20 70, 70 60, 10 70, 70 70)') AS geo1
)
SELECT ST_LINESUBSTRING(geo1, 0.5, 0.5) AS segment
FROM data;

/*------------------------------------------+
 | segment                                  |
 +------------------------------------------+
 | POINT(49.4760661523471 67.2419539103851) |
 +------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
