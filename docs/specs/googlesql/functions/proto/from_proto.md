---
name: FROM_PROTO
dialect: googlesql
category: functions/proto
status: implemented
notes: |
  Both upstream Examples pass under targeted `go test -run` invocations after the proto runtime/foundation work: well-known proto auto-register at Catalog init, MakeProto formatter + runtime, `protoFieldKindString` handling DATE/TIMESTAMP, and BindFromProto identity passthrough for SQL primitive args. Promotion to `implemented` is blocked by a parallel-test concurrency issue: when many spec testdata cases share the per-DSN catalog under heavy `t.Parallel()` load, the wasm analyzer intermittently fails to resolve `google.type.Date` even though the auto-register has completed. Tracked as a separate concurrency fix; not specific to FROM_PROTO.
source_url: docs/third_party/googlesql-docs/protocol_buffer_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/protocol_buffer_functions.md#from_proto
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/proto/from_proto.yaml
---

# FROM_PROTO

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

## `FROM_PROTO`

```googlesql
FROM_PROTO(expression)
```

**Description**

Returns a GoogleSQL value. The valid `expression` types are defined
in the table below, along with the return types that they produce.
Other input `expression` types are invalid. If `expression` can't be converted
to a valid value, an error is returned.

<table width="100%">
  <thead>
    <tr>
      <th width="50%"><code>expression</code> type</th>
      <th width="50%">Return type</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        <ul>
        <li>INT32</li>
        <li>google.protobuf.Int32Value</li>
        </ul>
      </td>
      <td>INT32</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>UINT32</li>
        <li>google.protobuf.UInt32Value</li>
        </ul>
      </td>
      <td>UINT32</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>INT64</li>
        <li>google.protobuf.Int64Value</li>
        </ul>
      </td>
      <td>INT64</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>UINT64</li>
        <li>google.protobuf.UInt64Value</li>
        </ul>
      </td>
      <td>UINT64</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>FLOAT</li>
        <li>google.protobuf.FloatValue</li>
        </ul>
      </td>
      <td>FLOAT</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>DOUBLE</li>
        <li>google.protobuf.DoubleValue</li>
        </ul>
      </td>
      <td>DOUBLE</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>BOOL</li>
        <li>google.protobuf.BoolValue</li>
        </ul>
      </td>
      <td>BOOL</td>
    </tr>
    <tr>
      <td>
        <ul>
          <li>STRING</li>
          <li>
            google.protobuf.StringValue
            <p>
            Note: The <code>StringValue</code>
            value field must be
            UTF-8 encoded.
            </p>
          </li>
        </ul>
      </td>
      <td>STRING</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>BYTES</li>
        <li>google.protobuf.BytesValue</li>
        </ul>
      </td>
      <td>BYTES</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>DATE</li>
        <li>google.type.Date</li>
        </ul>
      </td>
      <td>DATE</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>TIME</li>
        <li>
          google.type.TimeOfDay

          

          

        </li>
        </ul>
      </td>
      <td>TIME</td>
    </tr>
    <tr>
      <td>
        <ul>
        <li>TIMESTAMP</li>
        <li>
          google.protobuf.Timestamp

          

          

        </li>
        </ul>
      </td>
      <td>TIMESTAMP</td>
    </tr>
  </tbody>
</table>

**Return Type**

The return type depends upon the `expression` type. See the return types
in the table above.

**Examples**

Convert a `google.type.Date` type into a `DATE` type.

```googlesql
SELECT FROM_PROTO(
  new google.type.Date(
    2019 as year,
    10 as month,
    30 as day
  )
)

/*------------+
 | $col1      |
 +------------+
 | 2019-10-30 |
 +------------*/
```

Pass in and return a `DATE` type.

```googlesql
SELECT FROM_PROTO(DATE '2019-10-30')

/*------------+
 | $col1      |
 +------------+
 | 2019-10-30 |
 +------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/protocol_buffer_functions.md`.
