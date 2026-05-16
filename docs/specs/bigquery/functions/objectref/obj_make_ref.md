---
name: OBJ.MAKE_REF
dialect: bigquery
category: functions/objectref
status: implemented
notes: |
  Operates on GCS object references; googlesqlite has no GCS client and is intentionally fully offline. Revisit when the consumer pipes in a local-object-store implementation.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objmake_ref
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objmake_ref
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/objectref/obj_make_ref.yaml
---

# OBJ.MAKE_REF

## Summary
Constructs an `ObjectRef` value that carries reference information for a
Cloud Storage object. The function is typically used after writing
transformation output to Cloud Storage (for example, via a signed URL
returned by `OBJ.GET_ACCESS_URL`) so the resulting `ObjectRef` can be
persisted to a table column.

## Signatures
- `OBJ.MAKE_REF(uri [, authorizer] [, version => version_value] [, details => gcs_metadata_json])`
- `OBJ.MAKE_REF(objectref_json)`
- `OBJ.MAKE_REF(objectref, authorizer)`

## Behavior
- `uri` is a `STRING` URI of the Cloud Storage object (for example,
  `gs://mybucket/flowers/12345.jpg`); a column reference may be supplied
  in place of a string literal.
- `authorizer` is a `STRING` value identifying the Cloud Resource
  connection used for delegated access to the object; if omitted, the
  returned `ObjectRef` uses direct access, and permissions on the
  connection must be configured by the data administrator.
- `version_value` (named argument `version`) is a `STRING` representing
  the Cloud Storage object version.
- `gcs_metadata_json` (named argument `details`) is a `JSON` value with
  schema `{content_type: string, md5_hash: string, size: number, updated: number}`
  describing Cloud Storage metadata for the object.
- In the JSON-input form, `objectref_json` is a `JSON` value matching the
  schema `{uri: string, authorizer?: string, version?: string, details?: gcs_metadata_json}`,
  and the function emits an `ObjectRef` carrying all of those fields.
- In the `(objectref, authorizer)` form, the supplied top-level
  `authorizer` overwrites any `authorizer` already present in the input
  `ObjectRef`.
- Validation is performed on the formatting of the input only, not on
  the underlying object's existence or content.
- The return type is `ObjectRef`.

## Examples
```sql
-- Build an ObjectRef from a URI column and a Cloud resource connection.
CREATE OR REPLACE TABLE `mydataset.movies` AS (
  SELECT
    f.title,
    f.director,
    OBJ.MAKE_REF(p.uri, 'asia-south2.storage_connection') AS movie_poster
  FROM mydataset.movie_posters p
  JOIN mydataset.films f USING (title)
  WHERE region = 'US' AND release_year = 2024
);
-- expected: movies table populated with an ObjectRef column referencing
-- each poster via the asia-south2.storage_connection authorizer.

-- Build an ObjectRef from a JSON literal.
SELECT OBJ.MAKE_REF(JSON '''
  {
    "uri": "gs://cloud-samples-data/bigquery/tutorials/cymbal-pets/images/aquaclear-aquarium-background-poster.png",
    "authorizer": "asia-south2.storage_connection"
  }
''') AS poster;
-- expected: a single ObjectRef value carrying the supplied uri and
-- authorizer.

-- Re-authorize an existing ObjectRef with a new connection (named arg).
SELECT
  OBJ.MAKE_REF(movie_poster,
               authorizer => 'asia-south2.new_connection')
    AS movie_poster_updated
FROM mydataset.movies;
-- expected: each row's ObjectRef value is returned with its authorizer
-- replaced by 'asia-south2.new_connection'.
```

## Edge cases
- The input form determines the output: a URI input yields a reference
  to the named Cloud Storage object; a JSON input is reformatted into an
  `ObjectRef`; an `(objectref, authorizer)` input returns the existing
  `ObjectRef` with its authorizer replaced.
- When using the `(objectref, authorizer)` form, any authorizer already
  embedded inside the input `ObjectRef` is overwritten by the explicit
  top-level `authorizer` argument.
- Format validation is shallow: malformed `uri` / JSON shape is
  rejected, but the function does not verify the object exists, the
  connection has access, or the metadata fields match the live object.
- A project plus region cannot use more than 20 Cloud resource
  connections to access object data referenced as `ObjectRef` values in
  a single query; exceeding this limit is rejected.
- Without an `authorizer`, the returned `ObjectRef` relies on direct
  access permissions of the caller; with an `authorizer`, the caller
  must have permission to use that Cloud Resource connection.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/objectref_functions#objmake_ref>.
