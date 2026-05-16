package internal

import (
	"fmt"
	"sync"

	googlesql "github.com/goccy/go-googlesql"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	protofn "github.com/goccy/googlesqlite/internal/functions/proto"
)

// Proto / Enum types created from a registered descriptor cannot be
// rebuilt through TypeFactory.MakeSimpleType; the catalog has to hand
// back the same *ProtoType / *EnumType handle that was minted at
// registration time. Otherwise downstream paths like the analyzer's
// column-type resolution walk through a SimpleType placeholder and
// crash when they try to dereference the underlying descriptor.
//
// We mirror Catalog.registeredProtos / registeredEnums in a
// process-global map so call sites that don't already hold a Catalog
// pointer (in particular Type.ToGoogleSQLType, which receives a
// JSON-deserialised TableSpec from the catalog DB) can still reach
// the right handle by fully-qualified name.

var (
	protoRegistryMu sync.RWMutex
	protoRegistry   = map[string]*googlesql.ProtoType{}
	enumRegistry    = map[string]*googlesql.EnumType{}

	// goProtoFilesMu / goProtoFiles is the process-global Go-side
	// proto descriptor pool. Every FileDescriptorProto registered
	// through Catalog.RegisterProto is also Go-built via
	// protodesc.NewFile and added here, so runtime code (in
	// particular CAST(STRING AS Proto) which parses proto-text-format
	// against a live MessageType) can resolve descriptors without
	// crossing the wasm bridge.
	goProtoFilesMu sync.RWMutex
	goProtoFiles   = new(protoregistry.Files)
	goProtoTypesMu sync.RWMutex
	goProtoTypes   = map[string]protoreflect.MessageType{}
)

// registerProtoTypeGlobal pins a *ProtoType under its full name so
// later ToGoogleSQLType calls can recover the original handle. Safe to
// call multiple times — later writes overwrite earlier ones, which is
// the intended behaviour when a descriptor is re-registered.
func registerProtoTypeGlobal(name string, pt *googlesql.ProtoType) {
	if name == "" || pt == nil {
		return
	}
	protoRegistryMu.Lock()
	protoRegistry[name] = pt
	protoRegistryMu.Unlock()
}

func registerEnumTypeGlobal(name string, et *googlesql.EnumType) {
	if name == "" || et == nil {
		return
	}
	protoRegistryMu.Lock()
	enumRegistry[name] = et
	protoRegistryMu.Unlock()
}

func lookupRegisteredProtoType(name string) *googlesql.ProtoType {
	protoRegistryMu.RLock()
	defer protoRegistryMu.RUnlock()
	return protoRegistry[name]
}

func lookupRegisteredEnumType(name string) *googlesql.EnumType {
	protoRegistryMu.RLock()
	defer protoRegistryMu.RUnlock()
	return enumRegistry[name]
}

// registerGoProtoFileFromBytes parses a serialised
// google.protobuf.FileDescriptorProto and registers every message and
// enum it carries in the Go-side proto registry. The resolver chains
// the project-local goProtoFiles registry with protoregistry.GlobalFiles
// so imports of well-known types (google.protobuf.Timestamp, etc.)
// resolve against the binary's compiled-in descriptors.
//
// Safe to call multiple times for the same file — duplicate
// registrations are silently ignored.
func registerGoProtoFileFromBytes(fdBytes []byte) error {
	if len(fdBytes) == 0 {
		return nil
	}
	var fdProto descriptorpb.FileDescriptorProto
	if err := proto.Unmarshal(fdBytes, &fdProto); err != nil {
		return fmt.Errorf("registerGoProtoFile: unmarshal: %w", err)
	}
	path := fdProto.GetName()
	goProtoFilesMu.RLock()
	if existing, err := goProtoFiles.FindFileByPath(path); err == nil && existing != nil {
		goProtoFilesMu.RUnlock()
		return nil
	}
	goProtoFilesMu.RUnlock()
	resolver := chainedFileResolver{primary: goProtoFiles, fallback: protoregistry.GlobalFiles}
	fd, err := protodesc.NewFile(&fdProto, resolver)
	if err != nil {
		return fmt.Errorf("registerGoProtoFile: NewFile: %w", err)
	}
	goProtoFilesMu.Lock()
	if existing, err := goProtoFiles.FindFileByPath(path); err == nil && existing != nil {
		goProtoFilesMu.Unlock()
		return nil
	}
	if err := goProtoFiles.RegisterFile(fd); err != nil {
		goProtoFilesMu.Unlock()
		return fmt.Errorf("registerGoProtoFile: RegisterFile: %w", err)
	}
	goProtoFilesMu.Unlock()
	indexFileMessageTypes(fd)
	return nil
}

