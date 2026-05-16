---
name: DLP_DETERMINISTIC_ENCRYPT
dialect: bigquery
category: functions/dlp
status: implemented
notes: |
  Backed by Cloud DLP templates and transforms; there is no in-process analog of the DLP service. Revisit only when a consumer gains a DLP simulator.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_deterministic_encrypt
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_deterministic_encrypt
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/dlp/dlp_deterministic_encrypt.yaml
---

# DLP_DETERMINISTIC_ENCRYPT

## Summary
Encrypts a `STRING` value with a DLP-compatible deterministic algorithm,
deriving a data encryption key from a `DLP_KEY_CHAIN`-wrapped key and an
optional context, and returning the ciphertext as `STRING`.

## Signatures
- `DLP_DETERMINISTIC_ENCRYPT(key, plaintext, surrogate)`
- `DLP_DETERMINISTIC_ENCRYPT(key, plaintext, surrogate, context)`

## Behavior
- `key` is a serialized `BYTES` value returned by `DLP_KEY_CHAIN`; it
  wraps a Cloud KMS key that must be in the `ENABLED` state.
- `plaintext` is the `STRING` value to encrypt.
- `surrogate` is a `STRING` prepended to the encryption result; pass an
  empty string (`""`) when no surrogate prefix is desired.
- `context`, when supplied, is a user-provided `STRING` mixed with the
  Cloud KMS key to derive the actual data encryption key (matches the
  DLP `CryptoDeterministicConfig.context` field). The same `context`
  must be supplied to `DLP_DETERMINISTIC_DECRYPT` to recover the
  plaintext.
- The function is deterministic: the same `(key, plaintext, surrogate,
  context)` inputs always produce the same ciphertext.
- The return type is `STRING`. When a non-empty `surrogate` is provided
  the result has the form `<surrogate>(<len>):<ciphertext>` where
  `<len>` is the ciphertext length in characters.
- Using DLP functions requires first generating a Cloud KMS key and then
  obtaining a wrapped key for it (see `gcloud kms encrypt` and
  `DLP_KEY_CHAIN`).

## Examples
```sql
-- Wrapped key supplied as a BYTES literal; no surrogate prefix.
SELECT
  DLP_DETERMINISTIC_ENCRYPT(
    DLP_KEY_CHAIN(
      'gcp-kms://projects/myproject/locations/us/keyRings/kms-test/cryptoKeys/test-KEK',
      b'\012\044\000...'  -- serialized wrapped key bytes
    ),
    'Plaintext',
    '',
    'aad'
  ) AS results;
-- expected: AWDeSznl9C7+NzTaCgiqiEAZ8Y55fZSuvCQ=
```

```sql
-- Wrapped key supplied via base64; surrogate prefix included in output.
DECLARE DLP_KEY_VALUE BYTES;
SET DLP_KEY_VALUE = FROM_BASE64(
  'CiQA1W20a6Y5elj6xUMn4u4xPwwYVmchVmgKHhCCjSS3yNkMTpsSOQDz5Jg3vAfguw6KaZY0gP/Dh0NnKrcd6ARn9am5WzKppmp/DwW4JGGJTt8jHbNS4EjbtpD/p4SNmw==');

SELECT
  DLP_DETERMINISTIC_ENCRYPT(
    DLP_KEY_CHAIN(
      'gcp-kms://projects/myproject/locations/us/keyRings/kms-test/cryptoKeys/test-Kek',
      DLP_KEY_VALUE),
    'Plaintext',
    'your_surrogate',
    'aad'
  ) AS results;
-- expected: your_surrogate(36):AWDeSznl9C7+NzTaCgiqiEAZ8Y55fZSuvCQ=
```

## Edge cases
- The wrapped Cloud KMS key referenced by `key` must be `ENABLED`;
  otherwise the call fails. Disabling or destroying the underlying KMS
  key invalidates further encryption calls.
- An empty-string `surrogate` produces output with no prefix; a
  non-empty `surrogate` produces `<surrogate>(<len>):<ciphertext>`.
- Because the function is deterministic, equal plaintexts under the
  same `(key, surrogate, context)` produce equal ciphertexts: this
  enables equality joins on encrypted columns but leaks plaintext
  equality.
- The 4-argument form binds the ciphertext to `context`; decryption with
  a different `context` will not recover the original plaintext.
- The caller's identity must hold the IAM permissions required to use
  the referenced Cloud KMS key (`cloudkms.cryptoKeyVersions.useToEncrypt`
  via Cloud KMS, plus the BigQuery permissions needed to run the
  query). KMS region restrictions on the key apply.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_deterministic_encrypt>.
