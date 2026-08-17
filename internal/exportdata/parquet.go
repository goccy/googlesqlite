package exportdata

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress"
	parquetgzip "github.com/parquet-go/parquet-go/compress/gzip"
	"github.com/parquet-go/parquet-go/compress/snappy"
	"github.com/parquet-go/parquet-go/compress/uncompressed"
	"github.com/parquet-go/parquet-go/compress/zstd"
	"github.com/parquet-go/parquet-go/encoding"
	"github.com/parquet-go/parquet-go/format"

	"github.com/goccy/googlesqlite/internal/value"
)

var decimalTypePattern = regexp.MustCompile(`(?i)^(?:BIG)?NUMERIC(?:\((\d+)\s*,\s*(\d+)\))?$`)

// TypeSchema is the resolved GoogleSQL type information needed by Parquet.
type TypeSchema struct {
	Name        string
	Kind        googlesql.TypeKind
	ElementType *TypeSchema
	FieldTypes  []ColumnSchema
}

type ColumnSchema struct {
	Name string
	Type *TypeSchema
}

type ParquetEncodingConfig struct {
	columns      []parquetColumn
	writerConfig *parquet.WriterConfig
}

type parquetColumn struct {
	name  string
	type_ *parquetValueType
}

type parquetValueType struct {
	schema       *TypeSchema
	node         parquet.Node
	fields       []parquetColumn
	element      *parquetValueType
	precision    int
	scale        int
	decimalWidth int
}

func NewParquetEncodingConfig(columns []ColumnSchema, compression Compression) (cfg *ParquetEncodingConfig, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			cfg = nil
			err = fmt.Errorf("EXPORT DATA: build Parquet schema: %v", recovered)
		}
	}()

	fields := make([]parquet.Field, len(columns))
	parquetColumns := make([]parquetColumn, len(columns))
	for i := range columns {
		valueType, err := newParquetValueType(columns[i].Type, columns[i].Name)
		if err != nil {
			return nil, err
		}
		name := parquetFieldName(columns[i].Name, i)
		fields[i] = &orderedField{name: name, Node: parquet.Optional(valueType.node)}
		parquetColumns[i] = parquetColumn{name: name, type_: valueType}
	}
	schema := parquet.NewSchema("export", &orderedGroup{fields: fields})
	codec, err := parquetCompressionCodec(compression)
	if err != nil {
		return nil, err
	}
	writerOpts := []parquet.WriterOption{schema, parquet.Compression(codec)}
	if geoMetadata, ok, err := buildGeoParquetMetadata(columns); err != nil {
		return nil, err
	} else if ok {
		writerOpts = append(writerOpts, parquet.KeyValueMetadata("geo", geoMetadata))
	}
	writerConfig, err := parquet.NewWriterConfig(writerOpts...)
	if err != nil {
		return nil, fmt.Errorf("EXPORT DATA: configure Parquet writer: %w", err)
	}
	return &ParquetEncodingConfig{columns: parquetColumns, writerConfig: writerConfig}, nil
}

