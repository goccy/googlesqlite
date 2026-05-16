---
name: KEYS.KEYSET_CHAIN
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_chain
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_chain
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_keyset_chain.yaml
---

# KEYS.KEYSET_CHAIN

## Summary
Returns a `STRUCT` that wraps a Cloud KMS-encrypted Tink keyset together
with the KMS key resource path needed to decrypt it. The result can be
substituted for the `keyset` argument of the AEAD and deterministic
encryption functions, allowing those functions to be invoked without
embedding plaintext keys in the query.

## Signatures
- `KEYS.KEYSET_CHAIN(kms_resource_name, first_level_keyset)`

## Arguments
- `kms_resource_name`: `STRING` literal containing the resource path of
  the Cloud KMS key used to decrypt `first_level_keyset`. The KMS key
  must reside in the same Cloud region where the function is executed.
  The path takes the form
  `gcp-kms://projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>`.
- `first_level_keyset`: `BYTES` literal that holds a Tink keyset or a
  wrapped (KMS-encrypted) keyset.

## Return type
`STRUCT`. The struct is intended to be passed wherever a `keyset`
argument is expected by the AEAD or deterministic AEAD functions
(`AEAD.ENCRYPT`, `AEAD.DECRYPT_BYTES`, `AEAD.DECRYPT_STRING`,
`DETERMINISTIC_ENCRYPT`, `DETERMINISTIC_DECRYPT_BYTES`,
`DETERMINISTIC_DECRYPT_STRING`).

## Behavior
- Pairs a KMS resource path with a wrapped keyset so that AEAD
  functions can consume the keyset without the query handling the
  unwrapped key material directly.
- The KMS key referenced by `kms_resource_name` must live in the same
  Cloud region as the executing query.
- `first_level_keyset` may be either a plaintext Tink keyset or a
  wrapped keyset produced by `KEYS.NEW_WRAPPED_KEYSET`; in the wrapped
  case the engine uses `kms_resource_name` to decrypt it before use.
- For deterministic encryption, the wrapped keyset must have been
  created with key type `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`.
- Both arguments must be supplied as literals, not as values computed
  per row.

## Examples
```sql
-- Source data.
CREATE TABLE aead.RawCustomerData AS
SELECT 1 AS customer_id, b'jaguar' AS favorite_animal UNION ALL
SELECT 2 AS customer_id, b'zebra'  AS favorite_animal UNION ALL
SELECT 3 AS customer_id, b'zebra'  AS favorite_animal;

-- Encrypt with AEAD.ENCRYPT, supplying the keyset via KEYS.KEYSET_CHAIN.
DECLARE kms_resource_name STRING;
DECLARE first_level_keyset BYTES;
SET kms_resource_name = 'gcp-kms://projects/my-project/locations/us/keyRings/my-key-ring/cryptoKeys/my-crypto-key';
SET first_level_keyset = b'\012\044\000\107\275\360\176\264\206\332\235\215\304...';

CREATE TABLE aead.EncryptedCustomerData AS
SELECT
  customer_id,
  AEAD.ENCRYPT(
    KEYS.KEYSET_CHAIN(kms_resource_name, first_level_keyset),
    favorite_animal,
    CAST(CAST(customer_id AS STRING) AS BYTES)
  ) AS encrypted_animal
FROM aead.RawCustomerData;

-- Decrypt the same data with AEAD.DECRYPT_BYTES.
SELECT
  customer_id,
  AEAD.DECRYPT_BYTES(
    KEYS.KEYSET_CHAIN(kms_resource_name, first_level_keyset),
    encrypted_animal,
    CAST(CAST(customer_id AS STRING) AS BYTES)
  ) AS favorite_animal
FROM aead.EncryptedCustomerData;
-- expected: rows (1, b'jaguar'), (2, b'zebra'), (3, b'zebra')

-- Same pattern works with deterministic encryption when the wrapped
-- keyset uses DETERMINISTIC_AEAD_AES_SIV_CMAC_256; identical plaintexts
-- (rows 2 and 3 above) produce identical ciphertexts.
SELECT
  customer_id,
  DETERMINISTIC_ENCRYPT(
    KEYS.KEYSET_CHAIN(kms_resource_name, first_level_keyset),
    favorite_animal,
    CAST(CAST(customer_id AS STRING) AS BYTES)
  ) AS encrypted_animal
FROM aead.RawCustomerData;
```

## Edge cases
- Cross-region KMS keys are not allowed: the KMS key referenced by
  `kms_resource_name` must reside in the same Cloud region where the
  query runs.
- `kms_resource_name` and `first_level_keyset` must be literals; this
  precludes selecting either argument out of a per-row column.
- When `first_level_keyset` is a wrapped keyset, decryption is
  delegated to Cloud KMS, so KMS permission errors surface as query
  errors at execution time.
- Passing the result to a deterministic function with a non-SIV
  underlying key type is an error; the wrapped keyset must use the
  `DETERMINISTIC_AEAD_AES_SIV_CMAC_256` key type for deterministic use.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keyskeyset_chain>.
