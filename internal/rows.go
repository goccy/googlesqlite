package internal

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/goccy/go-json"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	"github.com/goccy/googlesqlite/internal/value"
)

type Rows struct {
	rows    *sql.Rows
	conn    *Conn
	columns []*ColumnSpec
	actions []StmtAction
}

func (r *Rows) ChangedCatalog() *ChangedCatalog {
	return r.conn.cc
}

func (r *Rows) SetActions(actions []StmtAction) {
	r.actions = actions
}

func (r *Rows) Columns() []string {
	colNames := make([]string, 0, len(r.columns))
	for _, col := range r.columns {
		colNames = append(colNames, col.Name)
	}
	return colNames
}

func (r *Rows) ColumnTypeDatabaseTypeName(i int) string {
	encodedType, _ := json.Marshal(r.columns[i].Type)
	return string(encodedType)
}

func (r *Rows) Close() (e error) {
	defer func() {
		eg := new(ErrorGroup)
		eg.Add(e)
		for _, action := range r.actions {
			eg.Add(action.Cleanup(context.Background(), r.conn))
		}
		if eg.HasError() {
			e = eg
		}
	}()
	if r.rows == nil {
		return nil
	}
	return r.rows.Close()
}

func (r *Rows) columnTypes() []*Type {
	ret := make([]*Type, 0, len(r.columns))
	for _, col := range r.columns {
		ret = append(ret, col.Type)
	}
	return ret
}

func (r *Rows) Next(dest []driver.Value) error {
	if r.rows == nil {
		return io.EOF
	}
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return unwrapSQLiteUserError(err)
		}
		return io.EOF
	}
	if err := r.rows.Err(); err != nil {
		return unwrapSQLiteUserError(err)
	}
	colTypes := r.columnTypes()
	values := make([]any, 0, len(dest))
	for range dest {
		var v any
		values = append(values, &v)
	}
	retErr := r.rows.Scan(values...)
	destV := reflect.ValueOf(dest)
	for idx, colType := range colTypes {
		src := reflect.ValueOf(values[idx]).Elem().Interface()
		dst := destV.Index(idx)
		if err := r.assignValue(src, dst, colType); err != nil {
			return err
		}
	}
	return retErr
}

func (r *Rows) assignValue(src any, dst reflect.Value, typ *Type) error {
	if src == nil {
		dst.Set(reflect.New(dst.Type()).Elem())
		return nil
	}
	decodedValue, err := DecodeValue(src)
	if err != nil {
		// SQLite TEXT-affinity columns (NUMERIC, BIGNUMERIC, DATE,
		// TIME, DATETIME, TIMESTAMP, JSON, INTERVAL, GEOGRAPHY) coerce
		// numeric arguments to their string form at write time, so a
		// row inserted with `?` bound to a Go int can come back as the
		// raw decimal string instead of our base64 layout. Rather
		// than reject those rows, hand the raw string to CastValue,
		// which knows how to parse e.g. "3" into a NumericValue or
		// "2024-01-15" into a DateValue.
		if s, ok := src.(string); ok {
			switch googlesql.TypeKind(typ.Kind) {
			case googlesql.TypeKindTypeNumeric,
				googlesql.TypeKindTypeBignumeric,
				googlesql.TypeKindTypeDate,
				googlesql.TypeKindTypeTime,
				googlesql.TypeKindTypeDatetime,
				googlesql.TypeKindTypeTimestamp,
				googlesql.TypeKindTypeJson,
				googlesql.TypeKindTypeInterval,
				googlesql.TypeKindTypeGeography,
				googlesql.TypeKindTypeRange:
				decodedValue = value.StringValue(s)
				err = nil
			}
		}
		if err != nil {
			return err
		}
	}
	t, err := typ.ToGoogleSQLType()
	if err != nil {
		return err
	}
	value, err := CastValue(t, decodedValue)
	if err != nil {
		return err
	}
	kind := dst.Type().Kind()
	if conv, ok := scalarValueConverters[kind]; ok {
		rv, err := conv(value)
		if err != nil {
			return err
		}
		dst.Set(rv)
		return nil
	}
	if kind == reflect.Interface {
		return r.assignInterfaceValue(value, dst, typ)
	}
	return fmt.Errorf("unexpected destination type %s for %T", kind, value)
}

