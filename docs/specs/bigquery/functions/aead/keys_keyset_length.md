---
name: KEYS.KEYSET_LENGTH
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_length
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_length
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_keyset_length.yaml
---

# KEYS.KEYSET_LENGTH

## Summary
Returns the number of keys contained in a Tink keyset.

## Signatures
- `KEYS.KEYSET_LENGTH(keyset)` -> `INT64`

## Behavior
- Takes a single `keyset` argument and returns the count of keys it
  contains as an `INT64`.
- `keyset` is the serialized `BYTES` representation of a Tink keyset,
  such as the value returned by another `KEYS.*` function (for example
  `KEYS.KEYSET_FROM_JSON` or `KEYS.NEW_KEYSET`).
- The count includes every key entry in the keyset regardless of its
  `status` (for example both `ENABLED` and `DISABLED` keys are counted).
- The result is the cardinality of the keyset's `key` array, not the
  size in bytes of the serialized keyset.

## Examples
```sql
-- json_keyset is a JSON-formatted STRING that describes a Tink keyset
-- with two key entries (one ENABLED, one DISABLED):
-- {
--   "primaryKeyId": 1354994251,
--   "key": [
--     { "keyData": {...}, "keyId": 1354994251, "status": "ENABLED" },
--     { "keyData": {...}, "keyId": 852264701,  "status": "DISABLED" }
--   ]
-- }
SELECT KEYS.KEYSET_LENGTH(KEYS.KEYSET_FROM_JSON(json_keyset)) AS key_count;
-- expected: key_count = 2
```

## Edge cases
- Disabled keys still contribute to the returned count; the function
  does not filter by `status`.

## Reference (upstream)

See the BigQuery AEAD encryption functions reference:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_length>
