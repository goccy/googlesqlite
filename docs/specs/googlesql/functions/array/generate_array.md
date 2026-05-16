---
name: GENERATE_ARRAY
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#generate_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/generate_array.yaml
---

# GENERATE_ARRAY

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

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `GENERATE_ARRAY`

```googlesql
GENERATE_ARRAY(start_expression, end_expression[, step_expression])
```

**Description**

Returns an array of values. The `start_expression` and `end_expression`
parameters determine the inclusive start and end of the array.

The `GENERATE_ARRAY` function accepts the following data types as inputs:

+ `INT64`
+ `UINT64`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `DOUBLE`

The `step_expression` parameter determines the increment used to
generate array values. The default value for this parameter is `1`.

This function returns an error if `step_expression` is set to 0, or if any
input is `NaN`.

If any argument is `NULL`, the function will return a `NULL` array.

**Return Data Type**

`ARRAY`

**Examples**

The following returns an array of integers, with a default step of 1.

```googlesql
SELECT GENERATE_ARRAY(1, 5) AS example_array;

/*-----------------+
 | example_array   |
 +-----------------+
 | [1, 2, 3, 4, 5] |
 +-----------------*/
```

The following returns an array using a user-specified step size.

```googlesql
SELECT GENERATE_ARRAY(0, 10, 3) AS example_array;

/*---------------+
 | example_array |
 +---------------+
 | [0, 3, 6, 9]  |
 +---------------*/
```

The following returns an array using a negative value, `-3` for its step size.

```googlesql
SELECT GENERATE_ARRAY(10, 0, -3) AS example_array;

/*---------------+
 | example_array |
 +---------------+
 | [10, 7, 4, 1] |
 +---------------*/
```

The following returns an array using the same value for the `start_expression`
and `end_expression`.

```googlesql
SELECT GENERATE_ARRAY(4, 4, 10) AS example_array;

/*---------------+
 | example_array |
 +---------------+
 | [4]           |
 +---------------*/
```

The following returns an empty array, because the `start_expression` is greater
than the `end_expression`, and the `step_expression` value is positive.

```googlesql
SELECT GENERATE_ARRAY(10, 0, 3) AS example_array;

/*---------------+
 | example_array |
 +---------------+
 | []            |
 +---------------*/
```

The following returns a `NULL` array because `end_expression` is `NULL`.

```googlesql
SELECT GENERATE_ARRAY(5, NULL, 1) AS example_array;

/*---------------+
 | example_array |
 +---------------+
 | NULL          |
 +---------------*/
```

The following returns multiple arrays.

```googlesql
SELECT GENERATE_ARRAY(start, 5) AS example_array
FROM UNNEST([3, 4, 5]) AS start;

/*---------------+
 | example_array |
 +---------------+
 | [3, 4, 5]     |
 | [4, 5]        |
 | [5]           |
 +---------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
