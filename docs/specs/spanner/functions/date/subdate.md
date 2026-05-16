---
name: SUBDATE
dialect: spanner
category: functions/date
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/date_functions#subdate
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/date_functions#subdate
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/date/subdate.yaml
---

# SUBDATE

## Summary

Subtracts an interval from a date or datetime value. Equivalent to `DATE_SUB`.

## Signatures

- `SUBDATE(date, INTERVAL value unit)`
- `SUBDATE(date, days)`

## Arguments

- `date`: `DATE` or `DATETIME`.
- `INTERVAL value unit`: an interval expression.
- `days`: `INT64` shorthand for `INTERVAL days DAY`.

## Return type

Same as the first argument.

## Behavior

- Returns `NULL` if any argument is `NULL`.
- Negative offsets add.

## Examples

```sql
SELECT SUBDATE(DATE "2024-03-15", 7);                    -- DATE "2024-03-08"
SELECT SUBDATE(DATE "2024-03-15", INTERVAL 1 MONTH);     -- DATE "2024-02-15"
```

## Edge cases

- Day-of-month snapping at month-end mirrors `ADDDATE`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/date_functions#subdate>.
