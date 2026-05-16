---
name: ST_CENTROID_AGG
dialect: bigquery
category: functions/geography
status: implemented
notes: |
  Pure-Go planar centroid aggregator. Honours BigQuery's dimension
  precedence: POLYGON / MULTIPOLYGON dominates LINESTRING /
  MULTILINESTRING dominates POINT / MULTIPOINT. Per-ring centroid
  uses the planar Shoelace formula (outer rings positive, inner
  rings negative); per-segment centroid is the midpoint weighted
  by segment length; point centroid is the equal-weight mean.
  Uses planar (lon/lat as Cartesian) math — the same caveat as
  the rest of the pure-Go GEOGRAPHY surface; spherical (S2)
  results may differ for large extents.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/geography_functions#st_centroid_agg
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/geography_functions#st_centroid_agg
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/geography/st_centroid_agg.yaml
---

# ST_CENTROID_AGG

## Summary

Aggregate that computes the centroid of a set of `GEOGRAPHY`
values as a single point `GEOGRAPHY`.

## Signatures

- `ST_CENTROID_AGG(geography)`

## Behavior

- The aggregate centroid is the weighted average of each input's
  centroid.
- Only inputs at the highest dimension contribute. If both points
  and lines are present, only line inputs participate in the
  centroid; standalone points are ignored.
- Skips `NULL` `GEOGRAPHY` inputs.
- Returns a point `GEOGRAPHY`.
- Not equivalent to `ST_CENTROID(ST_UNION_AGG(...))` because
  union deduplicates inputs first.

## Examples

```sql
SELECT ST_CENTROID_AGG(geo) AS centre
FROM mydataset.geo_table;
-- weighted by the highest dimension among the inputs
```

## Edge cases

- All-`NULL` input yields `NULL`.
- A mix of dimensions silently filters lower-dimensional inputs.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/geography_functions#st_centroid_agg>.
