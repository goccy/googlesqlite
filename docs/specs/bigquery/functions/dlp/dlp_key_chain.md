---
name: DLP_KEY_CHAIN
dialect: bigquery
category: functions/dlp
status: implemented
notes: |
  Backed by Cloud DLP templates and transforms; there is no in-process analog of the DLP service. Revisit only when a consumer gains a DLP simulator.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_key_chain
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_key_chain
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/dlp/dlp_key_chain.yaml
---

# DLP_KEY_CHAIN

## Summary
Builds a key descriptor that can be supplied in place of a plaintext `key`
argument to the DLP deterministic encryption functions, letting AES-SIV-based
DLP encryption run without exposing a raw key in the query text.

## Signatures
- `DLP_KEY_CHAIN(kms_resource_name, wrapped_key)`

## Behavior
- Returns a `STRUCT` value that pairs a Cloud KMS key reference with a
  wrapped (KMS-encrypted) data encryption key.
- The result is intended to be passed as the `key` argument of
  `DLP_DETERMINISTIC_ENCRYPT` and `DLP_DETERMINISTIC_DECRYPT`; those
  functions derive the actual data encryption key from this struct.
- `kms_resource_name` is a `STRING` literal naming the Cloud KMS key, in
  the form
  `gcp-kms://projects/<project>/locations/<region>/keyRings/<ring>/cryptoKeys/<key>`.
- `kms_resource_name` must not be `NULL` and must reside in the same
  Cloud region in which the query is executed.
- `wrapped_key` is a `BYTES` literal containing a user-chosen secret of
  16, 24, or 32 bytes that has already been wrapped (encrypted) with the
  named KMS key; wrapped keys are produced via `gcloud kms encrypt`.
- Because the wrapping is performed outside SQL, the plaintext key value
  never appears in the query — only the KMS reference and the wrapped
  bytes do.

## Examples
```sql
-- Wrapped key supplied as a BYTES literal.
SELECT
  DLP_DETERMINISTIC_ENCRYPT(
    DLP_KEY_CHAIN(
      'gcp-kms://projects/myproject/locations/us/keyRings/kms-test/cryptoKeys/test-Kek',
      b'\012\044\000\325\155\264\153\246\071\172\202\372\305\103\047\342\356\061\077\014\030\126\147\041\126\150\012\036\020\202\215\044\267\310\331\014\116\233\022\071\000\363\344\230\067\274\007\340\273\016\212\151\226\064\200\377\303\207\103\147\052\267\035\350\004\147\365\251\271\133\062\251\246\152\177\017\005\270\044\141\211\116\337\043\035\263\122\340\110\333\266\220\377\247\204\215\233'),
    'Plaintext',
    '',
    'aad') AS results;
-- expected results = 'AWDeSznl9C7+NzTaCgiqiEAZ8Y55fZSuvCQ='

-- Wrapped key materialised from a base64 string.
DECLARE DLP_KEY_VALUE BYTES;
SET DLP_KEY_VALUE =
  FROM_BASE64('CiQA1W20a6Y5elj6xUMn4u4xPwwYVmchVmgKHhCCjSS3yNkMTpsSOQDz5Jg3vAfguw6KaZY0gP/Dh0NnKrcd6ARn9am5WzKppmp/DwW4JGGJTt8jHbNS4EjbtpD/p4SNmw==');

SELECT
  DLP_DETERMINISTIC_ENCRYPT(
    DLP_KEY_CHAIN(
      'gcp-kms://projects/myproject/locations/us/keyRings/kms-test/cryptoKeys/test-Kek',
      DLP_KEY_VALUE),
    'Plaintext',
    '',
    'aad') AS results;
-- expected results = 'AWDeSznl9C7+NzTaCgiqiEAZ8Y55fZSuvCQ='
```

## Edge cases
- `kms_resource_name` cannot be `NULL`; passing `NULL` is rejected.
- The KMS key referenced by `kms_resource_name` must live in the same
  Cloud region as the query; cross-region usage is not supported.
- `wrapped_key` must decode to a 16-, 24-, or 32-byte secret after KMS
  unwrapping; other lengths are not valid AES-SIV key sizes.
- The caller must hold IAM permissions to use the referenced KMS key
  for unwrap operations, since the engine calls Cloud KMS at execution
  time to derive the data encryption key.

## Reference (upstream)

See the upstream BigQuery DLP encryption functions reference:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/dlp_functions#dlp_key_chain>
