package internal

import (
	"fmt"

	googlesql "github.com/goccy/go-googlesql"
)

// buildSimplePropertyGraph turns a resolved CREATE PROPERTY GRAPH
// statement into a go-googlesql SimplePropertyGraph, ready to be
// added to the analyzer catalog. Node-table aliases become the
// graph element names; key column ordinals are derived from the
// resolved KeyList against the underlying TableColumnList. Edge
// tables carry source / destination references to the previously
// registered node tables.
//
// Property definitions and dynamic labels are NOT propagated in
// this minimal pass — graph queries that only need to reach the
// underlying tables (e.g. `MATCH (l:L) RETURN l.population`
// using the implicit alias label) still resolve. MEASURE /
// expression-property plumbing is the next follow-up.
func buildSimplePropertyGraph(stmt *googlesql.ResolvedCreatePropertyGraphStmt) (*googlesql.SimplePropertyGraph, error) {
	namePath, err := stmt.NamePath()
	if err != nil {
		return nil, err
	}

	// Build property declarations first — shared across element
	// tables that expose the same property name. NewSimplePropertyGraph2
	// consumes them at the end.
	propDeclByName := map[string]*googlesql.SimpleGraphPropertyDeclaration{}
	makePropDecl := func(name string, t googlesql.Googlesql_TypeNode) (*googlesql.SimpleGraphPropertyDeclaration, error) {
		if d, ok := propDeclByName[name]; ok {
			return d, nil
		}
		d, err := googlesql.NewSimpleGraphPropertyDeclaration(name, namePath, t)
		if err != nil {
			return nil, err
		}
		propDeclByName[name] = d
		return d, nil
	}

	// Build labels next. Each label name appears at most once;
	// the graph constructor will own them. Element tables receive
	// the same label handles via setKeepAlive (no ownership
	// transfer at element-table creation time).
	labelByName := map[string]*googlesql.SimpleGraphElementLabel{}
	makeLabel := func(name string) (*googlesql.SimpleGraphElementLabel, error) {
		if l, ok := labelByName[name]; ok {
			return l, nil
		}
		l, err := googlesql.NewSimpleGraphElementLabel(name, namePath, nil)
		if err != nil {
			return nil, err
		}
		labelByName[name] = l
		return l, nil
	}

	// Walk a ResolvedGraphElementTable's property definitions
	// and return matching SimpleGraphPropertyDefinition handles,
	// creating any missing PropertyDeclarations along the way.
	propsForElement := func(element *googlesql.ResolvedGraphElementTable) ([]googlesql.GraphPropertyDefinitionNode, error) {
		defs, _ := element.PropertyDefinitionList()
		out := make([]googlesql.GraphPropertyDefinitionNode, 0, len(defs))
		for _, d := range defs {
			pname, _ := d.PropertyDeclarationName()
			expr, _ := d.Expr()
			if expr == nil {
				continue
			}
			t, err := expr.Type()
			if err != nil || t == nil {
				continue
			}
			// A MEASURE property's resolved expression has the
			// inner aggregation type; wrap it in MEASURE<T> so the
			// analyzer types `l.<measure_prop>` as MEASURE<T> and
			// AGG accepts it.
			isMeasure, _ := d.IsMeasure()
			if isMeasure {
				// In an ideal world we'd wrap the result type with
				// TypeFactory.MakeMeasureType so `l.<measure_prop>`
				// types as MEASURE<T>. The wasm bridge for that call
				// is buggy in the currently-published googlesql-wasm
				// snapshot (see internal/property_graph.go MEASURE
				// bridge bug note below), so the type stays as the
				// raw underlying T. AGG calls on this column fall
				// back to the agg() runtime registered through
				// function_register_aggregate.go, which is invoked
				// here as a plain aggregate over an INT64 / FLOAT64
				// argument.
				if measureT, merr := tf().MakeMeasureType(t); merr == nil && measureT != nil {
					t = measureT
				}
			}
			decl, err := makePropDecl(pname, t)
			if err != nil {
				return nil, err
			}
			sqlExpr, _ := d.Sql()
			def, err := googlesql.NewSimpleGraphPropertyDefinition(decl, sqlExpr)
			if err != nil {
				return nil, err
			}
			out = append(out, def)
		}
		return out, nil
	}

	labelsForElement := func(element *googlesql.ResolvedGraphElementTable, defaultName string) ([]googlesql.GraphElementLabelNode, error) {
		names, _ := element.LabelNameList()
		if len(names) == 0 && defaultName != "" {
			names = []string{defaultName}
		}
		out := make([]googlesql.GraphElementLabelNode, 0, len(names))
		for _, n := range names {
			l, err := makeLabel(n)
			if err != nil {
				return nil, err
			}
			out = append(out, l)
		}
		return out, nil
	}

	nodeTablesByAlias := map[string]*googlesql.SimpleGraphNodeTable{}
	var nodeTables []googlesql.GraphNodeTableNode
	rawNodeTables, err := stmt.NodeTableList()
	if err != nil {
		return nil, err
	}
	for _, nt := range rawNodeTables {
		alias, _ := nt.Alias()
		inputScan, err := nt.InputScan()
		if err != nil || inputScan == nil {
			continue
		}
		tableScan, ok := inputScan.(*googlesql.ResolvedTableScan)
		if !ok {
			continue
		}
		tbl, err := tableScan.Table()
		if err != nil || tbl == nil {
			continue
		}
		keyCols, err := keyColumnsForResolvedScan(tableScan, nt)
		if err != nil {
			return nil, fmt.Errorf("key columns for node %q: %w", alias, err)
		}
		labels, err := labelsForElement(nt, alias)
		if err != nil {
			return nil, fmt.Errorf("labels for node %q: %w", alias, err)
		}
		props, err := propsForElement(nt)
		if err != nil {
			return nil, fmt.Errorf("properties for node %q: %w", alias, err)
		}
		snt, err := googlesql.NewSimpleGraphNodeTable(
			alias, namePath, tbl, keyCols, labels, props, nil, nil,
		)
		if err != nil {
			return nil, fmt.Errorf("new graph node table %q: %w", alias, err)
		}
		nodeTablesByAlias[alias] = snt
		nodeTables = append(nodeTables, snt)
		meta := collectGraphElementMetadata(alias, nt)
		if tableScan != nil {
			if tbl, _ := tableScan.Table(); tbl != nil {
				meta.SQLTable, _ = tbl.Name()
				meta.KeyCols = primaryKeyColumnNames(tableScan, keyCols)
			}
		}
		registerGraphElementMetadata(snt, meta)
	}

	var edgeTables []googlesql.GraphEdgeTableNode
	rawEdgeTables, err := stmt.EdgeTableList()
	if err != nil {
		return nil, err
	}
	for _, et := range rawEdgeTables {
		alias, _ := et.Alias()
		inputScan, err := et.InputScan()
		if err != nil || inputScan == nil {
			continue
		}
		tableScan, ok := inputScan.(*googlesql.ResolvedTableScan)
		if !ok {
			continue
		}
		tbl, err := tableScan.Table()
		if err != nil || tbl == nil {
			continue
		}
		keyCols, err := keyColumnsForResolvedScan(tableScan, et)
		if err != nil {
			return nil, fmt.Errorf("key columns for edge %q: %w", alias, err)
		}
		srcRef, err := nodeReferenceFromResolved(et, true, nodeTablesByAlias, tableScan)
		if err != nil {
			return nil, fmt.Errorf("source ref for edge %q: %w", alias, err)
		}
		dstRef, err := nodeReferenceFromResolved(et, false, nodeTablesByAlias, tableScan)
		if err != nil {
			return nil, fmt.Errorf("destination ref for edge %q: %w", alias, err)
		}
		labels, err := labelsForElement(et, alias)
		if err != nil {
			return nil, fmt.Errorf("labels for edge %q: %w", alias, err)
		}
		props, err := propsForElement(et)
		if err != nil {
			return nil, fmt.Errorf("properties for edge %q: %w", alias, err)
		}
		set, err := googlesql.NewSimpleGraphEdgeTable(
			alias, namePath, tbl, keyCols, labels, props, srcRef, dstRef, nil, nil,
		)
		if err != nil {
			return nil, fmt.Errorf("new graph edge table %q: %w", alias, err)
		}
		edgeTables = append(edgeTables, set)
		meta := collectGraphElementMetadata(alias, et)
		meta.IsEdge = true
		if tableScan != nil {
			if tbl, _ := tableScan.Table(); tbl != nil {
				meta.SQLTable, _ = tbl.Name()
				meta.KeyCols = primaryKeyColumnNames(tableScan, keyCols)
			}
		}
		fillEdgeRefMetadata(et, meta)
		registerGraphElementMetadata(set, meta)
	}

	allLabels := make([]googlesql.GraphElementLabelNode, 0, len(labelByName))
	for _, l := range labelByName {
		allLabels = append(allLabels, l)
	}
	allPropDecls := make([]googlesql.GraphPropertyDeclarationNode, 0, len(propDeclByName))
	for _, d := range propDeclByName {
		allPropDecls = append(allPropDecls, d)
	}
	_ = nodeTablesByAlias // silence unused-warning in branches that skip edges
	pg, err := googlesql.NewSimplePropertyGraph2(namePath, nodeTables, edgeTables, allLabels, allPropDecls)
	if err != nil {
		return nil, fmt.Errorf("new property graph: %w", err)
	}
	return pg, nil
}

