---
name: JSON_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_array.yaml
---

# JSON_ARRAY

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

## `JSON_ARRAY`

```googlesql
JSON_ARRAY([value][, ...])
```

**Description**

Creates a JSON array from zero or more SQL values.

Arguments:

+   `value`: A [JSON encoding-supported][json-encodings] value to add
    to a JSON array.

**Return type**

`JSON`

**Examples**

The following query creates a JSON array with one value in it:

```googlesql
SELECT JSON_ARRAY(10) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | [10]      |
 +-----------*/
```

You can create a JSON array with an empty JSON array in it. For example:

```googlesql
SELECT JSON_ARRAY([]) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | [[]]      |
 +-----------*/
```

```googlesql
SELECT JSON_ARRAY(10, 'foo', NULL) AS json_data

/*-----------------+
 | json_data       |
 +-----------------+
 | [10,"foo",null] |
 +-----------------*/
```

```googlesql
SELECT JSON_ARRAY(STRUCT(10 AS a, 'foo' AS b)) AS json_data

/*----------------------+
 | json_data            |
 +----------------------+
 | [{"a":10,"b":"foo"}] |
 +----------------------*/
```

```googlesql
SELECT JSON_ARRAY(10, ['foo', 'bar'], [20, 30]) AS json_data

/*----------------------------+
 | json_data                  |
 +----------------------------+
 | [10,["foo","bar"],[20,30]] |
 +----------------------------*/
```

```googlesql
SELECT JSON_ARRAY(10, [JSON '20', JSON '"foo"']) AS json_data

/*-----------------+
 | json_data       |
 +-----------------+
 | [10,[20,"foo"]] |
 +-----------------*/
```

You can create an empty JSON array. For example:

```googlesql
SELECT JSON_ARRAY() AS json_data

/*-----------+
 | json_data |
 +-----------+
 | []        |
 +-----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
