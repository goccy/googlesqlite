---
name: AI.CLASSIFY
dialect: spanner
category: functions/ml
status: implemented
notes: |
  Spanner ML functions delegate to Vertex AI; pure-Go googlesqlite has no remote-model invocation path. Revisit only if a local-model harness is added.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_classify
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_classify
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/ml/ai_classify.yaml
---

# AI.CLASSIFY

## Summary

Classifies an input value into one of a fixed label set using a Spanner-managed AI model.

## Signatures

- `AI.CLASSIFY(input, labels[, options])`

## Arguments

- `input`: `STRING` natural-language input.
- `labels`: `ARRAY<STRING>` of candidate labels.
- `options`: optional `STRUCT` controlling model selection and confidence thresholds.

## Return type

`STRUCT<label STRING, score FLOAT64>` or similar — see upstream for the exact shape.

## Behavior

- Returns `NULL` if `input` is `NULL`.
- The selected label is always one of `labels` (no out-of-set classes).

## Examples

```sql
SELECT AI.CLASSIFY("This product is amazing", ["positive", "negative", "neutral"]);
```

## Edge cases

- Networked, billable function. Latency and pricing are model-dependent; consult upstream.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/ai_functions#ai_classify>.
