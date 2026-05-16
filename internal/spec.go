package internal

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/goccy/go-json"
)

// protoNameFromDebug strips the "PROTO<...>" / "ENUM<...>" wrapper that
// Googlesql_TypeNode.DebugString prints for proto / enum types, leaving
// just the fully-qualified type name. Returns the input unchanged when
// no wrapper is present.
func protoNameFromDebug(s string) string {
	for _, prefix := range []string{"PROTO<", "ENUM<"} {
		if strings.HasPrefix(s, prefix) && strings.HasSuffix(s, ">") {
			return s[len(prefix) : len(s)-1]
		}
	}
	return s
}

type NameWithType struct {
	Name string `json:"name"`
	Type *Type  `json:"type"`
}

func (t *NameWithType) FunctionArgumentType() (*googlesql.FunctionArgumentType, error) {
	if t.Type.SignatureKind != googlesql.SignatureArgumentKindArgTypeFixed {
		return m1(googlesql.NewFunctionArgumentType5(
			t.Type.SignatureKind,
			m1(googlesql.NewFunctionArgumentTypeOptions()),
			-1,
		)), nil
	}
	typ, err := t.Type.ToGoogleSQLType()
	if err != nil {
		return nil, err
	}
	opt := m1(googlesql.NewFunctionArgumentTypeOptions())
	// SetArgumentName isn't exposed on the bridge; argument names flow
	// through the FunctionSignature builder separately in modern API.
	_ = t.Name
	return m1(googlesql.NewFunctionArgumentType(typ, opt, -1)), nil
}

type FunctionSpec struct {
	IsTemp    bool            `json:"isTemp"`
	NamePath  []string        `json:"name"`
	Language  string          `json:"language"`
	Args      []*NameWithType `json:"args"`
	Return    *Type           `json:"return"`
	Body      string          `json:"body"`
	Code      string          `json:"code"`
	UpdatedAt time.Time       `json:"updatedAt"`
	CreatedAt time.Time       `json:"createdAt"`
}

func (s *FunctionSpec) FuncName() string {
	return formatPath(s.NamePath)
}

func (s *FunctionSpec) SQL() string {
	args := []string{}
	for _, arg := range s.Args {
		t, _ := arg.Type.ToGoogleSQLType()
		args = append(args, fmt.Sprintf("%s %s", arg.Name, m1(t.Kind())))
	}
	retType, _ := s.Return.ToGoogleSQLType()
	return fmt.Sprintf(
		"CREATE FUNCTION `%s`(%s) RETURNS %v AS (%s)",
		s.FuncName(),
		strings.Join(args, ", "),
		m1(retType.Kind()),
		s.Body,
	)
}

func (s *FunctionSpec) CallSQL(ctx context.Context, callNode *ResolvedBaseFunctionCallNode, argValues []string) (string, error) {
	args, _ := callNode.ArgumentList()
	var body string
	if s.Body == "" {
		// templated argument func
		definedArgs := make([]string, 0, len(args))
		for idx, arg := range args {
			typeName := newType(m1(arg.Type())).FormatType()
			definedArgs = append(
				definedArgs,
				fmt.Sprintf("%s %s", s.Args[idx].Name, typeName),
			)
		}
		funcName := strings.Join(s.NamePath, ".")
		runtimeDefinedFunc := fmt.Sprintf(
			"CREATE FUNCTION `%s`(%s) as (%s)",
			funcName,
			strings.Join(definedArgs, ","),
			s.Code,
		)
		analyzer := analyzerFromContext(ctx)
		runtimeSpec, err := analyzer.analyzeTemplatedFunctionWithRuntimeArgument(ctx, runtimeDefinedFunc)
		if err != nil {
			return "", err
		}
		body = runtimeSpec.Body
	} else {
		body = s.Body
	}
	for i := 0; i < len(s.Args); i++ {
		argRef := fmt.Sprintf("@%s", s.Args[i].Name)
		value := argValues[i]
		body = strings.Replace(body, argRef, value, -1)
	}
	return fmt.Sprintf("( %s )", body), nil
}

