package internal

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"

	"github.com/goccy/googlesqlite/internal/functions/window"
)

type Formatter interface {
	FormatSQL(context.Context) (string, error)
}

func New(node googlesql.ResolvedNode) Formatter {
	return newNode(node)
}

func getTableName(ctx context.Context, n googlesql.ResolvedNode) (string, error) {
	// Preferred path: pull the bound Table off a ResolvedTableScan and
	// ask it for its name. The go-googlesql version used a side-car
	// nodeMap that mapped ResolvedTableScan → ASTPathExpression to
	// recover the original user-supplied identifier, but that side
	// channel is not exposed through the wasm bridge. For the common
	// case (single-segment table names) the catalog-bound Table.Name()
	// is equivalent.
	if scan, ok := n.(*googlesql.ResolvedTableScan); ok {
		if table, err := scan.Table(); err == nil && table != nil {
			// Preferred: look the table's storage name (the SQLite
			// table identifier built from the spec's full namespace
			// path) up against our local catalog index. This works
			// regardless of which sub-catalog level the analyzer
			// resolved through, so it is robust to the user
			// supplying a path shorter than the namespace
			// (e.g. `dataset.table` while the conn's NamePath is
			// only the project).
			if a := analyzerFromContext(ctx); a != nil && a.catalog != nil {
				if storage := a.catalog.StorageNameForTable(table); storage != "" {
					return storage, nil
				}
			}
			if name, err := table.Name(); err == nil && name != "" {
				namePath := namePathFromContext(ctx)
				return namePath.format([]string{name}), nil
			}
		}
	}
	return "", fmt.Errorf("failed to find path node from table node %T", n)
}

func getFuncName(ctx context.Context, n googlesql.ResolvedNode) (string, error) {
	type fnCall interface {
		Function() (*googlesql.Function, error)
	}
	// Preferred: look the resolved Function up in our local catalog
	// index. This works regardless of which sub-catalog level the
	// analyzer resolved through, so it stays correct when the user
	// supplies a path shorter than the full namespace.
	if fc, ok := n.(fnCall); ok {
		if fn, err := fc.Function(); err == nil && fn != nil {
			if a := analyzerFromContext(ctx); a != nil && a.catalog != nil {
				if storage := a.catalog.StorageNameForFunction(fn); storage != "" {
					return storage, nil
				}
			}
		}
	}
	// Fallback: catalog index didn't have this function (e.g. a
	// built-in never registered through addFunctionSpec). Read the
	// function identity straight off the resolved node and merge
	// with the conn's NamePath.
	namePath := namePathFromContext(ctx)
	if fc, ok := n.(fnCall); ok {
		if fn, err := fc.Function(); err == nil && fn != nil {
			if name, err := fn.Name(); err == nil && name != "" {
				return namePath.format([]string{name}), nil
			}
		}
	}
	return "", fmt.Errorf("failed to find path node from function node %T", n)
}

func uniqueColumnName(ctx context.Context, col *googlesql.ResolvedColumn) string {
	colName, _ := col.Name()
	if useTableNameForColumn(ctx) {
		return fmt.Sprintf("%s.%s", m1(col.TableName()), colName)
	}
	if useColumnID(ctx) {
		colID, _ := col.ColumnId()
		return fmt.Sprintf("%s#%d", colName, colID)
	}
	return colName
}

type inputPattern int

const (
	InputKeep      inputPattern = 0
	InputNeedsWrap inputPattern = 1
	InputNeedsFrom inputPattern = 2
)

func getInputPattern(input string) inputPattern {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return InputKeep
	}
	if strings.HasPrefix(trimmed, "FROM") {
		return InputKeep
	}
	if strings.HasPrefix(trimmed, "SELECT") {
		return InputNeedsWrap
	}
	if strings.HasPrefix(trimmed, "WITH") {
		return InputNeedsWrap
	}
	return InputNeedsFrom
}

func formatInput(input string) (string, error) {
	switch getInputPattern(input) {
	case InputKeep:
		return input, nil
	case InputNeedsWrap:
		return fmt.Sprintf("FROM (%s)", input), nil
	case InputNeedsFrom:
		return fmt.Sprintf("FROM %s", input), nil
	}
	return "", fmt.Errorf("unexpected input pattern: %s", input)
}

func getFuncNameAndArgs(ctx context.Context, node *ResolvedBaseFunctionCallNode, isWindowFunc bool) (string, []string, error) {
	args := []string{}
	for _, a := range m1(node.ArgumentList()) {
		arg, err := newNode(a).FormatSQL(ctx)
		if err != nil {
			return "", nil, err
		}
		args = append(args, arg)
	}
	funcName := m1(m1(node.Function()).FullName(false))
	funcName = strings.Replace(funcName, ".", "_", -1)

	_, existsCurrentTimeFunc := currentTimeFuncMap[funcName]
	_, existsNormalFunc := normalFuncMap[funcName]
	_, existsAggregateFunc := aggregateFuncMap[funcName]
	_, existsWindowFunc := windowFuncMap[funcName]
	// $-prefixed builtin names (e.g. $array_at_offset, $divide) are
	// registered without the prefix; treat the unprefixed form as the
	// real key when deciding whether a safe variant exists.
	existsNormalFuncForSafe := existsNormalFunc
	if !existsNormalFuncForSafe && strings.HasPrefix(funcName, "$") {
		_, existsNormalFuncForSafe = normalFuncMap[funcName[1:]]
	}
	currentTime := CurrentTime(ctx)

	funcPrefix := "googlesqlite"
	if m1(node.ErrorMode()) == googlesql.ResolvedFunctionCallBaseEnums_ErrorModeSafeErrorMode {
		if !existsNormalFuncForSafe {
			return "", nil, fmt.Errorf("SAFE is not supported for function %s", funcName)
		}
		funcPrefix = "googlesqlite_safe"
	} else if inSafeEvalMode(ctx) && existsNormalFuncForSafe {
		// IFERROR / ISERROR / NULLIFERROR sub-context: route through the
		// safe variant so a runtime failure folds to NULL instead of
		// aborting the statement. Functions that have no safe variant
		// stay on the raising form — callers that hit them inside an
		// error-handling expression need to use SAFE.<func> explicitly.
		funcPrefix = "googlesqlite_safe"
	}

	if strings.HasPrefix(funcName, "$") {
		if isWindowFunc {
			funcName = fmt.Sprintf("%s_window_%s", funcPrefix, funcName[1:])
		} else {
			funcName = fmt.Sprintf("%s_%s", funcPrefix, funcName[1:])
		}
	} else if existsCurrentTimeFunc {
		if currentTime != nil {
			args = append(
				args,
				fmt.Sprint(currentTime.UnixNano()),
			)
		}
		funcName = fmt.Sprintf("%s_%s", funcPrefix, funcName)
	} else if existsNormalFunc {
		funcName = fmt.Sprintf("%s_%s", funcPrefix, funcName)
	} else if !isWindowFunc && existsAggregateFunc {
		funcName = fmt.Sprintf("%s_%s", funcPrefix, funcName)
	} else if isWindowFunc && existsWindowFunc {
		funcName = fmt.Sprintf("%s_window_%s", funcPrefix, funcName)
	} else {
		fname, err := getFuncName(ctx, node)
		if err != nil {
			return "", nil, err
		}
		funcName = fname
	}
	return funcName, args, nil
}

func (n *LiteralNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	return literalFromGoogleSQLValue(*m1(n.node.Value()))
}

func (n *ParameterNode) FormatSQL(ctx context.Context) (string, error) {
	if c := paramCollectorFromContext(ctx); c != nil && n.node != nil {
		// Named parameters dedup by name in getParamsFromNode; for
		// positional `?` parameters we want every occurrence in
		// AST order, so just append unconditionally here.
		c.params = append(c.params, n.node)
	}
	name, _ := n.node.Name()
	if name == "" {
		return "?", nil
	}
	return fmt.Sprintf("@%s", name), nil
}

func (n *ColumnRefNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	columnMap := columnRefMap(ctx)
	col, _ := n.node.Column()
	colName := uniqueColumnName(ctx, col)
	if ref, exists := columnMap[colName]; exists {
		delete(columnMap, colName)
		return ref, nil
	}
	return fmt.Sprintf("`%s`", colName), nil
}

func (n *SystemVariableNode) FormatSQL(ctx context.Context) (string, error) {
	if n == nil || n.node == nil {
		return "", nil
	}
	pathParts, _ := n.node.NamePath()
	name := strings.Join(pathParts, ".")
	// Read directly off the Conn so values written by an earlier SET
	// in the same connection (or from another driver entrypoint that
	// only sees Conn.SetSystemVariable) are observable. The
	// systemVarsKey snapshot is taken at Analyze-time and would miss
	// SETs issued during this exact analyze cycle.
	var raw string
	var have bool
	if conn := connFromContext(ctx); conn != nil {
		if v, ok := conn.SystemVariable(name); ok {
			raw, have = v, true
		}
	}
	if !have {
		if vars := systemVarsFromContext(ctx); vars != nil {
			if v, ok := vars[name]; ok {
				raw, have = v, true
			}
		}
	}
	if !have {
		return "NULL", nil
	}
	// Surrounding query will scan the result through the value
	// decoder, which expects the googlesqlite envelope. Emit a
	// StringValue literal so the round-trip works.
	return literalFromValue(value.StringValue(raw))
}

// FilterFieldNode lowers a ResolvedFilterField (the analyzer node
// FILTER_FIELDS resolves into) to a runtime call. Each filter-field
// arg carries an include/exclude flag plus a dotted path of
// FieldDescriptors. We flatten the path into the dot-separated field
// numbers our runtime expects ("1.2" for `type.award_name`), prefix
// the path with `+`/`-`, and concatenate every arg into a single
// path-spec string.
func (n *FilterFieldNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	exprNode := m1(n.node.Expr())
	exprSQL, err := newNode(exprNode).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	argList, err := n.node.FilterFieldArgList()
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(argList))
	for _, a := range argList {
		if a == nil {
			continue
		}
		include, _ := a.Include()
		fdPath, _ := a.FieldDescriptorPath()
		nums := make([]string, 0, len(fdPath))
		for _, fd := range fdPath {
			num, _ := fd.Number()
			nums = append(nums, strconv.Itoa(int(num)))
		}
		sign := "+"
		if !include {
			sign = "-"
		}
		parts = append(parts, sign+strings.Join(nums, "."))
	}
	pathSpec := strings.Join(parts, ",")
	resetReq, _ := n.node.ResetClearedRequiredFields()
	protoName := ""
	if t, _ := exprNode.Type(); t != nil {
		ds, _ := t.DebugString(false)
		protoName = protoNameFromDebug(ds)
	}
	return fmt.Sprintf("googlesqlite_filter_fields(%s, %s, %t, %s)",
		exprSQL,
		protoFormatLiteralString(pathSpec),
		resetReq,
		protoFormatLiteralString(protoName),
	), nil
}

func (n *FunctionCallNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	// Special-case the graph-element built-ins (LABELS, ELEMENT_ID,
	// PROPERTY_NAMES, SOURCE_NODE_ID, DESTINATION_NODE_ID,
	// ELEMENT_DEFINITION_NAME). They depend on the per-call graph
	// element metadata recorded in graphElementContext, so we lower
	// them here before the default funcName/args formatter kicks
	// in. See internal/graph_scan_node.go.
	if sql, ok, err := tryFormatGraphElementFunc(ctx, n.node); ok || err != nil {
		return sql, err
	}
	funcName, args, err := getFuncNameAndArgs(ctx, n.node.ResolvedFunctionCallBase, false)
	if err != nil {
		return "", err
	}
	switch funcName {
	case "googlesqlite_from_proto":
		return n.formatFromProto(ctx, args)
	case "googlesqlite_to_proto":
		return n.formatToProto(ctx, args)
	case "googlesqlite_filter_fields":
		return n.formatFilterFields(ctx, args)
	case "googlesqlite_replace_fields":
		return n.formatReplaceFields(ctx, args)
	case "googlesqlite_proto_modify_map":
		return n.formatProtoModifyMap(ctx, args)
	case "googlesqlite_proto_map_contains_key":
		return n.formatProtoMapContainsKey(ctx, args)
	case "googlesqlite_enum_value_descriptor_proto":
		return n.formatEnumValueDescriptorProto(ctx, args)
	case "googlesqlite_iferror", "googlesqlite_iserror", "googlesqlite_nulliferror":
		return n.formatErrorHandling(ctx, funcName, args)
	case "googlesqlite_error", "googlesqlite_safe_error":
		return n.formatErrorBuiltin(ctx, funcName, args)
	case "googlesqlite_ifnull":
		return fmt.Sprintf(
			"CASE WHEN %s IS NULL THEN %s ELSE %s END",
			args[0],
			args[1],
			args[0],
		), nil
	case "googlesqlite_if":
		return fmt.Sprintf(
			"CASE WHEN %s THEN %s ELSE %s END",
			args[0],
			args[1],
			args[2],
		), nil
	case "googlesqlite_case_no_value":
		return n.formatCaseNoValue(ctx, args)
	case "googlesqlite_case_with_value":
		return n.formatCaseWithValue(ctx, args)
	}
	funcMap := funcMapFromContext(ctx)
	if spec, exists := funcMap[funcName]; exists {
		return spec.CallSQL(ctx, n.node.ResolvedFunctionCallBase, args)
	}
	return fmt.Sprintf(
		"%s(%s)",
		funcName,
		strings.Join(args, ","),
	), nil
}

// formatFromProto lowers FROM_PROTO. It takes a proto-typed argument
// and returns a SQL primitive whose shape is determined by the resolved
// signature. Peek at the call's return type to dispatch.
func (n *FunctionCallNode) formatFromProto(_ context.Context, args []string) (string, error) {
	retType, _ := n.node.Type()
	if retType == nil {
		return fmt.Sprintf("googlesqlite_from_proto(%s, 'bytes')", args[0]), nil
	}
	kind, _ := retType.Kind()
	return fmt.Sprintf("googlesqlite_from_proto(%s, '%s')", args[0], protoFieldKindString(kind)), nil
}

