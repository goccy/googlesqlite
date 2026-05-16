---
name: NET.PUBLIC_SUFFIX
dialect: googlesql
category: functions/net
status: implemented
source_url: docs/third_party/googlesql-docs/net_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/net_functions.md#netpublic_suffix
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/net/net_public_suffix.yaml
---

# NET.PUBLIC_SUFFIX

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

Verbatim copy from `docs/third_party/googlesql-docs/net_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `NET.PUBLIC_SUFFIX`

```
NET.PUBLIC_SUFFIX(url)
```

**Description**

Takes a URL as a `STRING` value and returns the public suffix (such as `com`,
`org`, or `net`). A public suffix is an ICANN domain registered at
[publicsuffix.org][net-link-to-public-suffix]. For best results, URL values
should comply with the format as defined by
[RFC 3986][net-link-to-rfc-3986-appendix-a]. If the URL value doesn't comply
with RFC 3986 formatting, this function makes a best effort to parse the input
and return a relevant result.

This function returns `NULL` if any of the following is true:

+ It can't parse the host from the input;
+ The parsed host contains adjacent dots in the middle
  (not leading or trailing);
+ The parsed host doesn't contain any public suffix.

Before looking up the public suffix, this function temporarily normalizes the
host by converting uppercase English letters to lowercase and encoding all
non-ASCII characters with [Punycode][net-link-to-punycode].
The function then returns the public suffix as part of the original host instead
of the normalized host.

Note: The function doesn't perform
[Unicode normalization][unicode-normalization].

Note: The public suffix data at
[publicsuffix.org][net-link-to-public-suffix] also contains
private domains. This function ignores the private domains.

Note: The public suffix data may change over time. Consequently, input that
produces a `NULL` result now may produce a non-`NULL` value in the future.

**Return Data Type**

`STRING`

**Example**

```googlesql
SELECT
  FORMAT("%T", input) AS input,
  description,
  FORMAT("%T", NET.HOST(input)) AS host,
  FORMAT("%T", NET.PUBLIC_SUFFIX(input)) AS suffix,
  FORMAT("%T", NET.REG_DOMAIN(input)) AS domain
FROM (
  SELECT "" AS input, "invalid input" AS description
  UNION ALL SELECT "http://abc.xyz", "standard URL"
  UNION ALL SELECT "//user:password@a.b:80/path?query",
                   "standard URL with relative scheme, port, path and query, but no public suffix"
  UNION ALL SELECT "https://[::1]:80", "standard URL with IPv6 host"
  UNION ALL SELECT "http://例子.卷筒纸.中国", "standard URL with internationalized domain name"
  UNION ALL SELECT "    www.Example.Co.UK    ",
                   "non-standard URL with spaces, upper case letters, and without scheme"
  UNION ALL SELECT "mailto:?to=&subject=&body=", "URI rather than URL--unsupported"
);
```

| input                                                              | description                                                                   | host               | suffix  | domain         |
|--------------------------------------------------------------------|-------------------------------------------------------------------------------|--------------------|---------|----------------|
| ""                                                                 | invalid input                                                                 | NULL               | NULL    | NULL           |
| "http://abc.xyz"                                                   | standard URL                                                                  | "abc.xyz"          | "xyz"   | "abc.xyz"      |
| "//user:password@a.b:80/path?query"                                | standard URL with relative scheme, port, path and query, but no public suffix | "a.b"              | NULL    | NULL           |
| "https://[::1]:80"                                                 | standard URL with IPv6 host                                                   | "[::1]"            | NULL    | NULL           |
| "http://例子.卷筒纸.中国"                                            | standard URL with internationalized domain name                               | "例子.卷筒纸.中国"    | "中国"  | "卷筒纸.中国"     |
| "&nbsp;&nbsp;&nbsp;&nbsp;www.Example.Co.UK&nbsp;&nbsp;&nbsp;&nbsp;"| non-standard URL with spaces, upper case letters, and without scheme          | "www.Example.Co.UK"| "Co.UK" | "Example.Co.UK |
| "mailto:?to=&subject=&body="                                       | URI rather than URL--unsupported                                              | "mailto"           | NULL    | NULL           |

[unicode-normalization]: https://en.wikipedia.org/wiki/Unicode_equivalence

[net-link-to-punycode]: https://en.wikipedia.org/wiki/Punycode

[net-link-to-public-suffix]: https://publicsuffix.org/list/

[net-link-to-rfc-3986-appendix-a]: https://tools.ietf.org/html/rfc3986#appendix-A

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/net_functions.md`.
