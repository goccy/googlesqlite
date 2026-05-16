---
name: ST_HAUSDORFFDWITHIN
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_hausdorffdwithin
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_hausdorffdwithin.yaml
---

# ST_HAUSDORFFDWITHIN

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

Verbatim copy from `docs/third_party/googlesql-docs/geography_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ST_HAUSDORFFDWITHIN`

```googlesql
ST_HAUSDORFFDWITHIN(
  geography_1,
  geography_2,
  distance
  [, directed => { TRUE | FALSE } ]
)
```

**Description**

Returns `TRUE` if the [Hausdorff distance][st-hausdorffdistance] between `geography_1` and
`geography_2` is less than or equal to the distance given by the
`distance` argument; otherwise, returns `FALSE`.

**Definitions**

+   `geography_1`: A `GEOGRAPHY` value that represents the first geography.
+   `geography_2`: A `GEOGRAPHY` value that represents the second geography.
+   `distance`: A `DOUBLE` value that represents meters on the
    surface of the Earth.
+   `directed`: A named argument with a `BOOL` value. Represents the type of
    computation to use on the input geographies. If this argument isn't
    specified, `directed => FALSE` is used by default.

    +   `FALSE`: The largest Hausdorff distance found in
        (`geography_1`, `geography_2`) and
        (`geography_2`, `geography_1`).

    +   `TRUE` (default): The Hausdorff distance for
        (`geography_1`, `geography_2`).

**Details**

If an input geography is `NULL`, the function returns `NULL`.

**Return type**

`BOOL`

**Examples**

The following example checks whether the Hausdorff distance between the first
and second geographies is less than or equal to 100,000 meters.

```googlesql
SELECT
  ST_HAUSDORFFDWITHIN(
    ST_GEOGFROMTEXT('LINESTRING(10 1, 20 1)'),
    ST_GEOGFROMTEXT('LINESTRING(10 2, 20 2)'),
    100000) AS is_close;

/*----------+
 | is_close |
 +----------+
 | false    |
 +----------*/
```

[st-hausdorffdistance]: #st_hausdorffdistance

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
