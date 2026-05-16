---
name: KEYS.REWRAP_KEYSET
dialect: bigquery
category: functions/aead
status: implemented
notes: |
  Encrypted by an AES-GCM / deterministic-encryption keyset. BigQuery routes through Cloud KMS for keyset management; googlesqlite has no policy for accepting raw keysets safely (where to store, how to scope). The pure-Go cipher path is trivial, but turning it on without a KMS-equivalent surface would silently expose keys. Revisit when a consumer defines a key-management model.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysrewrap_keyset
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aead/keys_rewrap_keyset.yaml
---

# KEYS.REWRAP_KEYSET

## Summary
Re-wraps an existing Cloud KMS-wrapped keyset under a different Cloud KMS key,
returning the rewrapped keyset as a `BYTES` serialization of
`google.crypto.tink.Keyset`. The underlying cryptographic keys inside the
keyset are preserved unchanged; only the KMS key that protects them changes.

## Signatures
- `KEYS.REWRAP_KEYSET(source_kms_resource_name, target_kms_resource_name, wrapped_keyset)`

## Behavior
- `source_kms_resource_name` is a `STRING` literal naming the Cloud KMS key
  currently used to wrap `wrapped_keyset`, formatted as
  `gcp-kms://projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>`.
- `target_kms_resource_name` is a `STRING` literal naming the new Cloud KMS key
  that will be used to re-encrypt the keyset, in the same `gcp-kms://...`
  format.
- Both Cloud KMS keys must reside in the same Cloud region in which the query
  executes; cross-region references are not supported.
- `wrapped_keyset` is a `BYTES` literal containing the existing wrapped keyset
  to re-wrap.
- During the rewrap, the wrapped keyset is decrypted using
  `source_kms_resource_name` and re-encrypted using `target_kms_resource_name`;
  the cleartext keyset is never exposed to the caller.
- The cryptographic keys inside the keyset (primary plus any additional keys)
  are unchanged — only the wrapping KMS key changes — so all ciphertext
  previously encrypted under the keyset remains decryptable with the rewrapped
  keyset.
- Return type is `BYTES`.

## Examples
```sql
DECLARE kms_key_current STRING;
DECLARE kms_key_new STRING;
DECLARE wrapped_keyset BYTES;
SET kms_key_current = 'gcp-kms://projects/my-project/locations/us/keyRings/my-key-ring/cryptoKeys/old-key';
SET kms_key_new     = 'gcp-kms://projects/my-project/locations/us/keyRings/my-key-ring/cryptoKeys/new-key';
SET wrapped_keyset  = b'\012\044\000...';

-- Re-wrap an existing wrapped keyset under a different KMS key.
SELECT KEYS.REWRAP_KEYSET(kms_key_current, kms_key_new, wrapped_keyset);
```

## Edge cases
- `source_kms_resource_name` must be the KMS key originally used to wrap
  `wrapped_keyset`; otherwise the unwrap step fails.
- Both KMS keys must live in the same Cloud region as the query.
- All three arguments are required literals; the documented signature does not
  provide for NULL inputs.

## Reference (upstream)
- https://cloud.google.com/bigquery/docs/reference/standard-sql/aead_encryption_functions#keysrewrap_keyset
