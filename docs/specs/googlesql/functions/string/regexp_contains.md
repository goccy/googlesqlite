---
name: REGEXP_CONTAINS
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_contains
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_contains.yaml
---

# REGEXP_CONTAINS

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

## `REGEXP_CONTAINS`

```googlesql
REGEXP_CONTAINS(value, regexp)
```

**Description**

Returns `TRUE` if `value` is a partial match for the regular expression,
`regexp`.

If the `regexp` argument is invalid, the function returns an error.

You can search for a full match by using `^` (beginning of text) and `$` (end of
text). Due to regular expression operator precedence, it's good practice to use
parentheses around everything between `^` and `$`.

Note: GoogleSQL provides regular expression support using the
[re2][string-link-to-re2] library; see that documentation for its
regular expression syntax.

**Return type**

`BOOL`

**Examples**

The following queries check to see if an email is valid:

```googlesql
SELECT
  'foo@example.com' AS email,
  REGEXP_CONTAINS('foo@example.com', r'@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+') AS is_valid

/*-----------------+----------+
 | email           | is_valid |
 +-----------------+----------+
 | foo@example.com | TRUE     |
 +-----------------+----------*/
 ```

 ```googlesql
SELECT
  'www.example.net' AS email,
  REGEXP_CONTAINS('www.example.net', r'@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+') AS is_valid

/*-----------------+----------+
 | email           | is_valid |
 +-----------------+----------+
 | www.example.net | FALSE    |
 +-----------------+----------*/
 ```

The following queries check to see if an email is valid. They
perform a full match, using `^` and `$`. Due to regular expression operator
precedence, it's good practice to use parentheses around everything between `^`
and `$`.

```googlesql
SELECT
  'a@foo.com' AS email,
  REGEXP_CONTAINS('a@foo.com', r'^([\w.+-]+@foo\.com|[\w.+-]+@bar\.org)$') AS valid_email_address,
  REGEXP_CONTAINS('a@foo.com', r'^[\w.+-]+@foo\.com|[\w.+-]+@bar\.org$') AS without_parentheses;

/*----------------+---------------------+---------------------+
 | email          | valid_email_address | without_parentheses |
 +----------------+---------------------+---------------------+
 | a@foo.com      | true                | true                |
 +----------------+---------------------+---------------------*/
```

```googlesql
SELECT
  'a@foo.computer' AS email,
  REGEXP_CONTAINS('a@foo.computer', r'^([\w.+-]+@foo\.com|[\w.+-]+@bar\.org)$') AS valid_email_address,
  REGEXP_CONTAINS('a@foo.computer', r'^[\w.+-]+@foo\.com|[\w.+-]+@bar\.org$') AS without_parentheses;

/*----------------+---------------------+---------------------+
 | email          | valid_email_address | without_parentheses |
 +----------------+---------------------+---------------------+
 | a@foo.computer | false               | true                |
 +----------------+---------------------+---------------------*/
```

```googlesql
SELECT
  'b@bar.org' AS email,
  REGEXP_CONTAINS('b@bar.org', r'^([\w.+-]+@foo\.com|[\w.+-]+@bar\.org)$') AS valid_email_address,
  REGEXP_CONTAINS('b@bar.org', r'^[\w.+-]+@foo\.com|[\w.+-]+@bar\.org$') AS without_parentheses;

/*----------------+---------------------+---------------------+
 | email          | valid_email_address | without_parentheses |
 +----------------+---------------------+---------------------+
 | b@bar.org      | true                | true                |
 +----------------+---------------------+---------------------*/
```

```googlesql
SELECT
  '!b@bar.org' AS email,
  REGEXP_CONTAINS('!b@bar.org', r'^([\w.+-]+@foo\.com|[\w.+-]+@bar\.org)$') AS valid_email_address,
  REGEXP_CONTAINS('!b@bar.org', r'^[\w.+-]+@foo\.com|[\w.+-]+@bar\.org$') AS without_parentheses;

/*----------------+---------------------+---------------------+
 | email          | valid_email_address | without_parentheses |
 +----------------+---------------------+---------------------+
 | !b@bar.org     | false               | true                |
 +----------------+---------------------+---------------------*/
```

```googlesql
SELECT
  'c@buz.net' AS email,
  REGEXP_CONTAINS('c@buz.net', r'^([\w.+-]+@foo\.com|[\w.+-]+@bar\.org)$') AS valid_email_address,
  REGEXP_CONTAINS('c@buz.net', r'^[\w.+-]+@foo\.com|[\w.+-]+@bar\.org$') AS without_parentheses;

/*----------------+---------------------+---------------------+
 | email          | valid_email_address | without_parentheses |
 +----------------+---------------------+---------------------+
 | c@buz.net      | false               | false               |
 +----------------+---------------------+---------------------*/
```

[string-link-to-re2]: https://github.com/google/re2/wiki/Syntax

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
