---
name: INTERVAL
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#interval-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/interval.yaml
---

# INTERVAL

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

## Interval type 
<a id="interval_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Range</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>INTERVAL</code></td>
<td>
-10000-0 -3660000 -87840000:0:0 to 10000-0 3660000 87840000:0:0
</td>
</tr>
</tbody>
</table>

An `INTERVAL` object represents duration or amount of time, without referring
to any specific point in time.

##### Canonical format

```
[sign]Y-M [sign]D [sign]H:M:S[.F]
```

+ `sign`: `+` or `-`
+ `Y`: Year
+ `M`: Month
+ `D`: Day
+ `H`: Hour
+ `M`: Minute
+ `S`: Second
+ `[.F]`: Up to nine fractional
  digits (nanosecond precision)

To learn more about the literal representation of an interval type,
see [Interval literals][interval-literals].

### Constructing an interval 
<a id="construct_interval"></a>

You can construct an interval with an interval literal that supports
a [single datetime part][single-datetime-part-interval] or a
[datetime part range][range-datetime-part-interval].

#### Construct an interval with a single datetime part 
<a id="single_datetime_part_interval"></a>

```googlesql
INTERVAL int64_expression datetime_part
```

You can construct an `INTERVAL` object with an `INT64` expression and one
[interval-supported datetime part][interval-datetime-parts]. For example:

```googlesql
-- 1 year, 0 months, 0 days, 0 hours, 0 minutes, and 0 seconds (1-0 0 0:0:0)
INTERVAL 1 YEAR
INTERVAL 4 QUARTER
INTERVAL 12 MONTH

-- 0 years, 3 months, 0 days, 0 hours, 0 minutes, and 0 seconds (0-3 0 0:0:0)
INTERVAL 1 QUARTER
INTERVAL 3 MONTH

-- 0 years, 0 months, 42 days, 0 hours, 0 minutes, and 0 seconds (0-0 42 0:0:0)
INTERVAL 6 WEEK
INTERVAL 42 DAY

-- 0 years, 0 months, 0 days, 25 hours, 0 minutes, and 0 seconds (0-0 0 25:0:0)
INTERVAL 25 HOUR
INTERVAL 1500 MINUTE
INTERVAL 90000 SECOND

-- 0 years, 0 months, 0 days, 1 hours, 30 minutes, and 0 seconds (0-0 0 1:30:0)
INTERVAL 90 MINUTE

-- 0 years, 0 months, 0 days, 0 hours, 1 minutes, and 30 seconds (0-0 0 0:1:30)
INTERVAL 90 SECOND

-- 0 years, 0 months, -5 days, 0 hours, 0 minutes, and 0 seconds (0-0 -5 0:0:0)
INTERVAL -5 DAY
```

For additional examples, see [Interval literals][interval-literal-single].

#### Construct an interval with a datetime part range 
<a id="range_datetime_part_interval"></a>

```googlesql
INTERVAL datetime_parts_string starting_datetime_part TO ending_datetime_part
```

You can construct an `INTERVAL` object with a `STRING` that contains the
datetime parts that you want to include, a starting datetime part, and an ending
datetime part. The resulting `INTERVAL` object only includes datetime parts in
the specified range.

You can use one of the following formats with the
[interval-supported datetime parts][interval-datetime-parts]:

