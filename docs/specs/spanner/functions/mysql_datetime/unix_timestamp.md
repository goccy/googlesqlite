---
name: UNIX_TIMESTAMP
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Implemented as a `mysql.*` sub-catalog alias of the GoogleSQL
    equivalent (registerSpannerExtensionFunctions in
    internal/catalog.go registers the analyzer signature; the
    formatter flattens `mysql.<name>` to `mysql_<name>` at SQLite
    dispatch). The Spanner dialect surface is shared with the
    default dialect — there is no per-DSN dialect plumbing, so
    the aliases are always reachable.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#unix_timestamp
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#unix_timestamp
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/unix_timestamp.yaml
---

# UNIX_TIMESTAMP

## Summary

Returns the Unix epoch seconds for the current time, or for a specified `TIMESTAMP`/`DATETIME`/`DATE`/`STRING` value.

## Signatures

- `UNIX_TIMESTAMP()`
- `UNIX_TIMESTAMP(value)`

## Arguments

- `value`: optional `DATE`, `DATETIME`, `TIMESTAMP`, or parseable date `STRING`.

## Return type

`INT64` (seconds since `1970-01-01 00:00:00 UTC`). Negative values represent times before the epoch.

## Behavior

- The 0-argument form is equivalent to `UNIX_SECONDS(CURRENT_TIMESTAMP())`.
- The 1-argument form treats `DATE`/`DATETIME` arguments as being in the session time zone.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
SELECT UNIX_TIMESTAMP(TIMESTAMP "2024-03-15 14:40:00+00");   -- 1710518400
SELECT UNIX_TIMESTAMP();                                      -- current epoch seconds
```

## Edge cases

- Strings that fail to parse return `NULL` (or an error, depending on session strictness).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#unix_timestamp>.
