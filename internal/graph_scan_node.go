package internal

import (
	"context"
	"fmt"
	"strings"
	"sync"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/goccy/googlesqlite/internal/value"
)

// Wrapper node types for the resolved graph query AST. Each one
// implements the Formatter contract used by the rest of the
// driver pipeline. See internal/node.go for the dispatch table.
//
// Translation strategy for the simplest single-node-match shape:
//
//	GRAPH G MATCH (l:L) RETURN l.col
//
// becomes
//
//	SELECT <underlying-col-for-l.col> AS col
//	FROM   <underlying-table-for-L>
//
// Multi-element MATCH patterns (path patterns, edges, multiple
// hops) are not yet handled and return a clear error pointing at
// the missing piece. The pattern's first GraphNodeScan is the
// only one walked; everything else surfaces as a "not yet
// supported" error so we don't silently emit a broken query.

type GraphTableScanNode struct {
	node *googlesql.ResolvedGraphTableScan
}

type GraphLinearScanNode struct {
	node *googlesql.ResolvedGraphLinearScan
}

type GraphPathScanNode struct {
	node *googlesql.ResolvedGraphPathScan
}

type GraphNodeScanNode struct {
	node *googlesql.ResolvedGraphNodeScan
}

type GraphEdgeScanNode struct {
	node *googlesql.ResolvedGraphEdgeScan
}

type GraphGetElementPropertyNode struct {
	node *googlesql.ResolvedGraphGetElementProperty
}

type GraphRefScanNode struct {
	node *googlesql.ResolvedGraphRefScan
}

// GraphScanNode wraps a ResolvedGraphScan, which is the top-level
// MATCH stage produced inside the linear scan. It holds a list of
// path scans (typically one per comma-separated MATCH pattern).
type GraphScanNode struct {
	node *googlesql.ResolvedGraphScan
}

// graphElementContext is threaded through the SQL formatter on a
// per-query basis. It maps a graph-element column id to the SQL
// table alias and underlying element table the analyzer matched
// it to, so a subsequent GraphGetElementProperty can resolve
// `element.property` into a plain `alias.column` reference.
type graphElementContext struct {
	// columnAlias maps a column id (the graph element column
	// produced by a GraphNodeScan / GraphEdgeScan) to the SQL
	// alias under which the underlying table is exposed.
	columnAlias map[int32]string
	// elementTable maps a column id to the GraphElementTable
	// catalog object backing it, used to resolve property names
	// to their underlying-table column ordinals.
	elementTable map[int32]googlesql.GraphElementTableNode
	// lastTableRef is the most recent table-reference SQL emitted
	// by a graph scan stage. GraphRefScan uses it to satisfy
	// subsequent ProjectScan / FilterScan stages that ask for
	// their input scan's SQL.
	lastTableRef string
	// pathInfo maps a path-column id (from GraphPathScan.Path)
	// to the ordered list of node/edge aliases that make up the
	// path. Used by PATH_LENGTH / PATH_FIRST / NODES / EDGES /
	// IS_TRAIL / IS_ACYCLIC.
	pathInfo map[int32]*graphPathInfo
}

type graphPathInfo struct {
	NodeAliases []string
	EdgeAliases []string
}

type graphCtxKey struct{}

func graphCtxFromContext(ctx context.Context) *graphElementContext {
	if v, ok := ctx.Value(graphCtxKey{}).(*graphElementContext); ok {
		return v
	}
	return nil
}

func withGraphCtx(ctx context.Context, gc *graphElementContext) context.Context {
	return context.WithValue(ctx, graphCtxKey{}, gc)
}

