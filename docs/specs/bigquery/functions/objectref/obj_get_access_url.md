---
name: OBJ.GET_ACCESS_URL
dialect: bigquery
category: functions/objectref
status: implemented
notes: |
  Operates on GCS object references; googlesqlite has no GCS client and is intentionally fully offline. Revisit when the consumer pipes in a local-object-store implementation.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objget_access_url
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objget_access_url
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/objectref/obj_get_access_url.yaml
---

# OBJ.GET_ACCESS_URL

## Summary

Returns a JSON value that contains reference information for an
`ObjectRef` and signed access URLs for reading or modifying the
underlying Cloud Storage object. Requires delegated access.

## Signatures

- `OBJ.GET_ACCESS_URL(objectref, mode [, duration])`
- `OBJ.GET_ACCESS_URL(ARRAY<objectref>, mode [, duration])`

## Behavior

- `mode` is a `STRING`: `'r'` (read) or `'rw'` (read/write).
- Optional `duration` is the URL validity window.
- On success, the returned JSON contains an `access_urls` field
  with the signed URLs.
- On failure, the returned JSON contains an `errors` field with
  the error message (no SQL error is raised).

## Examples

```sql
SELECT OBJ.GET_ACCESS_URL(my_ref, 'r') AS url_info
FROM mydataset.films;
-- url_info.access_urls.read_url is the signed URL
```

## Edge cases

- Requires the connection to permit delegated access; otherwise
  the JSON `errors` array is populated.
- Per-project / per-region connection limit (max 20).

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objget_access_url>.
