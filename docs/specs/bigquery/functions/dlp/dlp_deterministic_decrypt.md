---
name: DLP_DETERMINISTIC_DECRYPT
dialect: bigquery
category: functions/dlp
status: implemented
notes: |
  Backed by Cloud DLP templates and transforms; there is no in-process analog of the DLP service. Revisit only when a consumer gains a DLP simulator.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_deterministic_decrypt
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_deterministic_decrypt
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/dlp/dlp_deterministic_decrypt.yaml
---

# DLP_DETERMINISTIC_DECRYPT

## Summary
Decrypts a `ciphertext` STRING that was produced by Cloud DLP deterministic
encryption, using a Cloud KMS-wrapped key returned by `DLP_KEY_CHAIN` and an
optional `context` value to derive the data encryption key. A `surrogate`
prefix that was used at encryption time can be supplied to be stripped from
the decrypted output.

## Signatures
- `DLP_DETERMINISTIC_DECRYPT(key, ciphertext, surrogate)`
- `DLP_DETERMINISTIC_DECRYPT(key, ciphertext, surrogate, context)`

## Behavior
- Decrypts `ciphertext` with an encryption key derived from `key` and, when
  supplied, `context`; the same `context` value used at encryption time must
  be passed to obtain the original plaintext.
- `key` is a serialized `BYTES` value returned by `DLP_KEY_CHAIN`; the
  underlying Cloud KMS key must be in the `ENABLED` state.
- `ciphertext` is the `STRING` value to decrypt; if `surrogate` is non-empty
  it identifies the prefix attached during encryption and is removed from
  the value before decryption.
- `surrogate` is a `STRING` that callers can use to tag and later strip a
  prefix from the encrypted token; pass an empty string `""` to opt out.
- `context` is a `STRING` whose bytes are mixed with the Cloud KMS key to
  derive the data encryption key, matching the `CryptoDeterministicConfig`
  context used at encryption.
- Returns a `STRING` containing the recovered plaintext.

## Examples
```sql
-- Wrapped key passed as a BYTES literal.
SELECT
  DLP_DETERMINISTIC_DECRYPT(
    DLP_KEY_CHAIN(
      'gcp-kms://projects/myproject/locations/us/keyRings/kms-test/cryptoKeys/test-Kek',
      b'\012\044\000\325...'  -- wrapped key BYTES
    ),
    'AWDeSznl9C7+NzTaCgiqiEAZ8Y55fZSuvCQ=',
    '',
    'aad'
  ) AS results;
-- expected: results = 'Plaintext'

-- Wrapped key supplied via a base64-decoded BYTES variable, with a surrogate.
DECLARE DLP_KEY_VALUE BYTES;
SET DLP_KEY_VALUE = FROM_BASE64(
  'CiQA1W20a6Y5elj6xUMn4u4xPwwYVmchVmgKHhCCjSS3yNkMTpsSOQDz5Jg3vAfguw6KaZY0gP/Dh0NnKrcd6ARn9am5WzKppmp/DwW4JGGJTt8jHbNS4EjbtpD/p4SNmw==');

SELECT
  DLP_DETERMINISTIC_DECRYPT(
    DLP_KEY_CHAIN(
      'gcp-kms://projects/myproject/locations/us/keyRings/kms-test/cryptoKeys/test-Kek',
      DLP_KEY_VALUE
    ),
    'your_surrogate(36):AWDeSznl9C7+NzTaCgiqiEAZ8Y55fZSuvCQ=',
    'your_surrogate',
    'aad'
  ) AS results;
-- expected: results = 'Plaintext'
```

## Edge cases
- Returns an error if the Cloud KMS key referenced by `key` is not in the
  `ENABLED` state, or if the caller lacks the IAM permissions Cloud KMS
  requires to unwrap the data encryption key.
- The `context` argument must match the value used at encryption time; a
  mismatched or missing `context` yields a different derived key, so
  decryption fails.
- When `surrogate` is non-empty, `ciphertext` is expected to carry the
  matching prefix produced at encryption time; supply `""` for `surrogate`
  if no prefix was used.
- Generating the wrapped `key` requires a separately created cryptographic
  key wrapped via `gcloud kms encrypt` (see `DLP_KEY_CHAIN`).

## Reference (upstream)

See the official BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_deterministic_decrypt>