// FormatSQL — GraphTableScan is the top-level scan produced by a
// GRAPH ... MATCH ... RETURN clause. ShapeExprList holds the
// computed columns that form the RETURN list; InputScan is the
// matched pattern.
func (n *GraphTableScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	gc := &graphElementContext{
		columnAlias:  map[int32]string{},
		elementTable: map[int32]googlesql.GraphElementTableNode{},
		pathInfo:     map[int32]*graphPathInfo{},
	}
	ctx = withGraphCtx(ctx, gc)
	inputScan, err := n.node.InputScan()
	if err != nil {
		return "", err
	}
	formatter := newNode(inputScan)
	if formatter == nil {
		k, _ := inputScan.NodeKind()
		return "", fmt.Errorf("graph table scan: unsupported input node kind %s", k)
	}
	innerSQL, err := formatter.FormatSQL(ctx)
	if err != nil {
		return "", err
	}
	// ShapeExprList may be empty when the analyzer already split
	// the RETURN projection out into a ProjectScan in the inner
	// pipeline. In that case the inner SQL already projects the
	// right columns; just pass it through (wrapped so the
	// enclosing query sees a parenthesised subquery).
	shape, err := n.node.ShapeExprList()
	if err != nil {
		return "", err
	}
	if len(shape) == 0 {
		return fmt.Sprintf("(%s)", innerSQL), nil
	}
	exprs := make([]string, 0, len(shape))
	for _, sc := range shape {
		col, err := sc.Column()
		if err != nil {
			return "", err
		}
		expr, err := sc.Expr()
		if err != nil {
			return "", err
		}
		exprSQL, err := newNode(expr).FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		exprs = append(exprs, fmt.Sprintf("%s AS `%s`", exprSQL, uniqueColumnName(ctx, col)))
	}
	return fmt.Sprintf("(SELECT %s FROM (%s))", strings.Join(exprs, ", "), innerSQL), nil
}

// FormatSQL — a graph linear scan is a pipeline of stages
// (typically [GraphScan, GraphRefScan] where the GraphScan
// performs the MATCH and the GraphRefScan represents the
// subsequent RETURN's reference back to the MATCH binding).
// Earlier stages establish graph element column bindings as a
// side-effect on the shared graphElementContext. We accumulate
// each stage's SQL fragment and propagate the last non-empty
// one — GraphRefScan returns "" by design.
func (n *GraphLinearScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	scans, err := n.node.ScanList()
	if err != nil {
		return "", err
	}
	if len(scans) == 0 {
		return "(SELECT NULL)", nil
	}
	var last string
	for i, s := range scans {
		fmtter := newNode(s)
		if fmtter == nil {
			kind, _ := s.NodeKind()
			return "", fmt.Errorf("graph linear scan stage %d: unsupported node kind %s", i, kind)
		}
		out, err := fmtter.FormatSQL(ctx)
		if err != nil {
			return "", err
		}
		if out != "" {
			last = out
		}
	}
	return last, nil
}

func (n *GraphPathScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	inputs, err := n.node.InputScanList()
	if err != nil {
		return "", err
	}
	if len(inputs) == 0 {
		return "(SELECT NULL)", nil
	}
	// Single-element pattern still goes through the legacy path so
	// existing tests that don't exercise edges keep working.
	if len(inputs) == 1 {
		return newNode(inputs[0]).FormatSQL(ctx)
	}
	// Multi-hop: walk the alternating node/edge/node/... sequence
	// and emit a `<table> AS <alias> JOIN ... ON ...` chain. The
	// individual node-scan formatters still run so they can stash
	// the alias→element-table mapping on graphCtx for downstream
	// property accesses.
	var b strings.Builder
	var prevNodeAlias string
	var prevNodeMeta *graphElementMetadata
	var pendingEdgeAlias string
	var pendingEdgeMeta *graphElementMetadata
	var pendingEdgeOrientation googlesql.ResolvedGraphEdgeScanEnums_EdgeOrientation
	pathNodeAliases := []string{}
	pathEdgeAliases := []string{}
	for i, in := range inputs {
		switch sub := in.(type) {
		case *googlesql.ResolvedGraphNodeScan:
			nodeSQL, err := (&GraphNodeScanNode{node: sub}).FormatSQL(ctx)
			if err != nil {
				return "", err
			}
			alias := lastBoundAlias(ctx)
			meta := lastBoundMeta(ctx)
			if i == 0 {
				b.WriteString(nodeSQL)
			} else if pendingEdgeAlias != "" {
				// We've already emitted prevNode then edge; now
				// close the chain by joining the edge to this
				// new node.
				joinCond, err := nodeEdgeJoinCondition(alias, meta, pendingEdgeAlias, pendingEdgeMeta, pendingEdgeOrientation, false)
				if err != nil {
					return "", err
				}
				fmt.Fprintf(&b, " JOIN %s ON %s", nodeSQL, joinCond)
				pendingEdgeAlias = ""
				pendingEdgeMeta = nil
			} else {
				// Two consecutive nodes (shouldn't happen, but
				// fall back to CROSS JOIN).
				fmt.Fprintf(&b, " CROSS JOIN %s", nodeSQL)
			}
			prevNodeAlias = alias
			prevNodeMeta = meta
			pathNodeAliases = append(pathNodeAliases, alias)
		case *googlesql.ResolvedGraphEdgeScan:
			edgeSQL, edgeAlias, edgeMeta, edgeOrient, err := formatEdgeScan(ctx, sub)
			if err != nil {
				return "", err
			}
			joinCond, err := nodeEdgeJoinCondition(prevNodeAlias, prevNodeMeta, edgeAlias, edgeMeta, edgeOrient, true)
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&b, " JOIN %s ON %s", edgeSQL, joinCond)
			pendingEdgeAlias = edgeAlias
			pendingEdgeMeta = edgeMeta
			pendingEdgeOrientation = edgeOrient
			pathEdgeAliases = append(pathEdgeAliases, edgeAlias)
		default:
			return "", fmt.Errorf("unsupported graph path element %T", sub)
		}
	}
	if gc := graphCtxFromContext(ctx); gc != nil {
		gc.lastTableRef = b.String()
		// If this path scan has a named path column, register the
		// alias sequence for downstream PATH_LENGTH / NODES /
		// EDGES / IS_TRAIL / IS_ACYCLIC lookups.
		if pcol, _ := n.node.Path(); pcol != nil {
			if col, _ := pcol.Column(); col != nil {
				id, _ := col.ColumnId()
				gc.pathInfo[id] = &graphPathInfo{
					NodeAliases: pathNodeAliases,
					EdgeAliases: pathEdgeAliases,
				}
			}
		}
	}
	return b.String(), nil
}

