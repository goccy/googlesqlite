---
name: ST_REGIONSTATS
dialect: bigquery
category: functions/geography
status: partial
notes: |
  Local-raster implementation in internal/functions/geography/
  st_regionstats.go. Accepts file paths
  (`/path/to/raster.tif` or `file:///path/...`) and `https://`
  URLs. Parses the GeoTIFF byte stream directly (header +
  IFD entries + GeoKey tags ModelPixelScaleTag /
  ModelTiepointTag) and accumulates count / min / max / sum /
  mean / std / area for the pixels whose centres fall inside
  the input GEOGRAPHY's bounding box. Returns a STRUCT with
  those seven fields.

  Caveats:
  - Pixel sampling is bounding-box-based (not the full polygon
    interior test) — adequate for axis-aligned regions, an
    over-estimate for arbitrarily shaped boundaries.
  - Per-pixel area uses a local equirectangular projection
    (m² = degX·cos(lat)·111195 × degY·111195), which is
    accurate for small / mid-latitude rasters.
  - Earth-Engine asset paths (`projects/.../raster/...`) are
    not resolvable locally and surface as a clear file-open
    error.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/geography_functions#st_regionstats
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/geography_functions#st_regionstats
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/geography/st_regionstats.yaml
---

# ST_REGIONSTATS

## Summary

Returns summary statistics over the pixels of a raster image that
intersect a `GEOGRAPHY`. Backed by Google Earth Engine.

## Signatures

- `ST_REGIONSTATS(geography, raster_id [, [band => ] value] [, include => value] [, options => value])`

## Behavior

- `geography`: the region of interest as a `GEOGRAPHY`.
- `raster_id`: a `STRING` identifying a raster image. Accepts an
  Analytics-Hub URI, a GeoTIFF URI, or an Earth Engine asset path.
- Optional named arguments select a band (`band =>`), filter
  pixels (`include =>`), or pass tuning `options =>`.
- Statistics returned: count, min, max, sum, standard deviation,
  mean, and area of valid pixels.
- Computation is delegated to Google Earth Engine — billed
  separately under the BigQuery Services SKU.

## Examples

```sql
SELECT ST_REGIONSTATS(boundary, 'projects/example/raster/landcover',
                     band => 'B4') AS stats
FROM mydataset.regions;
```

## Edge cases

- Function calls incur Earth Engine charges.
- `raster_id` access requires the appropriate IAM / sharing setup.
- Empty intersections produce zero-count statistics.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/geography_functions#st_regionstats>.
