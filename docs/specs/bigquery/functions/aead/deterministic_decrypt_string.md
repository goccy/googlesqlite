---
name: DETERMINISTIC_DECRYPT_STRING
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_decrypt_string
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_decrypt_string
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/aead/deterministic_decrypt_string.yaml
---

# DETERMINISTIC_DECRYPT_STRING

## Summary
Deterministically decrypts a `BYTES` ciphertext produced by `DETERMINISTIC_ENCRYPT` and returns the plaintext as a `STRING`. Behaves the same as `DETERMINISTIC_DECRYPT_BYTES` except that the recovered plaintext is interpreted as a `STRING`.

## Signatures
- `DETERMINISTIC_DECRYPT_STRING(keyset, ciphertext, additional_data)`

## Behavior
- Uses the matching key from `keyset` to decrypt `ciphertext` and verifies the data's integrity using `additional_data`; returns an error if decryption or verification fails.
- `keyset` is a serialized `BYTES` value or a `STRUCT` returned by one of the `KEYS.*` functions; it must contain the encrypting key, the key must be in the `ENABLED` state, and its algorithm must be `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`, otherwise the function errors.
- The matching key inside `keyset` is identified via the key ID embedded in the Tink-formatted prefix of `ciphertext`.
- `ciphertext` must be a `BYTES` value produced by `DETERMINISTIC_ENCRYPT` and follow Tink's wire format (Tink key version byte, 4-byte key hint, then IV/auth-tag/payload as defined by Deterministic AEAD; for SIV the IV and authentication tag are combined and prepended).
- `additional_data` is a `STRING` or `BYTES` value that binds the ciphertext to its encryption context; any `STRING` is cast to `BYTES`. It must match (ignoring type) the `additional_data` supplied to `DETERMINISTIC_ENCRYPT`, or the function errors.
- The return type is `STRING`; the only difference from `DETERMINISTIC_DECRYPT_BYTES` is that the decrypted plaintext is returned as a `STRING` instead of `BYTES`.

## Examples
```sql
-- Round-trip a STRING value through deterministic AEAD encryption.
WITH CustomerKeysets AS (
  SELECT 1 AS customer_id,
         KEYS.NEW_KEYSET('DETERMINISTIC_AEAD_AES_SIV_CMAC_256') AS keyset
),
Encrypted AS (
  SELECT ck.customer_id,
         ck.keyset,
         DETERMINISTIC_ENCRYPT(
           ck.keyset,
           'elephant',
           CAST(ck.customer_id AS STRING)
         ) AS encrypted_animal
  FROM CustomerKeysets AS ck
)
SELECT customer_id,
       DETERMINISTIC_DECRYPT_STRING(
         keyset,
         encrypted_animal,
         CAST(customer_id AS STRING)
       ) AS favorite_animal
FROM Encrypted;
-- expected: customer_id=1, favorite_animal='elephant'
```

## Edge cases
- Returns an error if the key referenced by `ciphertext` is missing from `keyset`, is not `ENABLED`, or is not of algorithm `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`.
- Returns an error if `additional_data` does not match the value supplied at encryption (after casting `STRING` to `BYTES`).
- Returns an error if `ciphertext` is malformed (e.g. wrong Tink wire-format prefix or truncated IV/authentication tag).
- The page does not document explicit `NULL` propagation for this function; see `DETERMINISTIC_ENCRYPT`, which states that it returns `NULL` if any input is `NULL`.

## Reference (upstream)

See the upstream reference page for full prose and additional context: <https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#deterministic_decrypt_string>.