// formatToProto lowers TO_PROTO. It takes a SQL primitive and returns
// the matching well-known proto wrapper bytes. Use the input argument's
// kind via the call's argument signature.
func (n *FunctionCallNode) formatToProto(_ context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("TO_PROTO: missing argument")
	}
	argList, _ := n.node.ArgumentList()
	argKind := "bytes"
	if len(argList) > 0 {
		if t, _ := argList[0].Type(); t != nil {
			if k, err := t.Kind(); err == nil {
				argKind = protoFieldKindString(k)
			}
		}
	}
	return fmt.Sprintf("googlesqlite_to_proto(%s, '%s')", args[0], argKind), nil
}

// formatFilterFields lowers FILTER_FIELDS(proto, '+a.b.c', '-d.e', ...).
// The analyzer surfaces each include/exclude path as a separate STRING
// literal argument. We pass the joined comma-separated path list to the
// runtime so it can decode include/exclude via the leading sign.
func (n *FunctionCallNode) formatFilterFields(_ context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("FILTER_FIELDS: missing argument")
	}
	paths := strings.Join(args[1:], ",")
	return fmt.Sprintf("googlesqlite_filter_fields(%s, %s)", args[0], protoFormatLiteralString(paths)), nil
}

// formatReplaceFields lowers
// REPLACE_FIELDS(proto, (value, 'field_path'), (value, 'field_path'), ...).
// The analyzer represents each replacement as a (value, path) pair
// surfaced as 2 ResolvedExpr args back-to-back. We pull the field
// number / kind from the path metadata embedded in the resolved
// signature; for the stub-grade implementation we accept a single
// (value, path) pair plus arbitrary additional pairs by walking
// args two at a time.
func (n *FunctionCallNode) formatReplaceFields(_ context.Context, args []string) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("REPLACE_FIELDS: needs at least one (value, path) pair")
	}
	// args: [proto, val1, path1, val2, path2, ...]
	// Lower to nested replace_fields_one(...) calls so the runtime
	// keeps a single field-replace primitive.
	inner := args[0]
	argList, _ := n.node.ArgumentList()
	pairOffset := 1
	for pairOffset+1 < len(args) {
		val := args[pairOffset]
		pathSql := args[pairOffset+1]
		// Look up the value argument's resolved type kind from the
		// FunctionCall's argument list (skip the proto arg at idx 0).
		kind := "bytes"
		if pairOffset < len(argList) {
			if t, _ := argList[pairOffset].Type(); t != nil {
				if k, err := t.Kind(); err == nil {
					kind = protoFieldKindString(k)
				}
			}
		}
		inner = fmt.Sprintf("googlesqlite_replace_fields(%s, %s, '%s', %s)", inner, pathSql, kind, val)
		pairOffset += 2
	}
	return inner, nil
}

// formatProtoModifyMap lowers
// PROTO_MODIFY_MAP(map_field, key1, value1, key2, value2, ...).
// RewriteProtoMapFns is disabled so the analyzer leaves the call shape
// untouched. We lower to
//
//	googlesqlite_proto_modify_map(parent_proto, map_field_tag,
//	  key_kind, val_kind, key1, value1, key2, value2, ...)
//
// and let the runtime walk + rewrite the parent message's map field
// entries.
func (n *FunctionCallNode) formatProtoModifyMap(ctx context.Context, args []string) (string, error) {
	argList, _ := n.node.ArgumentList()
	if len(args) < 3 || len(argList) < 3 {
		return "", fmt.Errorf("PROTO_MODIFY_MAP: invalid number of arguments: got %d, want at least 3", len(args))
	}
	if (len(args)-1)%2 != 0 {
		return "", fmt.Errorf("PROTO_MODIFY_MAP: key/value pairs must come in pairs (got %d trailing args)", len(args)-1)
	}
	parentSQL, mapFieldTag, err := mapFieldParentAndTag(ctx, argList[0])
	if err != nil {
		return "", fmt.Errorf("PROTO_MODIFY_MAP: %w", err)
	}
	keyKind := "bytes"
	valKind := "bytes"
	if t, _ := argList[1].Type(); t != nil {
		if k, err := t.Kind(); err == nil {
			keyKind = protoFieldKindString(k)
		}
	}
	if t, _ := argList[2].Type(); t != nil {
		if k, err := t.Kind(); err == nil {
			valKind = protoFieldKindString(k)
		}
	}
	out := []string{parentSQL, strconv.Itoa(int(mapFieldTag)), "'" + keyKind + "'", "'" + valKind + "'"}
	out = append(out, args[1:]...)
	return fmt.Sprintf("googlesqlite_proto_modify_map(%s)", strings.Join(out, ", ")), nil
}

// formatProtoMapContainsKey lowers PROTO_MAP_CONTAINS_KEY(map_field, key).
// Lower to googlesqlite_proto_map_contains_key(parent_proto,
// map_field_tag, key_kind, key) so the runtime can walk every
// occurrence of the map field's outer tag inside the parent message
// (GetProtoField on a repeated field only keeps the last occurrence's
// payload, which loses the entries needed for the key lookup).
// RewriteProtoMapFns is disabled so the analyzer leaves the call shape
// untouched.
func (n *FunctionCallNode) formatProtoMapContainsKey(ctx context.Context, args []string) (string, error) {
	argList, _ := n.node.ArgumentList()
	if len(args) < 2 || len(argList) < 2 {
		return "", fmt.Errorf("PROTO_MAP_CONTAINS_KEY: invalid number of arguments: got %d, want 2", len(args))
	}
	keyKind := "bytes"
	if t, _ := argList[1].Type(); t != nil {
		if k, err := t.Kind(); err == nil {
			keyKind = protoFieldKindString(k)
		}
	}
	parentSQL, mapFieldTag, err := mapFieldParentAndTag(ctx, argList[0])
	if err != nil {
		return "", fmt.Errorf("PROTO_MAP_CONTAINS_KEY: %w", err)
	}
	return fmt.Sprintf("googlesqlite_proto_map_contains_key(%s, %d, '%s', %s)", parentSQL, mapFieldTag, keyKind, args[1]), nil
}

// formatEnumValueDescriptorProto lowers ENUM_VALUE_DESCRIPTOR_PROTO(enum_value).
// Look up the enum's full name via the registered enum type so the
// runtime can emit name + number; without a registered enum we fall
// back to number-only.
func (n *FunctionCallNode) formatEnumValueDescriptorProto(_ context.Context, args []string) (string, error) {
	argList, _ := n.node.ArgumentList()
	enumName := ""
	if len(argList) > 0 {
		if t, _ := argList[0].Type(); t != nil {
			ds, _ := t.DebugString(false)
			enumName = protoNameFromDebug(ds)
		}
	}
	return fmt.Sprintf("googlesqlite_enum_value_descriptor_proto(%s, '%s')", args[0], enumName), nil
}

// formatErrorHandling lowers the IFERROR / ISERROR / NULLIFERROR family.
// SQLite has no UDF-level try / catch (the VDBE aborts the statement as
// soon as any function returns non-nil error), so we lower the
// error-handling family by reformatting the inner argument under a
// safe-eval sub-context. Every reachable function call switches to its
// `googlesqlite_safe_<name>` variant (NULL on failure), and a nested
// ERROR(msg) short-circuits to NULL via the inSafeEvalMode branch on
// `error`. NULL therefore stands in for "an error happened", which we
// surface back to the caller via simple SQL CASEs.
func (n *FunctionCallNode) formatErrorHandling(ctx context.Context, funcName string, args []string) (string, error) {
	safeArgs := m1(n.node.ArgumentList())
	if len(safeArgs) < 1 {
		return "", fmt.Errorf("%s: missing argument", funcName)
	}
	safeCtx := withSafeEvalMode(ctx)
	safeX, err := newNode(safeArgs[0]).FormatSQL(safeCtx)
	if err != nil {
		return "", err
	}
	switch funcName {
	case "googlesqlite_iferror":
		if len(args) < 2 {
			return "", fmt.Errorf("IFERROR: needs catch_expression")
		}
		return fmt.Sprintf("CASE WHEN (%s) IS NULL THEN %s ELSE (%s) END", safeX, args[1], safeX), nil
	case "googlesqlite_iserror":
		return fmt.Sprintf("((%s) IS NULL)", safeX), nil
	case "googlesqlite_nulliferror":
		return fmt.Sprintf("(%s)", safeX), nil
	}
	return safeX, nil
}

// formatErrorBuiltin lowers ERROR / SAFE_ERROR. Inside a safe-eval
// sub-context, ERROR(msg) folds to NULL so IFERROR / ISERROR /
// NULLIFERROR can observe "an error happened" without aborting the
// statement. Outside that context, fall through to the regular UDF that
// raises.
func (n *FunctionCallNode) formatErrorBuiltin(ctx context.Context, funcName string, args []string) (string, error) {
	if inSafeEvalMode(ctx) {
		return "NULL", nil
	}
	// Fall through to the default emission below.
	funcMap := funcMapFromContext(ctx)
	if spec, exists := funcMap[funcName]; exists {
		return spec.CallSQL(ctx, n.node.ResolvedFunctionCallBase, args)
	}
	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ",")), nil
}

// formatCaseNoValue lowers a searched CASE expression
// (CASE WHEN ... THEN ... [ELSE ...] END).
func (n *FunctionCallNode) formatCaseNoValue(_ context.Context, args []string) (string, error) {
	var whenStmts []string
	for i := 0; i < len(args)-1; i += 2 {
		whenStmts = append(whenStmts, fmt.Sprintf("WHEN %s THEN %s", args[i], args[i+1]))
	}
	stmt := fmt.Sprintf("CASE %s", strings.Join(whenStmts, " "))
	// if args length is odd number, else statement exists.
	if len(args) > (len(args)/2)*2 {
		stmt += fmt.Sprintf(" ELSE %s", args[len(args)-1])
	}
	stmt += " END"
	return stmt, nil
}

// formatCaseWithValue lowers a simple CASE expression
// (CASE value WHEN ... THEN ... [ELSE ...] END).
func (n *FunctionCallNode) formatCaseWithValue(_ context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("not enough arguments for case with value")
	}
	val := args[0]
	args = args[1:]
	var whenStmts []string
	for i := 0; i < len(args)-1; i += 2 {
		whenStmts = append(whenStmts, fmt.Sprintf("WHEN %s THEN %s", args[i], args[i+1]))
	}
	stmt := fmt.Sprintf("CASE %s %s", val, strings.Join(whenStmts, " "))
	// if args length is odd number, else statement exists.
	if len(args) > (len(args)/2)*2 {
		stmt += fmt.Sprintf(" ELSE %s", args[len(args)-1])
	}
	stmt += " END"
	return stmt, nil
}

