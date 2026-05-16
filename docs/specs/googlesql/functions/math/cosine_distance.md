---
name: COSINE_DISTANCE
dialect: googlesql
category: functions/math
status: implemented
source_url: docs/third_party/googlesql-docs/mathematical_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/mathematical_functions.md#cosine_distance
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/math/cosine_distance.yaml
---

# COSINE_DISTANCE

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

Verbatim copy from `docs/third_party/googlesql-docs/mathematical_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `COSINE_DISTANCE`

```googlesql
COSINE_DISTANCE(vector1, vector2)
```

**Description**

Computes the [cosine distance][wiki-cosine-distance] between two vectors.

**Definitions**

+   `vector1`: A vector that's represented by an
    `ARRAY<T>` value or a sparse vector that is
    represented by an `ARRAY<STRUCT<dimension,magnitude>>` value.
+   `vector2`: A vector that's represented by an
    `ARRAY<T>` value or a sparse vector that is
    represented by an `ARRAY<STRUCT<dimension,magnitude>>` value.

**Details**

+   `ARRAY<T>` can be used to represent a vector. Each zero-based index in this
    array represents a dimension. The value for each element in this array
    represents a magnitude.

    `T` can represent the following and must be the same for both
    vectors:

    
    
    

    + `FLOAT`
    + `DOUBLE`

    
    

    In the following example vector, there are four dimensions. The magnitude
    is `10.0` for dimension `0`, `55.0` for dimension `1`, `40.0` for
    dimension `2`, and `34.0` for dimension `3`:

    ```
    [10.0, 55.0, 40.0, 34.0]
    ```
+   `ARRAY<STRUCT<dimension,magnitude>>` can be used to represent a
    sparse vector. With a sparse vector, you only need to include
    dimension-magnitude pairs for non-zero magnitudes. If a magnitude isn't
    present in the sparse vector, the magnitude is implicitly understood to be
    zero.

    For example, if you have a vector with 10,000 dimensions, but only 10
    dimensions have non-zero magnitudes, then the vector is a sparse vector.
    As a result, it's more efficient to describe a sparse vector by only
    mentioning its non-zero magnitudes.

    In `ARRAY<STRUCT<dimension,magnitude>>`, `STRUCT<dimension,magnitude>`
    represents a dimension-magnitude pair for each non-zero magnitude in a
    sparse vector. These parts need to be included for each dimension-magnitude
    pair:

    + `dimension`: A `STRING` or `INT64` value that represents a
      dimension in a vector.

    + `magnitude`: A `DOUBLE` value that represents a
      non-zero magnitude for a specific dimension in a vector.

    You don't need to include empty dimension-magnitude pairs in a
    sparse vector. For example, the following sparse vector and
    non-sparse vector are equivalent:

    ```googlesql
    -- sparse vector ARRAY<STRUCT<INT64, DOUBLE>>
    [(1, 10.0), (2, 30.0), (5, 40.0)]
    ```

    ```googlesql
    -- vector ARRAY<DOUBLE>
    [0.0, 10.0, 30.0, 0.0, 0.0, 40.0]
    ```

    In a sparse vector, dimension-magnitude pairs don't need to be in any
    particular order. The following sparse vectors are equivalent:

    ```googlesql
    [('a', 10.0), ('b', 30.0), ('d', 40.0)]
    ```

    ```googlesql
    [('d', 40.0), ('a', 10.0), ('b', 30.0)]
    ```
+   Both non-sparse vectors
    in this function must share the same dimensions, and if they don't, an error
    is produced.
+   A vector can't be a zero vector. A vector is a zero vector if it has
    no dimensions or all dimensions have a magnitude of `0`, such as `[]` or
    `[0.0, 0.0]`. If a zero vector is encountered, an error is produced.
+   An error is produced if a magnitude in a vector is `NULL`.
+   If a vector is `NULL`, `NULL` is returned.

**Return type**

`DOUBLE`

**Examples**

In the following example, non-sparsevectors
are used to compute the cosine distance:

```googlesql
SELECT COSINE_DISTANCE([1.0, 2.0], [3.0, 4.0]) AS results;

/*----------+
 | results  |
 +----------+
 | 0.016130 |
 +----------*/
```

In the following example, sparse vectors are used to compute the
cosine distance:

```googlesql
SELECT COSINE_DISTANCE(
 [(1, 1.0), (2, 2.0)],
 [(2, 4.0), (1, 3.0)]) AS results;

 /*----------+
  | results  |
  +----------+
  | 0.016130 |
  +----------*/
```

The ordering of numeric values in a vector doesn't impact the results
produced by this function. For example these queries produce the same results
even though the numeric values in each vector is in a different order:

```googlesql
SELECT COSINE_DISTANCE([1.0, 2.0], [3.0, 4.0]) AS results;
```

```googlesql
SELECT COSINE_DISTANCE([2.0, 1.0], [4.0, 3.0]) AS results;
```

```googlesql
SELECT COSINE_DISTANCE([(1, 1.0), (2, 2.0)], [(1, 3.0), (2, 4.0)]) AS results;
```

```googlesql
 /*----------+
  | results  |
  +----------+
  | 0.016130 |
  +----------*/
```

In the following example, the function can't compute cosine distance against
the first vector, which is a zero vector:

```googlesql
-- ERROR
SELECT COSINE_DISTANCE([0.0, 0.0], [3.0, 4.0]) AS results;
```

```googlesql
-- ERROR
SELECT COSINE_DISTANCE([(1, 0.0), (2, 0.0)], [(1, 3.0), (2, 4.0)]) AS results;
```

Both non-sparse vectors must have the same
dimensions. If not, an error is produced. In the following example, the
first vector has two dimensions and the second vector has three:

```googlesql
-- ERROR
SELECT COSINE_DISTANCE([9.0, 7.0], [8.0, 4.0, 5.0]) AS results;
```

If you use sparse vectors and you repeat a dimension, an error is
produced:

```googlesql
-- ERROR
SELECT COSINE_DISTANCE(
  [(1, 9.0), (2, 7.0), (2, 8.0)], [(1, 8.0), (2, 4.0), (3, 5.0)]) AS results;
```

[wiki-cosine-distance]: https://en.wikipedia.org/wiki/Cosine_similarity#Cosine_distance

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/mathematical_functions.md`.
