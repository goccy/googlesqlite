---
name: CONTAINS_SUBSTR
dialect: bigquery
category: functions/string
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/string_functions#contains_substr
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/string_functions#contains_substr
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/string/contains_substr.yaml
---

# CONTAINS_SUBSTR

## Summary
Performs a normalized, case-insensitive search to see if a value exists as a
substring in an expression. Returns `TRUE` if the value is found, otherwise
`FALSE` (or `NULL` when the expression itself is `NULL` or no match is
found but a `NULL` field was encountered during a cross-field search).

## Signatures
- `CONTAINS_SUBSTR(expression, search_value_literal)`
- `CONTAINS_SUBSTR(expression, search_value_literal, json_scope => json_scope_value)`

## Behavior
- Both operands are normalized and case folded with NFKC normalization
  before comparison; wildcard searches are not supported.
- `search_value_literal` must be a `STRING` literal or a `STRING` constant
  expression (for example `CONCAT('a', 'b')`).
- `expression` may be a column or a table reference (a table reference is
  evaluated as a `STRUCT` whose fields are the table columns), and may
  resolve to `STRING`, `INT64`, `BOOL`, `NUMERIC`, `BIGNUMERIC`,
  `TIMESTAMP`, `TIME`, `DATE`, `DATETIME`, `ARRAY`, or `STRUCT`; the
  evaluated value is cast to `STRING` before the substring search.
- For `STRUCT` or `ARRAY` expressions the search is a cross-field search:
  each field/element (recursively for nested `STRUCT`s) is converted to a
  string individually and tested. Returns `TRUE` if any field matches;
  otherwise returns `NULL` if at least one field is `NULL`; otherwise
  `FALSE` when no field matches and all fields are non-`NULL`.
- When `expression` is `NULL`, the result is `NULL`.
- The optional named argument `json_scope` controls which parts of `JSON`
  data are searched; it has no effect when the expression is not `JSON`
  and contains no `JSON` field. Allowed values are `'JSON_VALUES'`
  (default — values only), `'JSON_KEYS'` (keys only), and
  `'JSON_KEYS_AND_VALUES'` (both).
- Return type is `BOOL`.

## Examples
```sql
SELECT CONTAINS_SUBSTR('the blue house', 'Blue house') AS result;
-- expected: true (case-insensitive match)

SELECT CONTAINS_SUBSTR('the blue house', CONCAT('Blue ', 'house')) AS result;
-- expected: true (search value may be a STRING constant expression)

SELECT CONTAINS_SUBSTR('the red house', 'blue') AS result;
-- expected: false

SELECT CONTAINS_SUBSTR('Ⅸ', 'IX') AS result;
-- expected: true (NFKC normalization treats Roman numeral Ⅸ as IX)

SELECT CONTAINS_SUBSTR((23, 35, 41), '35') AS result;
-- expected: true (matched inside a STRUCT field)

SELECT CONTAINS_SUBSTR(('abc', ['def', 'ghi', 'jkl'], 'mno'), 'jk') AS result;
-- expected: true (recursive search into a nested ARRAY)

SELECT CONTAINS_SUBSTR((23, NULL, 41), '41') AS result;
-- expected: true (NULL field ignored once a match is found)

SELECT CONTAINS_SUBSTR((23, NULL, 41), '35') AS result;
-- expected: NULL (no match, but a NULL field is present)

-- Search across all columns of a table reference.
SELECT * FROM Recipes WHERE CONTAINS_SUBSTR(Recipes, 'toast');

-- Search across a subset of columns by building an inline STRUCT.
SELECT * FROM Recipes WHERE CONTAINS_SUBSTR((Lunch, Dinner), 'potato');

-- Exclude columns by projecting a STRUCT with EXCEPT.
SELECT *
FROM Recipes
WHERE CONTAINS_SUBSTR(
  (SELECT AS STRUCT Recipes.* EXCEPT (Lunch, Dinner)),
  'potato'
);

-- json_scope controls which part of JSON is searched.
SELECT CONTAINS_SUBSTR(JSON '{"lunch":"soup"}', "lunch") AS result;
-- expected: false (default JSON_VALUES; "lunch" is a key)

SELECT CONTAINS_SUBSTR(JSON '{"lunch":"soup"}', "lunch",
                       json_scope => "JSON_VALUES") AS result;
-- expected: false

SELECT CONTAINS_SUBSTR(JSON '{"lunch":"soup"}', "lunch",
                       json_scope => "JSON_KEYS_AND_VALUES") AS result;
-- expected: true

SELECT CONTAINS_SUBSTR(JSON '{"lunch":"soup"}', "lunch",
                       json_scope => "JSON_KEYS") AS result;
-- expected: true
```

## Edge cases
- A literal `NULL` for `search_value_literal` raises an error
  (`SELECT CONTAINS_SUBSTR('hello', NULL)` throws).
- `search_value_literal` must be a `STRING` literal or `STRING` constant
  expression — non-constant string columns are not allowed.
- A `NULL` `expression` returns `NULL` (does not error).
- In cross-field searches over `STRUCT`/`ARRAY`, a `NULL` field that
  prevents a definitive `FALSE` result yields `NULL`, not `FALSE`.
- `json_scope` is ignored unless `expression` is or contains `JSON`; only
  the three documented values (`'JSON_VALUES'`, `'JSON_KEYS'`,
  `'JSON_KEYS_AND_VALUES'`) are accepted.
- Comparison uses NFKC normalization plus case folding, so
  compatibility-equivalent characters (for example `Ⅸ` vs `IX`,
  full-width vs half-width forms) match.

## Reference (upstream)

See the official BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/string_functions#contains_substr>
