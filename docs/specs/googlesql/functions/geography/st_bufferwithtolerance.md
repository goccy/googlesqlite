---
name: ST_BUFFERWITHTOLERANCE
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStBufferWithTolerance lays out the buffer polygon at the
  exact vertex count
  `n = ceil(pi / arccos(1 - tolerance/radius))` (sagitta of a
  regular n-gon inscribed in a circle of radius r), so the maximum
  deviation between the buffer's circular boundary and the
  polygon's straight-line approximation is at most
  `tolerance_meters`. ST_BUFFER's `num_seg_quarter_circle` knob is
  not used here because it implicitly multiplies by 4 and the
  upstream Example expects a 5-sided polygon (n=5) for
  tolerance=25 on radius=100.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_bufferwithtolerance
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_bufferwithtolerance.yaml
---

# ST_BUFFERWITHTOLERANCE

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

## `ST_BUFFERWITHTOLERANCE`

```googlesql
ST_BUFFERWITHTOLERANCE(
    geography,
    buffer_radius,
    tolerance_meters => tolerance
    [, use_spheroid => boolean_expression]
    [, endcap => endcap_style]
    [, side => line_side])
```

Returns a `GEOGRAPHY` that represents the buffer around the input `GEOGRAPHY`.
This function is similar to [`ST_BUFFER`][st-buffer],
but you provide tolerance instead of segments to determine how much the
resulting geography can deviate from the ideal buffer radius.

+   `geography`: The input `GEOGRAPHY` to encircle with the buffer radius.
+   `buffer_radius`: `DOUBLE` that represents the radius of the
    buffer around the input geography. The radius is in meters. Note that
    polygons contract when buffered with a negative `buffer_radius`. Polygon
    shells and holes that are contracted to a point are discarded.
+   `tolerance_meters`: `DOUBLE` specifies a tolerance in
    meters with which the shape is approximated. Tolerance determines how much a
    polygon can deviate from the ideal radius. Naming this argument is optional.
+   `endcap`: (Optional) `STRING` allows you to specify one of two endcap
    styles: `ROUND` and `FLAT`. The default value is `ROUND`. This option only
    affects the endcaps of buffered linestrings.
+   `side`: (Optional) `STRING` allows you to specify one of three possible line
    styles: `BOTH`, `LEFT`, and `RIGHT`. The default is `BOTH`. This option only
    affects the endcaps of buffered linestrings.
+   `use_spheroid`: (Optional) `BOOL` determines how this function measures
    distance. If `use_spheroid` is `FALSE`, the function measures distance on
    the surface of a perfect sphere. The `use_spheroid` parameter
    currently only supports the value `FALSE`. The default value of
    `use_spheroid` is `FALSE`.

**Return type**

Polygon `GEOGRAPHY`

**Example**

The following example shows the results of `ST_BUFFERWITHTOLERANCE` on a point,
given two different values for tolerance but with the same buffer radius of
`100`. A buffered point is an approximated circle. When `tolerance_meters=25`,
the tolerance is a large percentage of the buffer radius, and therefore only
five segments are used to approximate a circle around the input point. When
`tolerance_meters=1`, the tolerance is a much smaller percentage of the buffer
radius, and therefore twenty-four edges are used to approximate a circle around
the input point.

```googlesql
SELECT
  -- tolerance_meters=25, or 25% of the buffer radius.
  ST_NumPoints(ST_BUFFERWITHTOLERANCE(ST_GEOGFROMTEXT('POINT(1 2)'), 100, 25)) AS five_sides,
  -- tolerance_meters=1, or 1% of the buffer radius.
  st_NumPoints(ST_BUFFERWITHTOLERANCE(ST_GEOGFROMTEXT('POINT(100 2)'), 100, 1)) AS twenty_four_sides;

/*------------+-------------------+
 | five_sides | twenty_four_sides |
 +------------+-------------------+
 | 6          | 24                |
 +------------+-------------------*/
```

[wgs84-link]: https://en.wikipedia.org/wiki/World_Geodetic_System

[st-buffer]: #st_buffer

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
