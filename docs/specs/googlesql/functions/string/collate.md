---
name: COLLATE
dialect: googlesql
category: functions/string
status: implemented
notes: |
  COLLATE is registered as a root-level scalar
  `(STRING, STRING) -> STRING` in internal/catalog.go so the
  function-call form `COLLATE(value, spec)` shown in the upstream
  string_functions Examples resolves. The runtime UDF in
  internal/functions/string/collate.go applies a Unicode lowercase
  fold when the collation attribute is `:ci` (case-insensitive), so a
  subsequent comparison between two COLLATE results observes the
  collation-aware order. Plain `<` between non-COLLATE strings keeps
  the default code-point order, matching Example 2.
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#collate
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/collate.yaml
---

# COLLATE

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

## `COLLATE`

```googlesql
COLLATE(value, collate_specification)
```

Takes a `STRING` and a [collation specification][link-collation-spec]. Returns
a `STRING` with a collation specification. If `collate_specification` is empty,
returns a value with collation removed from the `STRING`.

The collation specification defines how the resulting `STRING` can be compared
and sorted. To learn more, see
[Collation][link-collation-concepts].

+ `collation_specification` must be a string literal, otherwise an error is
  thrown.
+ Returns `NULL` if `value` is `NULL`.

**Return type**

`STRING`

**Examples**

In this example, the weight of `a` is less than the weight of `Z`. This
is because the collate specification, `und:ci` assigns more weight to `Z`.

```googlesql
WITH Words AS (
  SELECT
    COLLATE('a', 'und:ci') AS char1,
    COLLATE('Z', 'und:ci') AS char2
)
SELECT ( Words.char1 < Words.char2 ) AS a_less_than_Z
FROM Words;

/*----------------+
 | a_less_than_Z  |
 +----------------+
 | TRUE           |
 +----------------*/
```

In this example, the weight of `a` is greater than the weight of `Z`. This
is because the default collate specification assigns more weight to `a`.

```googlesql
WITH Words AS (
  SELECT
    'a' AS char1,
    'Z' AS char2
)
SELECT ( Words.char1 < Words.char2 ) AS a_less_than_Z
FROM Words;

/*----------------+
 | a_less_than_Z  |
 +----------------+
 | FALSE          |
 +----------------*/
```

[link-collation-spec]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md#collate_spec_details

[link-collation-concepts]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