func newParquetValueType(t *TypeSchema, path string) (*parquetValueType, error) {
	if t == nil {
		return nil, fmt.Errorf("EXPORT DATA: Parquet column %q has no resolved type", path)
	}
	valueType := &parquetValueType{schema: t}
	switch t.Kind {
	case googlesql.TypeKindTypeInt64:
		valueType.node = parquet.Leaf(parquet.Int64Type)
	case googlesql.TypeKindTypeNumeric, googlesql.TypeKindTypeBignumeric:
		precision, scale, err := decimalPrecisionAndScale(t)
		if err != nil {
			return nil, fmt.Errorf("EXPORT DATA: Parquet column %q: %w", path, err)
		}
		valueType.precision = precision
		valueType.scale = scale
		valueType.decimalWidth = decimalByteWidth(precision)
		valueType.node = parquet.Decimal(scale, precision, parquet.FixedLenByteArrayType(valueType.decimalWidth))
	case googlesql.TypeKindTypeDouble:
		valueType.node = parquet.Leaf(parquet.DoubleType)
	case googlesql.TypeKindTypeBool:
		valueType.node = parquet.Leaf(parquet.BooleanType)
	case googlesql.TypeKindTypeString:
		valueType.node = parquet.String()
	case googlesql.TypeKindTypeBytes:
		valueType.node = parquet.Leaf(parquet.ByteArrayType)
	case googlesql.TypeKindTypeDate:
		valueType.node = parquet.Date()
	case googlesql.TypeKindTypeDatetime:
		valueType.node = parquet.TimestampAdjusted(parquet.Microsecond, false)
	case googlesql.TypeKindTypeTime:
		valueType.node = parquet.TimeAdjusted(parquet.Microsecond, true)
	case googlesql.TypeKindTypeTimestamp:
		valueType.node = parquet.TimestampAdjusted(parquet.Microsecond, true)
	case googlesql.TypeKindTypeGeography:
		valueType.node = parquet.Geography("OGC:CRS84", format.Spherical)
	case googlesql.TypeKindTypeStruct:
		fields := make([]parquet.Field, len(t.FieldTypes))
		valueType.fields = make([]parquetColumn, len(t.FieldTypes))
		for i := range t.FieldTypes {
			name := parquetFieldName(t.FieldTypes[i].Name, i)
			nested, err := newParquetValueType(t.FieldTypes[i].Type, path+"."+name)
			if err != nil {
				return nil, err
			}
			fields[i] = &orderedField{name: name, Node: parquet.Optional(nested.node)}
			valueType.fields[i] = parquetColumn{name: name, type_: nested}
		}
		valueType.node = &orderedGroup{fields: fields}
	case googlesql.TypeKindTypeArray:
		if t.ElementType == nil {
			return nil, fmt.Errorf("EXPORT DATA: Parquet ARRAY column %q has no element type", path)
		}
		if t.ElementType.Kind == googlesql.TypeKindTypeArray {
			return nil, fmt.Errorf("EXPORT DATA: GoogleSQL ARRAY values cannot contain ARRAY elements at %q", path)
		}
		element, err := newParquetValueType(t.ElementType, path+"[]")
		if err != nil {
			return nil, err
		}
		valueType.element = element
		valueType.node = parquet.List(element.node)
	default:
		return nil, fmt.Errorf("EXPORT DATA: GoogleSQL type %s in column %q has no documented Parquet mapping", t.Name, path)
	}
	return valueType, nil
}

// orderedGroup retains query/struct declaration order.
// parquet.Group is a map and intentionally sorts fields, which is unsuitable for EXPORT DATA schemas.
type orderedGroup struct {
	fields []parquet.Field
}

func (g *orderedGroup) ID() int                     { return 0 }
func (g *orderedGroup) String() string              { return "group" }
func (g *orderedGroup) Type() parquet.Type          { return (parquet.Group{}).Type() }
func (g *orderedGroup) Optional() bool              { return false }
func (g *orderedGroup) Repeated() bool              { return false }
func (g *orderedGroup) Required() bool              { return true }
func (g *orderedGroup) Leaf() bool                  { return false }
func (g *orderedGroup) Fields() []parquet.Field     { return g.fields }
func (g *orderedGroup) Encoding() encoding.Encoding { return nil }
func (g *orderedGroup) Compression() compress.Codec { return nil }
func (g *orderedGroup) GoType() reflect.Type        { return reflect.TypeFor[map[string]any]() }

type orderedField struct {
	parquet.Node
	name string
}

func (f *orderedField) Name() string { return f.name }

func (f *orderedField) Value(base reflect.Value) reflect.Value {
	for base.IsValid() && (base.Kind() == reflect.Interface || base.Kind() == reflect.Pointer) {
		if base.IsNil() {
			return reflect.Value{}
		}
		base = base.Elem()
	}
	if !base.IsValid() {
		return reflect.Value{}
	}
	return base.MapIndex(reflect.ValueOf(f.name))
}

