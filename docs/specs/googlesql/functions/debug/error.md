---
name: ERROR
dialect: googlesql
category: functions/debug
status: implemented
notes: |
  Two analyzer features (FeaturePipeIf, FeaturePipeAssert) plus a small
  formatter pass-through node (AssertScanNode) close out every upstream
  Example. The KeyValue compliance fixture is now seeded per-case via
  the testdata YAML `setup:` section, matching the `[prepare_database]`
  block in googlesql-wasm/compliance/testdata/pipe_if.test.
source_url: docs/third_party/googlesql-docs/debugging_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/debugging_functions.md#error
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/debug/error.yaml
---

# ERROR

## Summary

(TBD — refine from the upstream reference below.)

## Signatures

(TBD)

## Behavior

(TBD)

## Examples

(TBD)

## Edge cases

(TBD)

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/debugging_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ERROR`

```googlesql
ERROR(error_message)
```

**Description**

Returns an error.

**Definitions**

+   `error_message`: A `STRING` value that represents the error message to
    produce. Any whitespace characters beyond a
    single space are trimmed from the results.

**Details**

`ERROR` is treated like any other expression that may
result in an error: there is no special guarantee of evaluation order.

**Return Data Type**

GoogleSQL infers the return type in context.

**Examples**

In the following example, the query produces an error message:

```googlesql
-- ERROR: Show this error message (while evaluating error("Show this error message"))
SELECT ERROR('Show this error message')
```

In the following example, the query returns an error message if the value of the
row doesn't match one of two defined values.

```googlesql
SELECT
  CASE
    WHEN value = 'foo' THEN 'Value is foo.'
    WHEN value = 'bar' THEN 'Value is bar.'
    ELSE ERROR(CONCAT('Found unexpected value: ', value))
  END AS new_value
FROM (
  SELECT 'foo' AS value UNION ALL
  SELECT 'bar' AS value UNION ALL
  SELECT 'baz' AS value);

-- Found unexpected value: baz
```

The following example demonstrates bad usage of the `ERROR` function. In this
example, GoogleSQL might evaluate the `ERROR` function before or after
the <nobr>`x > 0`</nobr> condition, because GoogleSQL doesn't guarantee
ordering between `WHERE` clause conditions. Therefore, the results with the
`ERROR` function might vary.

```googlesql{.bad}
SELECT *
FROM (SELECT -1 AS x)
WHERE x > 0 AND ERROR('Example error');
```

In the next example, the `WHERE` clause evaluates an `IF` condition, which
ensures that GoogleSQL only evaluates the `ERROR` function if the
condition fails.

```googlesql
SELECT *
FROM (SELECT -1 AS x)
WHERE IF(x > 0, true, ERROR(FORMAT('Error: x must be positive but is %t', x)));

-- Error: x must be positive but is -1
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/debugging_functions.md`.
