---
name: ST_ASBINARY
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  OGC WKB encoding (little-endian) of the geography. Runtime entry: BindStAsBinary in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_asbinary
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_asbinary.yaml
---

# ST_ASBINARY

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

## `ST_ASBINARY`

```googlesql
ST_ASBINARY(geography_expression)
```

**Description**

Returns the [WKB][wkb-link] representation of an input
`GEOGRAPHY`.

See [`ST_GEOGFROMWKB`][st-geogfromwkb] to construct a
`GEOGRAPHY` from WKB.

**Return type**

`BYTES`

[wkb-link]: https://en.wikipedia.org/wiki/Well-known_text#Well-known_binary

[st-geogfromwkb]: #st_geogfromwkb

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