func parquetFieldName(name string, index int) string {
	if name == "" {
		return fmt.Sprintf("f%d_", index)
	}
	return name
}

func decimalPrecisionAndScale(t *TypeSchema) (int, int, error) {
	precision, scale := 38, 9
	if t.Kind == googlesql.TypeKindTypeBignumeric {
		precision, scale = 76, 38
	}
	match := decimalTypePattern.FindStringSubmatch(strings.TrimSpace(t.Name))
	if match == nil {
		return 0, 0, fmt.Errorf("invalid decimal type %q", t.Name)
	}
	if match[1] == "" {
		return precision, scale, nil
	}
	parsedPrecision, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid decimal precision in %q", t.Name)
	}
	parsedScale, err := strconv.Atoi(match[2])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid decimal scale in %q", t.Name)
	}
	return parsedPrecision, parsedScale, nil
}

func decimalByteWidth(precision int) int {
	return int(math.Ceil((math.Log10(2) + float64(precision)) / math.Log10(256)))
}

func parquetCompressionCodec(compression Compression) (compress.Codec, error) {
	switch compression {
	case "", CompressionNone:
		return &uncompressed.Codec{}, nil
	case CompressionSnappy:
		return &snappy.Codec{}, nil
	case CompressionGZIP:
		return &parquetgzip.Codec{Level: parquetgzip.DefaultCompression}, nil
	case CompressionZSTD:
		return &zstd.Codec{}, nil
	default:
		return nil, fmt.Errorf("EXPORT DATA: compression %s is not valid for format PARQUET", compression)
	}
}

// EncodeParquetRows streams rows into a Parquet file and writes its footer.
func EncodeParquetRows(w io.Writer, cfg *ParquetEncodingConfig, src RowSource) error {
	pw := parquet.NewWriter(w, cfg.writerConfig)
	var encodeErr error
	for {
		values, hasMore, err := src()
		if err != nil {
			encodeErr = err
			break
		}
		if !hasMore {
			break
		}
		row, err := parquetRow(cfg.columns, values)
		if err != nil {
			encodeErr = err
			break
		}
		if err := pw.Write(row); err != nil {
			encodeErr = fmt.Errorf("EXPORT DATA: write Parquet row: %w", err)
			break
		}
	}
	if err := pw.Close(); err != nil && encodeErr == nil {
		encodeErr = fmt.Errorf("EXPORT DATA: close Parquet writer: %w", err)
	}
	return encodeErr
}

func parquetRow(columns []parquetColumn, values []any) (map[string]any, error) {
	row := make(map[string]any, len(columns))
	for i := range columns {
		if i >= len(values) || values[i] == nil {
			row[columns[i].name] = nil
			continue
		}
		typed, ok := values[i].(value.Value)
		if !ok {
			return nil, fmt.Errorf("EXPORT DATA: encode Parquet column %q: expected typed value, got %T", columns[i].name, values[i])
		}
		converted, err := parquetValue(columns[i].type_, typed)
		if err != nil {
			return nil, fmt.Errorf("EXPORT DATA: encode Parquet column %q: %w", columns[i].name, err)
		}
		row[columns[i].name] = converted
	}
	return row, nil
}

