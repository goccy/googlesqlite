---
name: OBJ.GET_READ_URL
dialect: bigquery
category: functions/objectref
status: implemented
notes: |
  Operates on GCS object references; googlesqlite has no GCS client and is intentionally fully offline. Revisit when the consumer pipes in a local-object-store implementation.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objget_read_url
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objget_read_url
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/objectref/obj_get_read_url.yaml
---

# OBJ.GET_READ_URL

## Summary

Returns a `STRUCT<url STRING, status STRING>` containing a 45-minute
signed read URL for a Cloud Storage object referenced by an
`ObjectRef`. Requires delegated read access.

## Signatures

- `OBJ.GET_READ_URL(objectref)`

## Behavior

- Returns a `STRUCT` with two fields:
  - `url`: the signed read URL, or `NULL` on failure.
  - `status`: error message, or `NULL` on success.
- The signed URL expires after 45 minutes.
- Requires the underlying connection to permit delegated reads.

## Examples

```sql
SELECT OBJ.GET_READ_URL(poster) AS read_url
FROM mydataset.films;
-- read_url.url contains a https://storage.googleapis.com/... URL
```

## Edge cases

- On error, `url` is `NULL` and `status` carries the message; no
  SQL error is raised.
- Per-project / per-region connection limit (max 20).

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objget_read_url>.
