---
name: ADDDATE
dialect: spanner
category: functions/date
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/date_functions#adddate
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/date_functions#adddate
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/date/adddate.yaml
---

# ADDDATE

## Summary

Adds an interval to a date or datetime value. Equivalent to `DATE_ADD` (with date granularity) when the second argument is an integer.

## Signatures

- `ADDDATE(date, INTERVAL value unit)`
- `ADDDATE(date, days)`

## Arguments

- `date`: `DATE` or `DATETIME`.
- `INTERVAL value unit`: an interval expression (e.g. `INTERVAL 1 DAY`, `INTERVAL 3 MONTH`).
- `days`: `INT64` short-hand equivalent to `INTERVAL days DAY`.

## Return type

Same as the first argument (`DATE` in `DATE` out, `DATETIME` in `DATETIME` out).

## Behavior

- Returns `NULL` if any argument is `NULL`.
- Negative offsets subtract.

## Examples

```sql
SELECT ADDDATE(DATE "2024-03-15", 7);                    -- DATE "2024-03-22"
SELECT ADDDATE(DATE "2024-03-15", INTERVAL 1 MONTH);     -- DATE "2024-04-15"
```

## Edge cases

- `ADDDATE(date, INTERVAL n MONTH)` may "snap" the day-of-month at month-end (e.g. Jan 31 + 1 MONTH = Feb 29 in a leap year, Feb 28 otherwise).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/date_functions#adddate>.
