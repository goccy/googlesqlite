---
name: NORMALIZE
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#normalize
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/normalize.yaml
---

# NORMALIZE

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

## `NORMALIZE`

```googlesql
NORMALIZE(value[, normalization_mode])
```

**Description**

Takes a string value and returns it as a normalized string. If you don't
provide a normalization mode, `NFC` is used.

[Normalization][string-link-to-normalization-wikipedia] is used to ensure that
two strings are equivalent. Normalization is often used in situations in which
two strings render the same on the screen but have different Unicode code
points.

`NORMALIZE` supports four optional normalization modes:

| Value   | Name                                           | Description|
|---------|------------------------------------------------|------------|
| `NFC`   | Normalization Form Canonical Composition       | Decomposes and recomposes characters by canonical equivalence.|
| `NFKC`  | Normalization Form Compatibility Composition   | Decomposes characters by compatibility, then recomposes them by canonical equivalence.|
| `NFD`   | Normalization Form Canonical Decomposition     | Decomposes characters by canonical equivalence, and multiple combining characters are arranged in a specific order.|
| `NFKD`  | Normalization Form Compatibility Decomposition | Decomposes characters by compatibility, and multiple combining characters are arranged in a specific order.|

**Return type**

`STRING`

**Examples**

The following example normalizes different language characters:

```googlesql
SELECT
  NORMALIZE('\u00ea') as a,
  NORMALIZE('\u0065\u0302') as b,
  NORMALIZE('\u00ea') = NORMALIZE('\u0065\u0302') as normalized;

/*---+---+------------+
 | a | b | normalized |
 +---+---+------------+
 | ê | ê | TRUE       |
 +---+---+------------*/
```
The following examples normalize different space characters:

```googlesql
SELECT NORMALIZE('Raha\u2004Mahan', NFKC) AS normalized_name

/*-----------------+
 | normalized_name |
 +-----------------+
 | Raha Mahan      |
 +-----------------*/
```

```googlesql
SELECT NORMALIZE('Raha\u2005Mahan', NFKC) AS normalized_name

/*-----------------+
 | normalized_name |
 +-----------------+
 | Raha Mahan      |
 +-----------------*/
```

```googlesql
SELECT NORMALIZE('Raha\u2006Mahan', NFKC) AS normalized_name

/*-----------------+
 | normalized_name |
 +-----------------+
 | Raha Mahan      |
 +-----------------*/
```

```googlesql
SELECT NORMALIZE('Raha Mahan', NFKC) AS normalized_name

/*-----------------+
 | normalized_name |
 +-----------------+
 | Raha Mahan      |
 +-----------------*/
```

[string-link-to-normalization-wikipedia]: https://en.wikipedia.org/wiki/Unicode_equivalence#Normalization

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
