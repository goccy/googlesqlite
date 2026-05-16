---
name: DATETIME
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#datetime-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/datetime.yaml
---

# DATETIME

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

## Datetime type 
<a id="datetime_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Range</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>DATETIME</code></td>
<td>
    
        0001-01-01 00:00:00 to 9999-12-31 23:59:59.999999999<br/>
        <hr/>
        0001-01-01 00:00:00 to 9999-12-31 23:59:59.999999<br/>
    
</td>
</tr>
</tbody>
</table>

A datetime value represents a Gregorian date and a time,
as they might be displayed on a watch, independent of time zone.
It includes the year, month, day, hour, minute, second,
and subsecond.
The range of subsecond precision is determined by the SQL engine.
To represent an absolute point in time,
use a [timestamp][timestamp-type].

##### Canonical format 
<a id="canonical_format_for_datetime_literals"></a>

```
civil_date_part[time_part]

civil_date_part:
    YYYY-[M]M-[D]D

time_part:
    { |T|t}[H]H:[M]M:[S]S[.F]
```

+ <code>YYYY</code>: Four-digit year.
+ <code>[M]M</code>: One or two digit month.
+ <code>[D]D</code>: One or two digit day.
+ <code>{ |T|t}</code>: A space or a `T` or `t` separator. The `T` and `t`
  separators are flags for time.
+ <code>[H]H</code>: One or two digit hour (valid values from 00 to 23).
+ <code>[M]M</code>: One or two digit minutes (valid values from 00 to 59).
+ <code>[S]S</code>: One or two digit seconds (valid values from 00 to 60).
+ <code>[.F]</code>: Up to nine fractional
  digits (nanosecond precision).

To learn more about the literal representation of a datetime type,
see [Datetime literals][datetime-literals].

[timestamp-type]: #timestamp_type

[datetime-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#datetime_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
