---
name: SNIPPET
dialect: spanner
category: functions/search
status: implemented
notes: |
  Returns a +/-20 character window around the first match of the query in the raw text. Runtime entry: BindSpannerSnippet in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#snippet
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#snippet
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/snippet.yaml
---

# SNIPPET

## Summary

Returns a contextual snippet from the original text where the query matched, suitable for highlighted preview rendering.

## Signatures

- `SNIPPET(text, query[, options])`

## Arguments

- `text`: `STRING` the original document text.
- `query`: `STRING` query that produced the match.
- `options`: optional `STRUCT` controlling snippet length, highlight markers, etc.

## Return type

`STRING` containing the snippet, with matched terms wrapped in highlight markers per `options`.

## Behavior

- The function does not re-evaluate matching; it expects to be paired with an upstream `SEARCH` predicate that already filtered the row.
- Returns `NULL` if any required argument is `NULL`.

## Examples

```sql
SELECT SNIPPET(body, "machine learning") FROM Articles
WHERE SEARCH(body_full, "machine learning");
```

## Edge cases

- A query that does not appear in `text` still returns a non-`NULL` snippet (typically the start of `text`); guard with a `SEARCH` predicate.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#snippet>.
