---
name: INITCAP
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#initcap
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/initcap.yaml
---

# INITCAP

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

## `INITCAP`

```googlesql
INITCAP(value[, delimiters])
```

**Description**

Takes a `STRING` and returns it with the first character in each word in
uppercase and all other characters in lowercase. Non-alphabetic characters
remain the same.

`delimiters` is an optional string argument that's used to override the default
set of characters used to separate words. If `delimiters` isn't specified, it
defaults to the following characters: \
`<whitespace> [ ] ( ) { } / | \ < > ! ? @ " ^ # $ & ~ _ , . : ; * % + -`

If `value` or `delimiters` is `NULL`, the function returns `NULL`.

**Return type**

`STRING`

**Examples**

```googlesql
SELECT
  'Hello World-everyone!' AS value,
  INITCAP('Hello World-everyone!') AS initcap_value

/*-------------------------------+-------------------------------+
 | value                         | initcap_value                 |
 +-------------------------------+-------------------------------+
 | Hello World-everyone!         | Hello World-Everyone!         |
 +-------------------------------+-------------------------------*/
```

```googlesql
SELECT
  'Apples1oranges2pears' as value,
  '12' AS delimiters,
  INITCAP('Apples1oranges2pears' , '12') AS initcap_value

/*----------------------+------------+----------------------+
 | value                | delimiters | initcap_value        |
 +----------------------+------------+----------------------+
 | Apples1oranges2pears | 12         | Apples1Oranges2Pears |
 +----------------------+------------+----------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
