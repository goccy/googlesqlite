---
name: DATE
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#date-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/date.yaml
---

# DATE

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

## Date type 
<a id="date_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Range</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>DATE</code></td>
<td>0001-01-01 to 9999-12-31.</td>
</tr>
</tbody>
</table>

The date type represents a Gregorian calendar date, independent of time zone. A
date value doesn't represent a specific 24-hour time period. Rather, a given
date value represents a different 24-hour period when interpreted in different
time zones, and may represent a shorter or longer day during daylight saving
time (DST) transitions.
To represent an absolute point in time,
use a [timestamp][timestamp-type].

##### Canonical format 
<a id="canonical_format_for_date_literals"></a>

```
YYYY-[M]M-[D]D
```

+ `YYYY`: Four-digit year.
+ `[M]M`: One or two digit month.
+ `[D]D`: One or two digit day.

To learn more about the literal representation of a date type,
see [Date literals][date-literals].

[timestamp-type]: #timestamp_type

[date-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#date_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
