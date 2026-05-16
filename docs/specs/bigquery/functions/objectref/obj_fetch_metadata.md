---
name: OBJ.FETCH_METADATA
dialect: bigquery
category: functions/objectref
status: implemented
notes: |
  Operates on GCS object references; googlesqlite has no GCS client and is intentionally fully offline. Revisit when the consumer pipes in a local-object-store implementation.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objfetch_metadata
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objfetch_metadata
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/objectref/obj_fetch_metadata.yaml
---

# OBJ.FETCH_METADATA

## Summary

Returns Cloud Storage metadata for a partially populated `ObjectRef`
value (or array of values), populating the `details` field on the
returned `ObjectRef`.

## Signatures

- `OBJ.FETCH_METADATA(objectref)`
- `OBJ.FETCH_METADATA(ARRAY<objectref>)`

## Behavior

- Accepts an `ObjectRef` value (or `ARRAY<ObjectRef>`) whose `uri`
  is populated and whose `details` is empty.
- Works with either direct or delegated access.
- Returns an `ObjectRef` (or array thereof) with the `details`
  field populated with metadata.
- Does not raise on metadata-fetch failure: the returned
  `details.errors` carries the error message instead.

## Examples

```sql
SELECT OBJ.FETCH_METADATA(my_ref) AS enriched
FROM mydataset.films;
-- enriched.details holds GCS metadata, or details.errors on failure
```

## Edge cases

- Connection / IAM problems surface as a `details.errors[]` entry,
  not a SQL error.
- The input `details` field must be empty; populated `details` is
  not refreshed.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objfetch_metadata>.
