---
name: ST_ASTEXT
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_astext
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_astext.yaml
---

# ST_ASTEXT

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

## `ST_ASTEXT`

```googlesql
ST_ASTEXT(geography_expression)
```

**Description**

Returns the [WKT][wkt-link] representation of an input
`GEOGRAPHY`.

See [`ST_GEOGFROMTEXT`][st-geogfromtext] to construct a
`GEOGRAPHY` from WKT.

**Return type**

`STRING`

[wkt-link]: https://en.wikipedia.org/wiki/Well-known_text

[st-geogfromtext]: #st_geogfromtext

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