// primaryKeyColumnNames maps ordinal key positions (as returned
// by keyColumnsForResolvedScan) to the underlying SQL column
// names on the table scan. Used to plumb the JOIN keys for
// multi-hop GRAPH MATCH support.
func primaryKeyColumnNames(scan *googlesql.ResolvedTableScan, keyCols []int32) []string {
	tbl, err := scan.Table()
	if err != nil || tbl == nil {
		return nil
	}
	cols, _ := scan.ColumnList()
	idxs, _ := scan.ColumnIndexList()
	out := make([]string, 0, len(keyCols))
	for _, k := range keyCols {
		if int(k) >= len(cols) || int(k) >= len(idxs) {
			continue
		}
		colIdx := int(idxs[k])
		if c, cerr := tbl.GetColumn(int32(colIdx)); cerr == nil && c != nil {
			if name, _ := c.Name(); name != "" {
				out = append(out, name)
			}
		}
	}
	return out
}

// fillEdgeRefMetadata populates the source / destination reference
// fields on `meta` from the edge table's resolved node references.
func fillEdgeRefMetadata(et *googlesql.ResolvedGraphElementTable, meta *graphElementMetadata) {
	srcRef, _ := et.SourceNodeReference()
	dstRef, _ := et.DestNodeReference()
	if srcRef != nil {
		meta.SrcRefTable, _ = srcRef.NodeTableIdentifier()
		meta.SrcFKCols = edgeRefColNames(srcRef.EdgeTableColumnList)
		meta.SrcRefCols = edgeRefColNames(srcRef.NodeTableColumnList)
	}
	if dstRef != nil {
		meta.DstRefTable, _ = dstRef.NodeTableIdentifier()
		meta.DstFKCols = edgeRefColNames(dstRef.EdgeTableColumnList)
		meta.DstRefCols = edgeRefColNames(dstRef.NodeTableColumnList)
	}
}

