package internal

import (
	"fmt"

	"google.golang.org/genproto/googleapis/type/color"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/latlng"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// wellKnownProtoMessages lists the protobuf well-known types that
// every googlesqlite catalog auto-registers at construction time.
// Hosts (BigQuery / Spanner) treat these as built-in, so user SQL
// referencing them (e.g. `new google.type.Date(...)` or
// `FROM_PROTO(new google.protobuf.Int64Value(...))`) resolves without
// the caller having to wire descriptors manually.
//
// Each entry is a *zero-value* instance of the well-known message.
// We walk it once to extract the parent FileDescriptorProto, register
// the file with the catalog's DescriptorPool, then pin the message
// type on the catalog so the analyzer can look it up by full name.
func wellKnownProtoMessages() []protoreflect.Message {
	return []protoreflect.Message{
		// google.protobuf.* wrappers
		(&wrapperspb.Int32Value{}).ProtoReflect(),
		(&wrapperspb.Int64Value{}).ProtoReflect(),
		(&wrapperspb.UInt32Value{}).ProtoReflect(),
		(&wrapperspb.UInt64Value{}).ProtoReflect(),
		(&wrapperspb.FloatValue{}).ProtoReflect(),
		(&wrapperspb.DoubleValue{}).ProtoReflect(),
		(&wrapperspb.BoolValue{}).ProtoReflect(),
		(&wrapperspb.StringValue{}).ProtoReflect(),
		(&wrapperspb.BytesValue{}).ProtoReflect(),
		// google.protobuf.* timing
		(&timestamppb.Timestamp{}).ProtoReflect(),
		(&durationpb.Duration{}).ProtoReflect(),
		// google.protobuf.* misc
		(&emptypb.Empty{}).ProtoReflect(),
		(&fieldmaskpb.FieldMask{}).ProtoReflect(),
		(&anypb.Any{}).ProtoReflect(),
		(&structpb.Struct{}).ProtoReflect(),
		(&structpb.Value{}).ProtoReflect(),
		(&structpb.ListValue{}).ProtoReflect(),
		// google.type.* googleapis canonical types
		(&date.Date{}).ProtoReflect(),
		(&timeofday.TimeOfDay{}).ProtoReflect(),
		(&money.Money{}).ProtoReflect(),
		(&latlng.LatLng{}).ProtoReflect(),
		(&color.Color{}).ProtoReflect(),
	}
}

// wellKnownProtoEnums lists enum types that should also be visible to
// the analyzer without explicit registration. `google.protobuf.NullValue`
// is the canonical example used by struct.proto.
func wellKnownProtoEnums() []protoreflect.EnumType {
	return []protoreflect.EnumType{
		structpb.NullValue(0).Type(),
		dayofweek.DayOfWeek(0).Type(),
	}
}

// registerWellKnownProtos plugs the curated set of well-known proto
// messages and enums into the catalog's DescriptorPool. Called from
// NewCatalog once the pool exists. Failures are returned so the
// catalog construction can surface them, but in practice every
// well-known type already compiles against `google.golang.org/protobuf`
// or `google.golang.org/genproto/googleapis/type/*` so descriptor
// marshalling never errors at runtime.
func registerWellKnownProtos(c *Catalog) error {
	if c == nil || c.descriptorPool == nil {
		return nil
	}
	seenFiles := map[string]bool{}
	for _, msg := range wellKnownProtoMessages() {
		desc := msg.Descriptor()
		if err := registerFileTransitive(c, desc.ParentFile(), seenFiles); err != nil {
			return fmt.Errorf("register file for %s: %w", desc.FullName(), err)
		}
		if _, err := c.RegisterProtoMessage(string(desc.FullName())); err != nil {
			return fmt.Errorf("register message %s: %w", desc.FullName(), err)
		}
	}
	for _, et := range wellKnownProtoEnums() {
		desc := et.Descriptor()
		if err := registerFileTransitive(c, desc.ParentFile(), seenFiles); err != nil {
			return fmt.Errorf("register file for enum %s: %w", desc.FullName(), err)
		}
		if _, err := c.RegisterEnum(string(desc.FullName())); err != nil {
			return fmt.Errorf("register enum %s: %w", desc.FullName(), err)
		}
	}
	return nil
}

// registerFileTransitive walks the FileDescriptor's import graph and
// registers each file once via Catalog.RegisterProto. Dependencies are
// registered before their dependents so the underlying DescriptorPool
// always sees imports resolved by the time it parses a dependent
// FileDescriptorProto.
func registerFileTransitive(c *Catalog, fd protoreflect.FileDescriptor, seen map[string]bool) error {
	if fd == nil {
		return nil
	}
	path := fd.Path()
	if seen[path] {
		return nil
	}
	seen[path] = true
	imports := fd.Imports()
	for i := 0; i < imports.Len(); i++ {
		if err := registerFileTransitive(c, imports.Get(i).FileDescriptor, seen); err != nil {
			return err
		}
	}
	fdProto := protodesc.ToFileDescriptorProto(fd)
	bytes, err := proto.Marshal(fdProto)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	return c.RegisterProto(bytes)
}
