---
name: SAFE_TO_JSON
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#safe_to_json
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/safe_to_json.yaml
---

# SAFE_TO_JSON

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

Verbatim copy from `docs/third_party/googlesql-docs/json_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `SAFE_TO_JSON`

```googlesql
SAFE_TO_JSON(sql_value)
```

**Description**

Similar to the `TO_JSON` function, but for each unsupported field in the
input argument, produces a JSON null instead of an error.

Arguments:

+   `sql_value`: The SQL value to convert to a JSON value. You can review the
    GoogleSQL data types that this function supports and their
    [JSON encodings][json-encodings].

**Return type**

`JSON`

**Example**

The following queries are functionally the same, except that `SAFE_TO_JSON`
produces a JSON null instead of an error when a hypothetical unsupported
data type is encountered:

```googlesql
-- Produces a JSON null.
SELECT SAFE_TO_JSON(CAST(b'' AS UNSUPPORTED_TYPE)) as result;
```

```googlesql
-- Produces an error.
SELECT TO_JSON(CAST(b'' AS UNSUPPORTED_TYPE), stringify_wide_numbers=>TRUE) as result;
```

In the following query, the value for `ut` is ignored because the value is an
unsupported type:

```googlesql
SELECT SAFE_TO_JSON(STRUCT(CAST(b'' AS UNSUPPORTED_TYPE) AS ut) AS result;

/*--------------+
 | result       |
 +--------------+
 | {"ut": null} |
 +--------------*/
```

The following array produces a JSON null instead of an error because the data
type for the array isn't supported.

```googlesql
SELECT SAFE_TO_JSON([
        CAST(b'' AS UNSUPPORTED_TYPE),
        CAST(b'' AS UNSUPPORTED_TYPE),
        CAST(b'' AS UNSUPPORTED_TYPE),
    ]) AS result;

/*------------+
 | result     |
 +------------+
 | null       |
 +------------*/
```

[json-encodings]: #json_encodings

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