func parquetValue(t *parquetValueType, v value.Value) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch t.schema.Kind {
	case googlesql.TypeKindTypeInt64:
		return v.ToInt64()
	case googlesql.TypeKindTypeNumeric, googlesql.TypeKindTypeBignumeric:
		return parquetDecimal(t, v)
	case googlesql.TypeKindTypeDouble:
		return v.ToFloat64()
	case googlesql.TypeKindTypeBool:
		return v.ToBool()
	case googlesql.TypeKindTypeString:
		return v.ToString()
	case googlesql.TypeKindTypeBytes:
		return v.ToBytes()
	case googlesql.TypeKindTypeDate:
		tm, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		midnight := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.UTC)
		return int32(midnight.Unix() / int64((24*time.Hour)/time.Second)), nil
	case googlesql.TypeKindTypeDatetime:
		tm, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		local := time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond(), time.UTC)
		return local.UnixMicro(), nil
	case googlesql.TypeKindTypeTime:
		tm, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		return int64(tm.Hour())*int64(time.Hour/time.Microsecond) +
			int64(tm.Minute())*int64(time.Minute/time.Microsecond) +
			int64(tm.Second())*int64(time.Second/time.Microsecond) +
			int64(tm.Nanosecond())/int64(time.Microsecond), nil
	case googlesql.TypeKindTypeTimestamp:
		tm, err := v.ToTime()
		if err != nil {
			return nil, err
		}
		return tm.UTC().UnixMicro(), nil
	case googlesql.TypeKindTypeGeography:
		return geographyWKB(v)
	case googlesql.TypeKindTypeStruct:
		structValue, err := v.ToStruct()
		if err != nil {
			return nil, err
		}
		out := make(map[string]any, len(t.fields))
		for i := range t.fields {
			if i >= len(structValue.Values) || structValue.Values[i] == nil {
				out[t.fields[i].name] = nil
				continue
			}
			converted, err := parquetValue(t.fields[i].type_, structValue.Values[i])
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", t.fields[i].name, err)
			}
			out[t.fields[i].name] = converted
		}
		return out, nil
	case googlesql.TypeKindTypeArray:
		arrayValue, err := v.ToArray()
		if err != nil {
			return nil, err
		}
		out := make([]any, len(arrayValue.Values))
		for i := range arrayValue.Values {
			if arrayValue.Values[i] == nil {
				return nil, fmt.Errorf("ARRAY element %d is NULL", i)
			}
			converted, err := parquetValue(t.element, arrayValue.Values[i])
			if err != nil {
				return nil, fmt.Errorf("ARRAY element %d: %w", i, err)
			}
			out[i] = converted
		}
		return out, nil
	default:
		return nil, fmt.Errorf("GoogleSQL type %s has no documented Parquet mapping", t.schema.Name)
	}
}

func parquetDecimal(t *parquetValueType, v value.Value) ([]byte, error) {
	rat, err := v.ToRat()
	if err != nil {
		return nil, err
	}
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(t.scale)), nil)
	scaled := new(big.Rat).Mul(rat, new(big.Rat).SetInt(power))
	if !scaled.IsInt() {
		return nil, fmt.Errorf("decimal value %s has more than %d fractional digits", rat.RatString(), t.scale)
	}
	unscaled := new(big.Int).Set(scaled.Num())
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(t.precision)), nil)
	if new(big.Int).Abs(new(big.Int).Set(unscaled)).Cmp(limit) >= 0 {
		return nil, fmt.Errorf("decimal value %s exceeds precision %d", rat.RatString(), t.precision)
	}
	if unscaled.Sign() < 0 {
		modulus := new(big.Int).Lsh(big.NewInt(1), uint(t.decimalWidth*8))
		unscaled.Add(unscaled, modulus)
	}
	return unscaled.FillBytes(make([]byte, t.decimalWidth)), nil
}

type geoParquetMetadata struct {
	Version       string                      `json:"version"`
	PrimaryColumn string                      `json:"primary_column"`
	Columns       map[string]geoParquetColumn `json:"columns"`
}

type geoParquetColumn struct {
	Encoding      string   `json:"encoding"`
	GeometryTypes []string `json:"geometry_types"`
	CRS           any      `json:"crs"`
	Edges         string   `json:"edges"`
}

