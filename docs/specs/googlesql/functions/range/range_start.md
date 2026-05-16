---
name: RANGE_START
dialect: googlesql
category: functions/range
status: implemented
notes: Diverges from upstream Examples in testdata; downgraded after extract-testdata audit. Re-promote once cases pass.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range_start
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/range_start.yaml
---

# RANGE_START

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

Verbatim copy from `docs/third_party/googlesql-docs/range-functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `RANGE_START`

```googlesql
RANGE_START(range_to_check)
```

**Description**

Gets the lower bound of a range.

**Definitions**

+   `range_to_check`: The `RANGE<T>` value.

**Details**

Returns `NULL` if the lower bound of `range_value` is `UNBOUNDED`.

Returns `NULL` if `range_to_check` is `NULL`.

**Return type**

`T` in `range_value`

**Examples**

In the following query, the lower bound of the range is retrieved:

```googlesql
SELECT RANGE_START(RANGE<DATE> '[2022-12-01, 2022-12-31)') AS results;

/*------------+
 | results    |
 +------------+
 | 2022-12-01 |
 +------------*/
```

In the following query, the lower bound of the range is unbounded, so
`NULL` is returned:

```googlesql
SELECT RANGE_START(RANGE<DATE> '[UNBOUNDED, 2022-12-31)') AS results;

/*------------+
 | results    |
 +------------+
 | NULL       |
 +------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
