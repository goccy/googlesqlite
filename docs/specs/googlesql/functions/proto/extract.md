---
name: EXTRACT
dialect: googlesql
category: functions/proto
status: implemented
notes: |
  Testdata exercises `EXTRACT(FIELD(...) FROM ...)` and `EXTRACT(HAS(...) FROM ...)` over the auto-registered `google.type.Date` well-known proto. The upstream Examples (Album / chart_col / album_col / group_name) reference fictional proto types that have no Cloud BigQuery / Spanner counterpart; verbatim verification would need `CREATE PROTO BUNDLE` to ship user-defined `.proto` (tracked separately).
source_url: docs/third_party/googlesql-docs/protocol_buffer_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/protocol_buffer_functions.md#extract
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/proto/extract.yaml
---

# EXTRACT

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

## `EXTRACT` 
<a id="proto_extract"></a>

```googlesql
EXTRACT( extraction_type (proto_field) FROM proto_expression )

extraction_type:
  { FIELD | RAW | HAS | ONEOF_CASE }
```

**Description**

Extracts a value from a protocol buffer. `proto_expression` represents the
expression that returns a protocol buffer, `proto_field` represents the field of
the protocol buffer to extract from, and `extraction_type` determines the type
of data to return.

You can access most simple proto message fields idiomatically using the
[dot operator][dot-operator]. `EXTRACT` is a more general way to access fields
that can handle most cases. For instance, `EXTRACT` can access the values of
fields made ambiguous by tag reuse.

**Extraction Types**

You can choose the type of information to get with `EXTRACT`. Your choices are:

+  `FIELD`: Extract a value from a protocol buffer field.
+  `RAW`: Extract an uninterpreted value from a
    protocol buffer field. Raw values
    ignore any GoogleSQL type annotations.
+  `HAS`: Returns `TRUE` if a protocol buffer field is set in a proto message;
   otherwise, `FALSE`. Alternatively, use [`has_x`][has-value] to perform this
   task.
+  `ONEOF_CASE`: Returns the name of the set protocol buffer field in a Oneof.
   If no field is set, returns an empty string.

**Return Type**

The return type depends upon the extraction type in the query.

+  `FIELD`: Protocol buffer field type.
+  `RAW`: Protocol buffer field
    type. Format annotations are
    ignored.
+  `HAS`: `BOOL`
+  `ONEOF_CASE`: `STRING`

**Examples**

The examples in this section reference two protocol buffers called `Album` and
`Chart`, and one table called `AlbumList`.

```proto
message Album {
  optional string album_name = 1;
  repeated string song = 2;
  oneof group_name {
    string solo = 3;
    string duet = 4;
    string band = 5;
  }
}
```

```proto
message Chart {
  optional int64 date = 1 [(googlesql.format) = DATE];
  optional string chart_name = 2;
  optional int64 rank = 3;
}
```

```googlesql
WITH AlbumList AS (
  SELECT
    NEW Album(
      'Alana Yah' AS solo,
      'New Moon' AS album_name,
      ['Sandstorm','Wait'] AS song) AS album_col,
    NEW Chart(
      'Billboard' AS chart_name,
      '2016-04-23' AS date,
      1 AS rank) AS chart_col
    UNION ALL
  SELECT
    NEW Album(
      'The Roadlands' AS band,
      'Grit' AS album_name,
      ['The Way', 'Awake', 'Lost Things'] AS song) AS album_col,
    NEW Chart(
      'Billboard' AS chart_name,
      1 as rank) AS chart_col
)
SELECT * FROM AlbumList
```

The following example extracts the album names from a table called `AlbumList`
that contains a proto-typed column called `Album`.

```googlesql
SELECT EXTRACT(FIELD(album_name) FROM album_col) AS name_of_album
FROM AlbumList

/*------------------+
 | name_of_album    |
 +------------------+
 | New Moon         |
 | Grit             |
 +------------------*/
```

A table called `AlbumList` contains a proto-typed column called `Chart`.
`Chart` contains a field called `date`, which can store an integer. The
`date` field has an annotated format called `DATE` assigned to it, which means
that when you extract the value in this field, it returns a `DATE`, not an
`INT64`.

If you would like to return the value for `date` as an `INT64`, not
as a `DATE`, use the `RAW` extraction type in your query. For example:

```googlesql
SELECT
  EXTRACT(RAW(date) FROM chart_col) AS raw_date,
  EXTRACT(FIELD(date) FROM chart_col) AS formatted_date
FROM AlbumList

/*----------+----------------+
 | raw_date | formatted_date |
 +----------+----------------+
 | 16914    | 2016-04-23     |
 | 0        | 1970-01-01     |
 +----------+----------------*/
```

The following example checks to see if release dates exist in a table called
`AlbumList` that contains a protocol buffer called `Chart`.

```googlesql
SELECT EXTRACT(HAS(date) FROM chart_col) AS has_release_date
FROM AlbumList

/*------------------+
 | has_release_date |
 +------------------+
 | TRUE             |
 | FALSE            |
 +------------------*/
```

The following example extracts the group name that's assigned to an artist in
a table called `AlbumList`. The group name is set for exactly one
protocol buffer field inside of the `group_name` Oneof. The `group_name` Oneof
exists inside the `Album` protocol buffer.

```googlesql
SELECT EXTRACT(ONEOF_CASE(group_name) FROM album_col) AS artist_type
FROM AlbumList;

/*-------------+
 | artist_type |
 +-------------+
 | solo        |
 | band        |
 +-------------*/
```

[dot-operator]: https://github.com/google/googlesql/blob/master/docs/operators.md#field_access_operator

[has-value]: https://github.com/google/googlesql/blob/master/docs/protocol-buffers.md#checking_if_a_field_has_a_value

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/protocol_buffer_functions.md`.