func buildGeoParquetMetadata(columns []ColumnSchema) (string, bool, error) {
	var paths []string
	for i := range columns {
		collectGeographyPaths(columns[i].Type, parquetFieldName(columns[i].Name, i), &paths)
	}
	if len(paths) == 0 {
		return "", false, nil
	}
	metadata := geoParquetMetadata{
		Version:       "1.1.0",
		PrimaryColumn: paths[0],
		Columns:       make(map[string]geoParquetColumn, len(paths)),
	}
	for _, path := range paths {
		metadata.Columns[path] = geoParquetColumn{
			Encoding:      "WKB",
			GeometryTypes: []string{},
			CRS:           nil,
			Edges:         "spherical",
		}
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "", false, fmt.Errorf("EXPORT DATA: encode GeoParquet metadata: %w", err)
	}
	return string(encoded), true, nil
}

func collectGeographyPaths(t *TypeSchema, path string, paths *[]string) {
	if t == nil {
		return
	}
	switch t.Kind {
	case googlesql.TypeKindTypeGeography:
		*paths = append(*paths, path)
	case googlesql.TypeKindTypeStruct:
		for i := range t.FieldTypes {
			collectGeographyPaths(t.FieldTypes[i].Type, path+"."+parquetFieldName(t.FieldTypes[i].Name, i), paths)
		}
	case googlesql.TypeKindTypeArray:
		collectGeographyPaths(t.ElementType, path, paths)
	}
}

func geographyWKB(v value.Value) ([]byte, error) {
	geography, ok := v.(*value.GeographyValue)
	if !ok {
		wkt, err := v.ToString()
		if err != nil {
			return nil, err
		}
		geography, err = value.GeographyFromWKT(wkt)
		if err != nil {
			return nil, err
		}
	}
	var out bytes.Buffer
	if err := appendGeographyWKB(&out, geography); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func appendGeographyWKB(out *bytes.Buffer, geography *value.GeographyValue) error {
	if geography == nil {
		return fmt.Errorf("nil geography")
	}
	out.WriteByte(1)
	writeUint32 := func(v uint32) { _ = binary.Write(out, binary.LittleEndian, v) }
	writeFloat64 := func(v float64) { _ = binary.Write(out, binary.LittleEndian, v) }
	switch geography.Kind() {
	case "POINT":
		writeUint32(1)
		longitude, latitude, ok := geography.PointCoordinates()
		if !ok {
			longitude, latitude = math.NaN(), math.NaN()
		}
		writeFloat64(longitude)
		writeFloat64(latitude)
	case "LINESTRING":
		writeUint32(2)
		points, _ := geography.LineStringPoints()
		writeUint32(uint32(len(points)))
		for _, point := range points {
			writeFloat64(point[0])
			writeFloat64(point[1])
		}
	case "POLYGON":
		writeUint32(3)
		rings, _ := geography.PolygonRings()
		writeUint32(uint32(len(rings)))
		for _, ring := range rings {
			writeUint32(uint32(len(ring)))
			for _, point := range ring {
				writeFloat64(point[0])
				writeFloat64(point[1])
			}
		}
	case "MULTIPOINT":
		writeUint32(4)
		points, _ := geography.MultiPointPoints()
		writeUint32(uint32(len(points)))
		for _, point := range points {
			if err := appendGeographyWKB(out, value.NewGeographyPoint(point[0], point[1])); err != nil {
				return err
			}
		}
	case "MULTILINESTRING":
		writeUint32(5)
		lines, _ := geography.MultiLineStringLines()
		writeUint32(uint32(len(lines)))
		for _, line := range lines {
			if err := appendGeographyWKB(out, value.NewGeographyLineString(line)); err != nil {
				return err
			}
		}
	case "MULTIPOLYGON":
		writeUint32(6)
		polygons, _ := geography.MultiPolygonPolys()
		writeUint32(uint32(len(polygons)))
		for _, polygon := range polygons {
			if err := appendGeographyWKB(out, value.NewGeographyPolygon(polygon)); err != nil {
				return err
			}
		}
	case "GEOMETRYCOLLECTION":
		writeUint32(7)
		parts, _ := geography.CollectionParts()
		writeUint32(uint32(len(parts)))
		for _, part := range parts {
			if err := appendGeographyWKB(out, part); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported geography kind %q", geography.Kind())
	}
	return nil
}