// lastBoundAlias returns the most recently stored alias on the
// graphElementContext (set by GraphNodeScanNode after it emits
// its table reference).
func lastBoundAlias(ctx context.Context) string {
	gc := graphCtxFromContext(ctx)
	if gc == nil {
		return ""
	}
	// Find the highest column id in columnAlias — that's the most
	// recently registered scan.
	var bestAlias string
	var bestID int32 = -1
	for id, a := range gc.columnAlias {
		if id > bestID {
			bestID = id
			bestAlias = a
		}
	}
	return bestAlias
}

func lastBoundMeta(ctx context.Context) *graphElementMetadata {
	gc := graphCtxFromContext(ctx)
	if gc == nil {
		return &graphElementMetadata{}
	}
	var bestID int32 = -1
	var bestElem googlesql.GraphElementTableNode
	for id, e := range gc.elementTable {
		if id > bestID {
			bestID = id
			bestElem = e
		}
	}
	return graphElementMetadataFor(bestElem)
}

// formatEdgeScan emits the SQL for a ResolvedGraphEdgeScan: an
// underlying table aliased uniquely. Returns the SQL fragment,
// the alias, the edge metadata, and the path orientation.
func formatEdgeScan(ctx context.Context, n *googlesql.ResolvedGraphEdgeScan) (string, string, *graphElementMetadata, googlesql.ResolvedGraphEdgeScanEnums_EdgeOrientation, error) {
	targets, err := n.TargetElementTableList()
	if err != nil {
		return "", "", nil, 0, err
	}
	if len(targets) == 0 {
		return "", "", nil, 0, fmt.Errorf("graph edge scan with no target element table")
	}
	if len(targets) > 1 {
		return "", "", nil, 0, fmt.Errorf("graph edge scan over %d alternative element tables not yet supported", len(targets))
	}
	tbl := targets[0]
	underlying, err := tbl.GetTable()
	if err != nil || underlying == nil {
		return "", "", nil, 0, fmt.Errorf("graph edge table has no underlying SQL table")
	}
	tableName, err := underlying.Name()
	if err != nil {
		return "", "", nil, 0, err
	}
	colList, err := n.ColumnList()
	if err != nil {
		return "", "", nil, 0, err
	}
	orient, _ := n.Orientation()
	meta := graphElementMetadataFor(tbl)
	alias := fmt.Sprintf("__g%d", nextGraphAliasID(ctx))
	out := fmt.Sprintf("`%s` AS `%s`", tableName, alias)
	if gc := graphCtxFromContext(ctx); gc != nil {
		for _, col := range colList {
			id, _ := col.ColumnId()
			gc.columnAlias[id] = alias
			gc.elementTable[id] = tbl
		}
	}
	return out, alias, meta, orient, nil
}

