---
name: REPEAT
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#repeat
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/repeat.yaml
---

# REPEAT

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

## `REPEAT`

```googlesql
REPEAT(original_value, repetitions)
```

**Description**

Returns a `STRING` or `BYTES` value that consists of `original_value`, repeated.
The `repetitions` parameter specifies the number of times to repeat
`original_value`. Returns `NULL` if either `original_value` or `repetitions`
are `NULL`.

This function returns an error if the `repetitions` value is negative.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT REPEAT('abc', 3) AS results

/*-----------+
 | results   |
 |-----------|
 | abcabcabc |
 +-----------*/
```

```googlesql
SELECT REPEAT('abc', NULL) AS results

/*---------+
 | results |
 |---------|
 | NULL    |
 +---------*/
```

```googlesql
SELECT REPEAT(NULL, 3) AS results

/*---------+
 | results |
 |---------|
 | NULL    |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
