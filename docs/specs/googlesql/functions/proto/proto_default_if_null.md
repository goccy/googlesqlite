---
name: PROTO_DEFAULT_IF_NULL
dialect: googlesql
category: functions/proto
status: partial
notes: |
  Runtime entry (`googlesqlite_proto_default_if_null` → proto-defined
  default bytes) is in place; the gap is at analysis time. The
  upstream GoogleSQL analyzer rejects every field whose `FieldOptions`
  does not carry a typed `(googlesql.use_defaults) = true` extension.
  Verified against go-googlesql v0.2.1 and v0.2.2 — both produce
  `Field <name> is annotated to ignore proto defaults` for the
  auto-registered `google.type.Date` and for synthesised proto2 `Book`
  fixtures with `[default = ...]` set. The check is independent of
  `FEATURE_IGNORE_PROTO3_USE_DEFAULTS` (disabling it does not relax
  the rule) and is not satisfied by raw extension wire bytes appended
  to FieldOptions, even after registering a synthesised
  `googlesql/public/proto/type_annotation.proto` in the catalog's
  DescriptorPool — the wasm `Reflection->HasExtension` lookup does
  not elevate unknown bytes to typed extensions post hoc. Closing
  this spec needs either a Go binding for the extension that ships
  with go-googlesql so the typed bytes are emitted from our side, or
  an upstream wasm change that elevates extension unknown_fields when
  the declaring file is registered.
source_url: docs/third_party/googlesql-docs/protocol_buffer_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/protocol_buffer_functions.md#proto_default_if_null
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/proto/proto_default_if_null.yaml
---

# PROTO_DEFAULT_IF_NULL

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

Verbatim copy from `docs/third_party/googlesql-docs/protocol_buffer_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PROTO_DEFAULT_IF_NULL`

```googlesql
PROTO_DEFAULT_IF_NULL(proto_field_expression)
```

**Description**

Evaluates any expression that results in a proto field access.
If the `proto_field_expression` evaluates to `NULL`, returns the default
value for the field. Otherwise, returns the field value.

Stipulations:

+ The expression can't resolve to a required field.
+ The expression can't resolve to a message field.
+ The expression must resolve to a regular proto field access, not
  a virtual field.
+ The expression can't access a field with
  `googlesql.use_defaults=false`.

**Return Type**

Type of `proto_field_expression`.

**Example**

In the following example, each book in a library has a country of origin. If
the country isn't set, the country defaults to unknown.

In this statement, table `library_books` contains a column named `book`,
whose type is `Book`.

```googlesql
SELECT PROTO_DEFAULT_IF_NULL(book.country) AS origin FROM library_books;
```

`Book` is a type that contains a field called `country`.

```proto
message Book {
  optional string country = 4 [default = 'Unknown'];
}
```

This is the result if `book.country` evaluates to `Canada`.

```googlesql
/*-----------------+
 | origin          |
 +-----------------+
 | Canada          |
 +-----------------*/
```

This is the result if `book` is `NULL`. Since `book` is `NULL`,
`book.country` evaluates to `NULL` and therefore the function result is the
default value for `country`.

```googlesql
/*-----------------+
 | origin          |
 +-----------------+
 | Unknown         |
 +-----------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/protocol_buffer_functions.md`.
