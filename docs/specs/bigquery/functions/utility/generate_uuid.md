---
name: GENERATE_UUID
dialect: bigquery
category: functions/utility
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/utility-functions#generate_uuid
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/utility-functions#generate_uuid
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/utility/generate_uuid.yaml
---

# GENERATE_UUID

## Summary
Returns a freshly generated random universally unique identifier
(UUID) as a lowercase `STRING`, formatted per RFC 4122 section 4.4.

## Signatures
- `GENERATE_UUID()`

## Behavior
- Takes no arguments.
- Returns a `STRING` value.
- The returned string is exactly 36 characters: 32 hexadecimal digits
  arranged in five groups separated by hyphens in the form
  `8-4-4-4-12`.
- Of the 128 bits, 122 are random and 6 are fixed, in compliance with
  RFC 4122 section 4.4 (variant and version bits).
- All hexadecimal digits in the returned string are lowercase.
- Each invocation produces an independently generated value, so the
  function is non-deterministic.

## Examples
```sql
SELECT GENERATE_UUID() AS uuid;
-- expected: a single STRING column matching the regex
--           ^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$
-- example value: '4192bff0-e1e0-43ce-a4db-912808c32493'
```

## Edge cases
- Calling `GENERATE_UUID()` multiple times in the same query (or even
  the same row) yields independent random values; treat results as
  non-deterministic for testing purposes.
- The function never returns `NULL` and never returns an error under
  normal use; it has no inputs that could be `NULL` or out of range.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/utility-functions#generate_uuid>.
