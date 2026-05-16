---
name: ENUM_VALUE_DESCRIPTOR_PROTO
dialect: googlesql
category: functions/proto
status: partial
notes: |
  The basic ENUM_VALUE_DESCRIPTOR_PROTO call lowers and returns a
  proto2.EnumValueDescriptorProto (see BindEnumValueDescriptorProto in
  internal/functions/proto/proto.go), so user-defined enums whose
  value names are reachable via the enum registry render the `name`
  and `number` fields correctly. The upstream Example, however,
  reads `ENUM_VALUE_DESCRIPTOR_PROTO(feature).options.(googlesql.language_feature_options).ideally_enabled`
  — that path requires (a) the internal `googlesql.LanguageFeature`
  enum with specific tagged values (999991 →
  FEATURE_TEST_IDEALLY_ENABLED_BUT_IN_DEVELOPMENT, 999992 →
  FEATURE_TEST_IDEALLY_DISABLED), (b) the
  `googlesql.language_feature_options` extension on
  google.protobuf.EnumValueOptions with `ideally_enabled` /
  `in_development` boolean fields, and (c) each enum value's
  options annotated accordingly. Reproducing those internal
  GoogleSQL descriptors is out of scope for the runtime-only
  surface; this spec stays partial until either a fixture for those
  protos is added or the upstream test is replaced with an
  enum_value_descriptor_proto example over a user-supplied enum.
source_url: docs/third_party/googlesql-docs/protocol_buffer_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/protocol_buffer_functions.md#enum_value_descriptor_proto
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/proto/enum_value_descriptor_proto.yaml
---

# ENUM_VALUE_DESCRIPTOR_PROTO

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

## `ENUM_VALUE_DESCRIPTOR_PROTO`

```googlesql
ENUM_VALUE_DESCRIPTOR_PROTO(proto_enum)
```

**Description**

Gets the enum value descriptor proto
(`proto2.EnumValueDescriptorProto`) for an enum.

**Definitions**

+   `proto_enum`: An `ENUM` value that contains the descriptor to retrieve.

**Return type**

`proto2.EnumValueDescriptorProto PROTO`

**Example**

The following query gets the `ideally_enabled` and `in_development` options from
the value descriptors in the `LanguageFeature` enum, and then produces query
results that are based on these value descriptors.

```googlesql
WITH
  EnabledFeatures AS (
    SELECT CAST(999991 AS googlesql.LanguageFeature) AS feature UNION ALL
    SELECT CAST(999992 AS googlesql.LanguageFeature) AS feature
  )
SELECT
  CAST(feature AS STRING) AS feature_enum_name,
  CAST(feature AS INT64) AS feature_enum_id,
  IFNULL(
    ENUM_VALUE_DESCRIPTOR_PROTO(feature).options.(googlesql.language_feature_options).ideally_enabled,
    TRUE) AS feature_is_ideally_enabled,
  IFNULL(
    ENUM_VALUE_DESCRIPTOR_PROTO(feature).options.(googlesql.language_feature_options).in_development,
    FALSE) AS feature_is_in_development
FROM
  EnabledFeatures;

/*-------------------------------------------------+-----------------+----------------------------+---------------------------+
 | feature_enum_name                               | feature_enum_id | feature_is_ideally_enabled | feature_is_in_development |
 +-------------------------------------------------+-----------------+----------------------------+---------------------------+
 | FEATURE_TEST_IDEALLY_ENABLED_BUT_IN_DEVELOPMENT | 999991          | TRUE                       | TRUE                      |
 | FEATURE_TEST_IDEALLY_DISABLED                   | 999992          | FALSE                      | FALSE                     |
 +-------------------------------------------------+-----------------+----------------------------+---------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/protocol_buffer_functions.md`.