// nodeEdgeJoinCondition assembles the ON condition that links a
// node alias to an edge alias (or vice versa). When closing the
// chain (forwardFromNode=false), the node sits on the
// destination side of the edge; when opening (forwardFromNode=
// true), it sits on the source side. Both source and destination
// FK columns are read from the edge metadata.
func nodeEdgeJoinCondition(
	nodeAlias string,
	nodeMeta *graphElementMetadata,
	edgeAlias string,
	edgeMeta *graphElementMetadata,
	orient googlesql.ResolvedGraphEdgeScanEnums_EdgeOrientation,
	forwardFromNode bool,
) (string, error) {
	if nodeMeta == nil || edgeMeta == nil {
		return "1=1", nil
	}
	// Edge orientation determines which side the previous node sits
	// on. For our current single-direction implementation
	// (orientation=RIGHT), the previous node is the source side and
	// the next node is the destination side.
	useSource := forwardFromNode
	if orient == googlesql.ResolvedGraphEdgeScanEnums_EdgeOrientationLeft {
		useSource = !forwardFromNode
	}
	var fk, nodeRef []string
	if useSource {
		fk = edgeMeta.SrcFKCols
		nodeRef = edgeMeta.SrcRefCols
	} else {
		fk = edgeMeta.DstFKCols
		nodeRef = edgeMeta.DstRefCols
	}
	if len(fk) != len(nodeRef) || len(fk) == 0 {
		return "1=1", nil
	}
	parts := make([]string, 0, len(fk))
	for i := range fk {
		parts = append(parts, fmt.Sprintf("`%s`.`%s` = `%s`.`%s`", nodeAlias, nodeRef[i], edgeAlias, fk[i]))
	}
	return strings.Join(parts, " AND "), nil
}

// formatPathScalarFunc lowers PATH_LENGTH / IS_TRAIL / IS_ACYCLIC
// over a path column to constant or aggregate-style SQL.
func formatPathScalarFunc(ctx context.Context, name string, n *googlesql.ResolvedFunctionCall) (string, bool, error) {
	args, err := n.ArgumentList()
	if err != nil || len(args) != 1 {
		return "", false, nil
	}
	colRef, ok := args[0].(*googlesql.ResolvedColumnRef)
	if !ok {
		return "", false, nil
	}
	col, err := colRef.Column()
	if err != nil {
		return "", false, err
	}
	id, _ := col.ColumnId()
	gc := graphCtxFromContext(ctx)
	if gc == nil {
		return "NULL", true, nil
	}
	pi, ok := gc.pathInfo[id]
	if !ok || pi == nil {
		return "NULL", true, nil
	}
	switch name {
	case "path_length":
		return fmt.Sprintf("%d", len(pi.EdgeAliases)), true, nil
	case "is_trail":
		// All edges distinct ⇔ no two edge aliases reference the
		// same row. In our static lowering edges are all
		// different aliases, so the predicate equals
		// "every pair has different rowids".
		if len(pi.EdgeAliases) < 2 {
			return "TRUE", true, nil
		}
		parts := []string{}
		for i := 0; i < len(pi.EdgeAliases); i++ {
			for j := i + 1; j < len(pi.EdgeAliases); j++ {
				parts = append(parts, fmt.Sprintf("`%s`.rowid <> `%s`.rowid", pi.EdgeAliases[i], pi.EdgeAliases[j]))
			}
		}
		return "(" + strings.Join(parts, " AND ") + ")", true, nil
	case "is_acyclic":
		// All nodes distinct.
		if len(pi.NodeAliases) < 2 {
			return "TRUE", true, nil
		}
		parts := []string{}
		for i := 0; i < len(pi.NodeAliases); i++ {
			for j := i + 1; j < len(pi.NodeAliases); j++ {
				parts = append(parts, fmt.Sprintf("`%s`.rowid <> `%s`.rowid", pi.NodeAliases[i], pi.NodeAliases[j]))
			}
		}
		return "(" + strings.Join(parts, " AND ") + ")", true, nil
	case "nodes", "path_nodes":
		return graphElementIDArrayLiteral(pi.NodeAliases), true, nil
	case "edges", "path_edges":
		return graphElementIDArrayLiteral(pi.EdgeAliases), true, nil
	case "path_first":
		if len(pi.NodeAliases) == 0 {
			return "NULL", true, nil
		}
		return fmt.Sprintf("CAST(`%s`.rowid AS STRING)", pi.NodeAliases[0]), true, nil
	case "path_last":
		if len(pi.NodeAliases) == 0 {
			return "NULL", true, nil
		}
		return fmt.Sprintf("CAST(`%s`.rowid AS STRING)", pi.NodeAliases[len(pi.NodeAliases)-1]), true, nil
	}
	return "", false, nil
}

// must0 returns the first return value of a (T, error) pair,
// dropping the error. Used at format time where the error path
// is rare and we already have a fallback.
func must0[T any](v T, _ error) T { return v }

