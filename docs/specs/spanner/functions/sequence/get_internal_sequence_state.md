---
name: GET_INTERNAL_SEQUENCE_STATE
dialect: spanner
category: functions/sequence
status: implemented
notes: |
  Spanner sequences depend on Spanner's allocator + recovery semantics. googlesqlite has no sequence object; emulation under SQLite AUTOINCREMENT would diverge from the spec.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/sequence_functions#get_internal_sequence_state
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/sequence_functions#get_internal_sequence_state
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/sequence/get_internal_sequence_state.yaml
---

# GET_INTERNAL_SEQUENCE_STATE

## Summary

Returns the current internal counter state of a Spanner sequence object as an opaque integer.

## Signatures

- `GET_INTERNAL_SEQUENCE_STATE(sequence_name)`

## Arguments

- `sequence_name`: `STRING` fully-qualified Spanner sequence name.

## Return type

`INT64`.

## Behavior

- The result is a snapshot of the internal counter at evaluation time. Repeated invocations may observe the counter advance.
- Requires that the named sequence exists; otherwise an error is raised.

## Examples

```sql
SELECT GET_INTERNAL_SEQUENCE_STATE("MySequence");
```

## Edge cases

- Use `GET_NEXT_SEQUENCE_VALUE` to advance the sequence; this function only inspects state.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/sequence_functions#get_internal_sequence_state>.
