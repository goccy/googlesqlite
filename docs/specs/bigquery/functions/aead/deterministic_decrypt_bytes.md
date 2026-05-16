---
name: DETERMINISTIC_DECRYPT_BYTES
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_decrypt_bytes
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_decrypt_bytes
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/deterministic_decrypt_bytes.yaml
---

# DETERMINISTIC_DECRYPT_BYTES

## Summary
Uses the matching key from a deterministic AEAD keyset to decrypt a `BYTES`
ciphertext, verifying integrity with the supplied additional data, and returns
the original plaintext bytes.

## Signatures
- `DETERMINISTIC_DECRYPT_BYTES(keyset, ciphertext, additional_data)`

## Behavior
- Returns `BYTES` containing the plaintext that was passed to `DETERMINISTIC_ENCRYPT`.
- `keyset` is either a serialized `BYTES` value or a `STRUCT` value produced by
  one of the `KEYS` functions; the matching key is selected by key ID encoded
  in `ciphertext`.
- The matching key must be present in `keyset`, must be in the `'ENABLED'`
  state, and must be of type `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`; otherwise
  the function returns an error.
- `ciphertext` is a `BYTES` value originally produced by `DETERMINISTIC_ENCRYPT`
  with a `BYTES` plaintext input; it must follow Tink's deterministic AEAD wire
  format (a 1-byte Tink key version, a 4-byte key hint, then any IV/auth-tag
  bytes — for SIV the combined IV+tag sit at the start of the ciphertext;
  IV/tag are commonly 16 bytes but can vary).
- `additional_data` is a `STRING` or `BYTES` value that binds the ciphertext to
  its context; any `STRING` is cast to `BYTES` for comparison. It must equal
  the `additional_data` that was supplied to `DETERMINISTIC_ENCRYPT` (ignoring
  type), or the function returns an error.
- Returns an error if decryption or integrity verification fails for any other
  reason (wrong key, tampered ciphertext, malformed wire format, etc.).

## Examples
```sql
-- Build a keyset table for a few customers.
CREATE TABLE deterministic.CustomerKeysets AS
SELECT 1 AS customer_id,
       KEYS.NEW_KEYSET('DETERMINISTIC_AEAD_AES_SIV_CMAC_256') AS keyset,
       b'jaguar' AS favorite_animal
UNION ALL
SELECT 2, KEYS.NEW_KEYSET('DETERMINISTIC_AEAD_AES_SIV_CMAC_256'), b'zebra'
UNION ALL
SELECT 3, KEYS.NEW_KEYSET('DETERMINISTIC_AEAD_AES_SIV_CMAC_256'), b'nautilus';

-- Encrypt favorite_animal per customer using customer_id as additional data.
CREATE TABLE deterministic.EncryptedCustomerData AS
SELECT
  customer_id,
  DETERMINISTIC_ENCRYPT(
    ck.keyset,
    favorite_animal,
    CAST(CAST(customer_id AS STRING) AS BYTES)
  ) AS encrypted_animal
FROM deterministic.CustomerKeysets AS ck;

-- Decrypt back to the original BYTES plaintext.
SELECT
  ecd.customer_id,
  DETERMINISTIC_DECRYPT_BYTES(
    (SELECT ck.keyset
     FROM deterministic.CustomerKeysets AS ck
     WHERE ecd.customer_id = ck.customer_id),
    ecd.encrypted_animal,
    CAST(CAST(ecd.customer_id AS STRING) AS BYTES)
  ) AS favorite_animal
FROM deterministic.EncryptedCustomerData AS ecd;
-- expected favorite_animal column matches the original BYTES values
-- (b'jaguar', b'zebra', b'nautilus') for customer_id 1, 2, 3.
```

## Edge cases
- Errors if the matching key is missing from `keyset`, is not `'ENABLED'`, or
  is not of type `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`.
- Errors if `additional_data` (after `STRING`-to-`BYTES` casting) does not
  equal the value supplied at encryption time.
- Errors if `ciphertext` is malformed under Tink's deterministic AEAD wire
  format or if integrity verification fails.
- The companion `DETERMINISTIC_DECRYPT_STRING` is the same operation but
  returns the plaintext as `STRING`; mismatching plaintext type produces an
  error rather than a coerced result.

## Reference (upstream)

See: <https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_decrypt_bytes>
