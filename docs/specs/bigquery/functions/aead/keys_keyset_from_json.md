---
name: KEYS.KEYSET_FROM_JSON
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_from_json
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_from_json
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_keyset_from_json.yaml
---

# KEYS.KEYSET_FROM_JSON

## Summary
Converts a JSON-formatted `STRING` representation of a Tink keyset into the
serialized `BYTES` form that other `KEYS.*` and `AEAD.*` functions accept.

## Signatures
- `KEYS.KEYSET_FROM_JSON(json_keyset)` -> `BYTES`

## Behavior
- Accepts a single `STRING` argument `json_keyset` that encodes a keyset.
- The input must be JSON compatible with the `google.crypto.tink.Keyset`
  protocol buffer message, i.e. a JSON object whose fields and nested
  objects mirror that proto's "keyset" definition.
- Returns the keyset as serialized `BYTES`.
- The returned `BYTES` is a valid input wherever `KEYS.*` or `AEAD.*`
  functions expect a serialized keyset.
- Round-trips with `KEYS.KEYSET_TO_JSON`, which converts the serialized
  `BYTES` back into a JSON `STRING`.
- Typical keyset JSON contains a `primaryKeyId` and a `key` array whose
  entries carry `keyData` (`keyMaterialType`, `typeUrl`, `value`),
  `keyId`, `outputPrefixType`, and `status` (e.g. `ENABLED`).

## Examples
```sql
-- Given a JSON-formatted STRING `json_keyset` describing a Tink keyset,
-- e.g. an AES-GCM keyset:
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
SELECT KEYS.KEYSET_FROM_JSON(json_keyset);
-- expected: serialized BYTES form of the keyset, e.g.
-- \x08\x9d\x8e\x85\x82\x09\x12d\x0aX\x0a0type.googleapis.com/google.crypto.tink.AesGcmKey...
```

## Edge cases
- The JSON must conform to the `google.crypto.tink.Keyset` proto schema;
  unrelated JSON shapes are not valid inputs.
- The output is opaque serialized `BYTES`; downstream consumers should
  treat it as a keyset handle rather than parse it directly.

## Reference (upstream)

See the BigQuery AEAD encryption functions reference:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_from_json>
