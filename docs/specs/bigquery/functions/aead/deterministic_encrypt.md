---
name: DETERMINISTIC_ENCRYPT
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_encrypt
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_encrypt
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/deterministic_encrypt.yaml
---

# DETERMINISTIC_ENCRYPT

## Summary
Encrypts `plaintext` with the primary key of `keyset` using deterministic
AEAD, binding the ciphertext to `additional_data`. Identical inputs always
produce the same `BYTES` ciphertext, both within a query and across
queries.

## Signatures
- `DETERMINISTIC_ENCRYPT(keyset, plaintext, additional_data)`

## Behavior
- `keyset` is either a serialized `BYTES` value returned by one of the
  `KEYS` functions, or a `STRUCT` returned by `KEYS.KEYSET_CHAIN`.
- The primary key of `keyset` must use the
  `DETERMINISTIC_AEAD_AES_SIV_CMAC_256` algorithm; otherwise the function
  returns an error.
- `plaintext` is a `STRING` or `BYTES` value to encrypt.
- `additional_data` is a `STRING` or `BYTES` value bound to the
  ciphertext for context; the same value (matched as `BYTES`) must be
  supplied to `DETERMINISTIC_DECRYPT_*` to recover the plaintext.
- `plaintext` and `additional_data` must have the same type. Passing
  two `STRING` values is equivalent to casting both to `BYTES` first.
- The result is `BYTES` carrying a Tink-specific prefix that records the
  key ID used for the encryption, so the matching key can be found at
  decrypt time.
- For a fixed `keyset` and `plaintext`, the function returns the same
  ciphertext on every invocation (deterministic).
- Returns `NULL` if any argument is `NULL`.

## Examples
```sql
-- Encrypt each customer's favorite_animal under that customer's keyset,
-- binding the ciphertext to the customer_id as additional_data.
WITH CustomerKeysets AS (
  SELECT 1 AS customer_id,
         KEYS.NEW_KEYSET('DETERMINISTIC_AEAD_AES_SIV_CMAC_256') AS keyset UNION ALL
  SELECT 2, KEYS.NEW_KEYSET('DETERMINISTIC_AEAD_AES_SIV_CMAC_256') UNION ALL
  SELECT 3, KEYS.NEW_KEYSET('DETERMINISTIC_AEAD_AES_SIV_CMAC_256')
), PlaintextCustomerData AS (
  SELECT 1 AS customer_id, 'elephant' AS favorite_animal UNION ALL
  SELECT 2, 'walrus' UNION ALL
  SELECT 3, 'leopard'
)
SELECT
  pcd.customer_id,
  DETERMINISTIC_ENCRYPT(
    (SELECT keyset
     FROM CustomerKeysets AS ck
     WHERE ck.customer_id = pcd.customer_id),
    pcd.favorite_animal,
    CAST(pcd.customer_id AS STRING)
  ) AS encrypted_animal
FROM PlaintextCustomerData AS pcd;
-- expected: one BYTES ciphertext per customer_id; re-running yields the
-- same ciphertext bytes for the same (keyset, plaintext, additional_data).
```

## Edge cases
- Returns `NULL` if `keyset`, `plaintext`, or `additional_data` is
  `NULL`.
- Returns an error if the primary key of `keyset` is not of algorithm
  `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`.
- Returns an error if `plaintext` and `additional_data` are of
  incompatible types (one `STRING` and one `BYTES` is allowed only via
  the documented implicit cast pairing — both must be the same type).
- Because output is deterministic, equal plaintexts produce equal
  ciphertexts: this enables equality joins on encrypted columns but
  leaks plaintext equality. Prefer `AEAD.ENCRYPT` (randomized) when this
  leakage is unacceptable.
- A wrapped keyset (from `KEYS.NEW_WRAPPED_KEYSET`) must be unwrapped via
  `KEYS.KEYSET_CHAIN` before being passed as `keyset`.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_encrypt>.