// graphElementIDArrayLiteral builds a SQL expression that
// produces an ARRAY<STRING> of ELEMENT_ID-shaped values for the
// given alias list. Uses the `googlesqlite_make_array` runtime
// UDF so the rowids can be looked up at execution time and
// returned as a single ARRAY value.
func graphElementIDArrayLiteral(aliases []string) string {
	if len(aliases) == 0 {
		return "googlesqlite_make_array()"
	}
	parts := make([]string, 0, len(aliases))
	for _, a := range aliases {
		parts = append(parts, fmt.Sprintf("CAST(`%s`.rowid AS STRING)", a))
	}
	return "googlesqlite_make_array(" + strings.Join(parts, ", ") + ")"
}

// nextGraphAliasID returns a per-query monotonic alias suffix so
// edges and additional nodes don't collide with each other.
func nextGraphAliasID(ctx context.Context) int {
	gc := graphCtxFromContext(ctx)
	if gc == nil {
		return 0
	}
	max := 0
	for _, a := range gc.columnAlias {
		var n int
		fmt.Sscanf(a, "__g%d", &n)
		if n > max {
			max = n
		}
	}
	return max + 1
}

// FormatSQL for a node scan: emit the underlying SQL table the
// element table points at, aliased so the formatter can later
// resolve property accesses against it.
func (n *GraphNodeScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	targets, err := n.node.TargetElementTableList()
	if err != nil {
		return "", err
	}
	if len(targets) == 0 {
		return "", fmt.Errorf("graph node scan with no target element table")
	}
	if len(targets) > 1 {
		return "", fmt.Errorf("graph node scan over %d alternative element tables not yet supported", len(targets))
	}
	tbl := targets[0]
	underlying, err := tbl.GetTable()
	if err != nil || underlying == nil {
		return "", fmt.Errorf("graph element table has no underlying SQL table")
	}
	tableName, err := underlying.Name()
	if err != nil {
		return "", err
	}
	// The scan exposes one column of graph-element type. Track it
	// so subsequent GraphGetElementProperty lookups know which
	// alias to dereference.
	colList, err := n.node.ColumnList()
	if err != nil {
		return "", err
	}
	alias := fmt.Sprintf("__g%d", nextGraphAliasID(ctx))
	out := fmt.Sprintf("`%s` AS `%s`", tableName, alias)
	if gc := graphCtxFromContext(ctx); gc != nil {
		for _, col := range colList {
			id, _ := col.ColumnId()
			gc.columnAlias[id] = alias
			gc.elementTable[id] = tbl
		}
		gc.lastTableRef = out
	}
	return out, nil
}

func (n *GraphEdgeScanNode) FormatSQL(ctx context.Context) (string, error) {
	// Edge scans are formatted inline by GraphPathScanNode (see
	// formatEdgeScan) so the JOIN ON condition can be built
	// alongside the table reference. Reaching here means an edge
	// scan appeared outside a path scan, which the analyzer
	// shouldn't produce.
	return "", fmt.Errorf("graph edge scan must appear inside a path scan")
}

// FormatSQL — resolve an element-property access into an alias.col
// reference. Looks up the alias the enclosing scan registered for
// the graph-element column, then translates the property name to
// the underlying column.
func (n *GraphGetElementPropertyNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	gc := graphCtxFromContext(ctx)
	if gc == nil {
		return "", fmt.Errorf("graph property access outside a graph table scan")
	}
	inner, err := n.node.Expr()
	if err != nil {
		return "", err
	}
	colRef, ok := inner.(*googlesql.ResolvedColumnRef)
	if !ok {
		return "", fmt.Errorf("graph property access only supports direct column refs (got %T)", inner)
	}
	col, err := colRef.Column()
	if err != nil {
		return "", err
	}
	id, _ := col.ColumnId()
	alias, ok := gc.columnAlias[id]
	if !ok {
		return "", fmt.Errorf("graph element column %d not bound", id)
	}
	elemTable, ok := gc.elementTable[id]
	if !ok {
		return "", fmt.Errorf("graph element table for column %d not bound", id)
	}
	decl, err := n.node.Property()
	if err != nil || decl == nil {
		return "", fmt.Errorf("graph get-property has no declaration")
	}
	propName, err := decl.Name()
	if err != nil {
		return "", err
	}
	// For column-projection properties (PROPERTIES(col1, col2, ...))
	// the property name matches an underlying column. Look it up
	// on the element table; if the lookup succeeds and the value
	// expression is just an identifier, we can shorthand. For
	// anything more complex (e.g. MEASURE expressions) emit a
	// best-effort `alias.propname` and rely on the underlying
	// column existing under the same name.
	_, _ = elemTable.FindPropertyDefinitionByName(propName)
	_ = elemTable
	return fmt.Sprintf("`%s`.`%s`", alias, propName), nil
}

