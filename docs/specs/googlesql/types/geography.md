---
name: GEOGRAPHY
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#geography-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/geography.yaml
---

# GEOGRAPHY

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

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## Geography type 
<a id="geography_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>GEOGRAPHY</code></td>
<td>
  A collection of points, linestrings, and polygons, which is represented as a
  point set, or a subset of the surface of the Earth.
</td>
</tr>
</tbody>
</table>

The geography type is based on the [OGC Simple
Features specification (SFS)][ogc-sfs]{: class=external target=_blank },
and can contain the following objects:

<table>
  <thead>
    <tr>
      <th>Geography object</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>Point</code></td>
      <td>
        <p>
          A single location in coordinate space known as a point. A point has an
          x-coordinate value and a y-coordinate value, where the x-coordinate is
          longitude and the y-coordinate is latitude of the point
          on the
          <a href="https://en.wikipedia.org/wiki/World_Geodetic_System">WGS84 reference ellipsoid</a>.

        </p>
        <p>
          Syntax:
<pre class="lang-sql prettyprint">
POINT(x_coordinate y_coordinate)
</pre>
          Examples:
<pre class="lang-sql prettyprint">
POINT(32 210)
</pre>
<pre class="lang-sql prettyprint">
POINT EMPTY
</pre>
        </p>
      </td>
    </tr>
    <tr>
      <td><code>LineString</code></td>
      <td>
        <p>
          Represents a linestring, which is a one-dimensional geometric object,
          with a sequence of points and geodesic edges between them.
        </p>
        <p>
          Syntax:
<pre class="lang-sql prettyprint">
LINESTRING(point[, ...])
</pre>
          Examples:
<pre class="lang-sql prettyprint">
LINESTRING(1 1, 2 1, 3.1 2.88, 3 -3)
</pre>
<pre class="lang-sql prettyprint">
LINESTRING EMPTY
</pre>
        </p>
      </td>
    </tr>
    <tr>
      <td><code>Polygon</code></td>
      <td>
        <p>
          A polygon, which is represented as a planar surface defined by 1
          exterior boundary and 0 or more interior boundaries. Each
          interior boundary defines a hole in the polygon. The boundary loops of
          polygons are oriented so that if you traverse the boundary vertices in
          order, the interior of the polygon is on the left.
        </p>
        <p>
          Syntax:
<pre class="lang-sql prettyprint">
POLYGON(interior_ring[, ...])

interior_ring:
  (point[, ...])
</pre>
          Examples:
<pre class="lang-sql prettyprint">
POLYGON((0 0, 2 2, 2 0, 0 0), (2 2, 3 4, 2 4, 2 2))
</pre>
<pre class="lang-sql prettyprint">
POLYGON EMPTY
</pre>
        </p>
      </td>
    </tr>
    <tr>
      <td><code>MultiPoint</code></td>
      <td>
        <p>
          A collection of points.
        </p>
        <p>
          Syntax:
<pre class="lang-sql prettyprint">
MULTIPOINT(point[, ...])
</pre>
          Examples:
<pre class="lang-sql prettyprint">
MULTIPOINT(0 32, 123 9, 48 67)
</pre>
<pre class="lang-sql prettyprint">
MULTIPOINT EMPTY
</pre>
        </p>
      </td>
    </tr>
    <tr>
      <td><code>MultiLineString</code></td>
      <td>
        <p>
          Represents a multilinestring, which is a collection of linestrings.
        </p>
        <p>
          Syntax:
<pre class="lang-sql prettyprint">
MULTILINESTRING((linestring)[, ...])
</pre>
          Examples:
<pre class="lang-sql prettyprint">
MULTILINESTRING((2 2, 3 4), (5 6, 7 7))
</pre>
<pre class="lang-sql prettyprint">
MULTILINESTRING EMPTY
</pre>
        </p>
      </td>
    </tr>
    <tr>
      <td><code>MultiPolygon</code></td>
      <td>
        <p>
          Represents a multipolygon, which is a collection of polygons.
        </p>
        <p>
          Syntax:
<pre class="lang-sql prettyprint">
MULTIPOLYGON((polygon)[, ...])
</pre>
          Examples:
<pre class="lang-sql prettyprint">
MULTIPOLYGON(((0 -1, 1 0, 1 1, 0 -1)), ((0 0, 2 2, 3 0, 0 0), (2 2, 3 4, 2 4, 1 9)))
</pre>
<pre class="lang-sql prettyprint">
MULTIPOLYGON EMPTY
</pre>
        </p>
      </td>
    </tr>
    <tr>
      <td><code>GeometryCollection</code></td>
      <td>
        <p>
          Represents a geometry collection with elements of different dimensions
          or an empty geography.
        </p>
        <p>
          Syntax:
<pre class="lang-sql prettyprint">
GEOMETRYCOLLECTION(geography_object[, ...])
</pre>
          Examples:
<pre class="lang-sql prettyprint">
GEOMETRYCOLLECTION(MULTIPOINT(-1 2, 0 12), LINESTRING(-2 4, 0 6))
</pre>
<pre class="lang-sql prettyprint">
GEOMETRYCOLLECTION EMPTY
</pre>
        </p>
      </td>
    </tr>
  </tbody>
</table>

The points, linestrings and polygons of a geography value form a simple
arrangement on the [WGS84 reference ellipsoid][WGS84-reference-ellipsoid].
A simple arrangement is one where no point on the WGS84 surface is contained
by multiple elements of the collection. If self intersections exist, they
are automatically removed.

The geography that contains no points, linestrings or polygons is called an
empty geography. An empty geography isn't associated with a particular
geometry shape. For example, the following query produces the same results:

```googlesql
SELECT
  ST_GEOGFROMTEXT('POINT EMPTY') AS a,
  ST_GEOGFROMTEXT('GEOMETRYCOLLECTION EMPTY') AS b

/*--------------------------+--------------------------+
 | a                        | b                        |
 +--------------------------+--------------------------+
 | GEOMETRYCOLLECTION EMPTY | GEOMETRYCOLLECTION EMPTY |
 +--------------------------+--------------------------*/
```

The structure of compound geometry objects isn't preserved if a
simpler type can be produced. For example, in column `b`,
`GEOMETRYCOLLECTION` with `(POINT(1 1)` and `POINT(2 2)` is converted into the
simplest possible geometry, `MULTIPOINT(1 1, 2 2)`.

```googlesql
SELECT
  ST_GEOGFROMTEXT('MULTIPOINT(1 1, 2 2)') AS a,
  ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(POINT(1 1), POINT(2 2))') AS b

/*----------------------+----------------------+
 | a                    | b                    |
 +----------------------+----------------------+
 | MULTIPOINT(1 1, 2 2) | MULTIPOINT(1 1, 2 2) |
 +----------------------+----------------------*/
```

A geography is the result of, or an argument to, a
[Geography Function][geography-functions].

[ogc-sfs]: http://www.opengeospatial.org/standards/sfs#downloads

[WGS84-reference-ellipsoid]: https://en.wikipedia.org/wiki/World_Geodetic_System

[geography-functions]: https://github.com/google/googlesql/blob/master/docs/geography_functions.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
