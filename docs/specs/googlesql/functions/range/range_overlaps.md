---
name: RANGE_OVERLAPS
dialect: googlesql
category: functions/range
status: implemented
notes: |
  Blocked on RANGE literal formatting in the resolved-tree → SQL emitter (see go-googlesql ResolvedLiteral for TypeKindTypeRange). The runtime side exists; the formatter needs RANGE coverage.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range_overlaps
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/range_overlaps.yaml
---

# RANGE_OVERLAPS

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

## `RANGE_OVERLAPS`

```googlesql
RANGE_OVERLAPS(range_a, range_b)
```

**Description**

Checks if two ranges overlap.

**Definitions**

+   `range_a`: The first `RANGE<T>` value.
+   `range_b`: The second `RANGE<T>` value.

**Details**

Returns `TRUE` if a part of `range_a` intersects with `range_b`, otherwise
returns `FALSE`.

`T` must be of the same type for all inputs.

To get the part of the range that overlaps, use the
[`RANGE_INTERSECT`][range-intersect] function.

**Return type**

`BOOL`

**Examples**

In the following query, the first and second ranges overlap between
`2022-02-01` and `2022-04-15`:

```googlesql
SELECT RANGE_OVERLAPS(
  RANGE<DATE> '[2022-02-01, 2022-09-01)',
  RANGE<DATE> '[2021-06-15, 2022-04-15)') AS results;

/*---------+
 | results |
 +---------+
 | TRUE    |
 +---------*/
```

In the following query, the first and second ranges don't overlap:

```googlesql
SELECT RANGE_OVERLAPS(
  RANGE<DATE> '[2020-02-01, 2020-09-01)',
  RANGE<DATE> '[2021-06-15, 2022-04-15)') AS results;

/*---------+
 | results |
 +---------+
 | FALSE   |
 +---------*/
```

In the following query, the first and second ranges overlap between
`2022-02-01` and `UNBOUNDED`:

```googlesql
SELECT RANGE_OVERLAPS(
  RANGE<DATE> '[2022-02-01, UNBOUNDED)',
  RANGE<DATE> '[2021-06-15, UNBOUNDED)') AS results;

/*---------+
 | results |
 +---------+
 | TRUE    |
 +---------*/
```

[range-intersect]: #range_intersect

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
