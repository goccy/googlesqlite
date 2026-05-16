---
name: REGEXP_INSTR
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_instr
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_instr.yaml
---

# REGEXP_INSTR

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

## `REGEXP_INSTR`

```googlesql
REGEXP_INSTR(source_value, regexp [, position[, occurrence, [occurrence_position]]])
```

**Description**

Returns the lowest 1-based position of a regular expression, `regexp`, in
`source_value`. `source_value` and `regexp` must be the same type, either
`STRING` or `BYTES`.

If `position` is specified, the search starts at this position in
`source_value`, otherwise it starts at `1`, which is the beginning of
`source_value`. `position` is of type `INT64` and must be positive.

If `occurrence` is specified, the search returns the position of a specific
instance of `regexp` in `source_value`. If not specified, `occurrence` defaults
to `1` and returns the position of the first occurrence. For `occurrence` > 1,
the function searches for the next, non-overlapping occurrence.
`occurrence` is of type `INT64` and must be positive.

You can optionally use `occurrence_position` to specify where a position
in relation to an `occurrence` starts. Your choices are:

+  `0`: Returns the start position of `occurrence`.
+  `1`: Returns the end position of `occurrence` + `1`. If the
   end of the occurrence is at the end of `source_value `,
   `LENGTH(source_value) + 1` is returned.

Returns `0` if:

+ No match is found.
+ If `occurrence` is greater than the number of matches found.
+ If `position` is greater than the length of `source_value`.
+ The regular expression is empty.

Returns `NULL` if:

+ `position` is `NULL`.
+ `occurrence` is `NULL`.

Returns an error if:

+ `position` is `0` or negative.
+ `occurrence` is `0` or negative.
+ `occurrence_position` is neither `0` nor `1`.
+ The regular expression is invalid.
+ The regular expression has more than one capturing group.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT
  REGEXP_INSTR('ab@cd-ef',  '@[^-]*') AS instr_a,
  REGEXP_INSTR('ab@d-ef',   '@[^-]*') AS instr_b,
  REGEXP_INSTR('abc@cd-ef', '@[^-]*') AS instr_c,
  REGEXP_INSTR('abc-ef',    '@[^-]*') AS instr_d,

/*---------------------------------------+
 | instr_a | instr_b | instr_c | instr_d |
 +---------------------------------------+
 | 3       | 3       | 4       | 0       |
 +---------------------------------------*/
```

```googlesql
SELECT
  REGEXP_INSTR('a@cd-ef b@cd-ef', '@[^-]*', 1) AS instr_a,
  REGEXP_INSTR('a@cd-ef b@cd-ef', '@[^-]*', 2) AS instr_b,
  REGEXP_INSTR('a@cd-ef b@cd-ef', '@[^-]*', 3) AS instr_c,
  REGEXP_INSTR('a@cd-ef b@cd-ef', '@[^-]*', 4) AS instr_d,

/*---------------------------------------+
 | instr_a | instr_b | instr_c | instr_d |
 +---------------------------------------+
 | 2       | 2       | 10      | 10      |
 +---------------------------------------*/
```

```googlesql
SELECT
  REGEXP_INSTR('a@cd-ef b@cd-ef c@cd-ef', '@[^-]*', 1, 1) AS instr_a,
  REGEXP_INSTR('a@cd-ef b@cd-ef c@cd-ef', '@[^-]*', 1, 2) AS instr_b,
  REGEXP_INSTR('a@cd-ef b@cd-ef c@cd-ef', '@[^-]*', 1, 3) AS instr_c

/*-----------------------------+
 | instr_a | instr_b | instr_c |
 +-----------------------------+
 | 2       | 10      | 18      |
 +-----------------------------*/
```

```googlesql
SELECT
  REGEXP_INSTR('a@cd-ef', '@[^-]*', 1, 1, 0) AS instr_a,
  REGEXP_INSTR('a@cd-ef', '@[^-]*', 1, 1, 1) AS instr_b

/*-------------------+
 | instr_a | instr_b |
 +-------------------+
 | 2       | 5       |
 +-------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
