---
name: KEYS.ADD_KEY_FROM_RAW_BYTES
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_add_key_from_raw_bytes.yaml
---

# KEYS.ADD_KEY_FROM_RAW_BYTES

## Summary

Adds a key built from supplied raw key bytes to an existing keyset and
returns the updated keyset as a serialized `BYTES` value. The original
primary cryptographic key of the input keyset is preserved.

## Signatures

- `KEYS.ADD_KEY_FROM_RAW_BYTES(keyset, key_type, raw_key_bytes)`

## Arguments

- `keyset`: `BYTES`. A serialized keyset to which the new key is appended.
- `key_type`: `STRING` literal. Selects the key family for the new key.
  Supported values are `'AES_CBC_PKCS'` and `'AES_GCM'`.
- `raw_key_bytes`: `BYTES`. The raw key material. Its required length
  depends on `key_type` (see Behavior).

## Return type

`BYTES`.

## Behavior

- The returned keyset contains every key from the input `keyset` plus the
  new key derived from `raw_key_bytes`; the primary cryptographic key
  remains the same as in the input `keyset`.
- For `key_type = 'AES_CBC_PKCS'`, the new key is for AES decryption using
  cipher block chaining and PKCS padding. `raw_key_bytes` must be of
  length 16, 24, or 32 (128, 192, or 256 bits).
- GoogleSQL AEAD functions do not accept `'AES_CBC_PKCS'` keys for
  encryption; use `'AEAD_AES_GCM_256'` or `'AES_GCM'` keys for
  encryption instead.
- For `key_type = 'AES_GCM'`, the new key is for AES decryption or
  encryption using Galois/Counter Mode. `raw_key_bytes` must be of length
  16 or 32 (128 or 256 bits).
- Ciphertext produced by `AEAD.ENCRYPT` with an `'AES_GCM'` key has no
  Tink-specific prefix indicating which key was used.
- When the produced keyset is later passed to `AEAD.DECRYPT_STRING` or
  `AEAD.DECRYPT_BYTES`, decryption succeeds if any key in the keyset can
  decrypt the ciphertext.

## Examples

```sql
-- Adds an AES_CBC_PKCS key (built from per-customer raw key bytes) to a
-- freshly created AEAD_AES_GCM_256 keyset, one row per customer.
WITH CustomerRawKeys AS (
  SELECT 1 AS customer_id, b'0123456789012345' AS raw_key_bytes UNION ALL
  SELECT 2,                b'9876543210543210'                  UNION ALL
  SELECT 3,                b'0123012301230123'
), CustomerIds AS (
  SELECT 1 AS customer_id UNION ALL
  SELECT 2                UNION ALL
  SELECT 3
)
SELECT
  ci.customer_id,
  KEYS.ADD_KEY_FROM_RAW_BYTES(
    KEYS.NEW_KEYSET('AEAD_AES_GCM_256'),
    'AES_CBC_PKCS',
    (SELECT raw_key_bytes FROM CustomerRawKeys AS crk
       WHERE crk.customer_id = ci.customer_id)
  ) AS keyset
FROM CustomerIds AS ci;
-- expected: one row per customer_id with a serialized two-key keyset
```

## Edge cases

- If `raw_key_bytes` does not match a permitted length for the chosen
  `key_type`, the function fails.
- `key_type` values other than `'AES_CBC_PKCS'` and `'AES_GCM'` are not
  documented as supported here.
- Adding an `'AES_CBC_PKCS'` key does not enable encryption with that key:
  `AEAD.ENCRYPT` will continue to use the keyset's primary key (the
  unchanged `AEAD_AES_GCM_256` or other AEAD key).

## Reference (upstream)

See the BigQuery AEAD encryption functions reference, section
`KEYS.ADD_KEY_FROM_RAW_BYTES`:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysadd_key_from_raw_bytes>