// indexFileMessageTypes walks every message (including nested) in the
// given file and pins a dynamicpb-backed MessageType under its full
// name in goProtoTypes. The dynamicpb MessageType is enough for
// prototext.Unmarshal to construct the Message and for proto.Marshal
// to serialise it back to wire bytes.
func indexFileMessageTypes(fd protoreflect.FileDescriptor) {
	msgs := fd.Messages()
	for i := 0; i < msgs.Len(); i++ {
		walkProtoMessage(msgs.Get(i))
	}
}

func walkProtoMessage(md protoreflect.MessageDescriptor) {
	name := string(md.FullName())
	mt := dynamicpb.NewMessageType(md)
	goProtoTypesMu.Lock()
	goProtoTypes[name] = mt
	goProtoTypesMu.Unlock()
	nested := md.Messages()
	for i := 0; i < nested.Len(); i++ {
		walkProtoMessage(nested.Get(i))
	}
}

// lookupGoProtoMessageType returns the MessageType registered under
// the given full name, or nil when no descriptor for that name has
// been registered yet. protoregistry.GlobalTypes is consulted as a
// fallback so the well-known wrappers (google.protobuf.Int64Value, …)
// resolve without the catalog explicitly registering them.
func lookupGoProtoMessageType(name string) protoreflect.MessageType {
	goProtoTypesMu.RLock()
	mt := goProtoTypes[name]
	goProtoTypesMu.RUnlock()
	if mt != nil {
		return mt
	}
	if mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(name)); err == nil {
		return mt
	}
	return nil
}

// parseProtoTextLiteral parses a proto-text-format payload against
// the message descriptor registered under fullName and returns the
// binary-encoded (wire-format) bytes. Used by CastValue when lowering
// `CAST(STRING AS Proto)`.
//
// The resolver chains the project-local goProtoTypes registry with
// protoregistry.GlobalTypes so nested message references can be
// resolved transitively.
func parseProtoTextLiteral(fullName string, text string) ([]byte, error) {
	mt := lookupGoProtoMessageType(fullName)
	if mt == nil {
		return nil, fmt.Errorf("no Go-side descriptor for proto %q", fullName)
	}
	msg := mt.New().Interface()
	opts := prototext.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
		Resolver:       chainedTypeResolver{},
	}
	if err := opts.Unmarshal([]byte(text), msg); err != nil {
		return nil, fmt.Errorf("prototext.Unmarshal: %w", err)
	}
	return proto.MarshalOptions{AllowPartial: true, Deterministic: true}.Marshal(msg)
}

// protoNameFromGoogleSQLType extracts the dotted full name from a
// googlesql proto type. The underlying DebugString prints
// "PROTO<google.protobuf.Timestamp>"; protoNameFromDebug strips that
// wrapper. Returns the empty string for non-proto types.
func protoNameFromGoogleSQLType(t googlesql.Googlesql_TypeNode) string {
	if t == nil {
		return ""
	}
	k, err := t.Kind()
	if err != nil || k != googlesql.TypeKindTypeProto {
		return ""
	}
	ds, _ := t.DebugString(false)
	return protoNameFromDebug(ds)
}

// requiredFieldsOf returns each top-level REQUIRED field of `fullName`
// as a list of (number, wire-type, default-payload) records. Used to
// fill in FILTER_FIELDS' RESET_CLEARED_REQUIRED_FIELDS => TRUE
// behaviour without the runtime having to walk the descriptor itself.
func requiredFieldsOf(fullName string) []protofn.RequiredField {
	mt := lookupGoProtoMessageType(fullName)
	if mt == nil {
		return nil
	}
	desc := mt.Descriptor()
	fields := desc.Fields()
	var out []protofn.RequiredField
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if fd.Cardinality() != protoreflect.Required {
			continue
		}
		out = append(out, protofn.RequiredField{
			Number:  uint64(fd.Number()),
			Wire:    uint8(wireForFieldKind(fd.Kind())),
			Payload: defaultPayloadForFieldKind(fd.Kind()),
		})
	}
	return out
}

