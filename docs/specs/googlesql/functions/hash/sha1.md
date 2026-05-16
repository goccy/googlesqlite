---
name: SHA1
dialect: googlesql
category: functions/hash
status: implemented
source_url: docs/third_party/googlesql-docs/hash_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hash_functions.md#sha1
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hash/sha1.yaml
---

# SHA1

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

Verbatim copy from `docs/third_party/googlesql-docs/hash_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `SHA1`

```
SHA1(input)
```

**Description**

Computes the hash of the input using the
[SHA-1 algorithm][hash-link-to-sha-1-wikipedia]. The input can either be
`STRING` or `BYTES`. The string version treats the input as an array of bytes.

This function returns 20 bytes.

Warning: SHA1 is no longer considered secure.
For increased security, use another hashing function.

**Return type**

`BYTES`

**Example**

```googlesql
SELECT SHA1("Hello World") as sha1;

/*-----------------------------------------------------------+
 | sha1                                                      |
 +-----------------------------------------------------------+
 | \nMU\xa8\xd7x\xe5\x02/\xabp\x19w\xc5\xd8@\xbb\xc4\x86\xd0 |
 +-----------------------------------------------------------*/
```

[hash-link-to-sha-1-wikipedia]: https://en.wikipedia.org/wiki/SHA-1

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hash_functions.md`.
