---
name: PENDING_COMMIT_TIMESTAMP
dialect: spanner
category: functions/timestamp
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/timestamp_functions#pending_commit_timestamp
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/timestamp_functions#pending_commit_timestamp
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/timestamp/pending_commit_timestamp.yaml
---

# PENDING_COMMIT_TIMESTAMP

## Summary

Returns a sentinel value that, when stored in a `TIMESTAMP` column with `OPTIONS (allow_commit_timestamp = TRUE)`, is replaced at commit time with the actual transaction commit timestamp.

## Signatures

- `PENDING_COMMIT_TIMESTAMP()`

## Return type

`TIMESTAMP` (sentinel).

## Behavior

- The returned value is a placeholder, not a queryable timestamp; do not rely on its in-flight value.
- Only meaningful inside `INSERT`/`UPDATE` writes against a column declared with `allow_commit_timestamp = TRUE`. Reading the sentinel directly returns it unchanged.

## Examples

```sql
INSERT INTO Orders (OrderId, CreatedAt)
VALUES (1, PENDING_COMMIT_TIMESTAMP());
```

## Edge cases

- `SELECT PENDING_COMMIT_TIMESTAMP()` outside of a write returns the sentinel literally; client libraries typically refuse to materialize it.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/timestamp_functions#pending_commit_timestamp>.
