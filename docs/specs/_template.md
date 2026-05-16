---
name: EXAMPLE_FUNCTION
dialect: googlesql
category: functions/string
status: drafted
source_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#example_function
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/string_functions#example_function
last_synced: 2026-01-30
testdata: testdata/specs/googlesql/functions/string/example_function.yaml
---

# EXAMPLE_FUNCTION

> Replace this template with the real spec. Fields with curly braces are
> placeholders.

## Summary

A one-paragraph summary of what the function does.

## Signatures

- `EXAMPLE_FUNCTION(arg1, arg2)`

## Arguments

- `arg1`: STRING. ...
- `arg2`: INT64. ...

## Return type

STRING.

## Behavior

- Bullet points covering the function's defined behaviour.
- One bullet per observable rule.

## Examples

```sql
SELECT EXAMPLE_FUNCTION('a', 1);   -- 'a1'
SELECT EXAMPLE_FUNCTION(NULL, 0);  -- NULL
```

## Edge cases

- NULL handling.
- Type errors.
- Empty / boundary inputs.

## Reference (upstream)

The verbatim upstream section is preserved here so reviewers can spot
drift after `specctl upstream-sync`. Do not edit.

```text
(filled by `specctl normalize`)
```

## Status notes

(Required when `status` is `partial` or `unsupported`. Explain which
cases are intentionally unsupported and why.)

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/<file>.md`.
