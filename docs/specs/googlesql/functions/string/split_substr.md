---
name: SPLIT_SUBSTR
dialect: googlesql
category: functions/string
status: implemented
notes: |
  Lambda-style or grouped regex outputs that need a planner extension (REGEXP_EXTRACT_GROUPS returns ARRAY<STRUCT>, SPLIT_SUBSTR is variadic). Deferred.
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#split_substr
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/split_substr.yaml
---

# SPLIT_SUBSTR

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

## `SPLIT_SUBSTR`

```googlesql
SPLIT_SUBSTR(value, delimiter, start_split[, count])
```

**Description**

Returns a substring from an input `STRING` that's determined by a delimiter, a
location that indicates the first split of the substring to return, and the
number of splits to include in the returned substring.

The `value` argument is the supplied `STRING` value from which a substring is
returned.

The `delimiter` argument is the delimiter used to split the input `STRING`. It
must be a literal character or sequence of characters.

+ The `delimiter` argument can't be a regular expression.
+ Delimiter matching is from left to right.
+ If the delimiter is a sequence of characters, then two instances of the
  delimiter in the input string can't overlap. For example, if the delimiter is
  `**`, then the delimiters in the string `aa***bb***cc` are:
    + The first two asterisks after `aa`.
    + The first two asterisks after `bb`.

The `start_split` argument is an integer that specifies the first split of the
substring to return.

+ If `start_split` is `1`, then the returned substring starts from the first
  split.
+ If `start_split` is `0` or less than the negative of the number of splits,
  then `start_split` is treated as if it's `1` and returns a substring that
  starts with the first split.
+ If `start_split` is greater than the number of splits, then an empty string is
  returned.
+ If `start_split` is negative, then the splits are counted from the end of the
  input string. If `start_split` is `-1`, then the last split in the input
  string is returned.

The optional `count` argument is an integer that specifies the maximum number
of splits to include in the returned substring.

+ If `count` isn't specified, then the substring from the `start_split`
  position to the end of the input string is returned.
+ If `count` is `0`, an empty string is returned.
+ If `count` is negative, an error is returned.
+ If the sum of `count` plus `start_split` is greater than the number of splits,
  then a substring from `start_split` to the end of the input string is
  returned.

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Return type**

`STRING`

**Examples**

The following example returns an empty string because `count` is `0`:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", 1, 0) AS example

/*---------+
 | example |
 +---------+
 |         |
 +---------*/
```

The following example returns two splits starting with the first split:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", 1, 2) AS example

/*---------+
 | example |
 +---------+
 | www.abc |
 +---------*/
```

The following example returns one split starting with the first split:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", 1, 1) AS example

/*---------+
 | example |
 +---------+
 | www     |
 +---------*/
```

The following example returns splits from the right because `start_split` is a
negative value:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", -1, 1) AS example

/*---------+
 | example |
 +---------+
 | com     |
 +---------*/
```

The following example returns a substring with three splits, starting with the
first split:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", 1, 3) AS example

/*-------------+
 | example     |
 +-------------+
 | www.abc.xyz |
 +------------*/
```

If `start_split` is zero, then it's treated as if it's `1`. The following
example returns three substrings starting with the first split:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", 0, 3) AS example

/*-------------+
 | example     |
 +-------------+
 | www.abc.xyz |
 +------------*/
```

If `start_split` is greater than the number of splits, then an empty string is
returned:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", 5, 3) AS example

/*---------+
 | example |
 +---------+
 |         |
 +--------*/
```

In the following example, the `start_split` value (`-5`) is less than the
negative of the number of splits (`-4`), so `start_split` is treated as `1`:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", -5, 3) AS example

/*-------------+
 | example     |
 +-------------+
 | www.abc.xyz |
 +------------*/
```

In the following example, the substring from `start_split` to the end of the
string is returned because `count` isn't specified:

```googlesql
SELECT SPLIT_SUBSTR("www.abc.xyz.com", ".", 3) AS example

/*---------+
 | example |
 +---------+
 | xyz.com |
 +--------*/
```

The following two examples demonstrate how `SPLIT_SUBSTR` works with a
multi-character delimiter that has overlapping matches in the input string. In
each example, the input string contains instances of three asterisks in a row
(`***`) and the delimiter is two asterisks (`**`).

```googlesql
SELECT SPLIT_SUBSTR('aaa***bbb***ccc', '**', 1, 2) AS example

/*-----------+
 | example   |
 +-----------+
 | aaa***bbb |
 +----------*/
```

```googlesql
SELECT SPLIT_SUBSTR('aaa***bbb***ccc', '**', 2, 2) AS example

/*------------+
 | example    |
 +------------+
 | *bbb***ccc |
 +-----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
