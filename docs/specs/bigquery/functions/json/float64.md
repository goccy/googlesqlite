---
name: FLOAT64
dialect: bigquery
category: functions/json
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#double_for_json
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#double_for_json
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/json/float64.yaml
---

# FLOAT64

## Summary
Converts a JSON number to a SQL `FLOAT64` value. An optional `wide_number_mode` named argument controls behavior when the JSON number can't be represented as `FLOAT64` without loss of precision.

## Signatures
- `FLOAT64(json_expr)`
- `FLOAT64(json_expr, wide_number_mode => { 'exact' | 'round' })`

## Behavior
- `json_expr` must be a JSON value of type number; if it is any other JSON type (e.g. string, JSON `null`, boolean, object, array), the function raises an error.
- If `json_expr` is a SQL `NULL`, the function returns SQL `NULL`.
- `wide_number_mode` is a named argument with a `STRING` value that controls handling of numbers that can't be represented as `FLOAT64` without loss of precision.
- `wide_number_mode` values are case-sensitive and must be exactly `'exact'` or `'round'`; any other spelling or casing raises an error.
- With `wide_number_mode => 'exact'`, the function fails if the result can't be represented as a `FLOAT64` without loss of precision.
- With `wide_number_mode => 'round'` (the default when the argument is omitted), the JSON numeric value is rounded to `FLOAT64`; if rounding still isn't possible, the function fails.
- The return type is `FLOAT64`.
- `SAFE.FLOAT64(...)` returns SQL `NULL` instead of raising on errors that this function would otherwise produce.

## Examples
```sql
SELECT FLOAT64(JSON '9.8') AS velocity;
-- expected: 9.8

SELECT FLOAT64(JSON_QUERY(JSON '{"vo2_max": 39.1, "age": 18}', "$.vo2_max")) AS vo2_max;
-- expected: 39.1

SELECT FLOAT64(JSON '18446744073709551615', wide_number_mode=>'round') AS result;
-- expected: 1.8446744073709552e+19

SELECT FLOAT64(JSON '18446744073709551615') AS result;
-- expected: 1.8446744073709552e+19 (default mode is 'round')
```

## Edge cases
- `FLOAT64(JSON '"strawberry"')` raises: the JSON value is a string, not a number.
- `FLOAT64(JSON 'null')` raises: JSON `null` is not a number (this is distinct from a SQL `NULL` json_expr, which yields SQL `NULL`).
- `FLOAT64(JSON '123.4', wide_number_mode=>'EXACT')` raises: `wide_number_mode` is case-sensitive.
- `FLOAT64(JSON '123.4', wide_number_mode=>'exac')` raises: only `'exact'` and `'round'` are accepted.
- `FLOAT64(JSON '18446744073709551615', wide_number_mode=>'exact')` raises: the value can't be converted to `FLOAT64` without loss of precision.
- `SAFE.FLOAT64(JSON '"strawberry"')` returns SQL `NULL` instead of raising.

## Reference (upstream)

See the upstream documentation: <https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#double_for_json>
