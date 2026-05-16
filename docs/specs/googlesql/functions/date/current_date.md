---
name: CURRENT_DATE
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#current_date
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/current_date.yaml
---

# CURRENT_DATE

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

Verbatim copy from `docs/third_party/googlesql-docs/date_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `CURRENT_DATE`

```googlesql
CURRENT_DATE()
```

```googlesql
CURRENT_DATE(time_zone_expression)
```

```googlesql
CURRENT_DATE
```

**Description**

Returns the current date as a `DATE` object. Parentheses are optional when
called with no arguments.

This function supports the following arguments:

+ `time_zone_expression`: A `STRING` expression that represents a
  [time zone][date-timezone-definitions]. If no time zone is specified, the
  default time zone, which is implementation defined, is used. If this expression is
  used and it evaluates to `NULL`, this function returns `NULL`.

The current date value is set at the start of the query statement that contains
this function. All invocations of `CURRENT_DATE()` within a query statement
yield the same value.

**Return Data Type**

`DATE`

**Examples**

The following query produces the current date in the default time zone:

```googlesql
SELECT CURRENT_DATE() AS the_date;

/*--------------+
 | the_date     |
 +--------------+
 | 2016-12-25   |
 +--------------*/
```

The following queries produce the current date in a specified time zone:

```googlesql
SELECT CURRENT_DATE('America/Los_Angeles') AS the_date;

/*--------------+
 | the_date     |
 +--------------+
 | 2016-12-25   |
 +--------------*/
```

```googlesql
SELECT CURRENT_DATE('-08') AS the_date;

/*--------------+
 | the_date     |
 +--------------+
 | 2016-12-25   |
 +--------------*/
```

The following query produces the current date in the default time zone.
Parentheses aren't needed if the function has no arguments.

```googlesql
SELECT CURRENT_DATE AS the_date;

/*--------------+
 | the_date     |
 +--------------+
 | 2016-12-25   |
 +--------------*/
```

[date-timezone-definitions]: https://github.com/google/googlesql/blob/master/docs/data-types.md#time_zones

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
