---
name: KEYS.KEYSET_TO_JSON
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keys.keyset_to_json
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keys.keyset_to_json
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_keyset_to_json.yaml
---

# KEYS.KEYSET_TO_JSON

## Summary
Returns a JSON `STRING` representation of a Tink keyset that is compatible
with the `google.crypto.tink.Keyset` protocol buffer message.

## Signatures
- `KEYS.KEYSET_TO_JSON(keyset)` -> `STRING`

## Behavior
- Accepts a single serialized keyset argument `keyset` (the same `BYTES`
  form produced by other `KEYS.*` constructors such as `KEYS.NEW_KEYSET`).
- Returns a JSON-formatted `STRING` whose shape matches the
  `google.crypto.tink.Keyset` proto: a top-level `primaryKeyId` plus a
  `key` array whose entries carry `keyData` (with `keyMaterialType`,
  `typeUrl`, and `value`), `keyId`, `outputPrefixType`, and `status`.
- The returned `STRING` round-trips through `KEYS.KEYSET_FROM_JSON`, which
  converts the JSON representation back to the serialized `BYTES` form.
- Intended for inspecting or persisting keyset material as JSON; the
  encoded `value` field still contains raw key bytes (Base64-encoded), so
  the output should be treated as sensitive.

## Examples
```sql
-- Build a fresh AEAD_AES_GCM_256 keyset and emit it as JSON.
SELECT KEYS.KEYSET_TO_JSON(KEYS.NEW_KEYSET('AEAD_AES_GCM_256'));
-- expected: a STRING shaped like
-- {
--   "key":[{
--     "keyData":{
--       "keyMaterialType":"SYMMETRIC",
--       "typeUrl":"type.googleapis.com/google.crypto.tink.AesGcmKey",
--       "value":"GiD80Z8kL6AP3iSNHhqseZGAIvq7TVQzClT7FQy8YwK3OQ=="
--     },
--     "keyId":3101427138,
--     "outputPrefixType":"TINK",
--     "status":"ENABLED"
--   }],
--   "primaryKeyId":3101427138
-- }
```

## Edge cases
- The output JSON exposes raw key material (Base64-encoded inside
  `keyData.value`); guard the result like a secret rather than logging it.
- Input must be a valid serialized Tink keyset; arbitrary `BYTES` are not
  meaningful inputs and behaviour for malformed keysets is unspecified.

## Reference (upstream)

See the BigQuery AEAD encryption functions reference:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keys.keyset_to_json>
