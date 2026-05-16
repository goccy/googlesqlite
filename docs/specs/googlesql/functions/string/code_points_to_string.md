---
name: CODE_POINTS_TO_STRING
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#code_points_to_string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/code_points_to_string.yaml
---

# CODE_POINTS_TO_STRING

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

## `CODE_POINTS_TO_STRING`

```googlesql
CODE_POINTS_TO_STRING(unicode_code_points)
```

**Description**

Takes an array of Unicode [code points][string-link-to-code-points-wikipedia]
as `ARRAY<INT64>` and returns a `STRING`.

To convert from a string to an array of code points, see
[TO_CODE_POINTS][string-link-to-code-points].

**Return type**

`STRING`

**Examples**

The following are basic examples using `CODE_POINTS_TO_STRING`.

```googlesql
SELECT CODE_POINTS_TO_STRING([65, 255, 513, 1024]) AS string;

/*--------+
 | string |
 +--------+
 | AÿȁЀ   |
 +--------*/
```

```googlesql
SELECT CODE_POINTS_TO_STRING([97, 0, 0xF9B5]) AS string;

/*--------+
 | string |
 +--------+
 | a例    |
 +--------*/
```

```googlesql
SELECT CODE_POINTS_TO_STRING([65, 255, NULL, 1024]) AS string;

/*--------+
 | string |
 +--------+
 | NULL   |
 +--------*/
```

The following example computes the frequency of letters in a set of words.

```googlesql
WITH Words AS (
  SELECT word
  FROM UNNEST(['foo', 'bar', 'baz', 'giraffe', 'llama']) AS word
)
SELECT
  CODE_POINTS_TO_STRING([code_point]) AS letter,
  COUNT(*) AS letter_count
FROM Words,
  UNNEST(TO_CODE_POINTS(word)) AS code_point
GROUP BY 1
ORDER BY 2 DESC;

/*--------+--------------+
 | letter | letter_count |
 +--------+--------------+
 | a      | 5            |
 | f      | 3            |
 | r      | 2            |
 | b      | 2            |
 | l      | 2            |
 | o      | 2            |
 | g      | 1            |
 | z      | 1            |
 | e      | 1            |
 | m      | 1            |
 | i      | 1            |
 +--------+--------------*/
```

[string-link-to-code-points-wikipedia]: https://en.wikipedia.org/wiki/Code_point

[string-link-to-code-points]: #to_code_points

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
