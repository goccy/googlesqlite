---
name: KEYS.ROTATE_WRAPPED_KEYSET
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysrotate_wrapped_keyset
last_synced: 2026-05-06
testdata: testdata/specs/bigquery/functions/aead/keys_rotate_wrapped_keyset.yaml
---

# KEYS.ROTATE_WRAPPED_KEYSET

## Summary
Takes an existing Cloud KMS-wrapped keyset, adds a freshly generated key, and
returns the rewrapped keyset as a `BYTES` serialization of
`google.crypto.tink.Keyset`. The previous primary key is retained as an
additional (non-primary) key in the rotated keyset.

## Signatures
- `KEYS.ROTATE_WRAPPED_KEYSET(kms_resource_name, wrapped_keyset, key_type)`

## Behavior
- `kms_resource_name` is a `STRING` literal naming the Cloud KMS key that was
  used to wrap `wrapped_keyset`, formatted as
  `gcp-kms://projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>`.
- The referenced Cloud KMS key must reside in the same Cloud region in which
  the query executes.
- `wrapped_keyset` is a `BYTES` literal containing the existing wrapped keyset
  to rotate.
- `key_type` is a `STRING` literal selecting the keyset algorithm and must
  match the key type of the existing keys inside `wrapped_keyset` (for
  example `AEAD_AES_GCM_256` or `DETERMINISTIC_AEAD_AES_SIV_CMAC_256`).
- During rotation the wrapped keyset is decrypted, a new key is generated and
  installed as the new primary, and the keyset is re-encrypted; the cleartext
  keyset is never exposed to the caller.
- The previous primary key from `wrapped_keyset` is preserved in the returned
  keyset as an additional key, so ciphertext encrypted under the old key
  remains decryptable.
- Each query invocation yields a distinct rotated wrapped keyset, but multiple
  calls with the same arguments within a single query return the same value
  (so the rotation is computed once and broadcast across rows).
- Return type is `BYTES`.

## Examples
```sql
DECLARE kms_resource_name STRING;
DECLARE wrapped_keyset BYTES;
SET kms_resource_name = 'gcp-kms://projects/my-project/locations/us/keyRings/my-key-ring/cryptoKeys/my-crypto-key';
SET wrapped_keyset = b'\012\044\000...';

-- Rotate a wrapped keyset; a new wrapped keyset is produced per query.
SELECT KEYS.ROTATE_WRAPPED_KEYSET(kms_resource_name, wrapped_keyset, 'AEAD_AES_GCM_256');

-- Same rotated wrapped keyset is reused for every row of my_table.
SELECT *,
       KEYS.ROTATE_WRAPPED_KEYSET(kms_resource_name, wrapped_keyset, 'AEAD_AES_GCM_256')
FROM my_table;
```

## Edge cases
- `key_type` must match the existing key type embedded in `wrapped_keyset`;
  a mismatch is an error.
- The Cloud KMS key referenced by `kms_resource_name` must be the same key
  originally used to wrap `wrapped_keyset` and must live in the same region
  as the query; cross-region references are not supported.
- All three arguments are required literals; the documented signature does
  not provide for NULL inputs.

## Reference (upstream)
- https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysrotate_wrapped_keyset
