---
name: AI.IF
dialect: spanner
category: functions/ml
status: implemented
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_if
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_if
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/ml/ai_if.yaml
---

# AI.IF

## Summary

Evaluates a natural-language predicate using a Spanner-managed AI model and returns one of two branches.

## Signatures

- `AI.IF(predicate_text, then_value, else_value[, options])`

## Arguments

- `predicate_text`: `STRING` natural-language condition.
- `then_value`, `else_value`: any matching SQL types.
- `options`: optional `STRUCT`.

## Return type

Same type as `then_value` / `else_value`.

## Behavior

- Returns `then_value` if the model judges the predicate true; `else_value` otherwise.
- Networked, billable function.

## Examples

```sql
SELECT AI.IF("The customer review expresses dissatisfaction", "escalate", "ignore");
```

## Edge cases

- Non-determinism: identical predicates may not always classify identically across model versions.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_if>.
