---
name: KEYS.NEW_WRAPPED_KEYSET
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysnew_wrapped_keyset
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_new_wrapped_keyset.yaml
---

# KEYS.NEW_WRAPPED_KEYSET

## Summary
Creates a new keyset and wraps (encrypts) it with a Cloud KMS key, returning
the wrapped keyset as a `BYTES` serialization of a `google.crypto.tink.Keyset`
containing a single primary cryptographic key and no additional keys.

## Signatures
- `KEYS.NEW_WRAPPED_KEYSET(kms_resource_name, key_type)`

## Behavior
- `kms_resource_name` is a non-NULL `STRING` literal naming a Cloud KMS key,
  formatted as `gcp-kms://projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>`.
- The referenced Cloud KMS key must reside in the same Cloud region in which
  the query executes.
- `key_type` is a non-NULL `STRING` literal selecting the keyset algorithm,
  one of:
  - `AEAD_AES_GCM_256` — 256-bit key generated with the boringSSL PRNG; uses
    AES-GCM for encryption/decryption.
  - `DETERMINISTIC_AEAD_AES_SIV_CMAC_256` — 512-bit AES-SIV-CMAC key (a
    256-bit AES-CTR key plus a 256-bit AES-CMAC key) generated with the
    boringSSL PRNG; uses AES-SIV for encryption/decryption.
- Returns a fresh wrapped keyset on each query invocation; calling the
  function multiple times across separate queries yields distinct wrapped
  keysets even when the arguments are identical.
- Within a single query, multiple calls to this function with the same
  arguments return the same value (so a wrapped keyset projected over a
  table is generated once and broadcast to every row).
- Return type is `BYTES`.

## Examples
```sql
DECLARE kms_resource_name STRING;
SET kms_resource_name = 'gcp-kms://projects/my-project/locations/us/keyRings/my-key-ring/cryptoKeys/my-crypto-key';

-- Create a wrapped keyset.
SELECT KEYS.NEW_WRAPPED_KEYSET(kms_resource_name, 'AEAD_AES_GCM_256');

-- Same wrapped keyset is reused for every row of my_table.
SELECT *,
       KEYS.NEW_WRAPPED_KEYSET(kms_resource_name, 'AEAD_AES_GCM_256')
FROM my_table;
```

## Edge cases
- `kms_resource_name` and `key_type` must both be non-NULL string literals;
  NULL values are rejected.
- The Cloud KMS key must be in the same region as the query; cross-region
  references are not supported.
- Only the two `key_type` values listed above are accepted.

## Reference (upstream)
- https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysnew_wrapped_keyset
