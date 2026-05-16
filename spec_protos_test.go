package googlesqlite_test

import (
	"database/sql"
	"fmt"

	"github.com/goccy/googlesqlite"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// registerProtoOnConn unwraps the *sql.Conn to the underlying
// *googlesqlite.Conn and registers the proto bundle on it. Returns an
// error when the conn is not from the googlesqlite driver or the
// registration itself fails. RegisterProto is exposed to this
// black-box test package through export_test.go.
func registerProtoOnConn(conn *sql.Conn, bytes []byte) error {
	return conn.Raw(func(driverConn any) error {
		gc, ok := driverConn.(*googlesqlite.Conn)
		if !ok {
			return fmt.Errorf("spectest: not a googlesqlite.Conn: %T", driverConn)
		}
		return gc.RegisterProto(bytes)
	})
}

// testProtoBundles is the set of FileDescriptorProto byte payloads
// the spec-test runner can register via Conn.RegisterProto on
// demand. Names match the `register_protos:` keys callers use in
// testdata YAML cases. The descriptors are hand-built so the
// upstream proto-function Examples (`Item`, `googlesql.examples.music.*`,
// `googlesql.LanguageFeature`, etc.) can be exercised without
// shelling out to protoc.
var testProtoBundles = map[string][]byte{}

func init() {
	register := func(name string, fd *descriptorpb.FileDescriptorProto) {
		b, err := proto.Marshal(fd)
		if err != nil {
			panic("spectest: marshal " + name + ": " + err.Error())
		}
		testProtoBundles[name] = b
	}
	register("item", itemFile())
	register("award", awardFile())
	register("book", bookFile())
}

// awardFile mirrors the upstream FILTER_FIELDS Example proto:
//
//	package googlesql.examples.music;
//	message Award {
//	  required int32 year = 1;
//	  optional int32 month = 2;
//	  repeated Type type = 3;
//	  message Type {
//	    optional string award_name = 1;
//	    optional string category = 2;
//	  }
//	}
func awardFile() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("googlesqlite/spectest/award.proto"),
		Syntax:  proto.String("proto2"),
		Package: proto.String("googlesql.examples.music"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Award"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("year"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REQUIRED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
					},
					{
						Name:   proto.String("month"),
						Number: proto.Int32(2),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
					},
					{
						Name:     proto.String("type"),
						Number:   proto.Int32(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".googlesql.examples.music.Award.Type"),
					},
				},
				NestedType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("Type"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:   proto.String("award_name"),
								Number: proto.Int32(1),
								Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
							{
								Name:   proto.String("category"),
								Number: proto.Int32(2),
								Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
						},
					},
				},
			},
		},
	}
}

// bookFile mirrors the upstream REPLACE_FIELDS Example protos:
//
//	syntax = "proto2";
//	message Book {
//	  required string title = 1;
//	  repeated string reviews = 2;
//	  optional BookDetails details = 3;
//	};
//	message BookDetails {
//	  optional string author = 1;
//	  optional int32 chapters = 2;
//	};
func bookFile() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("googlesqlite/spectest/book.proto"),
		Syntax:  proto.String("proto2"),
		Package: proto.String(""),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Book"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("title"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REQUIRED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("reviews"),
						Number: proto.Int32(2),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:     proto.String("details"),
						Number:   proto.Int32(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".BookDetails"),
					},
				},
			},
			{
				Name: proto.String("BookDetails"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("author"),
						Number: proto.Int32(1),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					},
					{
						Name:   proto.String("chapters"),
						Number: proto.Int32(2),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
					},
				},
			},
		},
	}
}

// itemFile is the minimal `Item` proto from
// `proto_map_contains_key.yaml` / `proto_modify_map.yaml`:
//
//	syntax = "proto2";
//	message Item {
//	  optional map<string, int64> purchased = 1;
//	}
func itemFile() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("googlesqlite/spectest/item.proto"),
		Syntax:  proto.String("proto2"),
		Package: proto.String(""),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Item"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("purchased"),
						Number:   proto.Int32(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".Item.PurchasedEntry"),
					},
				},
				NestedType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("PurchasedEntry"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:   proto.String("key"),
								Number: proto.Int32(1),
								Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
							},
							{
								Name:   proto.String("value"),
								Number: proto.Int32(2),
								Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
								Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
							},
						},
						Options: &descriptorpb.MessageOptions{
							MapEntry: proto.Bool(true),
						},
					},
				},
			},
		},
	}
}