// scalarValueConverters maps a destination reflect.Kind to a function
// that converts a value.Value into a reflect.Value of that exact Go
// type. The integer / unsigned / float branches collapse the
// previously hand-written, near-identical cases; each closure preserves
// the original width cast (e.g. int8(i64), uint32(i64), float32(f64))
// and overflow / truncation semantics exactly.
var scalarValueConverters = map[reflect.Kind]func(value.Value) (reflect.Value, error){
	reflect.Int:     intConverter(func(i int64) any { return int(i) }),
	reflect.Int8:    intConverter(func(i int64) any { return int8(i) }),
	reflect.Int16:   intConverter(func(i int64) any { return int16(i) }),
	reflect.Int32:   intConverter(func(i int64) any { return int32(i) }),
	reflect.Int64:   intConverter(func(i int64) any { return i }),
	reflect.Uint:    intConverter(func(i int64) any { return uint(i) }),
	reflect.Uint8:   intConverter(func(i int64) any { return uint8(i) }),
	reflect.Uint16:  intConverter(func(i int64) any { return uint16(i) }),
	reflect.Uint32:  intConverter(func(i int64) any { return uint32(i) }),
	reflect.Uint64:  intConverter(func(i int64) any { return uint64(i) }),
	reflect.Float32: floatConverter(func(f float64) any { return float32(f) }),
	reflect.Float64: floatConverter(func(f float64) any { return f }),
	reflect.String: func(v value.Value) (reflect.Value, error) {
		s, err := v.ToString()
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(s), nil
	},
	reflect.Bool: func(v value.Value) (reflect.Value, error) {
		b, err := v.ToBool()
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(b), nil
	},
}

// intConverter builds a converter that reads an int64 from the value
// and applies cast to obtain the target-width Go integer.
func intConverter(cast func(int64) any) func(value.Value) (reflect.Value, error) {
	return func(v value.Value) (reflect.Value, error) {
		i64, err := v.ToInt64()
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(cast(i64)), nil
	}
}

// floatConverter builds a converter that reads a float64 from the value
// and applies cast to obtain the target-width Go float.
func floatConverter(cast func(float64) any) func(value.Value) (reflect.Value, error) {
	return func(v value.Value) (reflect.Value, error) {
		f64, err := v.ToFloat64()
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(cast(f64)), nil
	}
}

func (r *Rows) assignInterfaceValue(src value.Value, dst reflect.Value, typ *Type) error {
	switch googlesql.TypeKind(typ.Kind) {
	case googlesql.TypeKindTypeInt32, googlesql.TypeKindTypeInt64, googlesql.TypeKindTypeUint32, googlesql.TypeKindTypeUint64:
		i64, err := src.ToInt64()
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(i64))
	case googlesql.TypeKindTypeBool:
		b, err := src.ToBool()
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(b))
	case googlesql.TypeKindTypeFloat, googlesql.TypeKindTypeDouble:
		f64, err := src.ToFloat64()
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(f64))
	case googlesql.TypeKindTypeBytes,
		googlesql.TypeKindTypeString,
		googlesql.TypeKindTypeNumeric,
		googlesql.TypeKindTypeBignumeric,
		googlesql.TypeKindTypeInterval,
		googlesql.TypeKindTypeGeography,
		googlesql.TypeKindTypeRange:
		// String-valued kinds: render via ToString and assign the
		// raw Go string to the interface destination.
		s, err := src.ToString()
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(s))
	case googlesql.TypeKindTypeDate,
		googlesql.TypeKindTypeDatetime,
		googlesql.TypeKindTypeTime,
		googlesql.TypeKindTypeJson:
		// JSON-rendered kinds: ToJSON yields the canonical scalar
		// representation the driver hands back for these column types.
		v, err := src.ToJSON()
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(v))
	case googlesql.TypeKindTypeTimestamp:
		t, err := src.ToTime()
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(formatTimestampCanonical(t)))
	case googlesql.TypeKindTypeStruct:
		s, err := src.ToStruct()
		if err != nil {
			return err
		}
		// STRUCT scans to a positional `[]any` whose elements are each
		// recursively converted via assignInterfaceValue. Field names
		// live on the column type, not in the row value (the
		// equivalent is `SELECT s.field` in SQL). Callers that want
		// the canonical short-form string can render it via
		// FormatStructAsBraceString below.
		anyType := reflect.TypeOf((*any)(nil)).Elem()
		sliceRef := reflect.MakeSlice(reflect.SliceOf(anyType), len(s.Values), len(s.Values))
		for i, fv := range s.Values {
			elem := sliceRef.Index(i)
			if fv == nil {
				continue
			}
			var fieldType *Type
			if i < len(typ.FieldTypes) && typ.FieldTypes[i] != nil {
				fieldType = typ.FieldTypes[i].Type
			}
			if fieldType == nil {
				// The analyzer is expected to attach a per-field type
				// to every STRUCT field column; reaching this branch
				// means a producer regressed and dropped that metadata.
				return fmt.Errorf("rows: STRUCT field %d has no Type metadata", i)
			}
			if err := r.assignInterfaceValue(fv, elem, fieldType); err != nil {
				return err
			}
		}
		dst.Set(sliceRef)
	case googlesql.TypeKindTypeArray:
		array, err := src.ToArray()
		if err != nil {
			return err
		}
		sliceType := reflect.SliceOf(reflect.TypeOf((*any)(nil)).Elem())
		sliceRef := reflect.New(sliceType)
		sliceRef.Elem().Set(reflect.MakeSlice(sliceType, 0, len(array.Values)))
		for _, v := range array.Values {
			refV := reflect.New(sliceType.Elem())
			if v == nil {
				sliceRef.Elem().Set(reflect.Append(sliceRef.Elem(), refV.Elem()))
				continue
			}
			if err := r.assignInterfaceValue(v, refV.Elem(), typ.ElementType); err != nil {
				return err
			}
			sliceRef.Elem().Set(reflect.Append(sliceRef.Elem(), refV.Elem()))
		}
		dst.Set(sliceRef.Elem())
	case googlesql.TypeKindTypeProto:
		// PROTO column: the runtime stored wire bytes. Render via the
		// global protobuf type registry so well-known types
		// (google.protobuf.*, google.type.*) come out in the canonical
		// text-format BigQuery uses for proto-typed columns.
		dst.Set(reflect.ValueOf(protoWireBytesToText(src, typ)))
	}
	return nil
}

