---
name: TEXT_ANALYZE
dialect: bigquery
category: functions/text_analysis
status: implemented
notes: |
  Self-registered through registerBigQueryExtensionFunctions.
  LOG_ANALYZER performs NFKC normalisation, lowercase folding,
  splits on any non-alphanumeric Unicode codepoint, and emits one
  token per CJK ideograph (Han / Hiragana / Katakana / Hangul) to
  match BigQuery's per-character CJK tokenisation. NO_OP_ANALYZER
  returns the input verbatim; PATTERN_ANALYZER tokenises via the
  regex list in analyzer_options ({"patterns": ["regex", ...]}).
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#text_analyze
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#text_analyze
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/text_analysis/text_analyze.yaml
---

# TEXT_ANALYZE

## Summary

Tokenises a `STRING` into an `ARRAY<STRING>` (tokenized document)
using the chosen text analyzer.

## Signatures

- `TEXT_ANALYZE(text [, analyzer => 'LOG_ANALYZER' | 'NO_OP_ANALYZER' | 'PATTERN_ANALYZER'] [, analyzer_options => '<json>'])`

## Behavior

- `analyzer` defaults to `'LOG_ANALYZER'` (delimiter-based with
  normalisation).
- `'NO_OP_ANALYZER'` returns the input as a single token without
  normalisation.
- `'PATTERN_ANALYZER'` extracts tokens matching a regex supplied
  via `analyzer_options`.
- `analyzer_options` is a JSON-formatted `STRING` containing
  analyzer-specific rules.
- Token order is unspecified.

## Examples

```sql
SELECT TEXT_ANALYZE('The quick brown FOX')
-- LOG_ANALYZER yields tokens like ['the', 'quick', 'brown', 'fox']
```

## Edge cases

- Token order is not guaranteed.
- `PATTERN_ANALYZER` requires a regex via `analyzer_options`.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#text_analyze>.
