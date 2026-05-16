---
name: NORMALIZE_AND_CASEFOLD
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#normalize_and_casefold
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/normalize_and_casefold.yaml
---

# NORMALIZE_AND_CASEFOLD

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

## `NORMALIZE_AND_CASEFOLD`

```googlesql
NORMALIZE_AND_CASEFOLD(value[, normalization_mode])
```

**Description**

Takes a string value and returns it as a normalized string. If you don't
provide a normalization mode, `NFC` is used.

[Normalization][string-link-to-normalization-wikipedia] is used to ensure that
two strings are equivalent. Normalization is often used in situations in which
two strings render the same on the screen but have different Unicode code
points.

[Case folding][string-link-to-case-folding-wikipedia] is used for the caseless
comparison of strings. If you need to compare strings and case shouldn't be
considered, use `NORMALIZE_AND_CASEFOLD`, otherwise use
[`NORMALIZE`][string-link-to-normalize].

`NORMALIZE_AND_CASEFOLD` supports four optional normalization modes:

| Value   | Name                                           | Description|
|---------|------------------------------------------------|------------|
| `NFC`   | Normalization Form Canonical Composition       | Decomposes and recomposes characters by canonical equivalence.|
| `NFKC`  | Normalization Form Compatibility Composition   | Decomposes characters by compatibility, then recomposes them by canonical equivalence.|
| `NFD`   | Normalization Form Canonical Decomposition     | Decomposes characters by canonical equivalence, and multiple combining characters are arranged in a specific order.|
| `NFKD`  | Normalization Form Compatibility Decomposition | Decomposes characters by compatibility, and multiple combining characters are arranged in a specific order.|

**Return type**

`STRING`

**Examples**

```googlesql
SELECT
  NORMALIZE('The red barn') = NORMALIZE('The Red Barn') AS normalized,
  NORMALIZE_AND_CASEFOLD('The red barn')
    = NORMALIZE_AND_CASEFOLD('The Red Barn') AS normalized_with_case_folding;

/*------------+------------------------------+
 | normalized | normalized_with_case_folding |
 +------------+------------------------------+
 | FALSE      | TRUE                         |
 +------------+------------------------------*/
```

```googlesql
SELECT
  '\u2168' AS a,
  'IX' AS b,
  NORMALIZE_AND_CASEFOLD('\u2168', NFD)=NORMALIZE_AND_CASEFOLD('IX', NFD) AS nfd,
  NORMALIZE_AND_CASEFOLD('\u2168', NFC)=NORMALIZE_AND_CASEFOLD('IX', NFC) AS nfc,
  NORMALIZE_AND_CASEFOLD('\u2168', NFKD)=NORMALIZE_AND_CASEFOLD('IX', NFKD) AS nfkd,
  NORMALIZE_AND_CASEFOLD('\u2168', NFKC)=NORMALIZE_AND_CASEFOLD('IX', NFKC) AS nfkc;

/*---+----+-------+-------+------+------+
 | a | b  | nfd   | nfc   | nfkd | nfkc |
 +---+----+-------+-------+------+------+
 | Ⅸ | IX | false | false | true | true |
 +---+----+-------+-------+------+------*/
```

```googlesql
SELECT
  '\u0041\u030A' AS a,
  '\u00C5' AS b,
  NORMALIZE_AND_CASEFOLD('\u0041\u030A', NFD)=NORMALIZE_AND_CASEFOLD('\u00C5', NFD) AS nfd,
  NORMALIZE_AND_CASEFOLD('\u0041\u030A', NFC)=NORMALIZE_AND_CASEFOLD('\u00C5', NFC) AS nfc,
  NORMALIZE_AND_CASEFOLD('\u0041\u030A', NFKD)=NORMALIZE_AND_CASEFOLD('\u00C5', NFKD) AS nkfd,
  NORMALIZE_AND_CASEFOLD('\u0041\u030A', NFKC)=NORMALIZE_AND_CASEFOLD('\u00C5', NFKC) AS nkfc;

/*---+----+-------+-------+------+------+
 | a | b  | nfd   | nfc   | nkfd | nkfc |
 +---+----+-------+-------+------+------+
 | Å | Å  | true  | true  | true | true |
 +---+----+-------+-------+------+------*/
```

[string-link-to-normalization-wikipedia]: https://en.wikipedia.org/wiki/Unicode_equivalence#Normalization

[string-link-to-case-folding-wikipedia]: https://en.wikipedia.org/wiki/Letter_case#Case_folding

[string-link-to-normalize]: #normalize

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