// protoWireBytesToText renders proto wire bytes as BigQuery's
// curly-brace proto-text-format display string. Uses
// protoregistry.GlobalTypes to look up the message by full name
// (Type.Name carries `google.type.Date` etc. for the well-known
// auto-registered types). Falls back to a base64-of-bytes form when
// the message type is unknown.
func protoWireBytesToText(src value.Value, typ *Type) string {
	raw, err := src.ToBytes()
	if err != nil || len(raw) == 0 {
		return ""
	}
	if typ == nil || typ.Name == "" {
		return base64.StdEncoding.EncodeToString(raw)
	}
	// Type.Name comes back as `PROTO<full.name>` for proto-typed
	// columns. Strip the wrapper to look up the descriptor by its
	// canonical full name.
	name := typ.Name
	if strings.HasPrefix(name, "PROTO<") && strings.HasSuffix(name, ">") {
		name = name[len("PROTO<") : len(name)-1]
	}
	mt := lookupGoProtoMessageType(name)
	if mt == nil {
		return base64.StdEncoding.EncodeToString(raw)
	}
	msg := mt.New().Interface()
	if err := proto.Unmarshal(raw, msg); err != nil {
		return base64.StdEncoding.EncodeToString(raw)
	}
	out, err := (prototext.MarshalOptions{}).Marshal(msg)
	if err != nil {
		return base64.StdEncoding.EncodeToString(raw)
	}
	// Wrap in `{ ... }` to match BigQuery's display form and collapse
	// the prototext multiline output onto a single line. prototext
	// omits the space after `:` and emits two spaces between fields
	// when forced onto one line; re-normalise to a single space and
	// re-insert the BigQuery-canonical space after `:`. BigQuery
	// quotes string fields with single quotes, prototext with double
	// quotes — rewrite for consistency.
	body := strings.TrimSpace(string(out))
	body = strings.ReplaceAll(body, "\n", " ")
	body = strings.Join(strings.Fields(body), " ")
	body = strings.ReplaceAll(body, ":", ": ")
	body = strings.Join(strings.Fields(body), " ")
	body = strings.ReplaceAll(body, "\"", "'")
	// Submessage fields print as `name: { ... }` in prototext but
	// BigQuery's display uses `name { ... }` (no colon before the
	// opening brace). Strip the redundant `: ` immediately preceding
	// a `{`.
	body = strings.ReplaceAll(body, ": {", " {")
	return "{" + body + "}"
}

// formatTimestampCanonical renders a time.Time as the BigQuery /
// GoogleSQL canonical UTC string -- "YYYY-MM-DD HH:MM:SS+00" for
// instant-second timestamps, "YYYY-MM-DD HH:MM:SS.ffffff+00" when the
// sub-second part is non-zero. This matches the upstream Examples
// output (e.g. TIMESTAMP_FROM_UNIX_MICROS(1230219000000000) ->
// "2008-12-25 15:30:00+00").
func formatTimestampCanonical(t time.Time) string {
	utc := t.UTC()
	if utc.Nanosecond() == 0 {
		return utc.Format("2006-01-02 15:04:05+00")
	}
	// Render microsecond precision (BigQuery's default TIMESTAMP
	// fractional resolution). Trailing zeros are kept to match the
	// upstream textual form.
	return utc.Format("2006-01-02 15:04:05.000000+00")
}
