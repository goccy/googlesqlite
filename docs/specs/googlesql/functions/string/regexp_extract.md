---
name: REGEXP_EXTRACT
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_extract
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_extract.yaml
---

# REGEXP_EXTRACT

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

## `REGEXP_EXTRACT`

```googlesql
REGEXP_EXTRACT(value, regexp[, position[, occurrence]])
```

**Description**

Returns the substring in `value` that matches the
[re2 regular expression][string-link-to-re2], `regexp`.
Returns `NULL` if there is no match.

If the regular expression contains a capturing group (`(...)`), and there is a
match for that capturing group, that match is returned. If there
are multiple matches for a capturing group, the first match is returned.

To extract matches for multiple capturing groups in a single call, use
[`REGEXP_EXTRACT_GROUPS`][regexp-extract-groups].

If `position` is specified, the search starts at this
position in `value`, otherwise it starts at the beginning of `value`. The
`position` must be a positive integer and can't be 0. If `position` is greater
than the length of `value`, `NULL` is returned.

If `occurrence` is specified, the search returns a specific occurrence of the
`regexp` in `value`, otherwise returns the first match. If `occurrence` is
greater than the number of matches found, `NULL` is returned. For
`occurrence` > 1, the function searches for additional occurrences beginning
with the character following the previous occurrence.

Returns an error if:

+ The regular expression is invalid
+ The regular expression has more than one capturing group
+ The `position` isn't a positive integer
+ The `occurrence` isn't a positive integer

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT REGEXP_EXTRACT('foo@example.com', r'^[a-zA-Z0-9_.+-]+') AS user_name

/*-----------+
 | user_name |
 +-----------+
 | foo       |
 +-----------*/
```

```googlesql
SELECT REGEXP_EXTRACT('foo@example.com', r'^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.([a-zA-Z0-9-.]+$)')

/*------------------+
 | top_level_domain |
 +------------------+
 | com              |
 +------------------*/
```

```googlesql
SELECT
  REGEXP_EXTRACT('ab', '.b') AS result_a,
  REGEXP_EXTRACT('ab', '(.)b') AS result_b,
  REGEXP_EXTRACT('xyztb', '(.)+b') AS result_c,
  REGEXP_EXTRACT('ab', '(z)?b') AS result_d

/*-------------------------------------------+
 | result_a | result_b | result_c | result_d |
 +-------------------------------------------+
 | ab       | a        | t        | NULL     |
 +-------------------------------------------*/
```

```googlesql
WITH example AS
(SELECT 'Hello Helloo and Hellooo' AS value, 'H?ello+' AS regex, 1 as position,
1 AS occurrence UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 1, 2 UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 1, 3 UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 1, 4 UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 2, 1 UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 3, 1 UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 3, 2 UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 3, 3 UNION ALL
SELECT 'Hello Helloo and Hellooo', 'H?ello+', 20, 1 UNION ALL
SELECT 'cats&dogs&rabbits' ,'\\w+&', 1, 2 UNION ALL
SELECT 'cats&dogs&rabbits', '\\w+&', 2, 3
)
SELECT value, regex, position, occurrence, REGEXP_EXTRACT(value, regex,
position, occurrence) AS regexp_value FROM example;

/*--------------------------+---------+----------+------------+--------------+
 | value                    | regex   | position | occurrence | regexp_value |
 +--------------------------+---------+----------+------------+--------------+
 | Hello Helloo and Hellooo | H?ello+ | 1        | 1          | Hello        |
 | Hello Helloo and Hellooo | H?ello+ | 1        | 2          | Helloo       |
 | Hello Helloo and Hellooo | H?ello+ | 1        | 3          | Hellooo      |
 | Hello Helloo and Hellooo | H?ello+ | 1        | 4          | NULL         |
 | Hello Helloo and Hellooo | H?ello+ | 2        | 1          | ello         |
 | Hello Helloo and Hellooo | H?ello+ | 3        | 1          | Helloo       |
 | Hello Helloo and Hellooo | H?ello+ | 3        | 2          | Hellooo      |
 | Hello Helloo and Hellooo | H?ello+ | 3        | 3          | NULL         |
 | Hello Helloo and Hellooo | H?ello+ | 20       | 1          | NULL         |
 | cats&dogs&rabbits        | \w+&    | 1        | 2          | dogs&        |
 | cats&dogs&rabbits        | \w+&    | 2        | 3          | NULL         |
 +--------------------------+---------+----------+------------+--------------*/
```

[string-link-to-re2]: https://github.com/google/re2/wiki/Syntax

[regexp-extract-groups]: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_extract_groups

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
