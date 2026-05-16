---
name: CODE_POINTS_TO_BYTES
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#code_points_to_bytes
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/code_points_to_bytes.yaml
---

# CODE_POINTS_TO_BYTES

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

## `CODE_POINTS_TO_BYTES`

```googlesql
CODE_POINTS_TO_BYTES(ascii_code_points)
```

**Description**

Takes an array of extended ASCII
[code points][string-link-to-code-points-wikipedia]
as `ARRAY<INT64>` and returns `BYTES`.

To convert from `BYTES` to an array of code points, see
[TO_CODE_POINTS][string-link-to-code-points].

**Return type**

`BYTES`

**Examples**

The following is a basic example using `CODE_POINTS_TO_BYTES`.

```googlesql
SELECT CODE_POINTS_TO_BYTES([65, 98, 67, 100]) AS bytes;

/*----------+
 | bytes    |
 +----------+
 | AbCd     |
 +----------*/
```

The following example uses a rotate-by-13 places (ROT13) algorithm to encode a
string.

```googlesql
SELECT CODE_POINTS_TO_BYTES(ARRAY_AGG(
  (SELECT
      CASE
        WHEN chr BETWEEN b'a' and b'z'
          THEN TO_CODE_POINTS(b'a')[offset(0)] +
            MOD(code+13-TO_CODE_POINTS(b'a')[offset(0)],26)
        WHEN chr BETWEEN b'A' and b'Z'
          THEN TO_CODE_POINTS(b'A')[offset(0)] +
            MOD(code+13-TO_CODE_POINTS(b'A')[offset(0)],26)
        ELSE code
      END
   FROM
     (SELECT code, CODE_POINTS_TO_BYTES([code]) chr)
  ) ORDER BY OFFSET)) AS encoded_string
FROM UNNEST(TO_CODE_POINTS(b'Test String!')) code WITH OFFSET;

/*------------------+
 | encoded_string   |
 +------------------+
 | Grfg Fgevat!     |
 +------------------*/
```

[string-link-to-code-points-wikipedia]: https://en.wikipedia.org/wiki/Code_point

[string-link-to-code-points]: #to_code_points

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