<table>
  <thead>
    <tr>
      <th>Datetime part string</th>
      <th>Datetime parts</th>
      <th>Example</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>Y-M</code></td>
      <td><code>YEAR TO MONTH</code></td>
      <td><code>INTERVAL '2-11' YEAR TO MONTH</code></td>
    </tr>
    <tr>
      <td><code>Y-M D</code></td>
      <td><code>YEAR TO DAY</code></td>
      <td><code>INTERVAL '2-11 28' YEAR TO DAY</code></td>
    </tr>
    <tr>
      <td><code>Y-M D H</code></td>
      <td><code>YEAR TO HOUR</code></td>
      <td><code>INTERVAL '2-11 28 16' YEAR TO HOUR</code></td>
    </tr>
    <tr>
      <td><code>Y-M D H:M</code></td>
      <td><code>YEAR TO MINUTE</code></td>
      <td><code>INTERVAL '2-11 28 16:15' YEAR TO MINUTE</code></td>
    </tr>
    <tr>
      <td><code>Y-M D H:M:S</code></td>
      <td><code>YEAR TO SECOND</code></td>
      <td><code>INTERVAL '2-11 28 16:15:14' YEAR TO SECOND</code></td>
    </tr>
    <tr>
      <td><code>M D</code></td>
      <td><code>MONTH TO DAY</code></td>
      <td><code>INTERVAL '11 28' MONTH TO DAY</code></td>
    </tr>
    <tr>
      <td><code>M D H</code></td>
      <td><code>MONTH TO HOUR</code></td>
      <td><code>INTERVAL '11 28 16' MONTH TO HOUR</code></td>
    </tr>
    <tr>
      <td><code>M D H:M</code></td>
      <td><code>MONTH TO MINUTE</code></td>
      <td><code>INTERVAL '11 28 16:15' MONTH TO MINUTE</code></td>
    </tr>
    <tr>
      <td><code>M D H:M:S</code></td>
      <td><code>MONTH TO SECOND</code></td>
      <td><code>INTERVAL '11 28 16:15:14' MONTH TO SECOND</code></td>
    </tr>
    <tr>
      <td><code>D H</code></td>
      <td><code>DAY TO HOUR</code></td>
      <td><code>INTERVAL '28 16' DAY TO HOUR</code></td>
    </tr>
    <tr>
      <td><code>D H:M</code></td>
      <td><code>DAY TO MINUTE</code></td>
      <td><code>INTERVAL '28 16:15' DAY TO MINUTE</code></td>
    </tr>
    <tr>
      <td><code>D H:M:S</code></td>
      <td><code>DAY TO SECOND</code></td>
      <td><code>INTERVAL '28 16:15:14' DAY TO SECOND</code></td>
    </tr>
    <tr>
      <td><code>H:M</code></td>
      <td><code>HOUR TO MINUTE</code></td>
      <td><code>INTERVAL '16:15' HOUR TO MINUTE</code></td>
    </tr>
    <tr>
      <td><code>H:M:S</code></td>
      <td><code>HOUR TO SECOND</code></td>
      <td><code>INTERVAL '16:15:14' HOUR TO SECOND</code></td>
    </tr>
    <tr>
      <td><code>M:S</code></td>
      <td><code>MINUTE TO SECOND</code></td>
      <td><code>INTERVAL '15:14' MINUTE TO SECOND</code></td>
    </tr>
  </tbody>
</table>

For example:

```googlesql
-- 0 years, 8 months, 20 days, 17 hours, 0 minutes, and 0 seconds (0-8 20 17:0:0)
INTERVAL '8 20 17' MONTH TO HOUR

-- 0 years, 8 months, -20 days, 17 hours, 0 minutes, and 0 seconds (0-8 -20 17:0:0)
INTERVAL '8 -20 17' MONTH TO HOUR
```

For additional examples, see [Interval literals][interval-literal-range].

#### Interval-supported date and time parts 
<a id="interval_datetime_parts"></a>

You can use the following date parts to construct an interval:

+ `YEAR`: Number of years, `Y`.
+ `QUARTER`: Number of quarters; each quarter is converted to `3` months, `M`.
+ `MONTH`: Number of months, `M`. Each `12` months is converted to `1` year.
+ `WEEK`: Number of weeks; Each week is converted to `7` days, `D`.
+ `DAY`: Number of days, `D`.

You can use the following time parts to construct an interval:

+ `HOUR`: Number of hours, `H`.
+ `MINUTE`: Number of minutes, `M`. Each `60` minutes is converted to `1` hour.
+ `SECOND`: Number of seconds, `S`. Each `60` seconds is converted to
  `1` minute. Can include up to nine fractional
  digits (nanosecond precision).
+ `MILLISECOND`: Number of milliseconds.
+ `MICROSECOND`: Number of microseconds.
+ `NANOSECOND`: Number of nanoseconds.

[interval-datetime-parts]: #interval_datetime_parts

[single-datetime-part-interval]: #single_datetime_part_interval

[range-datetime-part-interval]: #range_datetime_part_interval

[interval-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#interval_literals

[interval-literal-single]: https://github.com/google/googlesql/blob/master/docs/lexical.md#interval_literal_single

[interval-literal-range]: https://github.com/google/googlesql/blob/master/docs/lexical.md#interval_literal_range

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
