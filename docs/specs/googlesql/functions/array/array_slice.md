---
name: ARRAY_SLICE
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_slice
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_slice.yaml
---

# ARRAY_SLICE

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

## `ARRAY_SLICE`

```googlesql
ARRAY_SLICE(array_to_slice, start_offset, end_offset)
```

**Description**

Returns an array containing zero or more consecutive elements from the
input array.

+ `array_to_slice`: The array that contains the elements you want to slice.
+ `start_offset`: The inclusive starting offset.
+ `end_offset`: The inclusive ending offset.

An offset can be positive or negative. A positive offset starts from the
beginning of the input array and is 0-based. A negative offset starts from
the end of the input array. Out-of-bounds offsets are supported. Here are some
examples:

  <table>
  <thead>
    <tr>
      <th width="150px">Input offset</th>
      <th width="200px">Final offset in array</th>
      <th>Notes</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>0</td>
      <td>[<b>'a'</b>, 'b', 'c', 'd']</td>
      <td>The final offset is <code>0</code>.</td>
    </tr>
    <tr>
      <td>3</td>
      <td>['a', 'b', 'c', <b>'d'</b>]</td>
      <td>The final offset is <code>3</code>.</td>
    </tr>
    <tr>
      <td>5</td>
      <td>['a', 'b', 'c', <b>'d'</b>]</td>
      <td>
        Because the input offset is out of bounds,
        the final offset is <code>3</code> (<code>array length - 1</code>).
      </td>
    </tr>
    <tr>
      <td>-1</td>
      <td>['a', 'b', 'c', <b>'d'</b>]</td>
      <td>
        Because a negative offset is used, the offset starts at the end of the
        array. The final offset is <code>3</code>
        (<code>array length - 1</code>).
      </td>
    </tr>
    <tr>
      <td>-2</td>
      <td>['a', 'b', <b>'c'</b>, 'd']</td>
      <td>
        Because a negative offset is used, the offset starts at the end of the
        array. The final offset is <code>2</code>
        (<code>array length - 2</code>).
      </td>
    </tr>
    <tr>
      <td>-4</td>
      <td>[<b>'a'</b>, 'b', 'c', 'd']</td>
      <td>
        Because a negative offset is used, the offset starts at the end of the
        array. The final offset is <code>0</code>
        (<code>array length - 4</code>).
      </td>
    </tr>
    <tr>
      <td>-5</td>
      <td>[<b>'a'</b>, 'b', 'c', 'd']</td>
      <td>
        Because the offset is negative and out of bounds, the final offset is
        <code>0</code> (<code>array length - array length</code>).
      </td>
    </tr>
  </tbody>
</table>

Additional details:

+ The input array can contain `NULL` elements. `NULL` elements are included
  in the resulting array.
+ Returns `NULL` if `array_to_slice`, `start_offset`, or `end_offset` is
  `NULL`.
+ Returns an empty array if `array_to_slice` is empty.
+ Returns an empty array if the position of the `start_offset` in the array is
  after the position of the `end_offset`.

**Return type**

`ARRAY`

**Examples**

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], 1, 3) AS result

/*-----------+
 | result    |
 +-----------+
 | [b, c, d] |
 +-----------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], -1, 3) AS result

/*-----------+
 | result    |
 +-----------+
 | []        |
 +-----------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], 1, -3) AS result

/*--------+
 | result |
 +--------+
 | [b, c] |
 +--------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], -1, -3) AS result

/*-----------+
 | result    |
 +-----------+
 | []        |
 +-----------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], -3, -1) AS result

/*-----------+
 | result    |
 +-----------+
 | [c, d, e] |
 +-----------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], 3, 3) AS result

/*--------+
 | result |
 +--------+
 | [d]    |
 +--------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], -3, -3) AS result

/*--------+
 | result |
 +--------+
 | [c]    |
 +--------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], 1, 30) AS result

/*--------------+
 | result       |
 +--------------+
 | [b, c, d, e] |
 +--------------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], 1, -30) AS result

/*-----------+
 | result    |
 +-----------+
 | []        |
 +-----------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], -30, 30) AS result

/*-----------------+
 | result          |
 +-----------------+
 | [a, b, c, d, e] |
 +-----------------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], -30, -5) AS result

/*--------+
 | result |
 +--------+
 | [a]    |
 +--------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], 5, 30) AS result

/*--------+
 | result |
 +--------+
 | []     |
 +--------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', 'c', 'd', 'e'], 1, NULL) AS result

/*-----------+
 | result    |
 +-----------+
 | NULL      |
 +-----------*/
```

```googlesql
SELECT ARRAY_SLICE(['a', 'b', NULL, 'd', 'e'], 1, 3) AS result

/*--------------+
 | result       |
 +--------------+
 | [b, NULL, d] |
 +--------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