// FormatSQL — a GraphRefScan re-uses the table reference set
// up by a previous graph scan stage. Returning the last bound
// table ref lets enclosing ProjectScan / FilterScan stages
// produce a valid FROM clause.
func (n *GraphRefScanNode) FormatSQL(ctx context.Context) (string, error) {
	if gc := graphCtxFromContext(ctx); gc != nil {
		return gc.lastTableRef, nil
	}
	return "", nil
}

// FormatSQL — a GraphScan holds the InputScanList (path scans).
// For single-pattern MATCH the list has one entry; cross-pattern
// MATCH would require a CROSS JOIN which is not yet implemented.
func (n *GraphScanNode) FormatSQL(ctx context.Context) (string, error) {
	if n.node == nil {
		return "", nil
	}
	paths, err := n.node.InputScanList()
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "(SELECT NULL)", nil
	}
	if len(paths) > 1 {
		return "", fmt.Errorf("graph scan with %d comma-separated patterns not yet supported", len(paths))
	}
	return newNode(paths[0]).FormatSQL(ctx)
}

// tryFormatMeasureAGG handles AGG(<graph_get_element_property>) by
// looking up the measure property's underlying SQL expression on
// the graph element table, extracting the outer aggregate kind
// (SUM / AVG / MIN / MAX / COUNT / COUNT_DISTINCT / ANY_VALUE) and
// the inner column expression, then emitting
// `googlesqlite_agg(<inner>, <locking_key>, '<kind>')`. Returns
// (sql, true, nil) when handled; (_, false, nil) when not
// applicable so callers can fall through to the default formatter.
func tryFormatMeasureAGG(ctx context.Context, n *googlesql.ResolvedAggregateFunctionCall) (string, bool, error) {
	if n == nil {
		return "", false, nil
	}
	fn, err := n.Function()
	if err != nil || fn == nil {
		return "", false, nil
	}
	full, _ := fn.FullName(false)
	name := strings.ToLower(full)
	// FullName(false) returns "<group>:<name>" form (e.g.
	// "GoogleSQL:AGG") or just the bare name; strip any prefix
	// before colon.
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = name[idx+1:]
	}
	if name != "agg" {
		return "", false, nil
	}
	args, err := n.ArgumentList()
	if err != nil || len(args) != 1 {
		return "", false, nil
	}
	prop, ok := args[0].(*googlesql.ResolvedGraphGetElementProperty)
	if !ok {
		return "", false, nil
	}
	gc := graphCtxFromContext(ctx)
	if gc == nil {
		return "", false, fmt.Errorf("AGG over graph property outside a graph table scan")
	}
	inner, err := prop.Expr()
	if err != nil {
		return "", false, err
	}
	colRef, ok := inner.(*googlesql.ResolvedColumnRef)
	if !ok {
		return "", false, fmt.Errorf("AGG over graph property: expected direct element column ref (got %T)", inner)
	}
	col, err := colRef.Column()
	if err != nil {
		return "", false, err
	}
	id, _ := col.ColumnId()
	alias, ok := gc.columnAlias[id]
	if !ok {
		return "", false, fmt.Errorf("AGG: graph element column %d not bound to a SQL alias", id)
	}
	elemTbl, ok := gc.elementTable[id]
	if !ok {
		return "", false, fmt.Errorf("AGG: graph element table for column %d not bound", id)
	}
	decl, err := prop.Property()
	if err != nil || decl == nil {
		return "", false, fmt.Errorf("AGG: graph property declaration missing")
	}
	propName, _ := decl.Name()
	def, err := elemTbl.FindPropertyDefinitionByName(propName)
	if err != nil || def == nil {
		return "", false, fmt.Errorf("AGG: property definition for %q not found on element table", propName)
	}
	exprSQL, err := def.ExpressionSql()
	if err != nil {
		return "", false, err
	}
	kind, innerExpr, perr := parseMeasureSQL(exprSQL)
	if perr != nil {
		return "", false, fmt.Errorf("AGG: %w", perr)
	}
	// Locking key: project the element table's primary key.
	keyExpr, err := keySQLForElement(elemTbl, alias)
	if err != nil {
		return "", false, err
	}
	// Inner expression: column references inside `innerExpr` are
	// against the underlying SQLite table, so prefix bare
	// identifiers with the alias. The simple parser only supports
	// a single column / "*", so wrap it directly.
	innerSQL := innerExpr
	if innerSQL != "*" {
		innerSQL = fmt.Sprintf("`%s`.`%s`", alias, innerExpr)
	} else {
		innerSQL = "1"
	}
	// Encode the kind discriminator using the driver's STRING-literal
	// layout (base64-encoded ValueLayout) so the runtime aggregator's
	// arg decoder can roundtrip it like any other string value.
	kindLit, err := literalFromValue(value.StringValue(kind))
	if err != nil {
		return "", false, err
	}
	return fmt.Sprintf("googlesqlite_agg(%s, %s, %s)", innerSQL, keyExpr, kindLit), true, nil
}

