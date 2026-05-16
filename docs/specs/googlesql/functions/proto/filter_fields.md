---
name: FILTER_FIELDS
dialect: googlesql
category: functions/proto
status: implemented
source_url: docs/third_party/googlesql-docs/protocol_buffer_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/protocol_buffer_functions.md#filter_fields
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/proto/filter_fields.yaml
---

# FILTER_FIELDS

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

Verbatim copy from `docs/third_party/googlesql-docs/protocol_buffer_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `FILTER_FIELDS`

```googlesql
FILTER_FIELDS(
  proto_expression,
  proto_field_list
  [, reset_cleared_required_fields => { TRUE | FALSE } ]
)

proto_field_list:
  {+|-}proto_field_path[, ...]
```

**Description**

Takes a protocol buffer and a list of its fields to include or exclude.
Returns a version of that protocol buffer with unwanted fields removed.
Returns `NULL` if the protocol buffer is `NULL`.

Input values:

+ `proto_expression`: The protocol buffer to filter.
+ `proto_field_list`: The fields to exclude or include in the resulting
  protocol buffer.
+ `+`: Include a protocol buffer field and its children in the results.
+ `-`: Exclude a protocol buffer field and its children in the results.
+ `proto_field_path`: The protocol buffer field to include or exclude.
  If the field represents an [extension][querying-proto-extensions], you can use
  syntax for that extension in the path.
+ `reset_cleared_required_fields`: Named argument with a `BOOL` value.
  If not explicitly set, `FALSE` is used implicitly.
  If `FALSE`, you must include all protocol buffer `required` fields in the
  `FILTER_FIELDS` function. If `TRUE`, you don't need to include all required
  protocol buffer fields and the value of required fields
  defaults to these values:

  Type                    | Default value
  ----------------------- | --------
  Floating point          | `0.0`
  Integer                 | `0`
  Boolean                 | `FALSE`
  String, byte            | `""`
  Protocol buffer message | Empty message

Protocol buffer field expression behavior:

+ The first field in `proto_field_list` determines the default
  inclusion/exclusion. By default, when you include the first field, all other
  fields are excluded. Or by default, when you exclude the first field, all
  other fields are included.
+ A required field in the protocol buffer can't be excluded explicitly or
  implicitly, unless you have the
  `RESET_CLEARED_REQUIRED_FIELDS` named argument set as `TRUE`.
+ If a field is included, its child fields and descendants are implicitly
  included in the results.
+ If a field is excluded, its child fields and descendants are
  implicitly excluded in the results.
+ A child field must be listed after its parent field in the argument list,
  but doesn't need to come right after the parent field.

Caveats:

+ If you attempt to exclude/include a field that already has been
  implicitly excluded/included, an error is produced.
+ If you attempt to explicitly include/exclude a field that has already
  implicitly been included/excluded, an error is produced.

**Return type**

Type of `proto_expression`

**Examples**

The examples in this section reference a protocol buffer called `Award` and
a table called `MusicAwards`.

```proto
message Award {
  required int32 year = 1;
  optional int32 month = 2;
  repeated Type type = 3;

  message Type {
    optional string award_name = 1;
    optional string category = 2;
  }
}
```

```googlesql
WITH
  MusicAwards AS (
    SELECT
      CAST(
        '''
        year: 2001
        month: 9
        type { award_name: 'Best Artist' category: 'Artist' }
        type { award_name: 'Best Album' category: 'Album' }
        '''
        AS googlesql.examples.music.Award) AS award_col
    UNION ALL
    SELECT
      CAST(
        '''
        year: 2001
        month: 12
        type { award_name: 'Best Song' category: 'Song' }
        '''
        AS googlesql.examples.music.Award) AS award_col
  )
SELECT *
FROM MusicAwards

/*---------------------------------------------------------+
 | award_col                                               |
 +---------------------------------------------------------+
 | {                                                       |
 |   year: 2001                                            |
 |   month: 9                                              |
 |   type { award_name: "Best Artist" category: "Artist" } |
 |   type { award_name: "Best Album" category: "Album" }   |
 | }                                                       |
 | {                                                       |
 |   year: 2001                                            |
 |   month: 12                                             |
 |   type { award_name: "Best Song" category: "Song" }     |
 | }                                                       |
 +---------------------------------------------------------*/
```

The following example returns protocol buffers that only include the `year`
field.

```googlesql
SELECT FILTER_FIELDS(award_col, +year) AS filtered_fields
FROM MusicAwards

/*-----------------+
 | filtered_fields |
 +-----------------+
 | {year: 2001}    |
 | {year: 2001}    |
 +-----------------*/
```

The following example returns protocol buffers that include all but the `type`
field.

```googlesql
SELECT FILTER_FIELDS(award_col, -type) AS filtered_fields
FROM MusicAwards

/*------------------------+
 | filtered_fields        |
 +------------------------+
 | {year: 2001 month: 9}  |
 | {year: 2001 month: 12} |
 +------------------------*/
```

The following example returns protocol buffers that only include the `year` and
`type.award_name` fields.

```googlesql
SELECT FILTER_FIELDS(award_col, +year, +type.award_name) AS filtered_fields
FROM MusicAwards

/*--------------------------------------+
 | filtered_fields                      |
 +--------------------------------------+
 | {                                    |
 |   year: 2001                         |
 |   type { award_name: "Best Artist" } |
 |   type { award_name: "Best Album" }  |
 | }                                    |
 | {                                    |
 |   year: 2001                         |
 |   type { award_name: "Best Song" }   |
 | }                                    |
 +--------------------------------------*/
```

The following example returns the `year` and `type` fields, but excludes the
`award_name` field in the `type` field.

```googlesql
SELECT FILTER_FIELDS(award_col, +year, +type, -type.award_name) AS filtered_fields
FROM MusicAwards

/*---------------------------------+
 | filtered_fields                 |
 +---------------------------------+
 | {                               |
 |   year: 2001                    |
 |   type { category: "Artist" }   |
 |   type { category: "Album" }    |
 | }                               |
 | {                               |
 |   year: 2001                    |
 |   type { category: "Song" }     |
 | }                               |
 +---------------------------------*/
```

The following example produces an error because `year` is a required field
and can't be excluded explicitly or implicitly from the results.

```googlesql
SELECT FILTER_FIELDS(award_col, -year) AS filtered_fields
FROM MusicAwards

-- Error
```

The following example produces an error because when `year` was included,
`month` was implicitly excluded. You can't explicitly exclude a field that
has already been implicitly excluded.

```googlesql
SELECT FILTER_FIELDS(award_col, +year, -month) AS filtered_fields
FROM MusicAwards

-- Error
```

When `RESET_CLEARED_REQUIRED_FIELDS` is set as `TRUE`, `FILTER_FIELDS` doesn't
need to include required fields. In the example below, `MusicAwards` has a
required field called `year`, but this isn't added as an argument for
`FILTER_FIELDS`. `year` is added to the results with its default value, `0`.

```googlesql
SELECT FILTER_FIELDS(
  award_col,
  +month,
  RESET_CLEARED_REQUIRED_FIELDS => TRUE) AS filtered_fields
FROM MusicAwards;

/*---------------------------------+
 | filtered_fields                 |
 +---------------------------------+
 | {                               |
 |   year: 0,                      |
 |   month: 9                      |
 | }                               |
 | {                               |
 |   year: 0,                      |
 |   month: 12                     |
 | }                               |
 +---------------------------------*/
```

[querying-proto-extensions]: https://github.com/google/googlesql/blob/master/docs/protocol-buffers.md#extensions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/protocol_buffer_functions.md`.
