---
name: TIME
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#time-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/time.yaml
---

# TIME

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

## Time type 
<a id="time_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Range</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>TIME</code></td>

    <td>
        00:00:00 to 23:59:59.999999999<br/>
        <hr/>
        00:00:00 to 23:59:59.999999<br/>
    </td>

</tr>
</tbody>
</table>

A time value represents a time of day, as might be displayed on a clock,
independent of a specific date and time zone.
The range of
subsecond precision is determined by the
SQL engine. To represent
an absolute point in time, use a [timestamp][timestamp-type].

##### Canonical format 
<a id="canonical_format_for_time_literals"></a>

```
[H]H:[M]M:[S]S[.F]
```

+ <code>[H]H</code>: One or two digit hour (valid values from 00 to 23).
+ <code>[M]M</code>: One or two digit minutes (valid values from 00 to 59).
+ <code>[S]S</code>: One or two digit seconds (valid values from 00 to 60).
+ <code>[.F]</code>: Up to nine fractional
  digits (nanosecond precision).

To learn more about the literal representation of a time type,
see [Time literals][time-literals].

[timestamp-type]: #timestamp_type

[time-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#time_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