// parseMeasureSQL extracts the outer aggregate kind and the inner
// argument from a simple measure expression like "SUM(population)"
// or "COUNT(*)". Returns ("ANY_VALUE", "*", nil) for unrecognised
// shapes so AGG still produces a result via its fallback path.
func parseMeasureSQL(s string) (kind, inner string, err error) {
	t := strings.TrimSpace(s)
	open := strings.Index(t, "(")
	if open < 0 || !strings.HasSuffix(t, ")") {
		return "ANY_VALUE", "*", nil
	}
	kind = strings.ToUpper(strings.TrimSpace(t[:open]))
	inner = strings.TrimSpace(t[open+1 : len(t)-1])
	if inner == "" {
		inner = "*"
	}
	if strings.HasPrefix(strings.ToLower(inner), "distinct ") {
		inner = strings.TrimSpace(inner[len("distinct "):])
		if kind == "COUNT" {
			kind = "COUNT_DISTINCT"
		}
	}
	switch kind {
	case "SUM", "AVG", "MIN", "MAX", "COUNT", "COUNT_DISTINCT", "ANY_VALUE":
		// ok
	default:
		kind = "ANY_VALUE"
	}
	return kind, inner, nil
}

// keySQLForElement projects a row-unique key for the underlying
// SQLite table behind a graph element. SQLite's `rowid` is the
// simplest such key — it's the integer primary key alias unique
// per row, exactly what the AGG locking-key dedup needs.
func keySQLForElement(_ googlesql.GraphElementTableNode, alias string) (string, error) {
	return fmt.Sprintf("`%s`.rowid", alias), nil
}

// tryFormatGraphElementFunc lowers the Spanner graph-element
// built-ins to constant SQL expressions derived from the static
// graph metadata captured in graphElementContext. Returns
// (sql, true, nil) when handled; (_, false, nil) when not
// applicable.
//
// Supported:
//
//	LABELS(elem)                  → ARRAY<STRING> of label names
//	ELEMENT_ID(elem)              → STRING (alias-qualified rowid)
//	ELEMENT_DEFINITION_NAME(elem) → STRING (element-table alias)
//	PROPERTY_NAMES(elem)          → ARRAY<STRING> of property names
func tryFormatGraphElementFunc(ctx context.Context, n *googlesql.ResolvedFunctionCall) (string, bool, error) {
	if n == nil {
		return "", false, nil
	}
	fn, err := n.Function()
	if err != nil || fn == nil {
		return "", false, nil
	}
	full, _ := fn.FullName(false)
	name := strings.ToLower(full)
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = name[idx+1:]
	}
	switch name {
	case "labels", "element_id", "element_definition_name", "property_names",
		"source_node_id", "destination_node_id":
		// continue below
	case "path_length", "is_trail", "is_acyclic",
		"nodes", "edges", "path_first", "path_last",
		"path_nodes", "path_edges":
		return formatPathScalarFunc(ctx, name, n)
	default:
		return "", false, nil
	}
	args, err := n.ArgumentList()
	if err != nil || len(args) != 1 {
		return "", false, nil
	}
	// Allow ELEMENT_ID(PATH_FIRST(p)) and similar wrappers: when
	// the argument is itself a PATH_FIRST/PATH_LAST call, format
	// the inner call directly (both produce the same element-id
	// shape in our lowering).
	if inner, ok := args[0].(*googlesql.ResolvedFunctionCall); ok && name == "element_id" {
		if fn2, _ := inner.Function(); fn2 != nil {
			innerName := strings.ToLower(must0(fn2.FullName(false)))
			if i := strings.Index(innerName, ":"); i >= 0 {
				innerName = innerName[i+1:]
			}
			if innerName == "path_first" || innerName == "path_last" {
				return formatPathScalarFunc(ctx, innerName, inner)
			}
		}
	}
	colRef, ok := args[0].(*googlesql.ResolvedColumnRef)
	if !ok {
		return "", false, nil
	}
	col, err := colRef.Column()
	if err != nil {
		return "", false, err
	}
	id, _ := col.ColumnId()
	gc := graphCtxFromContext(ctx)
	if gc == nil {
		return "", false, nil
	}
	alias, ok := gc.columnAlias[id]
	if !ok {
		return "", false, nil
	}
	elemTbl, ok := gc.elementTable[id]
	if !ok {
		return "", false, nil
	}
	// Source the static metadata via our own driver-side storage.
	// We keep label names + property names on each element table at
	// SimplePropertyGraph construction time (see property_graph.go),
	// because the binding's slice-write-back APIs on
	// GraphElementTableNode are not usable from Go.
	meta := graphElementMetadataFor(elemTbl)
	switch name {
	case "labels":
		return graphStringArrayLiteral(meta.Labels), true, nil
	case "element_id":
		return fmt.Sprintf("CAST(`%s`.rowid AS STRING)", alias), true, nil
	case "element_definition_name":
		lit, _ := literalFromValue(value.StringValue(meta.Name))
		return lit, true, nil
	case "property_names":
		return graphStringArrayLiteral(meta.PropertyNames), true, nil
	case "source_node_id":
		if !meta.IsEdge || len(meta.SrcFKCols) == 0 {
			return "NULL", true, nil
		}
		return fmt.Sprintf("CAST(`%s`.`%s` AS STRING)", alias, meta.SrcFKCols[0]), true, nil
	case "destination_node_id":
		if !meta.IsEdge || len(meta.DstFKCols) == 0 {
			return "NULL", true, nil
		}
		return fmt.Sprintf("CAST(`%s`.`%s` AS STRING)", alias, meta.DstFKCols[0]), true, nil
	}
	return "", false, nil
}

