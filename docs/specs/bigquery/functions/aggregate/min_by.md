---
name: MIN_BY
dialect: bigquery
category: functions/aggregate
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#min_by
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#min_by
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aggregate/min_by.yaml
---

# MIN_BY

## Summary
Returns the value of `x` from the row in the group that has the smallest
value of `y`. The page documents `MIN_BY` as a synonym for
`ANY_VALUE(x HAVING MIN y)`.

## Signatures
- `MIN_BY(x, y)`

## Arguments
- `x`: the expression whose value is returned for the chosen row. Any
  type permitted by `ANY_VALUE`.
- `y`: the ordering key. The row with the minimum value of `y` in the
  group is chosen.

## Return type
Matches the data type of the input `x`.

## Behavior
- Equivalent to `ANY_VALUE(x HAVING MIN y)`; the row with the smallest
  `y` is selected and its `x` is returned.
- When several rows tie for the minimum `y`, the choice among the tied
  rows is nondeterministic (not random) — inherited from `ANY_VALUE`'s
  documented semantics.
- Inherits `ANY_VALUE`'s NULL handling: returns `NULL` when the input
  has no rows, and returns `NULL` when `x` or `y` is `NULL` for every
  row in the group. When `x` has at least one non-NULL value the
  function behaves as if `IGNORE NULLS` were specified.
- Because the function expands to an `ANY_VALUE ... HAVING` form, it
  cannot be combined with an `OVER` clause (the `HAVING MIN/MAX`
  modifier disallows windowing).

## Examples
```sql
WITH fruits AS (
  SELECT "apple"  AS fruit, 3.55 AS price UNION ALL
  SELECT "banana" AS fruit, 2.10 AS price UNION ALL
  SELECT "pear"   AS fruit, 4.30 AS price
)
SELECT MIN_BY(fruit, price) AS fruit
FROM fruits;
-- expected: fruit = 'banana'
```

## Edge cases
- Returns `NULL` when the input is empty, or when `y` is `NULL` for
  every row of the group.
- Ties on `y` are resolved nondeterministically; do not rely on a
  stable winner across runs.
- Cannot be used as a window function; there is no `OVER` form.

## Reference (upstream)
- https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#min_by
