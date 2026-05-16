---
name: CONCAT
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#concat
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/concat.yaml
---

# CONCAT

## Summary

(TBD — refine from the upstream reference below.)

## Signatures

(TBD)

## Behavior

(TBD)

## Examples

(TBD)

## Edge cases

(TBD)

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `CONCAT`

```googlesql
CONCAT(value1[, ...])
```

**Description**

Concatenates one or more values into a single result. All values must be
`BYTES` or data types that can be cast to `STRING`.

The function returns `NULL` if any input argument is `NULL`.

Note: You can also use the
[|| concatenation operator][string-link-to-operators] to concatenate
values into a string.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT CONCAT('T.P.', ' ', 'Bar') as author;

/*---------------------+
 | author              |
 +---------------------+
 | T.P. Bar            |
 +---------------------*/
```

```googlesql
SELECT CONCAT('Summer', ' ', 1923) as release_date;

/*---------------------+
 | release_date        |
 +---------------------+
 | Summer 1923         |
 +---------------------*/
```

```googlesql

With Employees AS
  (SELECT
    'John' AS first_name,
    'Doe' AS last_name
  UNION ALL
  SELECT
    'Jane' AS first_name,
    'Smith' AS last_name
  UNION ALL
  SELECT
    'Joe' AS first_name,
    'Jackson' AS last_name)

SELECT
  CONCAT(first_name, ' ', last_name)
  AS full_name
FROM Employees;

/*---------------------+
 | full_name           |
 +---------------------+
 | John Doe            |
 | Jane Smith          |
 | Joe Jackson         |
 +---------------------*/
```

[string-link-to-operators]: https://github.com/google/googlesql/blob/master/docs/operators.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
