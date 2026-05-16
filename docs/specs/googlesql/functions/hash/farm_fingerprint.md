---
name: FARM_FINGERPRINT
dialect: googlesql
category: functions/hash
status: implemented
source_url: docs/third_party/googlesql-docs/hash_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hash_functions.md#farm_fingerprint
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hash/farm_fingerprint.yaml
---

# FARM_FINGERPRINT

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

Verbatim copy from `docs/third_party/googlesql-docs/hash_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `FARM_FINGERPRINT`

```
FARM_FINGERPRINT(value)
```

**Description**

Computes the fingerprint of the `STRING` or `BYTES` input using the
`Fingerprint64` function from the
[open-source FarmHash library][hash-link-to-farmhash-github]. The output
of this function for a particular input will never change.

**Return type**

INT64

**Examples**

```googlesql
WITH example AS (
  SELECT 1 AS x, "foo" AS y, true AS z UNION ALL
  SELECT 2 AS x, "apple" AS y, false AS z UNION ALL
  SELECT 3 AS x, "" AS y, true AS z
)
SELECT
  *,
  FARM_FINGERPRINT(CONCAT(CAST(x AS STRING), y, CAST(z AS STRING)))
    AS row_fingerprint
FROM example;
/*---+-------+-------+----------------------+
 | x | y     | z     | row_fingerprint      |
 +---+-------+-------+----------------------+
 | 1 | foo   | true  | -1541654101129638711 |
 | 2 | apple | false | 2794438866806483259  |
 | 3 |       | true  | -4880158226897771312 |
 +---+-------+-------+----------------------*/
```

[hash-link-to-farmhash-github]: https://github.com/google/farmhash

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hash_functions.md`.
