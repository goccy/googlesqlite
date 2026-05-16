---
name: KEYS.ROTATE_KEYSET
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysrotate_keyset
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysrotate_keyset
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_rotate_keyset.yaml
---

# KEYS.ROTATE_KEYSET

## Summary
Adds a new key to an existing serialized keyset and promotes it to the
primary cryptographic key, returning the rotated keyset as `BYTES`. The
previous primary key from the input keyset is retained as an additional
(non-primary) key in the returned keyset so that data encrypted with the
old primary can still be decrypted.

## Signatures
- `KEYS.ROTATE_KEYSET(keyset, key_type)` -> `BYTES`

## Behavior
- `keyset` is a `BYTES` value that holds a serialized
  `google.crypto.tink.Keyset` produced by `KEYS.NEW_KEYSET` (or another
  function in the `KEYS` family that returns a plaintext keyset).
- `key_type` is a `STRING` literal naming the kind of key to add. It
  must match the key type of the existing keys in `keyset`.
- A new key of the requested `key_type` is generated and inserted into
  the keyset, and that new key becomes the keyset's primary
  cryptographic key.
- The former primary key from the input `keyset` is preserved in the
  returned keyset as an additional, non-primary key, so prior
  ciphertexts can still be decrypted with the rotated keyset.
- The returned value is the new keyset serialized as `BYTES` in the
  `google.crypto.tink.Keyset` form, suitable for the `AEAD.ENCRYPT`,
  `AEAD.DECRYPT_BYTES`, `AEAD.DECRYPT_STRING`, and other `KEYS.*`
  functions.

## Examples
```sql
-- Build a small table of (customer_id, keyset) rows, then rotate the
-- primary key of each keyset, keeping the prior primary as a
-- secondary key for legacy ciphertexts.
WITH ExistingKeysets AS (
  SELECT 1 AS customer_id, KEYS.NEW_KEYSET('AEAD_AES_GCM_256') AS keyset
    UNION ALL
  SELECT 2, KEYS.NEW_KEYSET('AEAD_AES_GCM_256') UNION ALL
  SELECT 3, KEYS.NEW_KEYSET('AEAD_AES_GCM_256')
)
SELECT customer_id, KEYS.ROTATE_KEYSET(keyset, 'AEAD_AES_GCM_256') AS keyset
FROM ExistingKeysets;
-- expected: one row per customer_id, each with a rotated AEAD_AES_GCM_256 keyset (BYTES).
```

## Edge cases
- `key_type` must match the key type of the existing keys in `keyset`;
  passing a mismatched key type produces an error.
- `key_type` must be one of the supported plaintext keyset types
  documented for `KEYS.NEW_KEYSET` (for example, `AEAD_AES_GCM_256` or
  `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`); other values are rejected.
- `keyset` must be a valid serialized `google.crypto.tink.Keyset`;
  malformed or non-keyset `BYTES` are rejected.
- The newly generated primary key is freshly randomised, so repeated
  calls on the same input keyset return non-deterministic results.

## Reference (upstream)

See the BigQuery AEAD encryption functions reference for the full
upstream description and any updates:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysrotate_keyset>