func (n *AggregateFunctionCallNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	// Special-case AGG(<graph_get_element_property>) where the
	// property is a MEASURE. Upstream's measure rewriter refuses
	// to handle this shape (it wants AGG over a direct ColumnRef
	// only), so we lower it ourselves: emit
	//   agg(<measure_inner_expr>, <locking_key>, '<kind>')
	// against the locally-registered `googlesqlite_agg` aggregate.
	// See internal/functions/aggregate/agg.go for the runtime.
	if sql, ok, err := tryFormatMeasureAGG(ctx, n.node); ok || err != nil {
		return sql, err
	}
	funcName, args, err := getFuncNameAndArgs(ctx, n.node.ResolvedFunctionCallBase, false)
	if err != nil {
		return "", err
	}
	// HAVING { MAX | MIN } modifier rewrite. The resolved tree
	// carries the modifier on ResolvedAggregateFunctionCall, but
	// SQLite has no equivalent. For ANY_VALUE we rewrite to a
	// custom googlesqlite_having_any_value(value, having_expr,
	// 'MAX'|'MIN') that filters rows down to those matching the
	// extremum of having_expr at Done() time.
	if havingMod := m1(n.node.HavingModifier()); havingMod != nil && (funcName == "any_value" || funcName == "googlesqlite_any_value") {
		havingExpr, herr := newNode(m1(havingMod.HavingExpr())).FormatSQL(ctx)
		if herr != nil {
			return "", herr
		}
		kind := "MAX"
		if k, kerr := havingMod.Kind(); kerr == nil {
			if k == googlesql.ResolvedAggregateHavingModifierEnums_HavingModifierKindMin {
				kind = "MIN"
			}
		}
		kindLit, _ := literalFromValue(value.StringValue(kind))
		return fmt.Sprintf(
			"googlesqlite_having_any_value(%s,%s,%s)",
			strings.Join(args, ","), havingExpr, kindLit,
		), nil
	}
	// When formatting inside a DifferentialPrivacyAggregateScan /
	// AnonymizedAggregateScan, append (epsilon, delta) to every
	// $differential_privacy_* or $anon_* call so the runtime UDF
	// can size the Laplace noise from the scan's privacy budget.
	if dp := dpOptionsFromContext(ctx); dp != nil &&
		(strings.HasPrefix(funcName, "googlesqlite_differential_privacy_") ||
			strings.HasPrefix(funcName, "googlesqlite_anon_")) {
		epsLit, _ := literalFromValue(value.FloatValue(dp.Epsilon))
		delLit, _ := literalFromValue(value.FloatValue(dp.Delta))
		args = append(args, epsLit, delLit)
	}
	funcMap := funcMapFromContext(ctx)
	if spec, exists := funcMap[funcName]; exists {
		return spec.CallSQL(ctx, n.node.ResolvedFunctionCallBase, args)
	}
	var opts []string
	for _, item := range m1(n.node.OrderByItemList()) {
		columnRef := m1(item.ColumnRef())
		colName := uniqueColumnName(ctx, m1(columnRef.Column()))
		if m1(item.IsDescending()) {
			opts = append(opts, fmt.Sprintf("googlesqlite_order_by(`%s`, false)", colName))
		} else {
			opts = append(opts, fmt.Sprintf("googlesqlite_order_by(`%s`, true)", colName))
		}
	}
	if m1(n.node.Distinct()) {
		opts = append(opts, "googlesqlite_distinct()")
	}
	if m1(n.node.Limit()) != nil {
		limitValue, err := newNode(m1(n.node.Limit())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		opts = append(opts, fmt.Sprintf("googlesqlite_limit(%s)", limitValue))
	}
	switch m1(n.node.NullHandlingModifier()) {
	case googlesql.ResolvedNonScalarFunctionCallBaseEnums_NullHandlingModifierIgnoreNulls:
		opts = append(opts, "googlesqlite_ignore_nulls()")
	case googlesql.ResolvedNonScalarFunctionCallBaseEnums_NullHandlingModifierRespectNulls:
	}
	args = append(args, opts...)
	return fmt.Sprintf(
		"%s(%s)",
		funcName,
		strings.Join(args, ","),
	), nil
}

// nativeWindowFuncMap maps GoogleSQL window function names that have a
// direct SQLite built-in equivalent. When a function is in this map,
// AnalyticFunctionCallNode emits native `<sqlite_name>(args) OVER
// (PARTITION BY ... ORDER BY ... FRAME)` syntax — avoiding the
// per-row full-scan emulation that the predecessor used because the
// underlying SQLite binding (mattn) could not register custom window
// functions.
var nativeWindowFuncMap = map[string]string{
	// Numbering functions — SQLite built-ins return identical results.
	"row_number":   "row_number",
	"rank":         "rank",
	"dense_rank":   "dense_rank",
	"percent_rank": "percent_rank",
	"cume_dist":    "cume_dist",
	"ntile":        "ntile",
	// Navigation functions — SQLite returns the underlying value
	// unchanged, so encoded values flow through correctly.
	"lag":         "lag",
	"lead":        "lead",
	"first_value": "first_value",
	"last_value":  "last_value",
	"nth_value":   "nth_value",
	// Cumulative aggregates against native SQLite types. Encoded
	// values would not survive these (SUM of base64 strings would
	// fail), but parity tests exercise them on INT64/FLOAT64 where
	// SQLite arithmetic suffices.
	"sum":   "sum",
	"avg":   "avg",
	"count": "count",
	"min":   "min",
	"max":   "max",
}

// customNativeWindowFuncMap lists googlesql window functions whose
// semantics don't have a SQLite built-in counterpart but whose
// googlesqlite_window_<name> implementation has been refactored to
// run incrementally (Step / Done over the active frame), making them
// safe to invoke through native OVER syntax instead of the
// per-output-row correlated-subquery emulation. The value is the
// custom SQLite function name registered through sqlitex.
var customNativeWindowFuncMap = map[string]string{
	"array_agg":  "googlesqlite_window_array_agg",
	"string_agg": "googlesqlite_window_string_agg",
	"countif":    "googlesqlite_window_countif",
	"count_star": "googlesqlite_window_count_star",
	"any_value":  "googlesqlite_window_any_value",
	// Statistical aggregators. Each runs over a sliding buffer of
	// floats with Step / Inverse / Done — see
	// internal/function_window_stats.go.
	"corr":        "googlesqlite_window_corr",
	"covar_pop":   "googlesqlite_window_covar_pop",
	"covar_samp":  "googlesqlite_window_covar_samp",
	"stddev":      "googlesqlite_window_stddev_samp",
	"stddev_pop":  "googlesqlite_window_stddev_pop",
	"stddev_samp": "googlesqlite_window_stddev_samp",
	"variance":    "googlesqlite_window_var_samp",
	"var_pop":     "googlesqlite_window_var_pop",
	"var_samp":    "googlesqlite_window_var_samp",
	// PERCENTILE_CONT / PERCENTILE_DISC have no SQLite built-in;
	// custom natives buffer the active window and sort at Done.
	"percentile_cont": "googlesqlite_window_percentile_cont",
	"percentile_disc": "googlesqlite_window_percentile_disc",
	// Boolean / bitwise aggregators in OVER context. Their custom
	// natives implement Step / Inverse / Done over the active frame
	// with per-row tristate tracking (NULL / FALSE / TRUE) for the
	// LOGICAL_* family and direct &/|/^ folding for the BIT_* family.
	"logical_or":  "googlesqlite_window_logical_or",
	"logical_and": "googlesqlite_window_logical_and",
	"bit_and":     "googlesqlite_window_bit_and",
	"bit_or":      "googlesqlite_window_bit_or",
	"bit_xor":     "googlesqlite_window_bit_xor",
	// ARRAY_CONCAT_AGG flattens per-row ARRAY<T> arguments into a
	// single ARRAY<T> over the active frame.
	"array_concat_agg": "googlesqlite_window_array_concat_agg",
}

// nativeWindowFuncForName returns the SQLite native name for a
// predecessor window function, or "" when no native equivalent
// applies.
func nativeWindowFuncForName(name string) string {
	if v, ok := nativeWindowFuncMap[name]; ok {
		return v
	}
	return ""
}

func (n *AnalyticFunctionCallNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	orderColumnNames := analyticOrderColumnNamesFromContext(ctx)
	orderColumns := orderColumnNames.values

	// Resolve the function name first so we know whether to emit
	// native OVER syntax or the predecessor's correlated-subquery
	// emulation. Some BigQuery-specific modifiers (DISTINCT,
	// IGNORE NULLS) and one tied-row case in CUME_DIST diverge from
	// SQLite's native semantics, so those still take the emulation
	// path.
	rawName := strings.ReplaceAll(m1(m1(n.node.Function()).FullName(false)), ".", "_")
	// googlesql analyzer prefixes some operator-like functions with
	// `$` (e.g. `$count_star` for COUNT(*)). Strip it for lookup.
	rawName = strings.TrimPrefix(rawName, "$")

	// SUM(DISTINCT x) / COUNT(DISTINCT x) / AVG(DISTINCT x) — SQLite
	// rejects DISTINCT in OVER, but our custom natives know how to
	// dedupe. Take precedence over the SQLite-native fast path.
	if m1(n.node.Distinct()) {
		if custom, ok := distinctAwareNativeWindowFuncs[rawName]; ok {
			return n.formatNative(ctx, custom, orderColumns, true)
		}
	}
	if native := nativeWindowFuncForName(rawName); native != "" && !n.requiresPredecessorEmulation() {
		return n.formatNative(ctx, native, orderColumns, false)
	}
	if custom, ok := customNativeWindowFuncMap[rawName]; ok && !n.requiresPredecessorEmulation() {
		return n.formatNative(ctx, custom, orderColumns, true)
	}

	funcName, args, err := getFuncNameAndArgs(ctx, n.node.ResolvedFunctionCallBase, true)
	if err != nil {
		return "", err
	}
	var opts []string
	if m1(n.node.Distinct()) {
		opts = append(opts, "googlesqlite_distinct()")
	}
	switch m1(n.node.NullHandlingModifier()) {
	case googlesql.ResolvedNonScalarFunctionCallBaseEnums_NullHandlingModifierRespectNulls:
		// do nothing
	default:
		opts = append(opts, "googlesqlite_ignore_nulls()")
	}
	args = append(args, opts...)
	for _, column := range analyticPartitionColumnNamesFromContext(ctx) {
		args = append(args, window.GetWindowPartitionOptionFuncSQL(column))
	}
	for _, col := range orderColumns {
		args = append(args, window.GetWindowOrderByOptionFuncSQL(col.column, col.isAsc))
	}
	windowFrame, _ := n.node.WindowFrame()
	if windowFrame != nil {
		args = append(args, window.GetWindowFrameUnitOptionFuncSQL(m1(windowFrame.FrameUnit())))
		startSQL, err := n.getWindowBoundaryOptionFuncSQL(ctx, m1(windowFrame.StartExpr()), true)
		if err != nil {
			return "", err
		}
		endSQL, err := n.getWindowBoundaryOptionFuncSQL(ctx, m1(windowFrame.EndExpr()), false)
		if err != nil {
			return "", err
		}
		args = append(args, startSQL, endSQL)
	}
	args = append(args, window.GetWindowRowIDOptionFuncSQL())
	input := analyticInputScanFromContext(ctx)
	funcMap := funcMapFromContext(ctx)
	// Anything that reaches here uses the predecessor's per-output-row
	// correlated-subquery emulation, which references `row_id`. Tell
	// the surrounding AnalyticScanNode to keep the row_id-wrap.
	markAnalyticEmulationUsed(ctx)
	if spec, exists := funcMap[funcName]; exists {
		return spec.CallSQL(ctx, n.node.ResolvedFunctionCallBase, args)
	}
	return fmt.Sprintf(
		"( SELECT %s(%s) %s )",
		funcName,
		strings.Join(args, ","),
		input,
	), nil
}

// distinctAwareNativeWindowFuncs maps `<name>` to the custom function
// registered through Conn.CreateWindowFunction that handles
// `<name>(DISTINCT x) OVER (...)`. Used when the formatter sees the
// DISTINCT modifier on functions whose plain native form is a SQLite
// built-in (which doesn't accept DISTINCT in OVER).
var distinctAwareNativeWindowFuncs = map[string]string{
	"sum":   "googlesqlite_window_sum_distinct",
	"count": "googlesqlite_window_count_distinct",
	"avg":   "googlesqlite_window_avg_distinct",
}

// requiresPredecessorEmulation reports whether this analytic call
// needs to keep using the per-row full-scan emulation rather than the
// new native OVER path.
func (n *AnalyticFunctionCallNode) requiresPredecessorEmulation() bool {
	rawName := strings.TrimPrefix(strings.ReplaceAll(m1(m1(n.node.Function()).FullName(false)), ".", "_"), "$")

	if m1(n.node.Distinct()) {
		// MIN(DISTINCT x) and MAX(DISTINCT x) are equivalent to
		// MIN(x) / MAX(x) — the DISTINCT marker is a no-op, so
		// SQLite native is fine.
		switch rawName {
		case "min", "max":
			return false
		}
		// SUM/COUNT/AVG with DISTINCT take the custom-native path
		// (see distinctAwareNativeWindowFuncs lookup below). For
		// every other DISTINCT-using analytic call we still need
		// emulation — including the existing customNativeWindowFuncs
		// (array_agg, string_agg, etc.) which already strip the
		// DISTINCT marker themselves but only when they hit the
		// custom path naturally.
		if _, ok := distinctAwareNativeWindowFuncs[rawName]; ok {
			return false
		}
		if _, ok := customNativeWindowFuncMap[rawName]; ok {
			return false
		}
		return true
	}
	switch m1(n.node.NullHandlingModifier()) {
	case googlesql.ResolvedNonScalarFunctionCallBaseEnums_NullHandlingModifierIgnoreNulls:
		// SQLite's SUM/AVG/COUNT/MIN/MAX/COUNT_STAR already skip
		// NULLs unconditionally, which matches the IGNORE NULLS
		// modifier exactly. Navigation functions (LAG/LEAD/
		// FIRST_VALUE/LAST_VALUE/NTH_VALUE) honour the modifier
		// natively.
		switch rawName {
		case "sum", "count", "count_star", "avg", "min", "max",
			"lag", "lead", "first_value", "last_value", "nth_value":
			return false
		}
		// Custom natives parse `googlesqlite_ignore_nulls()` out of
		// their argument list and honour the modifier themselves —
		// no need to bounce through the emulation path.
		if _, ok := customNativeWindowFuncMap[rawName]; ok {
			return false
		}
		return true
	}
	return false
}

// formatNative emits a SQLite-native window function call:
//
//	<sqliteName>(<value-args>) OVER (PARTITION BY ... ORDER BY ... <frame>)
//
// The OVER clause is built from the same partition / order / frame
// information the predecessor tucked into function arguments. SQLite
// drives Step/Value over rows in the frame, which is dramatically
// cheaper than the predecessor's per-output-row full-scan subquery.
//
// includeOpts forwards the DISTINCT / IGNORE NULLS markers as
// trailing arguments. SQLite built-ins reject those, so we only
// include them when calling our custom googlesqlite_window_<name>
// implementations that know how to parse them.
func (n *AnalyticFunctionCallNode) formatNative(ctx context.Context, sqliteName string, orderColumns []*analyticOrderBy, includeOpts bool) (string, error) {
	// Collect the user-supplied value arguments (without our window
	// option markers).
	var valueArgs []string
	for _, a := range m1(n.node.ResolvedFunctionCallBase.ArgumentList()) {
		arg, err := newNode(a).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		valueArgs = append(valueArgs, arg)
	}
	if includeOpts {
		if m1(n.node.Distinct()) {
			valueArgs = append(valueArgs, "googlesqlite_distinct()")
		}
		switch m1(n.node.NullHandlingModifier()) {
		case googlesql.ResolvedNonScalarFunctionCallBaseEnums_NullHandlingModifierRespectNulls:
			// no marker needed
		default:
			valueArgs = append(valueArgs, "googlesqlite_ignore_nulls()")
		}
	}
	call := fmt.Sprintf("%s(%s)", sqliteName, strings.Join(valueArgs, ","))

	var clauses []string
	if cols := analyticPartitionColumnNamesFromContext(ctx); len(cols) > 0 {
		clauses = append(clauses, "PARTITION BY "+strings.Join(cols, ","))
	}
	if len(orderColumns) > 0 {
		var ob []string
		for _, col := range orderColumns {
			switch col.nullOrder {
			case nullOrderFirst:
				ob = append(ob, fmt.Sprintf("(%s IS NOT NULL)", col.column))
			case nullOrderLast:
				ob = append(ob, fmt.Sprintf("(%s IS NULL)", col.column))
			}
			suffix := " COLLATE googlesqlite_collate"
			if !col.isAsc {
				suffix += " DESC"
			}
			ob = append(ob, col.column+suffix)
		}
		clauses = append(clauses, "ORDER BY "+strings.Join(ob, ","))
	}
	if frameSQL, err := n.formatNativeFrame(ctx); err != nil {
		return "", err
	} else if frameSQL != "" {
		clauses = append(clauses, frameSQL)
	}

	if len(clauses) == 0 {
		return call + " OVER ()", nil
	}
	return call + " OVER (" + strings.Join(clauses, " ") + ")", nil
}

// formatNativeFrame turns the resolved-tree frame into a SQLite frame
// clause: `ROWS|RANGE BETWEEN <start> AND <end>`. Returns ""
// when no frame is specified (SQLite picks the appropriate default
// for the function — RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT
// ROW for ordered aggregates).
func (n *AnalyticFunctionCallNode) formatNativeFrame(ctx context.Context) (string, error) {
	windowFrame, _ := n.node.WindowFrame()
	if windowFrame == nil {
		return "", nil
	}
	unitText := ""
	switch m1(windowFrame.FrameUnit()) {
	case googlesql.ResolvedWindowFrameEnums_FrameUnitRows:
		unitText = "ROWS"
	case googlesql.ResolvedWindowFrameEnums_FrameUnitRange:
		unitText = "RANGE"
	default:
		return "", nil
	}
	startSQL, err := nativeBoundaryClause(ctx, m1(windowFrame.StartExpr()), true)
	if err != nil {
		return "", err
	}
	endSQL, err := nativeBoundaryClause(ctx, m1(windowFrame.EndExpr()), false)
	if err != nil {
		return "", err
	}
	return unitText + " BETWEEN " + startSQL + " AND " + endSQL, nil
}

func nativeBoundaryClause(ctx context.Context, expr *googlesql.ResolvedWindowFrameExpr, isStart bool) (string, error) {
	typ := m1(expr.BoundaryType())
	switch typ {
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedPreceding:
		return "UNBOUNDED PRECEDING", nil
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeCurrentRow:
		return "CURRENT ROW", nil
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedFollowing:
		return "UNBOUNDED FOLLOWING", nil
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetPreceding:
		offset, err := newNode(m1(expr.Expression())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		return offset + " PRECEDING", nil
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetFollowing:
		offset, err := newNode(m1(expr.Expression())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		return offset + " FOLLOWING", nil
	}
	if isStart {
		return "UNBOUNDED PRECEDING", nil
	}
	return "UNBOUNDED FOLLOWING", nil
}

func (n *AnalyticFunctionCallNode) getWindowBoundaryOptionFuncSQL(ctx context.Context, expr *googlesql.ResolvedWindowFrameExpr, isStart bool) (string, error) {
	typ := m1(expr.BoundaryType())
	switch typ {
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedPreceding, googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeCurrentRow, googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedFollowing:
		if isStart {
			return window.GetWindowBoundaryStartOptionFuncSQL(typ, ""), nil
		}
		return window.GetWindowBoundaryEndOptionFuncSQL(typ, ""), nil
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetPreceding, googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetFollowing:
		literal, err := newNode(m1(expr.Expression())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		if isStart {
			return window.GetWindowBoundaryStartOptionFuncSQL(typ, literal), nil
		}
		return window.GetWindowBoundaryEndOptionFuncSQL(typ, literal), nil
	}
	return "", fmt.Errorf("unexpected boundary type %d", typ)
}

func (n *CastNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	fromType := newType(m1(m1(n.node.Expr()).Type()))
	jsonEncodedFromType, err := json.Marshal(fromType)
	if err != nil {
		return "", err
	}
	toType := newType(m1(n.node.Type()))
	jsonEncodedToType, err := json.Marshal(toType)
	if err != nil {
		return "", err
	}
	encodedFromType, err := encodeGoValue(m1(tf().MakeSimpleType(googlesql.TypeKindTypeString)), string(jsonEncodedFromType))
	if err != nil {
		return "", err
	}
	encodedToType, err := encodeGoValue(m1(tf().MakeSimpleType(googlesql.TypeKindTypeString)), string(jsonEncodedToType))
	if err != nil {
		return "", err
	}
	expr, err := newNode(m1(n.node.Expr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"googlesqlite_cast(%s, '%s', '%s', %t)",
		expr, encodedFromType, encodedToType, m1(n.node.ReturnNullOnError()),
	), nil
}

func (n *MakeStructNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	typ := m1(m1(n.node.Type()).AsStruct())
	typFields := m1(typ.Fields())
	fields, _ := n.node.FieldList()
	args := make([]string, 0, len(typFields)*2)
	for i, sf := range typFields {
		fieldName := sf.Name
		key, err := literalFromValue(value.StringValue(fieldName))
		if err != nil {
			return "", err
		}
		args = append(args, key)
		field, err := newNode(fields[i]).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		args = append(args, field)
	}
	return fmt.Sprintf("googlesqlite_make_struct(%s)", strings.Join(args, ",")), nil
}

// MakeProtoNode formats `new MessageName(expr1 AS field1, expr2 AS field2, ...)`
// as a runtime call that produces proto wire bytes. The resolved tree
// hands us the field-tag / value-kind pairs via each MakeProtoField
// child; we forward them to `googlesqlite_make_proto`, which
// concatenates encodeSingleField(tag, value, kind) per pair.
func (n *MakeProtoNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	fields := m1(n.node.FieldList())
	args := make([]string, 0, len(fields)*3)
	for _, f := range fields {
		fd := m1(f.FieldDescriptor())
		if fd == nil {
			return "", fmt.Errorf("MakeProto: field descriptor unavailable")
		}
		tag, _ := fd.Number()
		expr := m1(f.Expr())
		exprSQL, err := newNode(expr).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		exprType, _ := expr.Type()
		var kindStr string
		if exprType != nil {
			k, _ := exprType.Kind()
			kindStr = protoFieldKindString(k)
		} else {
			kindStr = "bytes"
		}
		args = append(args,
			fmt.Sprintf("%d", tag),
			protoFormatLiteralString(kindStr),
			exprSQL,
		)
	}
	if len(args) == 0 {
		// Zero-field message: emit an empty BYTES literal cast through
		// the runtime to keep the type stable.
		return "googlesqlite_make_proto()", nil
	}
	return fmt.Sprintf("googlesqlite_make_proto(%s)", strings.Join(args, ",")), nil
}

// MakeProtoFieldNode never reaches FormatSQL directly — its parent
// MakeProtoNode walks the field list and lowers each one inline.
// Returning an empty string here is a safe no-op for the general
// dispatcher.
func (n *GetStructFieldNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	expr, err := newNode(m1(n.node.Expr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	idx, _ := n.node.FieldIdx()
	return fmt.Sprintf("googlesqlite_get_struct_field(%s, %d)", expr, idx), nil
}

func (n *GetProtoFieldNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	expr, err := newNode(m1(n.node.Expr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	fd, err := n.node.FieldDescriptor()
	if err != nil || fd == nil {
		return "", fmt.Errorf("GetProtoField: field descriptor unavailable: %w", err)
	}
	fieldNumber, _ := fd.Number()
	resultType, err := n.node.Type()
	if err != nil || resultType == nil {
		return "", fmt.Errorf("GetProtoField: result type unavailable: %w", err)
	}
	kind, _ := resultType.Kind()
	// Repeated proto fields surface as ARRAY<…> after the upstream
	// analyzer rewrite (map<K,V> becomes ARRAY<STRUCT<key, value>>).
	// The lowering picks up googlesqlite_decode_array on the result,
	// so the runtime side must return a real ArrayValue carrying one
	// element per wire occurrence rather than the last occurrence's
	// raw bytes. We route the repeated case to a dedicated runtime
	// that walks every match instead of overwriting on each hit.
	isRepeated, _ := fd.IsRepeated()
	if isRepeated && kind == googlesql.TypeKindTypeArray {
		var elemKindStr string
		if arr, ok := resultType.(*googlesql.ArrayType); ok && arr != nil {
			elem, err := arr.ElementType()
			if err == nil && elem != nil {
				ek, _ := elem.Kind()
				elemKindStr = protoFieldKindString(ek)
			}
		}
		if elemKindStr == "" {
			// Fallback: STRUCT-shaped map entries decode element-wise
			// through subsequent GetProtoField calls, so "message"
			// keeps each entry's payload intact.
			elemKindStr = "message"
		}
		return fmt.Sprintf("googlesqlite_get_proto_field_repeated(%s, %d, '%s')", expr, fieldNumber, elemKindStr), nil
	}
	kindStr := protoFieldKindString(kind)
	defaultLit := "''"
	if returnDefault, _ := n.node.ReturnDefaultValueWhenUnset(); returnDefault {
		defaultLit = protoFieldDefaultLiteral(fd, kindStr)
	}
	return fmt.Sprintf("googlesqlite_get_proto_field(%s, %d, '%s', %s)", expr, fieldNumber, kindStr, defaultLit), nil
}

// protoFormatLiteralString turns a Go string into a single-quoted
// SQL literal with embedded single quotes doubled. Used when the
// formatter has to splice an opaque metadata string (proto field
// path list, enum full name, …) into the lowered SQL.
func protoFormatLiteralString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// mapFieldParentAndTag inspects a ResolvedExpr that the analyzer
// produced as the map_field argument of PROTO_MAP_* (e.g.
// `m.purchased`). When it is a ResolvedGetProtoField we recover the
// parent message expression and the field number; the runtime can
// then walk the parent's wire bytes for every occurrence of that
// outer tag, which a non-accumulating GetProtoField extraction would
// have dropped.
//
// Returns the SQL for the parent message expression plus the field
// number. Errors when the argument isn't a ResolvedGetProtoField
// (a future caller might pass the map's wire bytes directly, but the
// current proto-runtime surface only emits the field-access form).
func mapFieldParentAndTag(ctx context.Context, arg googlesql.ResolvedExprNode) (string, int32, error) {
	if arg == nil {
		return "", 0, fmt.Errorf("map field argument is nil")
	}
	kind, err := arg.NodeKind()
	if err != nil || kind != googlesql.ResolvedNodeKindResolvedGetProtoField {
		return "", 0, fmt.Errorf("map field is not a proto field access (kind=%v)", kind)
	}
	gpf, ok := arg.(*googlesql.ResolvedGetProtoField)
	if !ok || gpf == nil {
		return "", 0, fmt.Errorf("map field has unexpected node type %T", arg)
	}
	parent, err := gpf.Expr()
	if err != nil || parent == nil {
		return "", 0, fmt.Errorf("map field parent unavailable: %w", err)
	}
	parentSQL, err := newNode(parent).FormatSQL(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("format map field parent: %w", err)
	}
	fd, err := gpf.FieldDescriptor()
	if err != nil || fd == nil {
		return "", 0, fmt.Errorf("map field descriptor unavailable: %w", err)
	}
	num, err := fd.Number()
	if err != nil {
		return "", 0, fmt.Errorf("map field number: %w", err)
	}
	return parentSQL, num, nil
}

// protoFieldKindString maps a TypeKind to the lowercase token the
// internal/functions/proto runtime expects.
func protoFieldKindString(k googlesql.TypeKind) string {
	switch k {
	case googlesql.TypeKindTypeBool:
		return "bool"
	case googlesql.TypeKindTypeInt32:
		return "int32"
	case googlesql.TypeKindTypeInt64:
		return "int64"
	case googlesql.TypeKindTypeUint32:
		return "uint32"
	case googlesql.TypeKindTypeUint64:
		return "uint64"
	case googlesql.TypeKindTypeFloat:
		return "float"
	case googlesql.TypeKindTypeDouble:
		return "double"
	case googlesql.TypeKindTypeString:
		return "string"
	case googlesql.TypeKindTypeBytes:
		return "bytes"
	case googlesql.TypeKindTypeEnum:
		return "enum"
	case googlesql.TypeKindTypeProto:
		return "message"
	case googlesql.TypeKindTypeDate:
		// FROM_PROTO(google.type.Date) returns a SQL DATE; the
		// runtime needs the well-known proto layout decoder.
		return "date"
	case googlesql.TypeKindTypeTimestamp:
		// FROM_PROTO(google.protobuf.Timestamp) returns SQL TIMESTAMP.
		return "timestamp"
	}
	return "bytes"
}

// protoFieldDefaultLiteral returns a SQL literal that the runtime
// proto UDF can decode as the field's proto-defined default. The
// literal is a base64-encoded STRING carrying the kind-specific
// representation (decimal for numerics, "true"/"false" for bool,
// raw string for string fields). Empty string when no default.
func protoFieldDefaultLiteral(fd *googlesql.FieldDescriptor, kindStr string) string {
	hasDefault, _ := fd.HasDefaultValue()
	if !hasDefault {
		return "''"
	}
	encode := func(s string) string {
		return "'" + base64.StdEncoding.EncodeToString([]byte(s)) + "'"
	}
	switch kindStr {
	case "bool":
		v, _ := fd.DefaultValueBool()
		if v {
			return encode("true")
		}
		return encode("false")
	case "int32":
		v, _ := fd.DefaultValueInt32()
		return encode(fmt.Sprintf("%d", v))
	case "int64":
		v, _ := fd.DefaultValueInt64()
		return encode(fmt.Sprintf("%d", v))
	case "uint32":
		v, _ := fd.DefaultValueUint32()
		return encode(fmt.Sprintf("%d", v))
	case "uint64":
		v, _ := fd.DefaultValueUint64()
		return encode(fmt.Sprintf("%d", v))
	case "float":
		v, _ := fd.DefaultValueFloat()
		return encode(fmt.Sprintf("%g", v))
	case "double":
		v, _ := fd.DefaultValueDouble()
		return encode(fmt.Sprintf("%g", v))
	case "string":
		// FieldDescriptor doesn't expose DefaultValueString directly
		// in this binding; fall back to DebugString parsing or empty.
		return "''"
	}
	return "''"
}

func (n *GetJsonFieldNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	expr, err := newNode(m1(n.node.Expr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	name, _ := n.node.FieldName()
	encodedName, err := encodeGoValue(m1(tf().MakeSimpleType(googlesql.TypeKindTypeString)), name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("googlesqlite_get_json_field(%s, '%s')", expr, encodedName), nil
}

// ReplaceFieldNode lowers a ResolvedReplaceField (the analyzer node
// REPLACE_FIELDS resolves into) to a chain of
// googlesqlite_replace_fields runtime calls, one per (value, path)
// pair. Each path is a dotted field-number string; the new value is
// the lowered SQL of the replacement expression. We pull the
// per-item leaf TypeKind from the path's last FieldDescriptor so the
// runtime knows how to encode the replacement value.
func (n *ReplaceFieldNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	exprSQL, err := newNode(m1(n.node.Expr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	items, err := n.node.ReplaceFieldItemList()
	if err != nil {
		return "", err
	}
	inner := exprSQL
	for _, it := range items {
		if it == nil {
			continue
		}
		valSQL, err := newNode(m1(it.Expr())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		fdPath, _ := it.ProtoFieldPath()
		nums := make([]string, 0, len(fdPath))
		for _, fd := range fdPath {
			num, _ := fd.Number()
			nums = append(nums, strconv.Itoa(int(num)))
		}
		// Read the replacement value's type kind via the value expr;
		// that's what the runtime needs to encode the wire payload.
		leafKindStr := "bytes"
		if vt, _ := m1(it.Expr()).Type(); vt != nil {
			if k, err := vt.Kind(); err == nil {
				leafKindStr = protoFieldKindString(k)
			}
		}
		path := strings.Join(nums, ".")
		inner = fmt.Sprintf("googlesqlite_replace_fields(%s, %s, '%s', %s)",
			inner, protoFormatLiteralString(path), leafKindStr, valSQL)
	}
	return inner, nil
}

// isStructTypedExpr reports whether the resolved expression's
// concrete type is STRUCT. The formatter uses this to decide whether
// to attach `COLLATE googlesqlite_collate` to an IN comparison so
// the runtime decoder makes the equality positional instead of
// byte-for-byte (anonymous-struct-field bug, runtime-side fix).
func isStructTypedExpr(expr googlesql.ResolvedExprNode) bool {
	if expr == nil {
		return false
	}
	typ, err := expr.Type()
	if err != nil || typ == nil {
		return false
	}
	kind, _ := typ.Kind()
	return kind == googlesql.TypeKindTypeStruct
}

func (n *SubqueryExprNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	columnNames := &arraySubqueryColumnNames{}
	ctx = withArraySubqueryColumnName(ctx, columnNames)
	sql, err := newNode(m1(n.node.Subquery())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	switch m1(n.node.SubqueryType()) {
	case googlesql.ResolvedSubqueryExprEnums_SubqueryTypeScalar:
		if inSafeEvalMode(ctx) {
			// GoogleSQL raises an error when a scalar subquery returns
			// more than one row; SQLite would silently take the first.
			// In IFERROR / ISERROR / NULLIFERROR contexts we want the
			// error to fold to NULL, so guard the value with an inline
			// COUNT(*) probe. Re-evaluates the inner sub-select twice
			// — acceptable in error-handling expressions.
			return fmt.Sprintf(
				"(SELECT CASE WHEN (SELECT COUNT(*) FROM (%s)) = 1 THEN (%s) ELSE NULL END)",
				sql, sql,
			), nil
		}
	case googlesql.ResolvedSubqueryExprEnums_SubqueryTypeArray:
		subCols := m1(m1(n.node.Subquery()).MutableColumnList())
		if len(subCols) == 0 {
			return "", fmt.Errorf("failed to find computed column names for array subquery")
		}
		colName := uniqueColumnName(ctx, subCols[0])
		return fmt.Sprintf("(SELECT googlesqlite_array(`%s`) FROM (%s))", colName, sql), nil
	case googlesql.ResolvedSubqueryExprEnums_SubqueryTypeExists:
		return fmt.Sprintf("EXISTS (%s)", sql), nil
	case googlesql.ResolvedSubqueryExprEnums_SubqueryTypeIn:
		expr, err := newNode(m1(n.node.InExpr())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		// SQLite's IN compares values byte-for-byte under the
		// BINARY collation. For STRUCT values the envelope encodes
		// the field names in JSON, so two structurally-equal but
		// differently-named structs (anonymous `(1,'a')` vs named
		// `STRUCT(1 AS x, 'a' AS y)`) miss each other. Force the
		// comparison through googlesqlite_collate, which decodes
		// each side into a StructValue and compares positionally
		// — matching BigQuery semantics. The collation attached to
		// the LHS is inherited by the entire IN comparison per
		// SQLite rules.
		if isStructTypedExpr(m1(n.node.InExpr())) {
			return fmt.Sprintf("(%s) COLLATE googlesqlite_collate IN (%s)", expr, sql), nil
		}
		return fmt.Sprintf("%s IN (%s)", expr, sql), nil
	case googlesql.ResolvedSubqueryExprEnums_SubqueryTypeLikeAny:
	case googlesql.ResolvedSubqueryExprEnums_SubqueryTypeLikeAll:
	}
	return fmt.Sprintf("(%s)", sql), nil
}

func (n *SingleRowScanNode) FormatSQL(ctx context.Context) (string, error) {
	return "", nil
}

func (n *TableScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	var columns []string
	for _, col := range m1(n.node.ColumnList()) {
		columns = append(
			columns,
			fmt.Sprintf("`%s` AS `%s`", m1(col.Name()), uniqueColumnName(ctx, col)),
		)
	}

	// If the underlying TableNode is actually a wildcard — recorded in
	// wildcardTableRegistry under the SimpleTable's wasm handle ptr —
	// rewrite the scan into the UNION-ALL pattern. Otherwise fall through
	// to a regular table reference.
	table := m1(n.node.Table())
	if wc := lookupWildcardTable(table); wc != nil {
		query, err := wc.FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(SELECT %s FROM (%s))", strings.Join(columns, ","), query), nil
	}
	// INFORMATION_SCHEMA scans route through the corresponding
	// vtab module. The dataset filter is pushed into the WHERE so
	// xBestIndex picks the constrained scan.
	if a := analyzerFromContext(ctx); a != nil && a.catalog != nil {
		if meta := a.catalog.infoSchemaTableMetaFor(table); meta != nil {
			return fmt.Sprintf(
				"(SELECT %s FROM %s WHERE __schema = '%s')",
				strings.Join(columns, ","),
				meta.view.Module,
				strings.ReplaceAll(meta.dataset, "'", "''"),
			), nil
		}
	}
	tableName, err := getTableName(ctx, n.node)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(SELECT %s FROM `%s`)", strings.Join(columns, ","), tableName), nil
}

func (n *JoinScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	left, err := newNode(m1(n.node.LeftScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	right, err := newNode(m1(n.node.RightScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	if getInputPattern(left) == InputNeedsWrap {
		left = fmt.Sprintf("(%s)", left)
	}
	if getInputPattern(right) == InputNeedsWrap {
		right = fmt.Sprintf("(%s)", right)
	}
	if m1(n.node.JoinExpr()) == nil {
		return fmt.Sprintf("%s CROSS JOIN %s", left, right), nil
	}
	joinExpr, err := newNode(m1(n.node.JoinExpr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	switch m1(n.node.JoinType()) {
	case googlesql.ResolvedJoinScanEnums_JoinTypeInner:
		return fmt.Sprintf("%s JOIN %s ON %s", left, right, joinExpr), nil
	case googlesql.ResolvedJoinScanEnums_JoinTypeLeft:
		return fmt.Sprintf("%s LEFT JOIN %s ON %s", left, right, joinExpr), nil
	case googlesql.ResolvedJoinScanEnums_JoinTypeRight:
		return fmt.Sprintf("%s RIGHT JOIN %s ON %s", left, right, joinExpr), nil
	case googlesql.ResolvedJoinScanEnums_JoinTypeFull:
		return fmt.Sprintf("%s FULL OUTER JOIN %s ON %s", left, right, joinExpr), nil
	}
	return "", fmt.Errorf("unexpected join type %d", m1(n.node.JoinType()))
}

func (n *ArrayScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	arrayExpr, err := newNode(m1(n.node.ArrayExpr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	colName := uniqueColumnName(ctx, m1(n.node.ElementColumn()))
	columns := []string{fmt.Sprintf("json_each.value AS `%s`", colName)}

	if offsetColumn, _ := n.node.ArrayOffsetColumn(); offsetColumn != nil {
		offsetColName := uniqueColumnName(ctx, m1(offsetColumn.Column()))
		columns = append(columns, fmt.Sprintf("json_each.key AS `%s`", offsetColName))
	}
	if m1(n.node.InputScan()) != nil {
		input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		formattedInput, err := formatInput(input)
		if err != nil {
			return "", err
		}

		array := fmt.Sprintf("json_each(googlesqlite_decode_array(%s))", arrayExpr)
		var arrayJoinExpr string
		if m1(n.node.JoinExpr()) != nil {
			arrayJoinExpr, err = newNode(m1(n.node.JoinExpr())).FormatSQL(ctx)
			if err != nil {
				return "", err
			}
			// RIGHT JOINs on array expressions are not supported by BigQuery
			var joinMode string
			if m1(n.node.IsOuter()) {
				joinMode = "LEFT OUTER JOIN"
			} else {
				joinMode = "INNER JOIN"
			}
			arrayJoinExpr = fmt.Sprintf("%s %s ON %s",
				joinMode,
				array,
				arrayJoinExpr,
			)
		} else {
			// If there is no join expression, use a CROSS JOIN
			arrayJoinExpr = fmt.Sprintf(", %s", array)
		}

		return fmt.Sprintf(
			"SELECT *, %s %s %s",
			strings.Join(columns, ","),
			formattedInput,
			arrayJoinExpr,
		), nil
	}
	return fmt.Sprintf(
		"SELECT %s FROM json_each(googlesqlite_decode_array(%s))",
		strings.Join(columns, ","),
		arrayExpr,
	), nil
}

var tokensAfterFromClause = [...]string{"WHERE", "GROUP BY", "HAVING", "QUALIFY", "WINDOW", "ORDER BY", "COLLATE"}
var removeExpressions = regexp.MustCompile(`\(.+?\)`)

func (n *FilterScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	filter, err := newNode(m1(n.node.FilterExpr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	currentQuery := removeExpressions.ReplaceAllString(input, "")

	// Qualify the statement if the input is not wrapped in parens
	queryWrappedInParens := currentQuery == ""
	containsTokens := false
	// and the input contains a token that would result in a syntax error
	for _, token := range tokensAfterFromClause {
		containsTokens = containsTokens || strings.Contains(currentQuery, token)
	}

	if !queryWrappedInParens && containsTokens {
		return fmt.Sprintf("( %s ) WHERE %s", input, filter), nil
	}
	return fmt.Sprintf("%s WHERE %s", input, filter), nil
}

// FormatSQL passes the AssertScan through to its InputScan. See the
// AssertScanNode doc comment in node.go for the rationale — the emulator
// skips the runtime assertion check, which matches every upstream
// Example because those Examples are written with conditions that the
// fixture data satisfies.
func (n *AssertScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	return newNode(m1(n.node.InputScan())).FormatSQL(ctx)
}

// FormatSQL on BarrierScanNode is a pass-through to InputScan. The
// barrier acts as an optimiser hint upstream; for the emulator it has
// no observable runtime semantics.
func (n *BarrierScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	return newNode(m1(n.node.InputScan())).FormatSQL(ctx)
}

// isGroupingCallColumn reports whether `name` is an output column for a
// GROUPING() call in the same AggregateScan.
func isGroupingCallColumn(name string, targets map[string]string) bool {
	_, ok := targets[name]
	return ok
}

// nullSetContains reports whether `name` is a group-by column that must
// be NULLed out for the current grouping set.
func nullSetContains(set map[string]struct{}, name string) bool {
	_, ok := set[name]
	return ok
}

func (n *AggregateScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	for _, agg := range m1(n.node.AggregateList()) {
		// assign sql to column ref map
		if _, err := newNode(agg).FormatSQL(ctx); err != nil {
			return "", err
		}
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	groupByColumns := []string{}
	groupByColumnMap := map[string]struct{}{}
	for _, col := range m1(n.node.GroupByList()) {
		if _, err := newNode(col).FormatSQL(ctx); err != nil {
			return "", err
		}
		colName := uniqueColumnName(ctx, m1(col.Column()))
		groupByColumns = append(groupByColumns, fmt.Sprintf("`%s`", colName))
		groupByColumnMap[colName] = struct{}{}
	}
	columns := []string{}
	columnMap := columnRefMap(ctx)
	columnNames := []string{}
	for _, col := range m1(n.node.ColumnList()) {
		colName := uniqueColumnName(ctx, col)
		columnNames = append(columnNames, colName)
		if ref, exists := columnMap[colName]; exists {
			columns = append(columns, ref)
			delete(columnMap, colName)
		} else {
			columns = append(columns, fmt.Sprintf("`%s`", colName))
		}
	}
	if gsList := m1(n.node.GroupingSetList()); len(gsList) != 0 {
		// GROUPING(col) maps to a per-grouping-set 0/1 literal: 0 if
		// `col` IS in the current grouping set (i.e. rows are grouped
		// by it), 1 if `col` is being aggregated away. Build the
		// mapping from output-column unique name → group-by-column
		// unique name so we can substitute the literal into each
		// branch's projection.
		groupingCallTargets := map[string]string{}
		for _, call := range m1(n.node.GroupingCallList()) {
			out, err := call.OutputColumn()
			if err != nil || out == nil {
				continue
			}
			ref, err := call.GroupByColumn()
			if err != nil || ref == nil {
				continue
			}
			groupingCallTargets[uniqueColumnName(ctx, out)] =
				uniqueColumnName(ctx, m1(ref.Column()))
		}
		columnPatterns := [][]string{}
		groupByColumnPatterns := [][]string{}
		for _, set := range gsList {
			concreteSet, ok := set.(*googlesql.ResolvedGroupingSet)
			if !ok {
				continue
			}
			groupBySetColumns := []string{}
			groupBySetColumnMap := map[string]struct{}{}
			for _, col := range m1(concreteSet.GroupByColumnList()) {
				colName := uniqueColumnName(ctx, m1(col.Column()))
				groupBySetColumns = append(groupBySetColumns, fmt.Sprintf("`%s`", colName))
				groupBySetColumnMap[colName] = struct{}{}
			}
			nullColumnNameMap := map[string]struct{}{}
			for colName := range groupByColumnMap {
				if _, exists := groupBySetColumnMap[colName]; !exists {
					nullColumnNameMap[colName] = struct{}{}
				}
			}
			groupBySetColumnPattern := []string{}
			for idx, col := range columnNames {
				switch {
				case isGroupingCallColumn(col, groupingCallTargets):
					target := groupingCallTargets[col]
					grouped := 1
					if _, exists := groupBySetColumnMap[target]; exists {
						grouped = 0
					}
					groupBySetColumnPattern = append(groupBySetColumnPattern, fmt.Sprintf("%d AS `%s`", grouped, col))
				case nullSetContains(nullColumnNameMap, col):
					groupBySetColumnPattern = append(groupBySetColumnPattern, fmt.Sprintf("NULL AS `%s`", col))
				default:
					groupBySetColumnPattern = append(groupBySetColumnPattern, columns[idx])
				}
			}
			columnPatterns = append(columnPatterns, groupBySetColumnPattern)
			annotatedGroupBySetColumns := make([]string, 0, len(groupBySetColumns))
			for _, column := range groupBySetColumns {
				annotatedGroupBySetColumns = append(
					annotatedGroupBySetColumns,
					fmt.Sprintf("googlesqlite_group_by(%s)", column),
				)
			}
			groupByColumnPatterns = append(groupByColumnPatterns, annotatedGroupBySetColumns)
		}
		stmts := []string{}
		for i := 0; i < len(columnPatterns); i++ {
			var groupBy string
			if len(groupByColumnPatterns[i]) != 0 {
				groupBy = fmt.Sprintf("GROUP BY %s", strings.Join(groupByColumnPatterns[i], ","))
			}
			formattedColumns := strings.Join(columnPatterns[i], ",")
			switch getInputPattern(input) {
			case InputKeep:
				stmts = append(stmts, fmt.Sprintf("SELECT %s %s %s", formattedColumns, input, groupBy))
			case InputNeedsWrap:
				stmts = append(stmts, fmt.Sprintf("SELECT %s FROM (%s) %s", formattedColumns, input, groupBy))
			case InputNeedsFrom:
				stmts = append(stmts, fmt.Sprintf("SELECT %s FROM %s %s", formattedColumns, input, groupBy))
			}
		}
		groupByWithCollates := make([]string, 0, len(groupByColumns))
		for _, groupByColumn := range groupByColumns {
			groupByWithCollates = append(
				groupByWithCollates,
				fmt.Sprintf("%s COLLATE googlesqlite_collate", groupByColumn),
			)
		}
		return fmt.Sprintf(
			"%s ORDER BY %s",
			strings.Join(stmts, " UNION ALL "),
			strings.Join(groupByWithCollates, ","),
		), nil
	}
	var groupBy string
	if len(groupByColumns) > 0 {
		annotatedGroupByColumns := make([]string, 0, len(groupByColumns))
		for _, groupByColumn := range groupByColumns {
			annotatedGroupByColumns = append(
				annotatedGroupByColumns,
				fmt.Sprintf("googlesqlite_group_by(%s)", groupByColumn),
			)
		}
		groupBy = fmt.Sprintf("GROUP BY %s", strings.Join(annotatedGroupByColumns, ","))
	}
	formattedColumns := strings.Join(columns, ",")
	switch getInputPattern(input) {
	case InputKeep:
		return fmt.Sprintf("SELECT %s %s %s", formattedColumns, input, groupBy), nil
	case InputNeedsWrap:
		return fmt.Sprintf("SELECT %s FROM (%s) %s", formattedColumns, input, groupBy), nil
	case InputNeedsFrom:
		return fmt.Sprintf("SELECT %s FROM %s %s", formattedColumns, input, groupBy), nil
	}
	return "", fmt.Errorf("unexpected input pattern: %s", input)
}

// FormatSQL lowers a ResolvedAnonymizedAggregateScan (the
// resolver's `SELECT WITH ANONYMIZATION` form) to SQL using the
// same shape as DifferentialPrivacyAggregateScan: epsilon / delta
// are pulled from OptionList and threaded through the context so
// inner $anon_* / $differential_privacy_* aggregate calls can pick
// them up; the aggregate list and group-by list are formatted as
// for a regular aggregate scan.
func (n *AnonymizedAggregateScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	dp := &dpOptions{Epsilon: 1, Delta: 1e-5}
	for _, opt := range m1(n.node.AnonymizationOptionList()) {
		name, _ := opt.Name()
		valExpr, _ := opt.Value()
		valSQL, err := newNode(valExpr).FormatSQL(ctx)
		if err != nil {
			continue
		}
		f, perr := strconv.ParseFloat(strings.TrimSpace(valSQL), 64)
		if perr != nil {
			continue
		}
		switch strings.ToLower(name) {
		case "epsilon":
			dp.Epsilon = f
		case "delta":
			dp.Delta = f
		}
	}
	ctx = withDPOptions(ctx, dp)
	for _, agg := range m1(n.node.AggregateList()) {
		if _, err := newNode(agg).FormatSQL(ctx); err != nil {
			return "", err
		}
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	groupByColumns := []string{}
	for _, col := range m1(n.node.GroupByList()) {
		if _, err := newNode(col).FormatSQL(ctx); err != nil {
			return "", err
		}
		colName := uniqueColumnName(ctx, m1(col.Column()))
		groupByColumns = append(groupByColumns, fmt.Sprintf("`%s`", colName))
	}
	columns := []string{}
	columnMap := columnRefMap(ctx)
	for _, col := range m1(n.node.ColumnList()) {
		colName := uniqueColumnName(ctx, col)
		if ref, exists := columnMap[colName]; exists {
			columns = append(columns, ref)
			delete(columnMap, colName)
		} else {
			columns = append(columns, fmt.Sprintf("`%s`", colName))
		}
	}
	formattedColumns := strings.Join(columns, ",")
	var groupBy string
	if len(groupByColumns) > 0 {
		groupBy = "GROUP BY " + strings.Join(groupByColumns, ",")
	}
	switch getInputPattern(input) {
	case InputKeep:
		return fmt.Sprintf("SELECT %s %s %s", formattedColumns, input, groupBy), nil
	case InputNeedsWrap:
		return fmt.Sprintf("SELECT %s FROM (%s) %s", formattedColumns, input, groupBy), nil
	case InputNeedsFrom:
		return fmt.Sprintf("SELECT %s FROM %s %s", formattedColumns, input, groupBy), nil
	}
	return "", fmt.Errorf("anon scan: unexpected input pattern: %s", input)
}

// dpOptions captures the (epsilon, delta) pair parsed from a
// DifferentialPrivacyAggregateScan's option_list. AggregateFunctionCall
// reads it at format time so it can append the privacy-budget args to
// $differential_privacy_* calls.
type dpOptions struct {
	Epsilon float64
	Delta   float64
}

type dpCtxKey struct{}

func dpOptionsFromContext(ctx context.Context) *dpOptions {
	if v, ok := ctx.Value(dpCtxKey{}).(*dpOptions); ok {
		return v
	}
	return nil
}

func withDPOptions(ctx context.Context, dp *dpOptions) context.Context {
	return context.WithValue(ctx, dpCtxKey{}, dp)
}

// FormatSQL lowers a ResolvedDifferentialPrivacyAggregateScan to
// SQL by formatting it like a regular AggregateScan, after
// extracting the epsilon/delta options so the inner
// $differential_privacy_* aggregate calls can append them as
// extra runtime arguments.
func (n *DifferentialPrivacyAggregateScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	dp := &dpOptions{Epsilon: 1, Delta: 1e-5}
	for _, opt := range m1(n.node.OptionList()) {
		name, _ := opt.Name()
		valExpr, _ := opt.Value()
		valSQL, err := newNode(valExpr).FormatSQL(ctx)
		if err != nil {
			continue
		}
		f, perr := strconv.ParseFloat(strings.TrimSpace(valSQL), 64)
		if perr != nil {
			continue
		}
		switch strings.ToLower(name) {
		case "epsilon":
			dp.Epsilon = f
		case "delta":
			dp.Delta = f
		}
	}
	ctx = withDPOptions(ctx, dp)
	for _, agg := range m1(n.node.AggregateList()) {
		if _, err := newNode(agg).FormatSQL(ctx); err != nil {
			return "", err
		}
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	// Format the GROUP BY entries so each ComputedColumn populates
	// the column-ref map with the right `<src> AS <out>` mapping;
	// otherwise the outer SELECT's reference to the group-by column
	// (e.g. `item#5`) cannot resolve to its source (e.g.
	// `item_partial#9`) emitted by the inner scan.
	groupByColumns := []string{}
	for _, col := range m1(n.node.GroupByList()) {
		if _, err := newNode(col).FormatSQL(ctx); err != nil {
			return "", err
		}
		colName := uniqueColumnName(ctx, m1(col.Column()))
		groupByColumns = append(groupByColumns, fmt.Sprintf("`%s`", colName))
	}
	columns := []string{}
	columnMap := columnRefMap(ctx)
	for _, col := range m1(n.node.ColumnList()) {
		colName := uniqueColumnName(ctx, col)
		if ref, exists := columnMap[colName]; exists {
			columns = append(columns, ref)
			delete(columnMap, colName)
		} else {
			columns = append(columns, fmt.Sprintf("`%s`", colName))
		}
	}
	formattedColumns := strings.Join(columns, ",")
	var groupBy string
	if len(groupByColumns) > 0 {
		groupBy = "GROUP BY " + strings.Join(groupByColumns, ",")
	}
	switch getInputPattern(input) {
	case InputKeep:
		return fmt.Sprintf("SELECT %s %s %s", formattedColumns, input, groupBy), nil
	case InputNeedsWrap:
		return fmt.Sprintf("SELECT %s FROM (%s) %s", formattedColumns, input, groupBy), nil
	case InputNeedsFrom:
		return fmt.Sprintf("SELECT %s FROM %s %s", formattedColumns, input, groupBy), nil
	}
	return "", fmt.Errorf("DP scan: unexpected input pattern: %s", input)
}

func (n *SetOperationItemNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	return newNode(m1(n.node.Scan())).FormatSQL(ctx)
}

func (n *SetOperationScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	var opType string
	switch m1(n.node.OpType()) {
	case googlesql.ResolvedSetOperationScanEnums_SetOperationTypeUnionAll:
		opType = "UNION ALL"
	case googlesql.ResolvedSetOperationScanEnums_SetOperationTypeUnionDistinct:
		opType = "UNION"
	case googlesql.ResolvedSetOperationScanEnums_SetOperationTypeIntersectAll:
		opType = "INTERSECT ALL"
	case googlesql.ResolvedSetOperationScanEnums_SetOperationTypeIntersectDistinct:
		opType = "INTERSECT"
	case googlesql.ResolvedSetOperationScanEnums_SetOperationTypeExceptAll:
		opType = "EXCEPT ALL"
	case googlesql.ResolvedSetOperationScanEnums_SetOperationTypeExceptDistinct:
		opType = "EXCEPT"
	default:
		opType = "UNKNOWN"
	}
	var queries []string
	for _, item := range m1(n.node.InputItemList()) {
		var outputColumns []string
		for _, outputColumn := range m1(item.OutputColumnList()) {
			outputColumns = append(outputColumns, fmt.Sprintf("`%s`", uniqueColumnName(ctx, outputColumn)))
		}
		query, err := newNode(item).FormatSQL(ctx)
		if err != nil {
			return "", err
		}

		formattedInput, err := formatInput(query)
		if err != nil {
			return "", err
		}

		queries = append(
			queries,
			fmt.Sprintf("SELECT %s %s",
				strings.Join(outputColumns, ", "),
				formattedInput,
			),
		)
	}
	columnMaps := []string{}
	if inputItems := m1(n.node.InputItemList()); len(inputItems) != 0 {
		for idx, col := range m1(inputItems[0].OutputColumnList()) {
			columnMaps = append(
				columnMaps,
				fmt.Sprintf(
					"`%s` AS `%s`",
					uniqueColumnName(ctx, col),
					uniqueColumnName(ctx, m1(n.node.ColumnList())[idx]),
				),
			)
		}
	}
	return fmt.Sprintf(
		"SELECT %s FROM (%s)",
		strings.Join(columnMaps, ","),
		strings.Join(queries, fmt.Sprintf(" %s ", opType)),
	), nil
}

func (n *OrderByScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	columns := []string{}
	columnMap := columnRefMap(ctx)
	for _, col := range m1(n.node.ColumnList()) {
		colName := uniqueColumnName(ctx, col)
		if ref, exists := columnMap[colName]; exists {
			columns = append(columns, ref)
			delete(columnMap, colName)
		} else {
			columns = append(
				columns,
				fmt.Sprintf("`%s`", colName),
			)
		}
	}
	orderByColumns := []string{}
	for _, item := range m1(n.node.OrderByItemList()) {
		colName := uniqueColumnName(ctx, m1(m1(item.ColumnRef()).Column()))
		switch m1(item.NullOrder()) {
		case googlesql.ResolvedOrderByItemEnums_NullOrderModeNullsFirst:
			orderByColumns = append(
				orderByColumns,
				fmt.Sprintf("(`%s` IS NOT NULL)", colName),
			)
		case googlesql.ResolvedOrderByItemEnums_NullOrderModeNullsLast:
			orderByColumns = append(
				orderByColumns,
				fmt.Sprintf("(`%s` IS NULL)", colName),
			)
		}
		if m1(item.IsDescending()) {
			orderByColumns = append(orderByColumns, fmt.Sprintf("`%s` COLLATE googlesqlite_collate DESC", colName))
		} else {
			orderByColumns = append(orderByColumns, fmt.Sprintf("`%s` COLLATE googlesqlite_collate", colName))
		}
	}
	formattedInput, err := formatInput(input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"SELECT %s %s ORDER BY %s",
		strings.Join(columns, ","),
		formattedInput,
		strings.Join(orderByColumns, ","),
	), nil
}

func (n *LimitOffsetScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	columns := []string{}
	columnMap := columnRefMap(ctx)
	for _, col := range m1(n.node.ColumnList()) {
		colName := uniqueColumnName(ctx, col)
		if ref, exists := columnMap[colName]; exists {
			columns = append(columns, ref)
			delete(columnMap, colName)
		} else {
			columns = append(
				columns,
				fmt.Sprintf("`%s`", colName),
			)
		}
	}
	formattedInput, err := formatInput(input)
	if err != nil {
		return "", err
	}
	var limitExpr string
	if m1(n.node.Limit()) != nil {
		expr, err := newNode(m1(n.node.Limit())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		limitExpr = fmt.Sprintf("LIMIT %s", expr)
	}
	var offsetExpr string
	if m1(n.node.Offset()) != nil {
		expr, err := newNode(m1(n.node.Offset())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		offsetExpr = fmt.Sprintf("OFFSET %s", expr)
	}
	return fmt.Sprintf(
		"SELECT %s %s %s %s",
		strings.Join(columns, ","),
		formattedInput,
		limitExpr,
		offsetExpr,
	), nil
}

func (n *WithRefScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	tableName, _ := n.node.WithQueryName()
	tableToColumnListMap := tableNameToColumnListMap(ctx)
	columnDefs := tableToColumnListMap[tableName]
	columns := m1(n.node.ColumnList())
	if len(columnDefs) != len(columns) {
		return "", fmt.Errorf(
			"column num mismatch. defined column num is %d but used %d column",
			len(columnDefs), len(columns),
		)
	}
	formattedColumns := []string{}
	for i := range columnDefs {
		formattedColumns = append(
			formattedColumns,
			fmt.Sprintf("`%s` AS `%s`", uniqueColumnName(ctx, columnDefs[i]), uniqueColumnName(ctx, columns[i])),
		)
	}
	return fmt.Sprintf("(SELECT %s FROM `%s`)", strings.Join(formattedColumns, ","), tableName), nil
}

func (n *AnalyticScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	formattedInput, err := formatInput(input)
	if err != nil {
		return "", err
	}
	ctx = withAnalyticInputScan(ctx, formattedInput)
	ctx, emulationUsed := withAnalyticEmulationFlag(ctx)
	orderColumnNames := analyticOrderColumnNamesFromContext(ctx)
	var scanOrderBy []*analyticOrderBy
	for _, group := range m1(n.node.FunctionGroupList()) {
		scanOrderBy = []*analyticOrderBy{}

		if m1(group.PartitionBy()) != nil {
			var partitionColumns []string
			for _, columnRef := range m1(m1(group.PartitionBy()).PartitionByList()) {
				colName := fmt.Sprintf("`%s`", uniqueColumnName(ctx, m1(columnRef.Column())))
				partitionColumns = append(
					partitionColumns,
					colName,
				)
				order := &analyticOrderBy{
					column: colName,
					isAsc:  true,
				}
				orderColumnNames.values = append(orderColumnNames.values, order)
				scanOrderBy = append(scanOrderBy, order)
			}
			ctx = withAnalyticPartitionColumnNames(ctx, partitionColumns)
		}
		if m1(group.OrderBy()) != nil {
			for _, item := range m1(m1(group.OrderBy()).OrderByItemList()) {
				colName := uniqueColumnName(ctx, m1(m1(item.ColumnRef()).Column()))
				formattedColName := fmt.Sprintf("`%s`", colName)
				nullOrder := nullOrderUnspecified
				switch m1(item.NullOrder()) {
				case googlesql.ResolvedOrderByItemEnums_NullOrderModeNullsFirst:
					nullOrder = nullOrderFirst
				case googlesql.ResolvedOrderByItemEnums_NullOrderModeNullsLast:
					nullOrder = nullOrderLast
				}
				order := &analyticOrderBy{
					column:    formattedColName,
					isAsc:     !m1(item.IsDescending()),
					nullOrder: nullOrder,
				}
				orderColumnNames.values = append(orderColumnNames.values, order)
				scanOrderBy = append(scanOrderBy, order)
			}
		}
		if _, err := newNode(group).FormatSQL(ctx); err != nil {
			return "", err
		}

		// Reset context after each analytic function group
		orderColumnNames.values = []*analyticOrderBy{}
		ctx = withAnalyticPartitionColumnNames(ctx, nil)
	}
	columns := []string{}
	columnMap := columnRefMap(ctx)
	for _, col := range m1(n.node.ColumnList()) {
		colName := uniqueColumnName(ctx, col)
		if ref, exists := columnMap[colName]; exists {
			columns = append(columns, ref)
			delete(columnMap, colName)
		} else {
			columns = append(
				columns,
				fmt.Sprintf("`%s`", colName),
			)
		}
	}
	var orderColumnFormattedNames []string
	for _, col := range scanOrderBy {
		// NULLS FIRST/LAST is implemented as a sort key prefix:
		// emit (col IS NOT NULL) for NULLS FIRST so nulls (0) sort
		// before non-nulls (1); emit (col IS NULL) for NULLS LAST so
		// non-nulls (0) come before nulls (1). Skipped when the
		// query did not request explicit null ordering.
		switch col.nullOrder {
		case nullOrderFirst:
			orderColumnFormattedNames = append(
				orderColumnFormattedNames,
				fmt.Sprintf("(%s IS NOT NULL)", col.column),
			)
		case nullOrderLast:
			orderColumnFormattedNames = append(
				orderColumnFormattedNames,
				fmt.Sprintf("(%s IS NULL)", col.column),
			)
		}
		if col.isAsc {
			orderColumnFormattedNames = append(
				orderColumnFormattedNames,
				fmt.Sprintf("%s COLLATE googlesqlite_collate", col.column),
			)
		} else {
			orderColumnFormattedNames = append(
				orderColumnFormattedNames,
				fmt.Sprintf("%s COLLATE googlesqlite_collate DESC", col.column),
			)
		}
	}
	var orderBy string
	if len(orderColumnFormattedNames) != 0 {
		orderBy = fmt.Sprintf("ORDER BY %s", strings.Join(orderColumnFormattedNames, ","))
	}
	orderColumnNames.values = []*analyticOrderBy{}
	if !*emulationUsed {
		// Every analytic call in this scan went through the native
		// OVER path; the row_id column the predecessor's emulation
		// needs is dead weight here, so skip the wrap and let
		// SQLite read the input directly.
		return fmt.Sprintf(
			"SELECT %s %s %s",
			strings.Join(columns, ","),
			formattedInput,
			orderBy,
		), nil
	}
	return fmt.Sprintf(
		"SELECT %s FROM (SELECT *, ROW_NUMBER() OVER() AS `row_id` %s) %s",
		strings.Join(columns, ","),
		formattedInput,
		orderBy,
	), nil
}

// FormatSQL passes the SampleScan through to its InputScan. We do
// not honour the sample size / method semantics — for the emulator's
// correctness contract, returning all rows of the input gives the
// same answer set for every aggregate the upstream Examples exercise
// (and the DP rewriter relies on SampleScan as a transparent node
// for `max_groups_contributed` limiting, not as a probabilistic
// sampler).
func (n *SampleScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	return newNode(m1(n.node.InputScan())).FormatSQL(ctx)
}

func (n *ComputedColumnNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	expr, err := newNode(m1(n.node.Expr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	col, _ := n.node.Column()
	uniqueName := uniqueColumnName(ctx, col)
	query := fmt.Sprintf("%s AS `%s`", expr, uniqueColumnName(ctx, col))
	columnMap := columnRefMap(ctx)
	columnMap[uniqueName] = query
	arraySubqueryColumnNames := arraySubqueryColumnNameFromContext(ctx)
	if arraySubqueryColumnNames != nil {
		arraySubqueryColumnNames.names = append(arraySubqueryColumnNames.names, fmt.Sprintf("`%s`", m1(col.Name())))
	}
	return query, nil
}

func (n *ProjectScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	for _, col := range m1(n.node.ExprList()) {
		// assign expr to columnRefMap
		if _, err := newNode(col).FormatSQL(ctx); err != nil {
			return "", err
		}
	}
	input, err := newNode(m1(n.node.InputScan())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	columns := []string{}
	columnMap := columnRefMap(ctx)
	for _, col := range m1(n.node.ColumnList()) {
		colName := uniqueColumnName(ctx, col)
		if ref, exists := columnMap[colName]; exists {
			columns = append(columns, ref)
			delete(columnMap, colName)
		} else {
			columns = append(
				columns,
				fmt.Sprintf("`%s`", colName),
			)
		}
	}
	formattedInput, err := formatInput(input)
	if err != nil {
		return "", err
	}
	formattedColumns := strings.Join(columns, ",")
	return fmt.Sprintf("SELECT %s %s", formattedColumns, formattedInput), nil
}

func (n *TVFScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n == nil || n.node == nil {
		return "", nil
	}
	tvf, err := n.node.Tvf()
	if err != nil || tvf == nil {
		return "", fmt.Errorf("TVF handle missing on TVFScan")
	}
	pathParts, _ := tvf.FunctionNamePath()
	if len(pathParts) == 0 {
		name, _ := tvf.Name()
		pathParts = []string{name}
	}
	tvfMap := tvfMapFromContext(ctx)
	var spec *TVFSpec
	for i := 0; i < len(pathParts) && spec == nil; i++ {
		spec = tvfMap[formatPath(pathParts[i:])]
	}
	if spec == nil {
		return "", fmt.Errorf("TVF spec not found for %s", strings.Join(pathParts, "."))
	}
	argValues := make([]string, 0, len(spec.Args))
	for _, argNode := range m1(n.node.ArgumentList()) {
		expr, _ := argNode.Expr()
		if expr == nil {
			argValues = append(argValues, "NULL")
			continue
		}
		formatted, err := newNode(expr).FormatSQL(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to format TVF argument: %w", err)
		}
		argValues = append(argValues, formatted)
	}
	callOutputColumns := make([]string, 0, len(spec.OutputColumns))
	for _, col := range m1(n.node.ColumnList()) {
		callOutputColumns = append(callOutputColumns, uniqueColumnName(ctx, col))
	}
	return spec.CallSQL(argValues, callOutputColumns)
}

// FormatSQL Formats the outermost query statement that runs and produces rows of output, like a SELECT
// The node's `OutputColumnList()` gives user-visible column names that should be returned. There may be duplicate names,
// and multiple output columns may reference the same column from `Query()`
// https://github.com/google/googlesql/blob/master/docs/resolved_ast.md#ResolvedQueryStmt
func (n *QueryStmtNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	input, err := newNode(m1(n.node.Query())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}

	var columns []string
	for _, outputColumnNode := range m1(n.node.OutputColumnList()) {
		columns = append(
			columns,
			fmt.Sprintf("`%s` AS `%s`",
				uniqueColumnName(ctx, m1(outputColumnNode.Column())),
				m1(outputColumnNode.Name()),
			),
		)
	}

	return fmt.Sprintf(
		"SELECT %s FROM (%s)",
		strings.Join(columns, ", "),
		input,
	), nil
}

func (n *DropStmtNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	namePath := namePathFromContext(ctx)
	tableName := namePath.format2(n.node.NamePath())
	objectType, _ := n.node.ObjectType()
	if m1(n.node.IsIfExists()) {
		return fmt.Sprintf("DROP %s IF EXISTS `%s`", objectType, tableName), nil
	}
	return fmt.Sprintf("DROP %s `%s`", objectType, tableName), nil
}

// RecursiveRefScanNode renders as the enclosing CTE's queryName so
// SQLite's native `WITH RECURSIVE name AS (... UNION ALL SELECT ...
// FROM name)` emits a self-reference correctly.
func (n *RecursiveRefScanNode) FormatSQL(ctx context.Context) (string, error) {
	name := recursiveCteName(ctx)
	if name == "" {
		return "", fmt.Errorf("recursive ref scan rendered outside a recursive WITH entry")
	}
	// Quote to match the WITH definition (WithEntryNode) so names with
	// hyphens and other non-bare-identifier characters resolve.
	return fmt.Sprintf("`%s`", name), nil
}

// findRecursiveRefScanCols walks a scan tree looking for the column
// list exposed by the (one and only) ResolvedRecursiveRefScan in a
// recursive UNION term. SQLite-side `WITH RECURSIVE t(...)` must use
// these IDs as the explicit column list so the recursive term's
// `FROM t` references resolve to the right columns.
//
// Returns nil when no RecursiveRefScan is reachable.
func findRecursiveRefScanCols(node googlesql.ResolvedNode) []*googlesql.ResolvedColumn {
	if node == nil {
		return nil
	}
	if kind, _ := node.NodeKind(); kind == googlesql.ResolvedNodeKindResolvedRecursiveRefScan {
		if scan, ok := node.(googlesql.ResolvedScanNode); ok {
			cols, _ := scan.MutableColumnList()
			return cols
		}
	}
	children, _ := node.GetChildNodes()
	for _, c := range children {
		if got := findRecursiveRefScanCols(c); len(got) > 0 {
			return got
		}
	}
	return nil
}

// RecursiveScanNode renders as `<non_recursive_branch> UNION [ALL]
// <recursive_branch>`. The recursive-cte reference must stay at the
// top-level FROM of the recursive branch (SQLite restriction). The
// enclosing WithEntryNode emits an explicit column list using the
// RecursiveScan canonical ColumnList so the outer query's references
// resolve. The recursive term internally uses RecursiveRefScan IDs;
// those are substituted with canonical IDs here so SQLite sees one
// consistent name set.
func (n *RecursiveScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	canonical := m1(n.node.ColumnList())
	formatBranch := func(item *googlesql.ResolvedSetOperationItem) (string, error) {
		scanSQL, err := newNode(m1(item.Scan())).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		trimmed := strings.TrimSpace(scanSQL)
		if strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH") {
			return scanSQL, nil
		}
		out := m1(item.OutputColumnList())
		cols := make([]string, 0, len(out))
		for _, c := range out {
			cols = append(cols, fmt.Sprintf("`%s`", uniqueColumnName(ctx, c)))
		}
		formatted, err := formatInput(scanSQL)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("SELECT %s %s",
			strings.Join(cols, ","),
			formatted,
		), nil
	}
	nonRec, err := formatBranch(m1(n.node.NonRecursiveTerm()))
	if err != nil {
		return "", err
	}
	rec, err := formatBranch(m1(n.node.RecursiveTerm()))
	if err != nil {
		return "", err
	}
	// Substitute RecursiveRefScan column IDs in the recursive branch
	// with the canonical RecursiveScan ColumnList IDs so SQLite's
	// single-column-name view of the CTE matches the outer query's
	// references. Same number of columns — pair by position.
	if rt := m1(n.node.RecursiveTerm()); rt != nil {
		if scan, _ := rt.Scan(); scan != nil {
			refCols := findRecursiveRefScanCols(scan)
			if len(refCols) == len(canonical) {
				for i, ref := range refCols {
					if ref == nil {
						continue
					}
					refName := fmt.Sprintf("`%s`", uniqueColumnName(ctx, ref))
					canName := fmt.Sprintf("`%s`", uniqueColumnName(ctx, canonical[i]))
					if refName != canName {
						rec = strings.ReplaceAll(rec, refName, canName)
					}
				}
			}
		}
	}
	op := "UNION ALL"
	if m1(n.node.OpType()) == googlesql.ResolvedRecursiveScanEnums_RecursiveSetOperationTypeUnionDistinct {
		op = "UNION"
	}
	return fmt.Sprintf("%s %s %s", nonRec, op, rec), nil
}

func (n *WithScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	keyword := "WITH"
	for _, entry := range m1(n.node.WithEntryList()) {
		sub := m1(entry.WithSubquery())
		if sub == nil {
			continue
		}
		if kind, _ := sub.NodeKind(); kind == googlesql.ResolvedNodeKindResolvedRecursiveScan {
			keyword = "WITH RECURSIVE"
			break
		}
	}
	// Count CTE references in the WITH body and across WithEntry
	// subqueries (a later CTE may reference an earlier one). Pass
	// the counts down so WithEntryNode.FormatSQL can decide whether
	// to emit `MATERIALIZED`.
	refCounts := map[string]int{}
	if conn := connFromContext(ctx); conn == nil || conn.MaterializeCTE() {
		query, _ := n.node.Query()
		if query != nil {
			collectCteRefCounts(query, refCounts)
		}
		for _, entry := range m1(n.node.WithEntryList()) {
			sub, _ := entry.WithSubquery()
			if sub == nil {
				continue
			}
			collectCteRefCounts(sub, refCounts)
		}
	}
	subCtx := withCteRefCounts(ctx, refCounts)
	queries := []string{}
	for _, entry := range m1(n.node.WithEntryList()) {
		sql, err := newNode(entry).FormatSQL(subCtx)
		if err != nil {
			return "", err
		}
		queries = append(queries, sql)
	}
	query, err := newNode(m1(n.node.Query())).FormatSQL(subCtx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"%s %s %s",
		keyword,
		strings.Join(queries, ", "),
		query,
	), nil
}

// collectCteRefCounts walks a resolved-scan subtree and increments
// counts[name] for every ResolvedWithRefScan it encounters. Used by
// WithScanNode to decide whether each CTE body should be emitted as
// `MATERIALIZED`.
func collectCteRefCounts(node googlesql.ResolvedNode, counts map[string]int) {
	if node == nil {
		return
	}
	if kind, _ := node.NodeKind(); kind == googlesql.ResolvedNodeKindResolvedWithRefScan {
		if ref, ok := node.(*googlesql.ResolvedWithRefScan); ok {
			name, _ := ref.WithQueryName()
			counts[name]++
		}
	}
	children, _ := node.GetChildNodes()
	for _, c := range children {
		collectCteRefCounts(c, counts)
	}
}

func (n *WithEntryNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	queryName, _ := n.node.WithQueryName()
	sub := m1(n.node.WithSubquery())
	subCtx := ctx
	subKind, _ := sub.NodeKind()
	if subKind == googlesql.ResolvedNodeKindResolvedRecursiveScan {
		// Make the queryName visible to any nested
		// RecursiveRefScanNode so it can render as the CTE name.
		subCtx = withRecursiveCteName(ctx, queryName)
	}
	subquery, err := newNode(sub).FormatSQL(subCtx)
	if err != nil {
		return "", err
	}
	tableToColumnList := tableNameToColumnListMap(ctx)
	tableToColumnList[queryName] = m1(sub.MutableColumnList())
	hint := cteMaterializeHint(ctx, queryName, subKind)
	if subKind == googlesql.ResolvedNodeKindResolvedRecursiveScan {
		// SQLite needs an explicit column list on the recursive CTE
		// so both UNION branches map their projected columns by
		// position. Use the RecursiveScan canonical ColumnList — the
		// outer query references those IDs. RecursiveScanNode.FormatSQL
		// substitutes RecursiveRefScan-internal IDs with these
		// canonical IDs so `FROM t` references resolve.
		canonical := m1(sub.MutableColumnList())
		cols := make([]string, 0, len(canonical))
		for _, c := range canonical {
			cols = append(cols, fmt.Sprintf("`%s`", uniqueColumnName(ctx, c)))
		}
		return fmt.Sprintf("`%s`(%s) AS%s ( %s )", queryName, strings.Join(cols, ","), hint, subquery), nil
	}
	return fmt.Sprintf("`%s` AS%s ( %s )", queryName, hint, subquery), nil
}

// cteMaterializeHint returns either " MATERIALIZED" (with a leading
// space) or "" depending on whether SQLite should be told to push
// this CTE into a transient temp table instead of inlining its body
// once per reference. Recursive CTEs are skipped — SQLite always
// materialises the recursion accumulator anyway, and the hint
// interacts oddly with the `WITH RECURSIVE` keyword.
func cteMaterializeHint(ctx context.Context, queryName string, subKind googlesql.ResolvedNodeKind) string {
	if subKind == googlesql.ResolvedNodeKindResolvedRecursiveScan {
		return ""
	}
	counts := cteRefCounts(ctx)
	if counts[queryName] >= 2 {
		return " MATERIALIZED"
	}
	return ""
}

func (n *AnalyticFunctionGroupNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}

	var queries []string
	for _, column := range m1(n.node.AnalyticFunctionList()) {
		sql, err := newNode(column).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		queries = append(queries, sql)
	}
	return strings.Join(queries, ","), nil
}

func (n *DMLValueNode) FormatSQL(ctx context.Context) (string, error) {
	if n == nil {
		return "", nil
	}
	return newNode(m1(n.node.Value())).FormatSQL(ctx)
}

func (n *InsertRowNode) FormatSQL(ctx context.Context) (string, error) {
	if n == nil {
		return "", nil
	}
	values := []string{}
	for _, value := range m1(n.node.ValueList()) {
		sql, err := newNode(value).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		values = append(values, sql)
	}
	return strings.Join(values, ","), nil
}

func (n *InsertStmtNode) FormatSQL(ctx context.Context) (string, error) {
	if n == nil {
		return "", nil
	}
	table, err := getTableName(ctx, m1(n.node.TableScan()))
	if err != nil {
		return "", err
	}
	insertCols := m1(n.node.InsertColumnList())
	insertedNames := make(map[string]struct{}, len(insertCols))
	columns := make([]string, 0, len(insertCols))
	for _, col := range insertCols {
		name := m1(col.Name())
		insertedNames[name] = struct{}{}
		columns = append(columns, fmt.Sprintf("`%s`", name))
	}
	// Collect any column DEFAULT expressions for columns the INSERT
	// omitted. Splicing them into the rewritten INSERT here is what
	// makes `CREATE TABLE t (... col TYPE DEFAULT '...')` actually
	// substitute the default — bqe#211.
	defaults := missingColumnDefaults(ctx, table, insertedNames)
	for _, d := range defaults {
		columns = append(columns, fmt.Sprintf("`%s`", d.name))
	}
	query, _ := n.node.Query()
	if query != nil {
		stmt, err := newNode(query).FormatSQL(withUseColumnID(ctx))
		if err != nil {
			return "", err
		}
		// INSERT ... SELECT form: extending the column list would
		// require rewriting the SELECT to project the defaults too.
		// Skip default substitution here — defaults still apply to
		// the VALUES form, which covers the common case.
		_ = defaults
		return fmt.Sprintf("INSERT INTO `%s` (%s) %s",
			table,
			strings.Join(columns[:len(insertCols)], ","),
			stmt,
		), nil
	}
	rows := []string{}
	for _, row := range m1(n.node.RowList()) {
		sql, err := newNode(row).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		full := sql
		for _, d := range defaults {
			full += "," + d.expr
		}
		rows = append(rows, fmt.Sprintf("(%s)", full))
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s",
		table,
		strings.Join(columns, ","),
		strings.Join(rows, ","),
	), nil
}

// missingColumnDefaults returns one entry per table column that has
// a DEFAULT expression and is not present in inserted. Callers append
// the names to the INSERT column list and the exprs to each row's
// VALUES tuple, which lets the standard SQLite path supply the
// default at execution time.
func missingColumnDefaults(ctx context.Context, storageTable string, inserted map[string]struct{}) []struct{ name, expr string } {
	a := analyzerFromContext(ctx)
	if a == nil || a.catalog == nil {
		return nil
	}
	spec, ok := a.catalog.tableMap[storageTable]
	if !ok || spec == nil {
		return nil
	}
	var out []struct{ name, expr string }
	for _, col := range spec.Columns {
		if _, present := inserted[col.Name]; present {
			continue
		}
		if col.DefaultExpr == "" {
			continue
		}
		out = append(out, struct{ name, expr string }{col.Name, col.DefaultExpr})
	}
	return out
}

func (n *DeleteStmtNode) FormatSQL(ctx context.Context) (string, error) {
	if n == nil {
		return "", nil
	}
	table, err := getTableName(ctx, m1(n.node.TableScan()))
	if err != nil {
		return "", err
	}
	where, err := newNode(m1(n.node.WhereExpr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"DELETE FROM `%s` WHERE %s",
		table,
		where,
	), nil
}

func (n *UpdateItemNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	targetCtx := unuseColumnID(withoutUseTableNameForColumn(ctx))
	targetNode := m1(n.node.Target())
	setValue, err := newNode(m1(n.node.SetValue())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	// SET col.field = value: SQL forbids assigning to a function
	// call. Rewrite into a whole-column replacement that goes
	// through googlesqlite_struct_with_field_set, which preserves
	// every other field of the struct (bqe#299, bqe#128).
	if rewritten, ok, err := rewriteStructFieldSet(targetCtx, targetNode, setValue); err != nil {
		return "", err
	} else if ok {
		return rewritten, nil
	}
	target, err := newNode(targetNode).FormatSQL(targetCtx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s=%s", target, setValue), nil
}

// rewriteStructFieldSet detects `target = ResolvedGetStructField(...)`
// chains and rewrites them as
//
//	col = googlesqlite_struct_with_field_set(
//	          ...struct_with_field_set(col, outerIdx, innerSet)... ,
//	          outermostIdx,
//	          assignedValue)
//
// so the assigned column is the original column and the value
// preserves all fields not on the assignment path. ok=false means
// the target is not a struct-field reference; the caller falls back
// to the regular `target = value` form.
func rewriteStructFieldSet(ctx context.Context, target googlesql.ResolvedExprNode, value string) (string, bool, error) {
	get, ok := target.(*googlesql.ResolvedGetStructField)
	if !ok {
		return "", false, nil
	}
	// Walk inwards, collecting (idx) at each level until we hit the
	// underlying ResolvedColumnRef.
	type level struct {
		idx int
	}
	var levels []level
	cur := googlesql.ResolvedExprNode(get)
	for {
		g, ok := cur.(*googlesql.ResolvedGetStructField)
		if !ok {
			break
		}
		idx, _ := g.FieldIdx()
		levels = append(levels, level{idx: int(idx)})
		next, _ := g.Expr()
		cur = next
	}
	colRef, ok := cur.(*googlesql.ResolvedColumnRef)
	if !ok {
		return "", false, nil
	}
	colSQL, err := newNode(colRef).FormatSQL(ctx)
	if err != nil {
		return "", false, err
	}
	// Build the wrapped value from the inside out. The outermost
	// (assigned) field is the LAST element appended above.
	expr := value
	for i := 0; i < len(levels); i++ {
		idx := levels[i].idx
		if i == len(levels)-1 {
			expr = fmt.Sprintf("googlesqlite_struct_with_field_set(%s, %d, %s)",
				colSQL, idx, expr)
		} else {
			// Nested struct field: wrap in get_struct_field for the
			// outer container, then set within. For now the simple
			// single-level path suffices for bqe#299 / #128; nested
			// `info.address.city` would need an additional builder
			// pass and is documented as a follow-up.
			return "", false, nil
		}
	}
	return fmt.Sprintf("%s=%s", colSQL, expr), true, nil
}

func (n *UpdateStmtNode) FormatSQL(ctx context.Context) (string, error) {
	if n == nil {
		return "", nil
	}
	table, err := getTableName(ctx, m1(n.node.TableScan()))
	if err != nil {
		return "", err
	}
	updateItems := []string{}
	for _, item := range m1(n.node.UpdateItemList()) {
		sql, err := newNode(item).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		updateItems = append(updateItems, sql)
	}
	where, err := newNode(m1(n.node.WhereExpr())).FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s",
		table,
		strings.Join(updateItems, ","),
		where,
	), nil
}

func (n *ArgumentRefNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	return fmt.Sprintf("@%s", m1(n.node.Name())), nil
}
