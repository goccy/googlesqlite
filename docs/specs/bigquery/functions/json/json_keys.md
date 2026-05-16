---
name: JSON_KEYS
dialect: bigquery
category: functions/json
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#json_keys
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#json_keys
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/json/json_keys.yaml
---

# JSON_KEYS

## Summary
Returns the unique JSON object keys reachable from a JSON expression as an
`ARRAY<STRING>`, optionally limited to a maximum nesting depth and controlled
by a mode that decides how keys inside arrays are handled.

## Signatures
- `JSON_KEYS(json_expr)`
- `JSON_KEYS(json_expr, max_depth)`
- `JSON_KEYS(json_expr, mode => { 'strict' | 'lax' | 'lax recursive' })`
- `JSON_KEYS(json_expr, max_depth, mode => { 'strict' | 'lax' | 'lax recursive' })`

## Behavior
- `json_expr` is a `JSON` value that the function walks to collect object keys.
- `max_depth` is an `INT64` that caps how deep into nested fields the search
  goes; when omitted the entire document is traversed.
- `mode` is a named `STRING` argument with three values: `strict` (default)
  ignores any key that appears inside an array; `lax` also includes keys inside
  non-consecutively nested arrays; `lax recursive` returns every key,
  including those in consecutively nested arrays.
- Returned keys are de-duplicated and sorted in alphabetical order.
- Nested keys are joined with `.` (e.g. `a.b`); array indices are not part of
  any returned key.
- Keys containing special characters are escaped using double quotes; keys are
  case sensitive and not normalised.
- The return type is `ARRAY<STRING>`.

## Examples
```sql
-- No arrays: every object key is returned
SELECT JSON_KEYS(JSON '{"a": {"b":1}}') AS json_keys;
-- expected: [a, a.b]
```

```sql
-- max_depth = 1 limits traversal to the top level
SELECT JSON_KEYS(JSON '{"a": {"b":1}}', 1) AS json_keys;
-- expected: [a]
```

```sql
-- Default 'strict' mode skips keys inside arrays
SELECT JSON_KEYS(JSON '{"a":[{"b":1}, {"c":2}], "d":3}') AS json_keys;
-- expected: [a, d]
```

```sql
-- 'lax' mode includes keys inside (non-consecutively nested) arrays
SELECT JSON_KEYS(
  JSON '{"a":[{"b":1}, {"c":2}], "d":3}',
  mode => "lax") AS json_keys;
-- expected: [a, a.b, a.c, d]
```

```sql
-- 'lax' mode still excludes keys inside consecutively nested arrays
SELECT JSON_KEYS(JSON '{"a":[[{"b":1}]]}', mode => "lax") AS json_keys;
-- expected: [a]
```

```sql
-- 'lax recursive' descends through consecutively nested arrays
SELECT JSON_KEYS(JSON '{"a":[[{"b":1}]]}', mode => "lax recursive") AS json_keys;
-- expected: [a, a.b]
```

```sql
-- 'lax' with separated single-level arrays returns nested keys
SELECT JSON_KEYS(JSON '{"a":[{"b":[{"c":1}]}]}', mode => "lax") AS json_keys;
-- expected: [a, a.b, a.b.c]
```

```sql
-- Mixed nesting: 'lax' stops at the consecutively nested array
SELECT JSON_KEYS(JSON '{"a":[{"b":[[{"c":1}]]}]}', mode => "lax") AS json_keys;
-- expected: [a, a.b]
```

```sql
-- Same input under 'lax recursive' yields every key
SELECT JSON_KEYS(
  JSON '{"a":[{"b":[[{"c":1}]]}]}',
  mode => "lax recursive") AS json_keys;
-- expected: [a, a.b, a.b.c]
```

## Edge cases
- If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
- If `mode` is SQL `NULL`, the function returns SQL `NULL`.
- If `max_depth` is SQL `NULL`, the argument is ignored (full traversal).
- If `max_depth` is less than or equal to `0`, the function raises an error.
- `mode` only accepts the literal strings `'strict'`, `'lax'`, and
  `'lax recursive'`.

## Reference (upstream)

See the upstream BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#json_keys>