type TableSpec struct {
	IsTemp     bool                                              `json:"isTemp"`
	IsView     bool                                              `json:"isView"`
	NamePath   []string                                          `json:"namePath"`
	Columns    []*ColumnSpec                                     `json:"columns"`
	PrimaryKey []string                                          `json:"primaryKey"`
	CreateMode googlesql.ResolvedCreateStatementEnums_CreateMode `json:"createMode"`
	Query      string                                            `json:"query"`
	UpdatedAt  time.Time                                         `json:"updatedAt"`
	CreatedAt  time.Time                                         `json:"createdAt"`
	// Options carries the CREATE TABLE ... OPTIONS(name=value, ...)
	// clause. Each entry has a stable on-disk identity in the JSON
	// schema (the option name) and an opaque value string; the
	// INFORMATION_SCHEMA.TABLE_OPTIONS view surfaces them
	// one-row-per-(table, option).
	Options []*tableOptionSpec `json:"options,omitempty"`
}

// TableOptionSpec captures a single CREATE TABLE OPTIONS entry. Type
// follows the BigQuery convention (`STRING`, `ARRAY<STRING>`, etc).
// Value is the rendered SQL form of the option value.
type tableOptionSpec struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (s *TableSpec) Column(name string) *ColumnSpec {
	for _, col := range s.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}

func (s *TableSpec) TableName() string {
	return formatPath(s.NamePath)
}

func (s *TableSpec) SQLiteSchema() string {
	if s.IsView {
		return viewSQLiteSchema(s)
	}
	if s.Query != "" {
		return fmt.Sprintf("CREATE TABLE `%s` AS %s", s.TableName(), s.Query)
	}
	columns := []string{}
	for _, c := range s.Columns {
		columns = append(columns, c.SQLiteSchema())
	}
	if len(s.PrimaryKey) != 0 {
		columns = append(
			columns,
			fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(s.PrimaryKey, ",")),
		)
	}
	var stmt string
	switch s.CreateMode {
	case googlesql.ResolvedCreateStatementEnums_CreateModeCreateDefault:
		stmt = "CREATE TABLE"
	case googlesql.ResolvedCreateStatementEnums_CreateModeCreateOrReplace:
		stmt = "CREATE TABLE"
	case googlesql.ResolvedCreateStatementEnums_CreateModeCreateIfNotExists:
		stmt = "CREATE TABLE IF NOT EXISTS"
	}
	return fmt.Sprintf("%s `%s` (%s)", stmt, s.TableName(), strings.Join(columns, ","))
}

func viewSQLiteSchema(s *TableSpec) string {
	var stmt string
	switch s.CreateMode {
	case googlesql.ResolvedCreateStatementEnums_CreateModeCreateDefault:
		stmt = "CREATE VIEW"
	case googlesql.ResolvedCreateStatementEnums_CreateModeCreateOrReplace:
		stmt = "CREATE VIEW"
	case googlesql.ResolvedCreateStatementEnums_CreateModeCreateIfNotExists:
		stmt = "CREATE VIEW IF NOT EXISTS"
	}
	return fmt.Sprintf("%s `%s` AS %s", stmt, s.TableName(), s.Query)
}

type ColumnSpec struct {
	Name      string `json:"name"`
	Type      *Type  `json:"type"`
	IsNotNull bool   `json:"isNotNull"`
	// DefaultExpr is the formatted SQL of the column's DEFAULT
	// clause, captured at CREATE TABLE time. The InsertStmt
	// formatter splices it in for any column omitted from the
	// INSERT column list. Empty string means no default; the
	// downstream insert path leaves the column NULL when omitted.
	DefaultExpr string `json:"defaultExpr,omitempty"`
}

