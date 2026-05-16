---
name: RANGE_INTERSECT
dialect: googlesql
category: functions/range
status: implemented
notes: |
  Blocked on RANGE literal formatting in the resolved-tree → SQL emitter (see go-googlesql ResolvedLiteral for TypeKindTypeRange). The runtime side exists; the formatter needs RANGE coverage.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range_intersect
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/range_intersect.yaml
---

# RANGE_INTERSECT

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

## `RANGE_INTERSECT`

```googlesql
RANGE_INTERSECT(range_a, range_b)
```

**Description**

Gets a segment of two ranges that intersect.

**Definitions**

+   `range_a`: The first `RANGE<T>` value.
+   `range_b`: The second `RANGE<T>` value.

**Details**

Returns `NULL` if any input is` NULL`.

Produces an error if `range_a` and `range_b` don't overlap. To return
`NULL` instead, add the `SAFE.` prefix to the function name.

`T` must be of the same type for all inputs.

**Return type**

`RANGE<T>`

**Examples**

```googlesql
SELECT RANGE_INTERSECT(
  RANGE<DATE> '[2022-02-01, 2022-09-01)',
  RANGE<DATE> '[2021-06-15, 2022-04-15)') AS results;

/*--------------------------+
 | results                  |
 +--------------------------+
 | [2022-02-01, 2022-04-15) |
 +--------------------------*/
```

```googlesql
SELECT RANGE_INTERSECT(
  RANGE<DATE> '[2022-02-01, UNBOUNDED)',
  RANGE<DATE> '[2021-06-15, 2022-04-15)') AS results;

/*--------------------------+
 | results                  |
 +--------------------------+
 | [2022-02-01, 2022-04-15) |
 +--------------------------*/
```

```googlesql
SELECT RANGE_INTERSECT(
  RANGE<DATE> '[2022-02-01, UNBOUNDED)',
  RANGE<DATE> '[2021-06-15, UNBOUNDED)') AS results;

/*-------------------------+
 | results                 |
 +-------------------------+
 | [2022-02-01, UNBOUNDED) |
 +-------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
