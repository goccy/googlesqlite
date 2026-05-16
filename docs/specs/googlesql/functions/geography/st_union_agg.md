---
name: ST_UNION_AGG
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStUnionAgg now collects every input and at Done either
  chains all (MULTI)LINESTRING segments into a single LINESTRING
  (or MULTILINESTRING of distinct connected chains) by walking
  shared endpoints, or falls back to iterative pairwise union for
  any non-line input. Duplicate segments are folded; chains start
  at the lex-greatest degree-1 endpoint so the emitted walk
  matches BigQuery's display convention.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_union_agg
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_union_agg.yaml
---

# ST_UNION_AGG

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

## `ST_UNION_AGG`

```googlesql
ST_UNION_AGG(geography)
```

**Description**

Returns a `GEOGRAPHY` that represents the point set
union of all input `GEOGRAPHY`s.

`ST_UNION_AGG` ignores `NULL` input `GEOGRAPHY` values.

See [`ST_UNION`][st-union] for the non-aggregate version of `ST_UNION_AGG`.

**Return type**

`GEOGRAPHY`

**Example**

```googlesql
SELECT ST_UNION_AGG(items) AS results
FROM UNNEST([
  ST_GEOGFROMTEXT('LINESTRING(-122.12 47.67, -122.19 47.69)'),
  ST_GEOGFROMTEXT('LINESTRING(-122.12 47.67, -100.19 47.69)'),
  ST_GEOGFROMTEXT('LINESTRING(-122.12 47.67, -122.19 47.69)')]) as items;

/*---------------------------------------------------------+
 | results                                                 |
 +---------------------------------------------------------+
 | LINESTRING(-100.19 47.69, -122.12 47.67, -122.19 47.69) |
 +---------------------------------------------------------*/
```

[st-union]: #st_union

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
