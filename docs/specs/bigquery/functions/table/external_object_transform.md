---
name: EXTERNAL_OBJECT_TRANSFORM
dialect: bigquery
category: functions/table
status: unsupported
notes: |
  Applies an external transform connection (typically a Cloud Run service) to object references. googlesqlite has no external-connection plumbing.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/table-functions-built-in
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/table-functions-built-in
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/table/external_object_transform.yaml
---

# EXTERNAL_OBJECT_TRANSFORM

## Summary
Returns a transformed object table that contains the original columns of the
input object table plus one or more additional columns derived from the
requested transform types.

## Signatures
- `EXTERNAL_OBJECT_TRANSFORM(TABLE object_table_name, transform_types_array)`

## Behavior
- Accepts only an object table as the first argument; subqueries and other
  table kinds are not supported.
- `object_table_name` must be passed with the `TABLE` keyword and use the
  `dataset_name.object_table_name` form.
- `transform_types_array` is an `ARRAY<STRING>` literal listing the transforms
  to apply to the input rows.
- Currently the only supported transform type is `SIGNED_URL`; any other
  value is not documented as supported.
- Specifying `SIGNED_URL` adds a `signed_url` column whose values are
  read-only signed URLs for each object referenced by the input table.
- Generated signed URLs are valid for 6 hours after generation.
- The return type is `TABLE`; the output preserves the original object-table
  columns and appends the columns produced by the requested transforms.

## Examples
```sql
-- Return URIs and signed URLs for the objects in mydataset.myobjecttable.
SELECT uri, signed_url
FROM EXTERNAL_OBJECT_TRANSFORM(TABLE mydataset.myobjecttable, ['SIGNED_URL']);
-- expected: rows like
--   ('gs://myobjecttable/1234_Main_St.jpeg',
--    'https://storage.googleapis.com/mybucket/1234_Main_St.jpeg?X-Goog-Algorithm=...')
--   ('gs://myobjecttable/345_River_Rd.jpeg',
--    'https://storage.googleapis.com/mybucket/345_River_Rd.jpeg?X-Goog-Algorithm=...')
```

## Edge cases
- Passing a non-object table (regular table, view, or subquery) is not
  supported and is rejected.
- The signed URLs returned for the `SIGNED_URL` transform expire after 6
  hours; queries cached or stored beyond that window will hold stale URLs.
- The signed URLs are read-only.

## Reference (upstream)

See the upstream BigQuery reference page for the canonical description and
any updates: <https://cloud.google.com/bigquery/docs/reference/standard-sql/table-functions-built-in#external_object_transform>
