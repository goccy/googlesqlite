---
name: RANGE_END
dialect: googlesql
category: functions/range
status: implemented
notes: Auto-extracted upstream Examples diverge from implementation; specctl downgrade pass after extract-testdata audit.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range_end
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/range_end.yaml
---

# RANGE_END

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

## `RANGE_END`

```googlesql
RANGE_END(range_to_check)
```

**Description**

Gets the upper bound of a range.

**Definitions**

+   `range_to_check`: The `RANGE<T>` value.

**Details**

Returns `NULL` if the upper bound in `range_value` is `UNBOUNDED`.

Returns `NULL` if `range_to_check` is `NULL`.

**Return type**

`T` in `range_value`

**Examples**

In the following query, the upper bound of the range is retrieved:

```googlesql
SELECT RANGE_END(RANGE<DATE> '[2022-12-01, 2022-12-31)') AS results;

/*------------+
 | results    |
 +------------+
 | 2022-12-31 |
 +------------*/
```

In the following query, the upper bound of the range is unbounded, so
`NULL` is returned:

```googlesql
SELECT RANGE_END(RANGE<DATE> '[2022-12-01, UNBOUNDED)') AS results;

/*------------+
 | results    |
 +------------+
 | NULL       |
 +------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