// TVFSpec captures a `CREATE TABLE FUNCTION` definition so the call
// site (`ResolvedTVFScan`) can be inlined as a subquery at format
// time. The body is the analyzed-and-formatted query of the TVF body
// in the same canonical form used by `TableSpec.Query` for views: a
// `SELECT col#id AS col, ... FROM (...)` envelope that exposes the
// TVF's declared output column names.
type TVFSpec struct {
	IsTemp        bool            `json:"isTemp"`
	NamePath      []string        `json:"namePath"`
	Args          []*NameWithType `json:"args"`
	OutputColumns []*ColumnSpec   `json:"outputColumns"`
	Body          string          `json:"body"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func (s *TVFSpec) TVFName() string {
	return formatPath(s.NamePath)
}

// CallSQL inlines the TVF body for a single ResolvedTVFScan
// invocation. argValues are the formatted SQL expressions for the
// TVF arguments, indexed by argument position. callOutputColumns are
// the canonical `name#id` names the analyzer assigned to the
// TVFScan output, in declaration order; we alias the body's output
// columns onto them so the surrounding query can reference them.
func (s *TVFSpec) CallSQL(argValues []string, callOutputColumns []string) (string, error) {
	if len(callOutputColumns) != len(s.OutputColumns) {
		return "", fmt.Errorf(
			"TVF %s: call site has %d output columns, spec declares %d",
			s.TVFName(), len(callOutputColumns), len(s.OutputColumns),
		)
	}
	body := s.Body
	for i, arg := range s.Args {
		if i >= len(argValues) {
			break
		}
		argRef := fmt.Sprintf("@%s", arg.Name)
		body = strings.Replace(body, argRef, argValues[i], -1)
	}
	projections := make([]string, 0, len(s.OutputColumns))
	for i, col := range s.OutputColumns {
		projections = append(
			projections,
			fmt.Sprintf("`%s` AS `%s`", col.Name, callOutputColumns[i]),
		)
	}
	return fmt.Sprintf(
		"SELECT %s FROM ( %s )",
		strings.Join(projections, ", "),
		body,
	), nil
}

type Type struct {
	Name          string                          `json:"name"`
	Kind          int                             `json:"kind"`
	SignatureKind googlesql.SignatureArgumentKind `json:"signatureKind"`
	ElementType   *Type                           `json:"elementType"`
	FieldTypes    []*NameWithType                 `json:"fieldTypes"`
}

func (t *Type) FunctionArgumentType() (*googlesql.FunctionArgumentType, error) {
	if t.SignatureKind != googlesql.SignatureArgumentKindArgTypeFixed {
		return m1(googlesql.NewFunctionArgumentType5(
			t.SignatureKind,
			m1(googlesql.NewFunctionArgumentTypeOptions()),
			-1,
		)), nil
	}
	typ, err := t.ToGoogleSQLType()
	if err != nil {
		return nil, err
	}
	opt := m1(googlesql.NewFunctionArgumentTypeOptions())
	return m1(googlesql.NewFunctionArgumentType(typ, opt, -1)), nil
}

// kindAs returns t.Kind as a googlesql.TypeKind for comparison with
// the enum constants the rest of the code uses.
func (t *Type) kindAs() googlesql.TypeKind { return googlesql.TypeKind(t.Kind) }

func (t *Type) IsArray() bool {
	return t.kindAs() == googlesql.TypeKindTypeArray
}

func (t *Type) IsStruct() bool {
	return t.kindAs() == googlesql.TypeKindTypeStruct
}

func (t *Type) AvailableAutoIndex() bool {
	switch t.kindAs() {
	case googlesql.TypeKindTypeBytes, googlesql.TypeKindTypeJson, googlesql.TypeKindTypeArray, googlesql.TypeKindTypeStruct,
		googlesql.TypeKindTypeGeography, googlesql.TypeKindTypeProto, googlesql.TypeKindTypeExtended:
		return false
	}
	return true
}

func (t *Type) GoReflectType() (reflect.Type, error) {
	switch t.kindAs() {
	case googlesql.TypeKindTypeInt32, googlesql.TypeKindTypeInt64, googlesql.TypeKindTypeUint32, googlesql.TypeKindTypeUint64:
		return reflect.TypeFor[int64](), nil
	case googlesql.TypeKindTypeBool:
		return reflect.TypeFor[bool](), nil
	case googlesql.TypeKindTypeFloat, googlesql.TypeKindTypeDouble:
		return reflect.TypeFor[float64](), nil
	case googlesql.TypeKindTypeBytes, googlesql.TypeKindTypeString, googlesql.TypeKindTypeNumeric, googlesql.TypeKindTypeBignumeric,
		googlesql.TypeKindTypeDate, googlesql.TypeKindTypeDatetime, googlesql.TypeKindTypeTime, googlesql.TypeKindTypeTimestamp, googlesql.TypeKindTypeInterval, googlesql.TypeKindTypeJson:
		return reflect.TypeFor[string](), nil
	case googlesql.TypeKindTypeArray:
		elem, err := t.ElementType.GoReflectType()
		if err != nil {
			return nil, err
		}
		return reflect.SliceOf(elem), nil
	case googlesql.TypeKindTypeStruct:
		return reflect.TypeOf(map[string]any{}), nil
	}
	return nil, fmt.Errorf("cannot convert %s to reflect.Type", t.Name)
}

func (t *Type) ToGoogleSQLType() (googlesql.Googlesql_TypeNode, error) {
	if t == nil {
		return nil, fmt.Errorf("nil Type cannot be converted to googlesql type")
	}
	switch t.kindAs() {
	case googlesql.TypeKindTypeProto:
		// Proto / Enum types cannot be rebuilt via MakeSimpleType.
		// They were originally created from a descriptor handle
		// registered through Catalog.RegisterProto / RegisterProtoMessage.
		// We cache those handles in process-global maps keyed by the
		// proto's fully-qualified name. Type.Name comes from
		// DebugString, which wraps the type in "PROTO<...>" or
		// "ENUM<...>"; strip that wrapper before looking up.
		name := protoNameFromDebug(t.Name)
		if pt := lookupRegisteredProtoType(name); pt != nil {
			return pt, nil
		}
		// Fallback: when the consumer has not registered the proto,
		// fall through to MakeSimpleType so existing call paths that
		// only care about the TypeKind continue to work.
	case googlesql.TypeKindTypeEnum:
		name := protoNameFromDebug(t.Name)
		if et := lookupRegisteredEnumType(name); et != nil {
			return et, nil
		}
	case googlesql.TypeKindTypeArray:
		if t.ElementType == nil {
			return nil, fmt.Errorf("ArrayType.ElementType is nil")
		}
		typ, err := t.ElementType.ToGoogleSQLType()
		if err != nil {
			return nil, err
		}
		return tf().MakeArrayType(typ)
	case googlesql.TypeKindTypeRange:
		if t.ElementType == nil {
			return nil, fmt.Errorf("RangeType.ElementType is nil")
		}
		elem, err := t.ElementType.ToGoogleSQLType()
		if err != nil {
			return nil, err
		}
		return tf().MakeRangeType3(elem)
	case googlesql.TypeKindTypeStruct:
		var fields []*googlesql.StructField
		for _, field := range t.FieldTypes {
			typ, err := field.Type.ToGoogleSQLType()
			if err != nil {
				return nil, err
			}
			fields = append(fields, &googlesql.StructField{Name: field.Name, Type_: typ})
		}
		return tf().MakeStructType(fields)
	}
	return m1(tf().MakeSimpleType(t.kindAs())), nil
}

func (t *Type) FormatType() string {
	switch t.kindAs() {
	case googlesql.TypeKindTypeStruct:
		formatTypes := make([]string, 0, len(t.FieldTypes))
		for _, field := range t.FieldTypes {
			formatTypes = append(formatTypes, fmt.Sprintf("`%s` %s", field.Name, field.Type.FormatType()))
		}
		return fmt.Sprintf("STRUCT<%s>", strings.Join(formatTypes, ","))
	case googlesql.TypeKindTypeArray:
		return fmt.Sprintf("ARRAY<%s>", t.ElementType.FormatType())
	}
	return typeKindToSQLName(t.kindAs())
}

// typeKindToSQLName returns the canonical SQL type-name string for a
// TypeKind — "INT64" rather than the enum's Go String() which is
// "TypeKindTypeInt64". The formatter embeds these names into
// re-analyzed SQL, where the extra prefix would crash the analyzer.
func typeKindToSQLName(kind googlesql.TypeKind) string {
	switch kind {
	case googlesql.TypeKindTypeInt32:
		return "INT32"
	case googlesql.TypeKindTypeInt64:
		return "INT64"
	case googlesql.TypeKindTypeUint32:
		return "UINT32"
	case googlesql.TypeKindTypeUint64:
		return "UINT64"
	case googlesql.TypeKindTypeBool:
		return "BOOL"
	case googlesql.TypeKindTypeFloat:
		return "FLOAT"
	case googlesql.TypeKindTypeDouble:
		return "DOUBLE"
	case googlesql.TypeKindTypeString:
		return "STRING"
	case googlesql.TypeKindTypeBytes:
		return "BYTES"
	case googlesql.TypeKindTypeDate:
		return "DATE"
	case googlesql.TypeKindTypeTimestamp:
		return "TIMESTAMP"
	case googlesql.TypeKindTypeDatetime:
		return "DATETIME"
	case googlesql.TypeKindTypeTime:
		return "TIME"
	case googlesql.TypeKindTypeInterval:
		return "INTERVAL"
	case googlesql.TypeKindTypeNumeric:
		return "NUMERIC"
	case googlesql.TypeKindTypeBignumeric:
		return "BIGNUMERIC"
	case googlesql.TypeKindTypeJson:
		return "JSON"
	case googlesql.TypeKindTypeGeography:
		return "GEOGRAPHY"
	}
	return fmt.Sprintf("%v", kind)
}

func (s *ColumnSpec) SQLiteSchema() string {
	var typ string
	switch googlesql.TypeKind(s.Type.Kind) {
	case googlesql.TypeKindTypeInt32, googlesql.TypeKindTypeInt64, googlesql.TypeKindTypeUint32, googlesql.TypeKindTypeUint64:
		typ = "INT"
	case googlesql.TypeKindTypeEnum:
		typ = "INT"
	case googlesql.TypeKindTypeBool:
		typ = "BOOLEAN"
	case googlesql.TypeKindTypeFloat:
		typ = "FLOAT"
	case googlesql.TypeKindTypeBytes:
		typ = "BLOB"
	case googlesql.TypeKindTypeDouble:
		typ = "DOUBLE"
	case googlesql.TypeKindTypeJson:
		typ = "JSON"
	case googlesql.TypeKindTypeString:
		typ = "TEXT"
	case googlesql.TypeKindTypeDate:
		typ = "TEXT"
	case googlesql.TypeKindTypeTimestamp:
		typ = "TEXT"
	case googlesql.TypeKindTypeArray:
		typ = "TEXT"
	case googlesql.TypeKindTypeStruct:
		typ = "TEXT"
	case googlesql.TypeKindTypeProto:
		typ = "TEXT"
	case googlesql.TypeKindTypeTime:
		typ = "TEXT"
	case googlesql.TypeKindTypeDatetime:
		typ = "TEXT"
	case googlesql.TypeKindTypeGeography:
		typ = "TEXT"
	case googlesql.TypeKindTypeNumeric:
		typ = "TEXT"
	case googlesql.TypeKindTypeBignumeric:
		typ = "TEXT"
	case googlesql.TypeKindTypeExtended:
		typ = "TEXT"
	case googlesql.TypeKindTypeInterval:
		typ = "TEXT"
	default:
		typ = "UNKNOWN"
	}
	schema := fmt.Sprintf("`%s` %s", s.Name, typ)
	if s.IsNotNull {
		schema += " NOT NULL"
	}
	return schema
}

func newTypeFromFunctionArgumentType(t *googlesql.FunctionArgumentType) *Type {
	if m1(t.IsTemplated()) {
		return &Type{SignatureKind: m1(t.Kind())}
	}
	return newType(m1(t.Type()))
}

func newFunctionSpec(ctx context.Context, namePath *NamePath, stmt *googlesql.ResolvedCreateFunctionStmt) (*FunctionSpec, error) {
	args := []*NameWithType{}
	signature, _ := stmt.Signature()
	for _, arg := range m1(signature.Arguments()) {
		args = append(args, &NameWithType{
			Name: m1(arg.ArgumentName()),
			Type: newTypeFromFunctionArgumentType(arg),
		})
	}

	var body string
	language, _ := stmt.Language()
	switch language {
	case "js":
		code, err := encodeGoValue(m1(tf().MakeSimpleType(googlesql.TypeKindTypeString)), m1(stmt.Code()))
		if err != nil {
			return nil, err
		}
		encodedType, err := json.Marshal(newType(m1(stmt.ReturnType())))
		if err != nil {
			return nil, err
		}
		retType, err := encodeGoValue(m1(tf().MakeSimpleType(googlesql.TypeKindTypeString)), string(encodedType))
		if err != nil {
			return nil, err
		}
		argParams := make([]string, 0, len(args))
		argNames := make([]string, 0, len(args))
		for _, arg := range args {
			argParams = append(argParams, fmt.Sprintf("@%s", arg.Name))
			argNames = append(argNames, arg.Name)
		}
		if len(argParams) == 0 {
			body = fmt.Sprintf("googlesqlite_eval_javascript('%s', '%s')", code, retType)
		} else {
			arr, err := encodeGoValue(m1(tf().MakeArrayType(m1(tf().MakeSimpleType(googlesql.TypeKindTypeString)))), argNames)
			if err != nil {
				return nil, err
			}
			body = fmt.Sprintf(
				"googlesqlite_eval_javascript('%s', '%s', '%s', %s)",
				code, retType, arr,
				strings.Join(argParams, ","),
			)
		}
	default:
		funcExpr, _ := stmt.FunctionExpression()
		if funcExpr != nil {
			bodyQuery, err := newNode(funcExpr).FormatSQL(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to format function expression: %w", err)
			}
			body = bodyQuery
		}
	}
	now := time.Now()
	return &FunctionSpec{
		IsTemp:    resolvedCreateScope(stmt) == googlesql.ResolvedCreateStatementEnums_CreateScopeCreateTemp,
		NamePath:  namePath.mergePath(m1(stmt.NamePath())),
		Args:      args,
		Return:    newType(m1(stmt.ReturnType())),
		Code:      m1(stmt.Code()),
		Body:      body,
		Language:  language,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func newTypeFromFunctionArgumentTypeByRealType(t *googlesql.FunctionArgumentType, realType googlesql.Googlesql_TypeNode) *Type {
	if m1(t.IsTemplated()) {
		if m1(realType.IsArray()) {
			return &Type{SignatureKind: googlesql.SignatureArgumentKindArgArrayTypeAny1}
		}
		return &Type{SignatureKind: googlesql.SignatureArgumentKindArgTypeAny1}
	}
	return newType(m1(t.Type()))
}

func newTemplatedFunctionSpec(ctx context.Context, namePath *NamePath, stmt *googlesql.ResolvedCreateFunctionStmt, realStmts []*googlesql.ResolvedCreateFunctionStmt) (*FunctionSpec, error) {
	signature, _ := stmt.Signature()
	arguments := m1(signature.Arguments())
	realStmt := realStmts[0]
	realSignature, _ := realStmt.Signature()
	realArguments := m1(realSignature.Arguments())
	resultType := newType(m1(m1(realSignature.ResultType()).Type()))
	resultTypeName := resultType.FormatType()

	allSameResultType := true
	for _, stmt := range realStmts {
		if newType(m1(m1(m1(stmt.Signature()).ResultType()).Type())).FormatType() != resultTypeName {
			allSameResultType = false
			break
		}
	}
	var retType *Type
	if allSameResultType {
		retType = resultType
	} else {
		retType = newTypeFromFunctionArgumentTypeByRealType(
			m1(signature.ResultType()),
			m1(m1(realSignature.ResultType()).Type()),
		)
	}
	args := []*NameWithType{}
	for i := range arguments {
		args = append(args, &NameWithType{
			Name: m1(arguments[i].ArgumentName()),
			Type: newTypeFromFunctionArgumentTypeByRealType(
				arguments[i],
				m1(realArguments[i].Type()),
			),
		})
	}
	funcExpr, _ := stmt.FunctionExpression()
	var body string
	if funcExpr != nil {
		bodyQuery, err := newNode(funcExpr).FormatSQL(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to format function expression: %w", err)
		}
		body = bodyQuery
	}
	now := time.Now()
	return &FunctionSpec{
		IsTemp:    resolvedCreateScope(stmt) == googlesql.ResolvedCreateStatementEnums_CreateScopeCreateTemp,
		NamePath:  namePath.mergePath(m1(stmt.NamePath())),
		Args:      args,
		Return:    retType,
		Code:      m1(stmt.Code()),
		Body:      body,
		Language:  m1(stmt.Language()),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func newColumnsFromDef(ctx context.Context, def []*googlesql.ResolvedColumnDefinition) []*ColumnSpec {
	columns := []*ColumnSpec{}
	for _, columnNode := range def {
		annotation, _ := columnNode.Annotations()
		var isNotNull bool
		if annotation != nil {
			// annotation.TypeParameters isn't exposed on the bridge
			// yet; keep the hook but skip type-param extraction.
			isNotNull, _ = annotation.NotNull()
		}
		var defaultExpr string
		if dv, _ := columnNode.DefaultValue(); dv != nil {
			if expr, _ := dv.Expression(); expr != nil {
				if sql, err := newNode(expr).FormatSQL(ctx); err == nil {
					defaultExpr = sql
				}
			}
		}
		columns = append(columns, &ColumnSpec{
			Name:        m1(columnNode.Name()),
			Type:        newType(m1(columnNode.Type())),
			IsNotNull:   isNotNull,
			DefaultExpr: defaultExpr,
		})
	}
	return columns
}

func newColumnsFromOutputColumns(def []*googlesql.ResolvedOutputColumn) []*ColumnSpec {
	columns := []*ColumnSpec{}
	for _, columnNode := range def {
		column, _ := columnNode.Column()

		columns = append(columns, &ColumnSpec{
			Name: m1(columnNode.Name()),
			Type: newType(m1(column.Type())),
		})
	}
	return columns
}

func newPrimaryKey(key *googlesql.ResolvedPrimaryKey) []string {
	if key == nil {
		return nil
	}
	names, _ := key.ColumnNameList()
	return names
}

func newTableSpecWithQuery(ctx context.Context, namePath *NamePath, query string, stmt *googlesql.ResolvedCreateTableStmt) *TableSpec {
	now := time.Now()
	return &TableSpec{
		IsTemp:     resolvedCreateScope(stmt) == googlesql.ResolvedCreateStatementEnums_CreateScopeCreateTemp,
		NamePath:   namePath.mergePath(m1(stmt.NamePath())),
		Columns:    newColumnsFromDef(ctx, m1(stmt.ColumnDefinitionList())),
		PrimaryKey: newPrimaryKey(m1(stmt.PrimaryKey())),
		CreateMode: m1(stmt.CreateMode()),
		Options:    newTableOptionsFromResolved(m1(stmt.OptionList())),
		UpdatedAt:  now,
		CreatedAt:  now,
	}
}

// normaliseStringLiteralQuotes rewrites a double-quoted STRING
// literal (`"foo"`) into the single-quoted BigQuery canonical form
// (`'foo'`), preserving any escape sequence and doubling any single
// quote that appears in the original payload. Other shapes (numeric,
// boolean, already single-quoted) pass through unchanged.
func normaliseStringLiteralQuotes(s string) string {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return s
	}
	body := s[1 : len(s)-1]
	body = strings.ReplaceAll(body, `\"`, `"`)
	body = strings.ReplaceAll(body, "'", "''")
	return "'" + body + "'"
}

// newTableOptionsFromResolved extracts the CREATE TABLE OPTIONS
// clause into the on-disk TableOptionSpec format. Only literal-valued
// options are preserved; non-literal value exprs (parameter refs,
// sub-queries, etc.) skip the value entry — `INFORMATION_SCHEMA.
// TABLE_OPTIONS` is purely informational, so a missing value just
// means the option is recorded by name only.
//
// Each call uses `Value.GetSQLLiteral()` for the value text and
// `Value.TypeKind()` for the declared type. We deliberately avoid
// touching `Value.Type()` (the wasm-side Type handle) here: pulling
// it out keeps a wasm sub-handle alive that another goroutine's
// finalizer can deadlock against if a process-global wasm Module
// mutex is contended.
func newTableOptionsFromResolved(opts []*googlesql.ResolvedOption) []*tableOptionSpec {
	if len(opts) == 0 {
		return nil
	}
	out := make([]*tableOptionSpec, 0, len(opts))
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		name, _ := opt.Name()
		expr, _ := opt.Value()
		spec := &tableOptionSpec{Name: name}
		if lit, ok := expr.(*googlesql.ResolvedLiteral); ok && lit != nil {
			if vp, err := lit.Value(); err == nil && vp != nil {
				v := *vp
				if k, err := v.TypeKind(); err == nil {
					spec.Type = typeKindToSQLName(k)
				}
				if s, err := v.GetSQLLiteral(); err == nil {
					spec.Value = normaliseStringLiteralQuotes(s)
				}
			}
		}
		out = append(out, spec)
	}
	return out
}

func newTVFSpec(ctx context.Context, namePath *NamePath, stmt *googlesql.ResolvedCreateTableFunctionStmt) (*TVFSpec, error) {
	args := []*NameWithType{}
	signature, _ := stmt.Signature()
	for _, arg := range m1(signature.Arguments()) {
		args = append(args, &NameWithType{
			Name: m1(arg.ArgumentName()),
			Type: newTypeFromFunctionArgumentType(arg),
		})
	}
	innerScan, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to read TVF body: %w", err)
	}
	if innerScan == nil {
		return nil, fmt.Errorf("TVF body is missing for %s", strings.Join(m1(stmt.NamePath()), "."))
	}
	innerBody, err := newNode(innerScan).FormatSQL(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to format TVF body: %w", err)
	}
	outList := m1(stmt.OutputColumnList())
	projections := make([]string, 0, len(outList))
	outputColumns := make([]*ColumnSpec, 0, len(outList))
	for _, column := range outList {
		colName, _ := column.Name()
		col, _ := column.Column()
		refColumnName, _ := col.Name()
		colID, _ := col.ColumnId()
		projections = append(
			projections,
			fmt.Sprintf("`%s#%d` AS `%s`", refColumnName, colID, colName),
		)
		outputColumns = append(outputColumns, &ColumnSpec{
			Name: colName,
			Type: newType(m1(col.Type())),
		})
	}
	body := fmt.Sprintf("SELECT %s FROM (%s)", strings.Join(projections, ","), innerBody)
	now := time.Now()
	return &TVFSpec{
		IsTemp:        resolvedCreateScope(stmt) == googlesql.ResolvedCreateStatementEnums_CreateScopeCreateTemp,
		NamePath:      namePath.mergePath(m1(stmt.NamePath())),
		Args:          args,
		OutputColumns: outputColumns,
		Body:          body,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func newTableAsViewSpec(namePath *NamePath, query string, stmt *googlesql.ResolvedCreateViewStmt) *TableSpec {
	var outputColumns []string
	outList, _ := stmt.OutputColumnList()
	for _, column := range outList {
		colName, _ := column.Name()
		col, _ := column.Column()
		refColumnName, _ := col.Name()
		colID, _ := col.ColumnId()
		outputColumns = append(
			outputColumns,
			fmt.Sprintf("`%s#%d` AS `%s`", refColumnName, colID, colName),
		)
	}
	now := time.Now()
	return &TableSpec{
		IsTemp:     resolvedCreateScope(stmt) == googlesql.ResolvedCreateStatementEnums_CreateScopeCreateTemp,
		IsView:     true,
		NamePath:   namePath.mergePath(m1(stmt.NamePath())),
		Columns:    newColumnsFromOutputColumns(m1(stmt.OutputColumnList())),
		CreateMode: m1(stmt.CreateMode()),
		Query:      fmt.Sprintf("SELECT %s FROM (%s)", strings.Join(outputColumns, ","), query),
		UpdatedAt:  now,
		CreatedAt:  now,
	}
}

func newTableAsSelectSpec(ctx context.Context, namePath *NamePath, query string, stmt *googlesql.ResolvedCreateTableAsSelectStmt) *TableSpec {
	var outputColumns []string
	for _, column := range m1(stmt.OutputColumnList()) {
		colName, _ := column.Name()
		refColumnName, _ := m1(column.Column()).Name()
		colID, _ := m1(column.Column()).ColumnId()
		outputColumns = append(
			outputColumns,
			fmt.Sprintf("`%s#%d` AS `%s`", refColumnName, colID, colName),
		)
	}
	now := time.Now()
	return &TableSpec{
		IsTemp:     resolvedCreateScope(stmt) == googlesql.ResolvedCreateStatementEnums_CreateScopeCreateTemp,
		NamePath:   namePath.mergePath(m1(stmt.NamePath())),
		Columns:    newColumnsFromDef(ctx, m1(stmt.ColumnDefinitionList())),
		PrimaryKey: newPrimaryKey(m1(stmt.PrimaryKey())),
		CreateMode: m1(stmt.CreateMode()),
		Query:      fmt.Sprintf("SELECT %s FROM (%s)", strings.Join(outputColumns, ","), query),
		UpdatedAt:  now,
		CreatedAt:  now,
	}
}

func newType(t googlesql.Googlesql_TypeNode) *Type {
	// Googlesql_TypeNode exposes the base-class accessors directly.
	kind := m1(t.Kind())
	var (
		elem       *Type
		fieldTypes []*NameWithType
	)
	// Composite types need the nested element information so
	// CAST(... AS ARRAY<T>) and friends can round-trip through the
	// formatter. ArrayType.ElementType and StructType.Fields are now
	// exposed on the bridge; recurse into them when the dynamic type
	// matches.
	switch kind {
	case googlesql.TypeKindTypeArray:
		if at, ok := t.(*googlesql.ArrayType); ok {
			if e, err := at.ElementType(); err == nil && e != nil {
				elem = newType(e)
			}
		}
	case googlesql.TypeKindTypeRange:
		if rt, ok := t.(*googlesql.RangeType); ok {
			if e, err := rt.ElementType(); err == nil && e != nil {
				elem = newType(e)
			}
		}
	case googlesql.TypeKindTypeStruct:
		if st, ok := t.(*googlesql.StructType); ok {
			fields, _ := st.Fields()
			for _, field := range fields {
				if field == nil || field.Type_ == nil {
					continue
				}
				fieldTypes = append(fieldTypes, &NameWithType{
					Name: field.Name,
					Type: newType(field.Type_),
				})
			}
		}
	}
	// Googlesql_TypeNode interface does not expose TypeName(mode); use
	// DebugString instead which gives an equivalent printable form.
	name, _ := t.DebugString(false)
	return &Type{
		Name:          name,
		Kind:          int(kind),
		SignatureKind: googlesql.SignatureArgumentKindArgTypeFixed,
		ElementType:   elem,
		FieldTypes:    fieldTypes,
	}
}