func wireForFieldKind(k protoreflect.Kind) protowire.Type {
	switch k {
	case protoreflect.BoolKind, protoreflect.EnumKind,
		protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind:
		return protowire.VarintType
	case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		return protowire.Fixed32Type
	case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		return protowire.Fixed64Type
	}
	return protowire.BytesType
}

func defaultPayloadForFieldKind(k protoreflect.Kind) []byte {
	switch k {
	case protoreflect.BoolKind, protoreflect.EnumKind,
		protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind:
		return protowire.AppendVarint(nil, 0)
	case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		return protowire.AppendFixed32(nil, 0)
	case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		return protowire.AppendFixed64(nil, 0)
	case protoreflect.StringKind, protoreflect.BytesKind, protoreflect.MessageKind, protoreflect.GroupKind:
		return protowire.AppendVarint(nil, 0)
	}
	return nil
}

func init() {
	protofn.RequiredFieldResolver = requiredFieldsOf
}

// chainedTypeResolver implements prototext's Resolver by chaining
// project-local goProtoTypes with protoregistry.GlobalTypes.
type chainedTypeResolver struct{}

func (chainedTypeResolver) FindMessageByName(name protoreflect.FullName) (protoreflect.MessageType, error) {
	if mt := lookupGoProtoMessageType(string(name)); mt != nil {
		return mt, nil
	}
	return protoregistry.GlobalTypes.FindMessageByName(name)
}

func (chainedTypeResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	return protoregistry.GlobalTypes.FindMessageByURL(url)
}

func (chainedTypeResolver) FindExtensionByName(name protoreflect.FullName) (protoreflect.ExtensionType, error) {
	return protoregistry.GlobalTypes.FindExtensionByName(name)
}

func (chainedTypeResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	return protoregistry.GlobalTypes.FindExtensionByNumber(message, field)
}

// messageFullNamesFromGoFile parses the FileDescriptorProto bytes
// and returns the fully-qualified names of every message it declares
// (including nested messages, except synthetic map-entry messages
// which are flagged map_entry=true). Used by Catalog.RegisterProto
// to auto-promote each message into a SQL-visible proto type without
// the caller having to enumerate them.
func messageFullNamesFromGoFile(fdBytes []byte) []string {
	if len(fdBytes) == 0 {
		return nil
	}
	var fdProto descriptorpb.FileDescriptorProto
	if err := proto.Unmarshal(fdBytes, &fdProto); err != nil {
		return nil
	}
	pkg := fdProto.GetPackage()
	var out []string
	var walk func(prefix string, msgs []*descriptorpb.DescriptorProto)
	walk = func(prefix string, msgs []*descriptorpb.DescriptorProto) {
		for _, m := range msgs {
			name := m.GetName()
			full := name
			if prefix != "" {
				full = prefix + "." + name
			}
			// Map entry synthetic types must be visible to the analyzer
			// (PROTO_MODIFY_MAP and friends type the map field as
			// ARRAY<PROTO<Outer.MapEntry>>), so register them too.
			out = append(out, full)
			walk(full, m.GetNestedType())
		}
	}
	walk(pkg, fdProto.GetMessageType())
	return out
}

// chainedFileResolver implements protodesc.Resolver by trying the
// primary resolver first and falling back to a second one on miss.
// This lets project-local descriptors resolve against
// protoregistry.GlobalFiles for well-known imports.
type chainedFileResolver struct {
	primary  protodesc.Resolver
	fallback protodesc.Resolver
}

func (r chainedFileResolver) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	if fd, err := r.primary.FindFileByPath(path); err == nil {
		return fd, nil
	}
	return r.fallback.FindFileByPath(path)
}

func (r chainedFileResolver) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	if d, err := r.primary.FindDescriptorByName(name); err == nil {
		return d, nil
	}
	return r.fallback.FindDescriptorByName(name)
}
