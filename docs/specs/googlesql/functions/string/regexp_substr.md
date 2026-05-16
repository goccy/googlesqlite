---
name: REGEXP_SUBSTR
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_substr
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_substr.yaml
---

# REGEXP_SUBSTR

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

## `REGEXP_SUBSTR`

```googlesql
REGEXP_SUBSTR(value, regexp[, position[, occurrence]])
```

**Description**

Synonym for [REGEXP_EXTRACT][string-link-to-regex].

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
WITH example AS
(SELECT 'Hello World Helloo' AS value, 'H?ello+' AS regex, 1 AS position, 1 AS
occurrence
)
SELECT value, regex, position, occurrence, REGEXP_SUBSTR(value, regex,
position, occurrence) AS regexp_value FROM example;

/*--------------------+---------+----------+------------+--------------+
 | value              | regex   | position | occurrence | regexp_value |
 +--------------------+---------+----------+------------+--------------+
 | Hello World Helloo | H?ello+ | 1        | 1          | Hello        |
 +--------------------+---------+----------+------------+--------------*/
```

[string-link-to-regex]: #regexp_extract

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
