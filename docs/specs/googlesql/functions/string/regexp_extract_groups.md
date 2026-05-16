---
name: REGEXP_EXTRACT_GROUPS
dialect: googlesql
category: functions/string
status: implemented
notes: |
  Lambda-style or grouped regex outputs that need a planner extension (REGEXP_EXTRACT_GROUPS returns ARRAY<STRUCT>, SPLIT_SUBSTR is variadic). Deferred.
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_extract_groups
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_extract_groups.yaml
---

# REGEXP_EXTRACT_GROUPS

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

## `REGEXP_EXTRACT_GROUPS`

```googlesql
REGEXP_EXTRACT_GROUPS(value, regexp)
```

**Description**

Returns a `STRUCT` where each field contains a substring from `value` that
matches a capturing group in the [re2 regular expression][string-link-to-re2],
`regexp`. The function returns the substrings from the first place in `value`
where the *entire* `regexp` pattern matches.

**Details**

This function is similar to [`REGEXP_EXTRACT`][regexp-extract], but it returns a
`STRUCT` with a field for each capturing group in the `regexp`.

The `regexp` must contain at least one capturing group. The fields in the
returned `STRUCT` correspond to these capturing groups:

+   If a capturing group is named (for example, `(?<name>...)` or `(?P<name>...)`),
    the corresponding `STRUCT` field will have that name. Both syntaxes are
    equivalent.
+   If a capturing group is unnamed, the corresponding `STRUCT` field is
    anonymous. These fields can be accessed by their 0-based position in the
    `STRUCT`.
+   The order of fields in the `STRUCT` matches the order of the capturing
    groups in `regexp` from left to right.

Returns `NULL` if `value` is `NULL` or if the overall `regexp` pattern doesn't
match at all. If a specific capturing group doesn't match (for example, if it's
part of an alternation or is optional), the corresponding `STRUCT` field is
`NULL`.

Returns an error if:

+ The `regexp` is invalid.
+ The `regexp` is not a string literal.
+ The `regexp` has no capturing groups.
+ A capturing group name is not a valid `STRUCT` field name (for example, starts
  with a digit or contains spaces). Valid names consist of letters, numbers,
  and underscores, and must start with a letter or underscore.
+ The same capturing group name is used more than once (case-insensitive).

**Return type**

`STRUCT<...>`

