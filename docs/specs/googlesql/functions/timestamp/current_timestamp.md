---
name: CURRENT_TIMESTAMP
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#current_timestamp
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/current_timestamp.yaml
---

# CURRENT_TIMESTAMP

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

Verbatim copy from `docs/third_party/googlesql-docs/timestamp_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `CURRENT_TIMESTAMP`

```googlesql
CURRENT_TIMESTAMP()
```

```googlesql
CURRENT_TIMESTAMP
```

**Description**

Returns the current date and time as a timestamp object. The timestamp is
continuous, non-ambiguous, has exactly 60 seconds per minute and doesn't repeat
values over the leap second. Parentheses are optional.

This function handles leap seconds by smearing them across a window of 20 hours
around the inserted leap second.

The current timestamp value is set at the start of the query statement that
contains this function. All invocations of `CURRENT_TIMESTAMP()` within a query
statement yield the same value.

**Supported Input Types**

Not applicable

**Result Data Type**

`TIMESTAMP`

**Examples**

```googlesql
SELECT CURRENT_TIMESTAMP() AS now;

/*---------------------------------------------+
 | now                                         |
 +---------------------------------------------+
 | 2020-06-02 17:00:53.110 America/Los_Angeles |
 +---------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
