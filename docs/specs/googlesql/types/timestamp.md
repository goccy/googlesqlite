---
name: TIMESTAMP
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#timestamp-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/timestamp.yaml
---

# TIMESTAMP

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

## Timestamp type 
<a id="timestamp_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Range</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>TIMESTAMP</code></td>

    <td>
      0001-01-01 00:00:00 to 9999-12-31 23:59:59.999999999 UTC<br/>
      <hr/>
      0001-01-01 00:00:00 to 9999-12-31 23:59:59.999999 UTC<br/>
    </td>

</tr>
</tbody>
</table>

A timestamp value represents an absolute point in time,
independent of any time zone or convention such as daylight saving time (DST),
with
microsecond, nanosecond, or picosecond
precision.
The range of subsecond precision is determined by the query engine.

A timestamp is typically represented internally as the number of elapsed
microseconds, nanoseconds, or picoseconds since a fixed initial point in time.

Note that a timestamp itself doesn't have a time zone; it represents the same
instant in time globally. However, the _display_ of a timestamp for human
readability usually includes a Gregorian date, a time, and a time zone, in an
implementation-dependent format. For example, the displayed values "2020-01-01
00:00:00 UTC", "2019-12-31 19:00:00 America/New_York", and "2020-01-01 05:30:00
Asia/Kolkata" all represent the same instant in time and therefore represent the
same timestamp value.

+  To represent a Gregorian date as it might appear on a calendar
   (a civil date), use a [date][date-type] value.
+  To represent a time as it might appear on a clock (a civil time),
   use a [time][time-type] value.
+  To represent a Gregorian date and time as they might appear on a watch,
   use a [datetime][datetime-type] value.

##### Canonical format 
<a id="canonical_format_for_timestamp_literals"></a>

The canonical format for a timestamp literal has the following parts:

```
{
  civil_date_part[time_part [time_zone]] |
  civil_date_part[time_part[time_zone_offset]] |
  civil_date_part[time_part[utc_time_zone]]
}

civil_date_part:
    YYYY-[M]M-[D]D

time_part:
    { |T|t}[H]H:[M]M:[S]S[.F]
```

+   <code>YYYY</code>: Four-digit year.
+   <code>[M]M</code>: One or two digit month.
+   <code>[D]D</code>: One or two digit day.
+   <code>{ |T|t}</code>: A space or a `T` or `t` separator. The `T` and `t`
    separators are flags for time.
+   <code>[H]H</code>: One or two digit hour (valid values from 00 to 23).
+   <code>[M]M</code>: One or two digit minutes (valid values from 00 to 59).
+   <code>[S]S</code>: One or two digit seconds (valid values from 00 to 60).
+   <code>[.F]</code>: Up to 12 fractional
    digits (picosecond precision).
+   <code>[time_zone]</code>: String representing the time zone. When a time
    zone isn't explicitly specified, the default time zone,
    which is implementation defined, is used. For details, see <a href="#time_zones">time
    zones</a>.
+   <code>[time_zone_offset]</code>: String representing the offset from the
    Coordinated Universal Time (UTC) time zone. For details, see
    <a href="#time_zones">time zones</a>.
+   <code>[utc_time_zone]</code>: String representing the Coordinated Universal
    Time (UTC), usually the letter `Z` or `z`. For details, see
    <a href="#time_zones">time zones</a>.

To learn more about the literal representation of a timestamp type,
see [Timestamp literals][timestamp-literals].

### Time zones 
<a id="time_zones"></a>

A time zone is used when converting from a civil date or time (as might appear
on a calendar or clock) to a timestamp (an absolute time), or vice versa. This
includes the operation of parsing a string containing a civil date and time like
"2020-01-01 00:00:00" and converting it to a timestamp. The resulting timestamp
value itself doesn't store a specific time zone, because it represents one
instant in time globally.

Time zones are represented by strings in one of these canonical formats:

+ Offset from Coordinated Universal Time (UTC), or the letter `Z` or `z` for
  UTC.
+ Time zone name from the
  [tz database][tz-database]{: class=external target=_blank }.

The following timestamps are identical because the time zone offset
for `America/Los_Angeles` is `-08` for the specified date and time.

```googlesql
SELECT UNIX_MILLIS(TIMESTAMP '2008-12-25 15:30:00 America/Los_Angeles') AS millis;
```

```googlesql
SELECT UNIX_MILLIS(TIMESTAMP '2008-12-25 15:30:00-08:00') AS millis;
```

#### Specify Coordinated Universal Time (UTC) 
<a id="utc"></a>

You can specify UTC using the following suffix:

```
{Z|z}
```

You can also specify UTC using the following time zone name:

```
{Etc/UTC}
```

The `Z` suffix is a placeholder that implies UTC when converting an [RFC
3339-format][rfc-3339-format] value to a `TIMESTAMP` value. The value `Z` isn't
a valid time zone for functions that accept a time zone. If you're specifying a
time zone, or you're unsure of the format to use to specify UTC, we recommend
using the `Etc/UTC` time zone name.

The `Z` suffix isn't case sensitive. When using the `Z` suffix, no space is
allowed between the `Z` and the rest of the timestamp. The following are
examples of using the `Z` suffix and the `Etc/UTC` time zone name:

