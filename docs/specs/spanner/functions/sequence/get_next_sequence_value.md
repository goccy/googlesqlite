---
name: GET_NEXT_SEQUENCE_VALUE
dialect: spanner
category: functions/sequence
status: implemented
notes: |
  Spanner sequences depend on Spanner's allocator + recovery semantics. googlesqlite has no sequence object; emulation under SQLite AUTOINCREMENT would diverge from the spec.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/sequence_functions#get_next_sequence_value
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/sequence_functions#get_next_sequence_value
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/sequence/get_next_sequence_value.yaml
---

# GET_NEXT_SEQUENCE_VALUE

## Summary

Atomically advances a Spanner sequence and returns the next value.

## Signatures

- `GET_NEXT_SEQUENCE_VALUE(sequence_name)`

## Arguments

- `sequence_name`: `STRING` fully-qualified sequence name.

## Return type

`INT64`.

## Behavior

- Each evaluation advances the sequence and returns the new value. Concurrent calls observe distinct values.
- Requires that the named sequence exists; otherwise an error is raised.

## Examples

```sql
INSERT INTO Orders (OrderId, ...) VALUES (GET_NEXT_SEQUENCE_VALUE("OrderSeq"), ...);
```

## Edge cases

- The increment direction and skip pattern are determined by the sequence definition (see `CREATE SEQUENCE`).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/sequence_functions#get_next_sequence_value>.
