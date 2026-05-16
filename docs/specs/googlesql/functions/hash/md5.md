---
name: MD5
dialect: googlesql
category: functions/hash
status: implemented
source_url: docs/third_party/googlesql-docs/hash_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hash_functions.md#md5
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hash/md5.yaml
---

# MD5

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

## `MD5`

```
MD5(input)
```

**Description**

Computes the hash of the input using the
[MD5 algorithm][hash-link-to-md5-wikipedia]. The input can either be
`STRING` or `BYTES`. The string version treats the input as an array of bytes.

This function returns 16 bytes.

Warning: MD5 is no longer considered secure.
For increased security use another hashing function.

**Return type**

`BYTES`

**Example**

```googlesql
SELECT MD5("Hello World") as md5;

/*-------------------------------------------------+
 | md5                                             |
 +-------------------------------------------------+
 | \xb1\n\x8d\xb1d\xe0uA\x05\xb7\xa9\x9b\xe7.?\xe5 |
 +-------------------------------------------------*/
```

[hash-link-to-md5-wikipedia]: https://en.wikipedia.org/wiki/MD5

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hash_functions.md`.