```
SELECT TIMESTAMP '2014-09-27T12:30:00.45Z'
SELECT TIMESTAMP '2014-09-27 12:30:00.45z'
SELECT TIMESTAMP '2014-09-27T12:30:00.45 Etc/UTC'
```

#### Specify an offset from Coordinated Universal Time (UTC) 
<a id="utc_offset"></a>

You can specify the offset from UTC using the following format:

```
{+|-}H[H][:M[M]]
```

Examples:

```
-08:00
-8:15
+3:00
+07:30
-7
```

When using this format, no space is allowed between the time zone and the rest
of the timestamp.

```
2014-09-27 12:30:00.45-8:00
```

#### Time zone name {: #time_zone_name}

Format:

```
tz_identifier
```

A time zone name is a tz identifier from the
[tz database][tz-database]{: class=external target=_blank }.
For a less comprehensive but simpler reference, see the
[List of tz database time zones][tz-database-list]{: class=external target=_blank }
on Wikipedia.

Examples:

```
America/Los_Angeles
America/Argentina/Buenos_Aires
Etc/UTC
Pacific/Auckland
```

When using a time zone name, a space is required between the name and the rest
of the timestamp:

```
2014-09-27 12:30:00.45 America/Los_Angeles
```

Note that not all time zone names are interchangeable even if they do happen to
report the same time during a given part of the year. For example,
`America/Los_Angeles` reports the same time as `UTC-7:00` during daylight
saving time (DST), but reports the same time as `UTC-8:00` outside of DST.

If a time zone isn't specified, the default time zone value is used.

#### Leap seconds

A timestamp is simply an offset from 1970-01-01 00:00:00 UTC, assuming there are
exactly 60 seconds per minute. Leap seconds aren't represented as part of a
stored timestamp.

If the input contains values that use ":60" in the seconds field to represent a
leap second, that leap second isn't preserved when converting to a timestamp
value. Instead that value is interpreted as a timestamp with ":00" in the
seconds field of the following minute.

Leap seconds don't affect timestamp computations. All timestamp computations
are done using Unix-style timestamps, which don't reflect leap seconds. Leap
seconds are only observable through functions that measure real-world time. In
these functions, it's possible for a timestamp second to be skipped or repeated
when there is a leap second.

#### Daylight saving time

A timestamp is unaffected by daylight saving time (DST) because it represents a
point in time. When you display a timestamp as a civil time,
with a timezone that observes DST, the following rules apply:

+ During the transition from standard time to DST, one hour is skipped. A
  civil time from the skipped hour is treated the same as if it were written
  an hour later. For example, in the `America/Los_Angeles` time zone, the hour
  between 2 AM and 3 AM on March 10, 2024 is skipped on a clock. The times
  2:30 AM and 3:30 AM on that date are treated as the same point in time:

  ```googlesql
  SELECT
  FORMAT_TIMESTAMP("%c %Z", "2024-03-10 02:30:00 America/Los_Angeles", "UTC") AS two_thirty,
  FORMAT_TIMESTAMP("%c %Z", "2024-03-10 03:30:00 America/Los_Angeles", "UTC") AS three_thirty;

  /*------------------------------+------------------------------+
   | two_thirty                   | three_thirty                 |
   +------------------------------+------------------------------+
   | Sun Mar 10 10:30:00 2024 UTC | Sun Mar 10 10:30:00 2024 UTC |
   +------------------------------+------------------------------*/
  ```
+ When there's ambiguity in how to represent a civil time in a particular
  timezone because of DST, the later time is chosen:

  ```googlesql
  SELECT
  FORMAT_TIMESTAMP("%c %Z", "2024-03-10 10:30:00 UTC", "America/Los_Angeles") as ten_thirty;

  /*--------------------------------+
   | ten_thirty                     |
   +--------------------------------+
   | Sun Mar 10 03:30:00 2024 UTC-7 |
   +--------------------------------*/
  ```
+ During the transition from DST to standard time, one hour is repeated. A
  civil time that shows a time during that hour is treated as if it's the
  earlier instance of that time. For example, in the `America/Los_Angeles` time
  zone, the hour between 1 AM and 2 AM on November 3, 2024, is repeated on a
  clock. The time 1:30 AM on that date is treated as the earlier (DST) instance
  of that time.

  ```googlesql
  SELECT
  FORMAT_TIMESTAMP("%c %Z", "2024-11-03 01:30:00 America/Los_Angeles", "UTC") as one_thirty,
  FORMAT_TIMESTAMP("%c %Z", "2024-11-03 02:30:00 America/Los_Angeles", "UTC") as two_thirty;

  /*------------------------------+------------------------------+
   | one_thirty                   | two_thirty                   |
   +------------------------------+------------------------------+
   | Sun Nov 3 08:30:00 2024 UTC  | Sun Nov 3 10:30:00 2024 UTC  |
   +------------------------------+------------------------------*/
  ```

[rfc-3339-format]: https://datatracker.ietf.org/doc/html/rfc3339#page-10

[tz-database]: http://www.iana.org/time-zones

[tz-database-list]: http://en.wikipedia.org/wiki/List_of_tz_database_time_zones

[time-type]: #time_type

[date-type]: #date_type

[datetime-type]: #datetime_type

[timestamp-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#timestamp_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
