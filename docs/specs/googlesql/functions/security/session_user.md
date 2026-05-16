---
name: SESSION_USER
dialect: googlesql
category: functions/security
status: implemented
source_url: docs/third_party/googlesql-docs/security_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/security_functions.md#session_user
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/security/session_user.yaml
---

# SESSION_USER

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

Verbatim copy from `docs/third_party/googlesql-docs/security_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `SESSION_USER`

```
SESSION_USER()
```

**Description**

For first-party users, returns the email address of the user that's running the
query.
For third-party users, returns the
[principal identifier](https://cloud.google.com/iam/docs/principal-identifiers)
of the user that's running the query.
For more information about identities, see
[Principals](https://cloud.google.com/docs/authentication#principal).

**Return Data Type**

`STRING`

**Example**

```googlesql
SELECT SESSION_USER() as user;

/*----------------------+
 | user                 |
 +----------------------+
 | jdoe@example.com     |
 +----------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/security_functions.md`.
