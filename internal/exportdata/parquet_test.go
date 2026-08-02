package exportdata

import (
	"bytes"
	"io"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/format"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestParquetTypeMappings(t *testing.T) {
	columns := []ColumnSchema{
		{Name: "z_int", Type: scalarType("INT64", googlesql.TypeKindTypeInt64)},
		{Name: "a_numeric", Type: scalarType("NUMERIC", googlesql.TypeKindTypeNumeric)},
		{Name: "bignumeric", Type: scalarType("BIGNUMERIC", googlesql.TypeKindTypeBignumeric)},
		{Name: "float", Type: scalarType("FLOAT64", googlesql.TypeKindTypeDouble)},
		{Name: "bool", Type: scalarType("BOOL", googlesql.TypeKindTypeBool)},
		{Name: "string", Type: scalarType("STRING", googlesql.TypeKindTypeString)},
		{Name: "bytes", Type: scalarType("BYTES", googlesql.TypeKindTypeBytes)},
		{Name: "date", Type: scalarType("DATE", googlesql.TypeKindTypeDate)},
		{Name: "datetime", Type: scalarType("DATETIME", googlesql.TypeKindTypeDatetime)},
		{Name: "time", Type: scalarType("TIME", googlesql.TypeKindTypeTime)},
		{Name: "timestamp", Type: scalarType("TIMESTAMP", googlesql.TypeKindTypeTimestamp)},
		{Name: "nullable", Type: scalarType("STRING", googlesql.TypeKindTypeString)},
		{
			Name: "record",
			Type: &TypeSchema{
				Name: "STRUCT<name STRING, scores ARRAY<INT64>>",
				Kind: googlesql.TypeKindTypeStruct,
				FieldTypes: []ColumnSchema{
					{Name: "name", Type: scalarType("STRING", googlesql.TypeKindTypeString)},
					{Name: "scores", Type: &TypeSchema{
						Name:        "ARRAY<INT64>",
						Kind:        googlesql.TypeKindTypeArray,
						ElementType: scalarType("INT64", googlesql.TypeKindTypeInt64),
					}},
				},
			},
		},
	}
	cfg, err := NewParquetEncodingConfig(columns, CompressionSnappy)
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2024, 2, 3, 0, 0, 0, 0, time.UTC)
	datetime := time.Date(2024, 2, 3, 4, 5, 6, 7000, time.UTC)
	timeOfDay := time.Date(1, 1, 1, 7, 8, 9, 10000, time.UTC)
	timestamp := time.Date(2024, 2, 3, 4, 5, 6, 11000, time.FixedZone("offset", 2*60*60))
	numeric := new(big.Rat)
	numeric.SetString("123.456789012")
	bignumeric := new(big.Rat)
	bignumeric.SetString("0.12345678901234567890123456789012345678")
	rows := [][]value.Value{{
		value.IntValue(42),
		&value.NumericValue{Rat: numeric},
		&value.NumericValue{Rat: bignumeric, IsBigNumeric: true},
		value.FloatValue(1.5),
		value.BoolValue(true),
		value.StringValue("hello"),
		value.BytesValue{0, 1, 2},
		value.DateValue(date),
		value.DatetimeValue(datetime),
		value.TimeValue(timeOfDay),
		value.TimestampValue(timestamp),
		nil,
		&value.StructValue{
			Keys: []string{"name", "scores"},
			Values: []value.Value{
				value.StringValue("nested"),
				&value.ArrayValue{Values: []value.Value{value.IntValue(1), value.IntValue(2)}},
			},
		},
	}}
	var encoded bytes.Buffer
	if err := EncodeParquetRows(&encoded, cfg, valueRows(rows)); err != nil {
		t.Fatal(err)
	}

	t.Run("schema", func(t *testing.T) {
		file, err := parquet.OpenFile(bytes.NewReader(encoded.Bytes()), int64(encoded.Len()))
		if err != nil {
			t.Fatalf("open encoded Parquet: %v", err)
		}
		if got, want := file.NumRows(), int64(1); got != want {
			t.Fatalf("NumRows = %d; want %d", got, want)
		}
		if got, want := topLevelNames(file.Schema()), []string{"z_int", "a_numeric", "bignumeric", "float", "bool", "string", "bytes", "date", "datetime", "time", "timestamp", "nullable", "record"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("schema order = %v; want %v", got, want)
		}
		schemaText := file.Schema().String()
		for _, fragment := range []string{
			"int64 z_int", "fixed_len_byte_array(16) a_numeric (DECIMAL(38,9))",
			"fixed_len_byte_array(32) bignumeric (DECIMAL(76,38))", "double float",
			"boolean bool", "binary string (STRING)", "int32 date (DATE)", "TIMESTAMP", "TIME", "LIST",
		} {
			if !strings.Contains(schemaText, fragment) {
				t.Errorf("schema %q does not contain %q", schemaText, fragment)
			}
		}
		for _, column := range file.Root().Columns() {
			assertLeafCompression(t, column, format.Snappy)
		}
	})

	t.Run("values", func(t *testing.T) {
		reader := parquet.NewGenericReader[any](bytes.NewReader(encoded.Bytes()))
		defer reader.Close()
		decoded := make([]any, 1)
		if n, err := reader.Read(decoded); n != 1 || err != nil && err != io.EOF {
			t.Fatalf("read rows = %d, %v", n, err)
		}
		row, ok := decoded[0].(map[string]any)
		if !ok {
			t.Fatalf("decoded row type = %T", decoded[0])
		}
		if row["z_int"] != int64(42) || row["float"] != 1.5 || row["bool"] != true ||
			row["string"] != "hello" || row["bytes"] != string([]byte{0, 1, 2}) || row["nullable"] != nil {
			t.Fatalf("decoded scalar values = %#v", row)
		}
		if row["date"] != int32(19756) || row["datetime"] != datetime.UnixMicro() ||
			row["time"] != int64(7*time.Hour/time.Microsecond+8*time.Minute/time.Microsecond+9*time.Second/time.Microsecond+10) ||
			row["timestamp"] != timestamp.UTC().UnixMicro() {
			t.Fatalf("decoded temporal values = %#v", row)
		}
		decodedNumeric, ok := row["a_numeric"].(*big.Float)
		if !ok || decodedNumeric.Text('f', 9) != "123.456789012" {
			t.Fatalf("decoded NUMERIC = %T %v", row["a_numeric"], row["a_numeric"])
		}
		decodedBigNumeric, ok := row["bignumeric"].(*big.Float)
		if !ok || decodedBigNumeric.Text('f', 38) != "0.12345678901234567890123456789012345678" {
			t.Fatalf("decoded BIGNUMERIC = %T %v", row["bignumeric"], row["bignumeric"])
		}
		record, ok := row["record"].(map[string]any)
		if !ok || record["name"] != "nested" {
			t.Fatalf("decoded record = %#v", row["record"])
		}
		scores, ok := record["scores"].([]any)
		if !ok || !reflect.DeepEqual(scores, []any{int64(1), int64(2)}) {
			t.Fatalf("decoded scores = %#v", record["scores"])
		}
	})
}

