---
name: RANGE_CONTAINS
dialect: googlesql
category: functions/range
status: implemented
notes: |
  Blocked on RANGE literal formatting in the resolved-tree → SQL emitter (see go-googlesql ResolvedLiteral for TypeKindTypeRange). The runtime side exists; the formatter needs RANGE coverage.
source_url: docs/third_party/googlesql-docs/range-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range_contains
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/range/range_contains.yaml
---

# RANGE_CONTAINS

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

## `RANGE_CONTAINS`

+   [Signature 1][range_contains-sig1]: Checks if every value in one range is
    in another range.
+   [Signature 2][range_contains-sig2]: Checks if a value is in a range.

#### Signature 1

```googlesql
RANGE_CONTAINS(outer_range, inner_range)
```

**Description**

Checks if the inner range is in the outer range.

**Definitions**

+   `outer_range`: The `RANGE<T>` value to search within.
+   `inner_range`: The `RANGE<T>` value to search for in `outer_range`.

**Details**

Returns `TRUE` if `inner_range` exists in `outer_range`.
Otherwise, returns `FALSE`.

`T` must be of the same type for all inputs.

**Return type**

`BOOL`

**Examples**

In the following query, the inner range is in the outer range:

```googlesql
SELECT RANGE_CONTAINS(
  RANGE<DATE> '[2022-01-01, 2023-01-01)',
  RANGE<DATE> '[2022-04-01, 2022-07-01)') AS results;

/*---------+
 | results |
 +---------+
 | TRUE    |
 +---------*/
```

In the following query, the inner range isn't in the outer range:

```googlesql
SELECT RANGE_CONTAINS(
  RANGE<DATE> '[2022-01-01, 2023-01-01)',
  RANGE<DATE> '[2023-01-01, 2023-04-01)') AS results;

/*---------+
 | results |
 +---------+
 | FALSE   |
 +---------*/
```

#### Signature 2

```googlesql
RANGE_CONTAINS(range_to_search, value_to_find)
```

**Description**

Checks if a value is in a range.

**Definitions**

+   `range_to_search`: The `RANGE<T>` value to search within.
+   `value_to_find`: The value to search for in `range_to_search`.

**Details**

Returns `TRUE` if `value_to_find` exists in `range_to_search`.
Otherwise, returns `FALSE`.

The data type for `value_to_find` must be the same data type as `T`in
`range_to_search`.

**Return type**

`BOOL`

**Examples**

In the following query, the value `2022-04-01` is found in the range
`[2022-01-01, 2023-01-01)`:

```googlesql
SELECT RANGE_CONTAINS(
  RANGE<DATE> '[2022-01-01, 2023-01-01)',
  DATE '2022-04-01') AS results;

/*---------+
 | results |
 +---------+
 | TRUE    |
 +---------*/
```

In the following query, the value `2023-04-01` isn't found in the range
`[2022-01-01, 2023-01-01)`:

```googlesql
SELECT RANGE_CONTAINS(
  RANGE<DATE> '[2022-01-01, 2023-01-01)',
  DATE '2023-04-01') AS results;

/*---------+
 | results |
 +---------+
 | FALSE   |
 +---------*/
```

[range_contains-sig1]: #signature_1

[range_contains-sig2]: #signature_2

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/range-functions.md`.