// graphStringArrayLiteral encodes the given STRING slice as the
// driver's ARRAY<STRING> literal form (base64-wrapped layout) so
// SQLite passes it through unchanged and the row decoder reads
// it back as an ARRAY value.
func graphStringArrayLiteral(items []string) string {
	values := make([]value.Value, 0, len(items))
	for _, it := range items {
		values = append(values, value.StringValue(it))
	}
	arr := &value.ArrayValue{Values: values}
	lit, err := literalFromValue(arr)
	if err != nil {
		return "NULL"
	}
	return lit
}

// graphElementMetadata is the Go-side snapshot of an element
// table's labels, properties, keys, and edge references. We
// populate it at SimplePropertyGraph construction time because
// the Go binding exposes a slice-write-back API on
// GraphElementTableNode that can't actually return data through
// wasmify.
type graphElementMetadata struct {
	Name          string
	Labels        []string
	PropertyNames []string
	SQLTable      string   // underlying SQL table name
	KeyCols       []string // primary-key column names on the SQL table

	// Edge-only fields.
	IsEdge      bool
	SrcRefTable string   // node-table name referenced by source FK
	DstRefTable string   // node-table name referenced by destination FK
	SrcFKCols   []string // edge-side FK column names (source)
	DstFKCols   []string // edge-side FK column names (destination)
	SrcRefCols  []string // node-side PK column names (source)
	DstRefCols  []string // node-side PK column names (destination)
}

var graphElementMetadataStore sync.Map // key: string (element-table name), value: *graphElementMetadata

// elementKey extracts the element-table's catalog name. This is
// stable across wrapper instances the analyzer may produce.
func elementKey(elem googlesql.GraphElementTableNode) string {
	if elem == nil {
		return ""
	}
	if n, err := elem.Name(); err == nil {
		return n
	}
	return ""
}

// registerGraphElementMetadata associates element-table metadata
// with the given element-table handle. Called from
// property_graph.go after the SimpleGraphNodeTable /
// SimpleGraphEdgeTable is built.
func registerGraphElementMetadata(elem googlesql.GraphElementTableNode, meta *graphElementMetadata) {
	if elem == nil || meta == nil {
		return
	}
	if k := elementKey(elem); k != "" {
		graphElementMetadataStore.Store(k, meta)
	}
}

// graphElementMetadataFor returns the previously-registered
// metadata for `elem`, or an empty record when none is found.
func graphElementMetadataFor(elem googlesql.GraphElementTableNode) *graphElementMetadata {
	if elem == nil {
		return &graphElementMetadata{}
	}
	if v, ok := graphElementMetadataStore.Load(elementKey(elem)); ok {
		return v.(*graphElementMetadata)
	}
	return &graphElementMetadata{}
}