The fields of the `STRUCT` are generally `STRING` (or `BYTES` if the inputs are
`BYTES`). However, fields can be [auto-casted](#auto_casting) to other types.

**Examples**

Extract unnamed groups:

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('abc123xyz', r'([a-z]+)([0-9]+)([a-z]+)') AS result

/*---------------------------------+
 | result                          |
 +---------------------------------+
 | {abc, 123, xyz}                 |
 +---------------------------------*/
```

Extract named groups:

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('2025-09-10', r'(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})') AS result

/*----------------------------------------------+
 | result                                       |
 +----------------------------------------------+
 | {2025 year, 09 month, 10 day}                |
 +----------------------------------------------*/
```

**Expand STRUCT fields into columns**

Because `REGEXP_EXTRACT_GROUPS` returns a `STRUCT`, you can use the `.*` operator
in the `SELECT` list to expand the fields of the `STRUCT` into separate columns.
Expanding `STRUCT` fields into columns is particularly useful when all capturing
groups are named.

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('PROD-WIDGET-1234', r'(?<env>\w+)-(?<product>\w+)-(?<id>\d+)').*

/*-------+-----------+------+
 | env   | product   | id   |
 +-------+-----------+------+
 | PROD  | WIDGET    | 1234 |
 +-------+-----------+------*/
```

Mix of named and unnamed groups:

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('id:123', r'(?<key>[a-z]+):([0-9]+)') AS result

/*-----------------------+
 | result                |
 +-----------------------+
 | {id key, 123}         |
 +-----------------------*/
```

No match returns `NULL`:

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('abc', r'(\d+)') AS result

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

Optional groups and empty matches:

```googlesql
WITH inputs AS (
  SELECT 'id:123:extra' AS t UNION ALL
  SELECT 'id:123:' AS t UNION ALL
  SELECT 'id:123' AS t
)
SELECT
  t,
  REGEXP_EXTRACT_GROUPS(t, r'(?<key>\w+):(?<val>\w+)(?::(?<opt>\w*))?') AS result
FROM inputs;

/*-----------------+--------------------------------------+
 | t               | result                               |
 +-----------------+--------------------------------------+
 | id:123:extra    | {id key, 123 val, extra opt}         |
 | id:123:         | {id key, 123 val,  opt}              |
 | id:123          | {id key, 123 val, NULL opt}          |
 +-----------------+--------------------------------------*/
```

Note that in the second row, the optional group `opt` matches an empty string,
which is different from the third row where the group doesn't match at all and
results in `NULL`.

Nested groups:

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('a=b=c', r'(\w+)=((\w+)=\w+)') AS result

/*-----------------------+
 | result                |
 +-----------------------+
 | {a, b=c, b}           |
 +-----------------------*/
```

Alternation with different groups:

```googlesql
WITH inputs AS (
  SELECT 'config_id=123' AS t UNION ALL
  SELECT 'option_name=ABC' AS t
)
SELECT
  t,
  REGEXP_EXTRACT_GROUPS(t, r'config_id=(?<id>\d+)|option_name=(?<name>\w+)') AS result
FROM inputs;

/*-----------------+--------------------------+
 | t               | result                   |
 +-----------------+--------------------------+
 | config_id=123   | {123 id, NULL name}      |
 | option_name=ABC | {NULL id, ABC name}      |
 +-----------------+--------------------------*/
```

The `STRUCT` result contains fields for all named capturing groups across all
alternatives in the regular expression. In each row, only the fields
corresponding to the alternative that matched are populated. Other fields are
`NULL`.

##### Auto-casting 
<a id="auto_casting"></a>

You can automatically cast the captured substring to a specific type by
suffixing the capturing group name with a double underscore (`__`) followed by
the type name.

Any type that can be cast from `STRING` (or `BYTES` for the `BYTES` version
of the function) is supported. Type names are case-insensitive.

The field name in the resulting `STRUCT` will have the `__TYPE` suffix removed.

If the captured substring can't be cast to the specified type, an error is
returned. This includes casting an empty string to a numeric or boolean type.
If the captured substring is `NULL` (due to an optional group not matching), the
cast result is also `NULL`.

**Examples of auto-casting**

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('val=0x1a', r'val=(?<val__INT64>0x[0-9a-fA-F]+)') AS result

/*-------------+
 | result      |
 +-------------+
 | {26 val}    |
 +-------------*/
```

Auto-casted values in expressions with Pipe syntax:

```googlesql
FROM UNNEST(['02:30:10', '01:02:03']) AS time_str
|> EXTEND REGEXP_EXTRACT_GROUPS(time_str, r'(?<h__INT64>\d{2}):(?<m__INT64>\d{2}):(?<s__INT64>\d{2})').*
|> SELECT time_str, h * 3600 + m * 60 + s AS total_seconds

/*----------+---------------+
 | time_str | total_seconds |
 +----------+---------------+
 | 02:30:10 | 9010          |
 | 01:02:03 | 3723          |
 +----------+---------------*/
```

Expand auto-casted fields into columns:

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('2025-09-10', r'(?<year__INT64>\d{4})-(?<month__INT64>\d{2})-(?<day__INT64>\d{2})').*

/*--------+---------+-------+
 | year   | month   | day   |
 +--------+---------+-------+
 | 2025   | 9       | 10    |
 +--------+---------+-------*/
```

Cast failure:

```googlesql {.bad}
-- Error: Bad INT64 value
SELECT REGEXP_EXTRACT_GROUPS('ID: ABC', r'ID: (?<item_id__INT64>\w+)')
```

Cast failure with empty string:

```googlesql {.bad}
-- Error: Bad INT64 value
SELECT REGEXP_EXTRACT_GROUPS('ID: ', r'ID: (?<item_id__INT64>\d*)')
```

Workaround for empty string cast failure by making the group optional:

```googlesql
SELECT REGEXP_EXTRACT_GROUPS('ID: ', r'ID: (?<item_id__INT64>\d+)?') AS result

/*-----------------+
 | result          |
 +-----------------+
 | {NULL item_id}  |
 +-----------------*/
```

[string-link-to-re2]: https://github.com/google/re2/wiki/Syntax

[regexp-extract]: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_extract

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
