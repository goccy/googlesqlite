---
name: AI.SCORE
dialect: spanner
category: functions/ml
status: implemented
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_score
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_score
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/ml/ai_score.yaml
---

# AI.SCORE

## Summary

Returns a numeric score (typically in `[0, 1]`) from an AI model evaluating a natural-language prompt about an input.

## Signatures

- `AI.SCORE(input, prompt[, options])`

## Return type

`FLOAT64`.

## Behavior

- Higher values mean stronger match per the models scale.
- Networked, billable function.

## Examples

```sql
SELECT AI.SCORE(review_text, "Is this review spam?") FROM Reviews;
```

## Edge cases

- Determinism caveat: model upgrades can shift score distributions; downstream thresholds may need recalibration.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_score>.
