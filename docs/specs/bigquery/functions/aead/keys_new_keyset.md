---
name: KEYS.NEW_KEYSET
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysnew_keyset
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysnew_keyset
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_new_keyset.yaml
---

# KEYS.NEW_KEYSET

## Summary
Returns a serialized keyset containing a single new primary cryptographic key
of the requested type. The serialized `BYTES` value is a `google.crypto.tink.Keyset`
suitable for use with the `AEAD.ENCRYPT`, `AEAD.DECRYPT_BYTES`, and
`AEAD.DECRYPT_STRING` functions, as well as with the rest of the `KEYS`
function family.

## Signatures
- `KEYS.NEW_KEYSET(key_type)` -> `BYTES`

## Behavior
- `key_type` is a `STRING` literal that names the kind of key to create.
- `key_type` is required and must not be `NULL`.
- Supported `key_type` values:
  - `AEAD_AES_GCM_256` — generates a 256-bit AES-GCM key, using the
    pseudo-random number generator provided by BoringSSL. The key is used with
    AES-GCM for encryption and decryption.
  - `DETERMINISTIC_AEAD_AES_SIV_CMAC_256` — generates a 512-bit AES-SIV-CMAC
    key (a 256-bit AES-CTR key plus a 256-bit AES-CMAC key), using the
    BoringSSL pseudo-random number generator. The key is used with AES-SIV
    for encryption and decryption.
- The returned keyset always contains exactly one primary key and no
  additional keys.
- The output is the serialized form of `google.crypto.tink.Keyset` returned
  as `BYTES`.

## Examples
```sql
-- Create a fresh AEAD_AES_GCM_256 keyset per customer.
SELECT
  customer_id,
  KEYS.NEW_KEYSET('AEAD_AES_GCM_256') AS keyset
FROM (
  SELECT 1 AS customer_id UNION ALL
  SELECT 2 UNION ALL
  SELECT 3
) AS CustomerIds;
-- expected: one row per customer_id, each with an independent random keyset (BYTES).
```

## Edge cases
- A `NULL` `key_type` is not allowed and produces an error.
- `key_type` must be a string literal; only the documented values
  (`AEAD_AES_GCM_256`, `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`) are accepted.
  Other values are rejected.
- Each invocation produces a fresh, independently-randomised keyset, so the
  output is non-deterministic across calls and rows.

## Reference (upstream)

See the BigQuery AEAD encryption functions reference for the full upstream
description and any updates:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysnew_keyset>
