---
name: REGEXP_MATCH
dialect: googlesql
category: functions/string
status: implemented
notes: |
  Lambda-style or grouped regex outputs that need a planner extension (REGEXP_EXTRACT_GROUPS returns ARRAY<STRUCT>, SPLIT_SUBSTR is variadic). Deferred.
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_match-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_match.yaml
---

# REGEXP_MATCH

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

## `REGEXP_MATCH` (Deprecated) 
<a id="regexp_match"></a>

```googlesql
REGEXP_MATCH(value, regexp)
```

**Description**

Returns `TRUE` if `value` is a full match for the regular expression, `regexp`.

If the `regexp` argument is invalid, the function returns an error.

This function is deprecated. When possible, use
[`REGEXP_CONTAINS`][regexp-contains] to find a partial match for a
regular expression.

Note: GoogleSQL provides regular expression support using the
[re2][string-link-to-re2] library; see that documentation for its
regular expression syntax.

**Return type**

`BOOL`

**Examples**

```googlesql
WITH email_addresses AS
  (SELECT 'foo@example.com' as email
  UNION ALL
  SELECT 'bar@example.org' as email
  UNION ALL
  SELECT 'notavalidemailaddress' as email)

SELECT
  email,
  REGEXP_MATCH(email,
               r'[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+')
               AS valid_email_address
FROM email_addresses;

/*-----------------------+---------------------+
 | email                 | valid_email_address |
 +-----------------------+---------------------+
 | foo@example.com       | true                |
 | bar@example.org       | true                |
 | notavalidemailaddress | false               |
 +-----------------------+---------------------*/
```

[string-link-to-re2]: https://github.com/google/re2/wiki/Syntax

[regexp-contains]: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_contains

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
