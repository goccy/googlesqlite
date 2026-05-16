---
name: LAX_FLOAT64
dialect: bigquery
category: functions/json
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#lax_double
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#lax_double
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/json/lax_float64.yaml
---

# LAX_FLOAT64

## Summary
Attempts to convert a JSON value to a SQL `FLOAT64` value, applying lenient
per-JSON-type conversion rules and returning SQL `NULL` (rather than raising)
when a value can't be converted.

## Signatures
- `LAX_FLOAT64(json_expr)`

## Behavior
- `json_expr` is a JSON value; the return type is `FLOAT64`.
- If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
- If the JSON value is a number, it is cast to `FLOAT64`; large numbers that
  exceed `FLOAT64` precision are rounded rather than raising an error.
- If the JSON value is a string, the string is parsed as a `BIGNUMERIC` and
  the result is then safe-cast to `FLOAT64`; strings that can't be parsed
  yield SQL `NULL`. Strings naming non-finite values (`"NaN"`, `"Inf"`,
  `"Infinity"`, including signed and mixed-case variants like `"-InfiNiTY"`)
  are accepted and produce the corresponding `FLOAT64` non-finite values.
- If the JSON value is a boolean, the function returns SQL `NULL` for both
  `true` and `false`.
- If the JSON value is JSON `null`, an array, or an object, the function
  returns SQL `NULL`.
- The function never raises a conversion error for non-`NULL` JSON input —
  unparseable cases are surfaced as SQL `NULL` instead.

## Examples
```sql
SELECT LAX_FLOAT64(JSON '9.8') AS result;
-- expected: 9.8

SELECT LAX_FLOAT64(JSON '9') AS result;
-- expected: 9.0

SELECT LAX_FLOAT64(JSON '9007199254740993') AS result;
-- expected: 9007199254740992.0 (rounded; exceeds FLOAT64 precision)

SELECT LAX_FLOAT64(JSON '1e100') AS result;
-- expected: 1e+100

SELECT LAX_FLOAT64(JSON 'true') AS result;
-- expected: NULL

SELECT LAX_FLOAT64(JSON 'false') AS result;
-- expected: NULL

SELECT LAX_FLOAT64(JSON '"10"') AS result;
-- expected: 10.0

SELECT LAX_FLOAT64(JSON '"1.1"') AS result;
-- expected: 1.1

SELECT LAX_FLOAT64(JSON '"1.1e2"') AS result;
-- expected: 110.0

SELECT LAX_FLOAT64(JSON '"9007199254740993"') AS result;
-- expected: 9007199254740992.0

SELECT LAX_FLOAT64(JSON '"+1.5"') AS result;
-- expected: 1.5

SELECT LAX_FLOAT64(JSON '"NaN"') AS result;
-- expected: NaN

SELECT LAX_FLOAT64(JSON '"Inf"') AS result;
-- expected: Infinity

SELECT LAX_FLOAT64(JSON '"-InfiNiTY"') AS result;
-- expected: -Infinity

SELECT LAX_FLOAT64(JSON '"foo"') AS result;
-- expected: NULL
```

## Edge cases
- JSON booleans (`true`, `false`) convert to SQL `NULL`, unlike the strict
  `FLOAT64` function which would raise on non-number JSON.
- JSON `null`, JSON arrays, and JSON objects all convert to SQL `NULL`.
- A JSON string that doesn't represent a parseable number (e.g. `"foo"`)
  returns SQL `NULL` rather than raising.
- Non-finite string forms `"NaN"`, `"Inf"`, `"Infinity"` (with optional sign
  and case-insensitive matching) are recognised and converted to the
  corresponding `FLOAT64` non-finite values.
- JSON numbers that don't fit in `FLOAT64` (e.g. `9007199254740993`) are
  rounded to the nearest representable `FLOAT64` rather than failing.
- A SQL `NULL` `json_expr` propagates as SQL `NULL`; this is distinct from
  a JSON `null` value, which also returns SQL `NULL` per the conversion
  rules.

## Reference (upstream)

See the upstream documentation: <https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#lax_double>
