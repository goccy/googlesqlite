---
name: JSON
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#json-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/json.yaml
---

# JSON

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

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## JSON type 
<a id="json_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>JSON</code></td>
<td>Represents JSON, a lightweight data-interchange format.</td>
</tr>
</tbody>
</table>

Expect these canonicalization behaviors when creating a value of JSON type:

+  Booleans, strings, and nulls are preserved exactly.
+  Whitespace characters aren't preserved.
+  A JSON value can store integers in the range of
   -9,223,372,036,854,775,808 (minimum signed 64-bit integer) to
   18,446,744,073,709,551,615 (maximum unsigned 64-bit integer) and
   floating point numbers within a domain of
   `DOUBLE`.
+  The order of elements in an array is preserved exactly.
+  The order of the members of an object isn't guaranteed or preserved.
+  If an object has duplicate keys, the first key that's found is preserved.
+  The format of the original string representation of a JSON number may not be
   preserved.

To learn more about the literal representation of a JSON type,
see [JSON literals][json-literals].

[json-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#json_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
