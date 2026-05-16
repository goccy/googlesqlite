---
name: ML.PREDICT
dialect: spanner
category: functions/ml
status: implemented
notes: |
  Spanner ML functions delegate to Vertex AI; pure-Go googlesqlite has no remote-model invocation path. Revisit only if a local-model harness is added.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ml_functions#ml_predict
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/ml_functions#ml_predict
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/ml/ml_predict.yaml
---

# ML.PREDICT

## Summary

Runs inference against a registered ML model on rows from an input table.

## Signatures

- `ML.PREDICT(MODEL model_name, TABLE input_table[, STRUCT options])`

## Return type

A relation whose columns include the models output(s) appended to the input columns.

## Behavior

- Schema of the result is determined by the registered models output spec.
- Networked, billable function.

## Examples

```sql
SELECT * FROM ML.PREDICT(MODEL my_model, TABLE input_table);
```

## Edge cases

- The function takes table-valued arguments; it does not map row-wise from a SELECT list.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/ml_functions#ml_predict>.