func TestParquetNullAndEmptyCollections(t *testing.T) {
	columns := []ColumnSchema{{
		Name: "items",
		Type: &TypeSchema{
			Name:        "ARRAY<STRING>",
			Kind:        googlesql.TypeKindTypeArray,
			ElementType: scalarType("STRING", googlesql.TypeKindTypeString),
		},
	}}
	cfg, err := NewParquetEncodingConfig(columns, CompressionNone)
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := EncodeParquetRows(&encoded, cfg, valueRows([][]value.Value{
		{&value.ArrayValue{Values: []value.Value{}}},
		{nil},
	})); err != nil {
		t.Fatal(err)
	}
	reader := parquet.NewGenericReader[any](bytes.NewReader(encoded.Bytes()))
	defer reader.Close()
	rows := make([]any, 2)
	if n, err := reader.Read(rows); n != 2 || err != nil && err != io.EOF {
		t.Fatalf("read rows = %d, %v", n, err)
	}
	first := rows[0].(map[string]any)
	second := rows[1].(map[string]any)
	items, ok := first["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("empty ARRAY decoded as %#v", first["items"])
	}
	if second["items"] != nil {
		t.Fatalf("NULL ARRAY decoded as %#v", second["items"])
	}
}

func TestParquetZeroRows(t *testing.T) {
	columns := []ColumnSchema{{Name: "id", Type: scalarType("INT64", googlesql.TypeKindTypeInt64)}}
	cfg, err := NewParquetEncodingConfig(columns, CompressionNone)
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := EncodeParquetRows(&encoded, cfg, valueRows(nil)); err != nil {
		t.Fatal(err)
	}
	file, err := parquet.OpenFile(bytes.NewReader(encoded.Bytes()), int64(encoded.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if file.NumRows() != 0 {
		t.Fatalf("NumRows = %d; want 0", file.NumRows())
	}
}

func TestParquetCompressionCodecs(t *testing.T) {
	columns := []ColumnSchema{{Name: "id", Type: scalarType("INT64", googlesql.TypeKindTypeInt64)}}
	for _, tc := range []struct {
		compression Compression
		codec       format.CompressionCodec
	}{
		{CompressionNone, format.Uncompressed},
		{CompressionSnappy, format.Snappy},
		{CompressionGZIP, format.Gzip},
		{CompressionZSTD, format.Zstd},
	} {
		t.Run(string(tc.compression), func(t *testing.T) {
			cfg, err := NewParquetEncodingConfig(columns, tc.compression)
			if err != nil {
				t.Fatal(err)
			}
			var encoded bytes.Buffer
			if err := EncodeParquetRows(&encoded, cfg, valueRows([][]value.Value{{value.IntValue(1)}})); err != nil {
				t.Fatal(err)
			}
			file, err := parquet.OpenFile(bytes.NewReader(encoded.Bytes()), int64(encoded.Len()))
			if err != nil {
				t.Fatal(err)
			}
			assertLeafCompression(t, file.Root(), tc.codec)
		})
	}
}

func TestParquetRejectsUnsupportedTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind googlesql.TypeKind
	}{
		{"JSON", googlesql.TypeKindTypeJson},
		{"INTERVAL", googlesql.TypeKindTypeInterval},
		{"RANGE<DATE>", googlesql.TypeKindTypeRange},
		{"PROTO<Foo>", googlesql.TypeKindTypeProto},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewParquetEncodingConfig([]ColumnSchema{{Name: "v", Type: scalarType(tc.name, tc.kind)}}, CompressionNone)
			if err == nil || !strings.Contains(err.Error(), "no documented Parquet mapping") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func scalarType(name string, kind googlesql.TypeKind) *TypeSchema {
	return &TypeSchema{Name: name, Kind: kind}
}

func valueRows(rows [][]value.Value) RowSource {
	index := 0
	return func() ([]any, bool, error) {
		if index == len(rows) {
			return nil, false, nil
		}
		row := make([]any, len(rows[index]))
		for i := range rows[index] {
			row[i] = rows[index][i]
		}
		index++
		return row, true, nil
	}
}

func topLevelNames(schema *parquet.Schema) []string {
	fields := schema.Fields()
	names := make([]string, len(fields))
	for i := range fields {
		names[i] = fields[i].Name()
	}
	return names
}

func assertLeafCompression(t *testing.T, column *parquet.Column, want format.CompressionCodec) {
	t.Helper()
	if column.Leaf() {
		if got := column.Compression().CompressionCodec(); got != want {
			t.Errorf("column %v compression = %v; want %v", column.Path(), got, want)
		}
		return
	}
	for _, child := range column.Columns() {
		assertLeafCompression(t, child, want)
	}
}
