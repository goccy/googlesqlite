---
name: AEAD.DECRYPT_STRING
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeaddecrypt_string
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeaddecrypt_string
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/aead_decrypt_string.yaml
---

# AEAD.DECRYPT_STRING

## Summary
Decrypts a BYTES `ciphertext` produced by `AEAD.ENCRYPT` and returns the
plaintext as a `STRING`. Behaves like `AEAD.DECRYPT_BYTES` but expects the
original plaintext to have been a `STRING` value and accepts `additional_data`
as a `STRING`.

## Signatures
- `AEAD.DECRYPT_STRING(keyset, ciphertext, additional_data)`

## Behavior
- Identifies the matching key inside `keyset` by the key ID embedded in
  `ciphertext` and uses it to decrypt, while verifying integrity against
  `additional_data`; returns an error if decryption or verification fails.
- `keyset` is a serialized `BYTES` value from one of the `KEYS` functions or a
  `STRUCT` returned by `KEYS.KEYSET_CHAIN`; it must contain the key used to
  encrypt `ciphertext`, and that key must be in the `'ENABLED'` state.
- `ciphertext` is the `BYTES` output of an `AEAD.ENCRYPT` call where the
  original `plaintext` was of type `STRING`.
- `ciphertext` follows Tink's wire format: an optional initialization vector
  (IV) at the front, an optional authentication tag at the end, or a combined
  SIV at the front; the IV and tag are commonly 16 bytes but may vary.
- `additional_data` is a `STRING` (paired type for this overload) and must
  match the `additional_data` value supplied to `AEAD.ENCRYPT`, ignoring its
  type, otherwise the function returns an error.
- Returns a `STRING` value containing the decrypted plaintext.

## Examples
```sql
-- Decrypt the encrypted_animal column using the per-customer keyset.
SELECT
  ecd.customer_id,
  AEAD.DECRYPT_STRING(
    (SELECT ck.keyset
     FROM aead.CustomerKeysets AS ck
     WHERE ecd.customer_id = ck.customer_id),
    ecd.encrypted_animal,
    CAST(ecd.customer_id AS STRING)
  ) AS favorite_animal
FROM aead.EncryptedCustomerData AS ecd;
-- expected: one STRING favorite_animal value per customer_id
```

## Edge cases
- Returns an error if `keyset` does not contain the key that encrypted
  `ciphertext`, or if that key is not in the `'ENABLED'` state.
- Returns an error if `additional_data` does not match the value used at
  encryption time (compared ignoring type).
- Returns an error if integrity verification or decryption otherwise fails
  (for example, malformed `ciphertext` or wrong key).
- The decrypted bytes must be valid for the `STRING` return type; this
  overload assumes the original `plaintext` was a `STRING`.

## Reference (upstream)

See the official BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeaddecrypt_string>
