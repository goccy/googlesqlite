---
name: RANGE
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#range-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/range.yaml
---

# RANGE

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

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## Range type 
<a id="range_type"></a>

<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Range</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>RANGE</code></td>
      <td>
        Contiguous range between two dates, datetimes, or timestamps.
        The lower and upper bound for the range are optional. The lower bound
        is inclusive and the upper bound is exclusive.
      </td>
    </tr>
  </tbody>
</table>

### Declare a range type

A range type can be declared as follows:

<table>
  <thead>
    <tr>
      <th>Type Declaration</th>
      <th>Meaning</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>RANGE&lt;DATE&gt;</code></td>
      <td>Contiguous range between two dates.</td>
    </tr>
    <tr>
      <td><code>RANGE&lt;DATETIME&gt;</code></td>
      <td>Contiguous range between two datetimes.</td>
    </tr>
    <tr>
      <td><code>RANGE&lt;TIMESTAMP&gt;</code></td>
      <td>Contiguous range between two timestamps.</td>
    </tr>
  </tbody>
</table>

### Construct a range 
<a id="constructing_a_range"></a>

You can construct a range with the [`RANGE` constructor][range-with-constructor]
or a [range literal][range-with-literal].

#### Construct a range with a constructor 
<a id="range_with_constructor"></a>

You can construct a range with the `RANGE` constructor. To learn more,
see [`RANGE`][range-constructor].

#### Construct a range with a literal 
<a id="range_with_literal"></a>

You can construct a range with a range literal. The canonical format for a
range literal has the following parts:

```googlesql
RANGE<T> '[lower_bound, upper_bound)'
```

+   `T`: The type of range. This can be `DATE`, `DATETIME`, or `TIMESTAMP`.
+   `lower_bound`: The range starts from this value. This can be a
    [date][date-literals], [datetime][datetime-literals], or
    [timestamp][timestamp-literals] literal. If this value is `UNBOUNDED` or
    `NULL`, the range doesn't include a lower bound.
+   `upper_bound`: The range ends before this value. This can be a
    [date][date-literals], [datetime][datetime-literals], or
    [timestamp][timestamp-literals] literal. If this value is `UNBOUNDED` or
    `NULL`, the range doesn't include an upper bound.

`T`, `lower_bound`, and `upper_bound` must be of the same data type.

To learn more about the literal representation of a range type,
see [Range literals][range-literals].

### Additional details

The range type doesn't support arithmetic operators.

[range-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#range_literals

[range-with-constructor]: #range_with_constructor

[range-constructor]: https://github.com/google/googlesql/blob/master/docs/range-functions.md#range

[range-with-literal]: #range_with_literal

[date-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#date_literals

[datetime-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#datetime_literals

[timestamp-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#timestamp_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
