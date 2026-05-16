---
name: ST_UNION
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Planar polygon union (convex hull of the combined vertex set) or MULTI* concatenation for disjoint inputs / non-polygon shapes. Runtime entry: BindStUnion in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_union
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_union.yaml
---

# ST_UNION

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

## `ST_UNION`

```googlesql
ST_UNION(geography_1, geography_2)
```

```googlesql
ST_UNION(array_of_geography)
```

**Description**

Returns a `GEOGRAPHY` that represents the point set
union of all input `GEOGRAPHY`s.

`ST_UNION` comes in two variants. For the first variant, input must be two
`GEOGRAPHY`s. For the second, the input is an
`ARRAY` of type `GEOGRAPHY`.

For the first variant of `ST_UNION`, if an input
`GEOGRAPHY` is `NULL`, `ST_UNION` returns `NULL`.
For the second variant, if the input `ARRAY` value
is `NULL`, `ST_UNION` returns `NULL`.
For a non-`NULL` input `ARRAY`, the union is computed
and `NULL` elements are ignored so that they don't affect the output.

See [`ST_UNION_AGG`][st-union-agg] for the aggregate version of `ST_UNION`.

**Return type**

`GEOGRAPHY`

**Example**

```googlesql
SELECT ST_UNION(
  ST_GEOGFROMTEXT('LINESTRING(-122.12 47.67, -122.19 47.69)'),
  ST_GEOGFROMTEXT('LINESTRING(-122.12 47.67, -100.19 47.69)')
) AS results

/*---------------------------------------------------------+
 | results                                                 |
 +---------------------------------------------------------+
 | LINESTRING(-100.19 47.69, -122.12 47.67, -122.19 47.69) |
 +---------------------------------------------------------*/
```

[st-union-agg]: #st_union_agg

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