func edgeRefColNames(list func() ([]googlesql.ResolvedExprNode, error)) []string {
	exprs, _ := list()
	out := make([]string, 0, len(exprs))
	for _, e := range exprs {
		ref, ok := e.(*googlesql.ResolvedColumnRef)
		if !ok {
			continue
		}
		col, _ := ref.Column()
		if col == nil {
			continue
		}
		name, _ := col.Name()
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

// collectGraphElementMetadata snapshots the labels and property
// names declared on the resolved element table so the format-time
// LABELS / PROPERTY_NAMES helpers can return them without going
// through the binding's slice-write-back accessors.
func collectGraphElementMetadata(alias string, e *googlesql.ResolvedGraphElementTable) *graphElementMetadata {
	meta := &graphElementMetadata{Name: alias}
	if labels, err := e.LabelNameList(); err == nil {
		meta.Labels = append(meta.Labels, labels...)
	}
	if meta.Labels == nil && alias != "" {
		meta.Labels = []string{alias}
	}
	if defs, err := e.PropertyDefinitionList(); err == nil {
		for _, d := range defs {
			if pn, perr := d.PropertyDeclarationName(); perr == nil && pn != "" {
				meta.PropertyNames = append(meta.PropertyNames, pn)
			}
		}
	}
	return meta
}

// keyColumnsForResolvedScan resolves the key-column ordinals for
// a graph element table. The resolved KeyList holds expressions
// (typically ResolvedColumnRef) referencing columns of the input
// scan; we map their column IDs back to ordinals in the input
// table's column list.
func keyColumnsForResolvedScan(tableScan *googlesql.ResolvedTableScan, element *googlesql.ResolvedGraphElementTable) ([]int32, error) {
	scanCols, err := tableScan.ColumnList()
	if err != nil {
		return nil, err
	}
	colIDToOrdinal := map[int32]int32{}
	for i, c := range scanCols {
		id, _ := c.ColumnId()
		colIDToOrdinal[id] = int32(i)
	}
	keyExprs, err := element.KeyList()
	if err != nil {
		return nil, err
	}
	out := make([]int32, 0, len(keyExprs))
	for _, e := range keyExprs {
		ref, ok := e.(*googlesql.ResolvedColumnRef)
		if !ok {
			continue
		}
		col, err := ref.Column()
		if err != nil || col == nil {
			continue
		}
		id, _ := col.ColumnId()
		if ord, ok := colIDToOrdinal[id]; ok {
			out = append(out, ord)
		}
	}
	return out, nil
}

// nodeReferenceFromResolved builds a SimpleGraphNodeTableReference
// for either the source (when source=true) or destination side of
// an edge.
func nodeReferenceFromResolved(
	edge *googlesql.ResolvedGraphElementTable,
	source bool,
	nodeTables map[string]*googlesql.SimpleGraphNodeTable,
	edgeScan *googlesql.ResolvedTableScan,
) (*googlesql.SimpleGraphNodeTableReference, error) {
	var ref *googlesql.ResolvedGraphNodeTableReference
	var err error
	if source {
		ref, err = edge.SourceNodeReference()
	} else {
		ref, err = edge.DestNodeReference()
	}
	if err != nil || ref == nil {
		return nil, nil
	}
	edgeColRefs, err := ref.EdgeTableColumnList()
	if err != nil {
		return nil, err
	}
	edgeScanCols, _ := edgeScan.ColumnList()
	edgeColIDToOrdinal := map[int32]int32{}
	for i, c := range edgeScanCols {
		id, _ := c.ColumnId()
		edgeColIDToOrdinal[id] = int32(i)
	}
	edgeCols := make([]int32, 0, len(edgeColRefs))
	for _, e := range edgeColRefs {
		colRef, ok := e.(*googlesql.ResolvedColumnRef)
		if !ok {
			continue
		}
		col, err := colRef.Column()
		if err != nil || col == nil {
			continue
		}
		id, _ := col.ColumnId()
		if ord, ok := edgeColIDToOrdinal[id]; ok {
			edgeCols = append(edgeCols, ord)
		}
	}
	nodeColRefs, err := ref.NodeTableColumnList()
	if err != nil {
		return nil, err
	}
	nodeCols := make([]int32, 0, len(nodeColRefs))
	for _, e := range nodeColRefs {
		colRef, ok := e.(*googlesql.ResolvedColumnRef)
		if !ok {
			continue
		}
		col, err := colRef.Column()
		if err != nil || col == nil {
			continue
		}
		id, _ := col.ColumnId()
		nodeCols = append(nodeCols, id)
	}
	refName, err := ref.NodeTableIdentifier()
	if err != nil {
		return nil, err
	}
	tbl := nodeTables[refName]
	if tbl == nil {
		return nil, fmt.Errorf("referenced node table %q not registered", refName)
	}
	return googlesql.NewSimpleGraphNodeTableReference(tbl, edgeCols, nodeCols)
}
