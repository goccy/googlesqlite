---
name: AEAD.ENCRYPT
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeadencrypt
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeadencrypt
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/aead_encrypt.yaml
---

# AEAD.ENCRYPT

## Summary
Encrypts `plaintext` with the primary key of `keyset` using authenticated
encryption with associated data (AEAD), binding the ciphertext to
`additional_data`. The encryption is randomized: each invocation generally
produces a different ciphertext for the same inputs because of a fresh
initialization vector.

## Signatures
- `AEAD.ENCRYPT(keyset, plaintext, additional_data)`

## Behavior
- `keyset` is either a serialized `BYTES` value returned by one of the
  `KEYS` functions, or a `STRUCT` returned by `KEYS.KEYSET_CHAIN`.
- The primary key of `keyset` must use the `AEAD_AES_GCM_256` algorithm;
  otherwise the function returns an error.
- `plaintext` is a `STRING` or `BYTES` value to encrypt.
- `additional_data` is a `STRING` or `BYTES` value bound to the ciphertext
  for context; the same value (matched as `BYTES`) must be supplied to
  `AEAD.DECRYPT_*` to recover the plaintext.
- `plaintext` and `additional_data` must have the same type. Calling
  `AEAD.ENCRYPT(keyset, string1, string2)` is equivalent to
  `AEAD.ENCRYPT(keyset, CAST(string1 AS BYTES), CAST(string2 AS BYTES))`.
- The result is `BYTES` carrying a Tink-specific prefix that records the
  key ID used for the encryption, so the matching key can be found at
  decrypt time.
- Encrypting the same `plaintext` more than once with the same `keyset`
  generally returns different ciphertexts because each call uses a fresh
  initialization vector (IV).
- Returns `NULL` if any argument is `NULL`.

## Examples
```sql
-- Encrypt each customer's favorite_animal under that customer's keyset,
-- binding the ciphertext to the customer_id as additional_data.
WITH CustomerKeysets AS (
  SELECT 1 AS customer_id,
         KEYS.NEW_KEYSET('AEAD_AES_GCM_256') AS keyset UNION ALL
  SELECT 2, KEYS.NEW_KEYSET('AEAD_AES_GCM_256') UNION ALL
  SELECT 3, KEYS.NEW_KEYSET('AEAD_AES_GCM_256')
), PlaintextCustomerData AS (
  SELECT 1 AS customer_id, 'elephant' AS favorite_animal UNION ALL
  SELECT 2, 'walrus' UNION ALL
  SELECT 3, 'leopard'
)
SELECT
  pcd.customer_id,
  AEAD.ENCRYPT(
    (SELECT keyset
     FROM CustomerKeysets AS ck
     WHERE ck.customer_id = pcd.customer_id),
    pcd.favorite_animal,
    CAST(pcd.customer_id AS STRING)
  ) AS encrypted_animal
FROM PlaintextCustomerData AS pcd;
-- expected: one BYTES ciphertext per customer_id; re-running yields
-- different ciphertext bytes for the same (keyset, plaintext,
-- additional_data) because a fresh IV is used per call.
```

## Edge cases
- Returns `NULL` if `keyset`, `plaintext`, or `additional_data` is `NULL`.
- Returns an error if the primary key of `keyset` is not of algorithm
  `AEAD_AES_GCM_256` (use `DETERMINISTIC_ENCRYPT` for the
  `DETERMINISTIC_AEAD_AES_SIV_CMAC_256` algorithm instead).
- `plaintext` and `additional_data` must be the same type (both `STRING`
  or both `BYTES`); a `STRING`/`BYTES` mix outside the documented implicit
  cast pairing is rejected.
- Because output is randomized, equal plaintexts produce different
  ciphertexts; this prevents equality joins on the encrypted column.
  Use `DETERMINISTIC_ENCRYPT` if equal-ciphertext-for-equal-plaintext is
  required, accepting the resulting plaintext-equality leakage.
- A wrapped keyset (from `KEYS.NEW_WRAPPED_KEYSET`) must be unwrapped via
  `KEYS.KEYSET_CHAIN` before being passed as `keyset`.
- Passing plaintext keysets inline in query text may cause the keyset to
  be logged as part of query text; prefer KMS-wrapped keysets resolved at
  query time via `KEYS.KEYSET_CHAIN`.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#aeadencrypt>.
