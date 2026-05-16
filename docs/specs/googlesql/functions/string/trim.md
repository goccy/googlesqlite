---
name: TRIM
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#trim
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/trim.yaml
---

# TRIM

## Summary

Removes leading and trailing characters or bytes from a `STRING` or `BYTES`
value. By default, whitespace is stripped from string inputs; an optional
set specifies which code points (or bytes) to remove.

## Signatures

- `TRIM(value_to_trim[, set_of_characters_to_remove])`

## Behavior

- Returns `STRING` when `value_to_trim` is `STRING`, and `BYTES` when it is `BYTES`.
- For `STRING` input, removes all leading and trailing Unicode code points
  that appear in `set_of_characters_to_remove`.
- When `set_of_characters_to_remove` is omitted for a `STRING`, all leading
  and trailing whitespace characters are removed.
- For `BYTES` input, removes all leading and trailing bytes that appear in
  `set_of_characters_to_remove`; the set is required for `BYTES`.
- The character set is interpreted as a set of Unicode code points, not as
  grapheme clusters, so combining marks may be stripped independently of
  their base letter.

## Examples

```googlesql
SELECT CONCAT('#', TRIM('   apple   '), '#') AS example
-- expected: #apple#
```

```googlesql
SELECT TRIM('***apple***', '*') AS example
-- expected: apple
```

```googlesql
SELECT TRIM('xzxapplexxy', 'xyz') AS example
-- expected: apple
```

```googlesql
SELECT TRIM(b'apple', b'na\xab') AS example
-- expected: b'pple'
```

## Edge cases

- Omitting `set_of_characters_to_remove` is only valid for `STRING` inputs;
  `BYTES` inputs require the set.
- Because trimming operates on Unicode code points, a combining diacritic
  in the trim set can strip the same mark from a different base letter
  (e.g. `TRIM('abaW̊', 'Y̊')` yields `abaW`).
- A character or byte not present in the input is a no-op rather than an error.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `TRIM`

```googlesql
TRIM(value_to_trim[, set_of_characters_to_remove])
```

**Description**

Takes a `STRING` or `BYTES` value to trim.

If the value to trim is a `STRING`, removes from this value all leading and
trailing Unicode code points in `set_of_characters_to_remove`.
The set of code points is optional. If it isn't specified, all
whitespace characters are removed from the beginning and end of the
value to trim.

If the value to trim is `BYTES`, removes from this value all leading and
trailing bytes in `set_of_characters_to_remove`. The set of bytes is required.

**Return type**

+ `STRING` if `value_to_trim` is a `STRING` value.
+ `BYTES` if `value_to_trim` is a `BYTES` value.

**Examples**

In the following example, all leading and trailing whitespace characters are
removed from `item` because `set_of_characters_to_remove` isn't specified.

```googlesql
SELECT CONCAT('#', TRIM( '   apple   '), '#') AS example

/*----------+
 | example  |
 +----------+
 | #apple#  |
 +----------*/
```

In the following example, all leading and trailing `*` characters are removed
from '***apple***'.

```googlesql
SELECT TRIM('***apple***', '*') AS example

/*---------+
 | example |
 +---------+
 | apple   |
 +---------*/
```

In the following example, all leading and trailing `x`, `y`, and `z` characters
are removed from 'xzxapplexxy'.

```googlesql
SELECT TRIM('xzxapplexxy', 'xyz') as example

/*---------+
 | example |
 +---------+
 | apple   |
 +---------*/
```

In the following example, examine how `TRIM` interprets characters as
Unicode code-points. If your trailing character set contains a combining
diacritic mark over a particular letter, `TRIM` might strip the
same diacritic mark from a different letter.

```googlesql
SELECT
  TRIM('abaW̊', 'Y̊') AS a,
  TRIM('W̊aba', 'Y̊') AS b,
  TRIM('abaŪ̊', 'Y̊') AS c,
  TRIM('Ū̊aba', 'Y̊') AS d

/*------+------+------+------+
 | a    | b    | c    | d    |
 +------+------+------+------+
 | abaW | W̊aba | abaŪ | Ūaba |
 +------+------+------+------*/
```

In the following example, all leading and trailing `b'n'`, `b'a'`, `b'\xab'`
bytes are removed from `item`.

```googlesql
SELECT b'apple', TRIM(b'apple', b'na\xab') AS example

/*----------------------+------------------+
 | item                 | example          |
 +----------------------+------------------+
 | apple                | pple             |
 +----------------------+------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
