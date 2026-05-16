---
name: MAX_BY
dialect: bigquery
category: functions/aggregate
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#max_by
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#max_by
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/aggregate/max_by.yaml
---

# MAX_BY

## Summary

Aggregate that returns the value of `x` from the row whose `y` is
the maximum across the group. Synonym for
`ANY_VALUE(x HAVING MAX y)`.

## Signatures

- `MAX_BY(x, y)`

## Behavior

- Return type matches the input type of `x`.
- Inherits `ANY_VALUE` tie semantics — if multiple rows share the
  same max `y`, the chosen `x` is non-deterministic.
- Cannot be used as a window function (no `OVER`).

## Examples

```sql
WITH fruits AS (
  SELECT 'apple'  AS fruit, 3.55 AS price UNION ALL
  SELECT 'banana' AS fruit, 2.10 AS price UNION ALL
  SELECT 'pear'   AS fruit, 4.30 AS price
)
SELECT MAX_BY(fruit, price) AS fruit FROM fruits;
-- expected: pear
```

## Edge cases

- Ties on `y` produce a non-deterministic result (any row that
  ties wins).
- Ignores `NULL` `y` per `ANY_VALUE` semantics.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#max_by>.
