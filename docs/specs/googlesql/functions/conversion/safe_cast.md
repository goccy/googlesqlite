---
name: SAFE_CAST
dialect: googlesql
category: functions/conversion
status: implemented
source_url: docs/third_party/googlesql-docs/conversion_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/conversion_functions.md#safe_cast
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/conversion/safe_cast.yaml
---

# SAFE_CAST

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

Verbatim copy from `docs/third_party/googlesql-docs/conversion_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `SAFE_CAST` 
<a id="safe_casting"></a>

<pre class="lang-sql prettyprint">
<code>SAFE_CAST(expression AS typename [format_clause])</code>
</pre>

**Description**

When using `CAST`, a query can fail if GoogleSQL is unable to perform
the cast. For example, the following query generates an error:

```googlesql
SELECT CAST("apple" AS INT64) AS not_a_number;
```

If you want to protect your queries from these types of errors, you can use
`SAFE_CAST`. `SAFE_CAST` replaces runtime errors with `NULL`s. However, during
static analysis, impossible casts between two non-castable types still produce
an error because the query is invalid.

```googlesql
SELECT SAFE_CAST("apple" AS INT64) AS not_a_number;

/*--------------+
 | not_a_number |
 +--------------+
 | NULL         |
 +--------------*/
```

Some casts can include a [format clause][formatting-syntax], which provides
instructions for how to conduct the
cast. For example, you could
instruct a cast to convert a sequence of bytes to a BASE64-encoded string
instead of a UTF-8-encoded string.

The structure of the format clause is unique to each type of cast and more
information is available in the section for that cast.

If you are casting from bytes to strings, you can also use the
function, [`SAFE_CONVERT_BYTES_TO_STRING`][SC_BTS]. Any invalid UTF-8 characters
are replaced with the unicode replacement character, `U+FFFD`.

[SC_BTS]: https://github.com/google/googlesql/blob/master/docs/string_functions.md#safe_convert_bytes_to_string

[formatting-syntax]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#formatting_syntax

[conversion-rules]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/conversion_functions.md`.
