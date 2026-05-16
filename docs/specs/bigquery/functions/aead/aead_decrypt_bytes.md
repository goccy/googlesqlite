---
name: AEAD.DECRYPT_BYTES
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeaddecrypt_bytes
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeaddecrypt_bytes
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/aead_decrypt_bytes.yaml
---

# AEAD.DECRYPT_BYTES

## Summary
Decrypts a `BYTES` ciphertext that was produced by `AEAD.ENCRYPT` over a
`BYTES` plaintext, using the matching key from a keyset and verifying
integrity against `additional_data`. Returns an error if decryption or
authentication fails.

## Signatures
- `AEAD.DECRYPT_BYTES(keyset, ciphertext, additional_data)`

## Behavior
- `keyset` is either a serialized `BYTES` value returned by one of the
  `KEYS` functions, or a `STRUCT` returned by `KEYS.KEYSET_CHAIN`.
- The keyset must contain the key originally used to encrypt
  `ciphertext`, and that key must be in the `'ENABLED'` state; otherwise
  the function returns an error.
- The matching key is selected by finding the entry in `keyset` whose
  key ID equals the key ID embedded in `ciphertext`.
- `ciphertext` must be a `BYTES` value produced by `AEAD.ENCRYPT` where
  the input `plaintext` was of type `BYTES`.
- If the ciphertext carries an initialization vector (IV), it occupies
  the first bytes of `ciphertext`; an authentication tag, if present,
  occupies the last bytes; for SIV constructions the combined IV/tag
  sits at the front. IV and tag are commonly 16 bytes but may vary.
- `additional_data` is a `STRING` or `BYTES` value that binds the
  ciphertext to its context; any `STRING` is cast to `BYTES`. It must
  match the `additional_data` passed to `AEAD.ENCRYPT` (ignoring its
  declared type) or the function returns an error.
- The return type is `BYTES`.

## Examples
```sql
-- Build a table of (id, keyset, plaintext favorite_animal).
CREATE TABLE aead.CustomerKeysets AS
SELECT 1 AS customer_id, KEYS.NEW_KEYSET('AEAD_AES_GCM_256') AS keyset, b'jaguar'   AS favorite_animal UNION ALL
SELECT 2 AS customer_id, KEYS.NEW_KEYSET('AEAD_AES_GCM_256') AS keyset, b'zebra'    AS favorite_animal UNION ALL
SELECT 3 AS customer_id, KEYS.NEW_KEYSET('AEAD_AES_GCM_256') AS keyset, b'nautilus' AS favorite_animal;

-- Encrypt the plaintext bytes, binding each row to its customer_id.
CREATE TABLE aead.EncryptedCustomerData AS
SELECT
  customer_id,
  AEAD.ENCRYPT(keyset, favorite_animal, CAST(CAST(customer_id AS STRING) AS BYTES))
    AS encrypted_animal
FROM aead.CustomerKeysets AS ck;

-- Decrypt back to the original BYTES plaintext.
SELECT
  ecd.customer_id,
  AEAD.DECRYPT_BYTES(
    (SELECT ck.keyset
     FROM aead.CustomerKeysets AS ck
     WHERE ecd.customer_id = ck.customer_id),
    ecd.encrypted_animal,
    CAST(CAST(customer_id AS STRING) AS BYTES)
  ) AS favorite_animal
FROM aead.EncryptedCustomerData AS ecd;
-- expected: rows (1, b'jaguar'), (2, b'zebra'), (3, b'nautilus')
```

## Edge cases
- Returns an error if the keyset does not contain the key used to
  encrypt `ciphertext`, or if the matching key is not in the `'ENABLED'`
  state.
- Returns an error if integrity verification fails — for example when
  `additional_data` differs (after the implicit `STRING` to `BYTES`
  cast) from the value supplied to `AEAD.ENCRYPT`, or when the
  ciphertext has been altered.
- Use `AEAD.DECRYPT_STRING` instead when the original plaintext was a
  `STRING`; `AEAD.DECRYPT_BYTES` is specifically for ciphertext produced
  from `BYTES` plaintext.
- A wrapped keyset (from `KEYS.NEW_WRAPPED_KEYSET`) must be unwrapped
  via `KEYS.KEYSET_CHAIN` before being passed as `keyset`.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeaddecrypt_bytes>.
