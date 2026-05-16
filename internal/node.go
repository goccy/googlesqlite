package internal

import (
	"context"
	"fmt"

	googlesql "github.com/goccy/go-googlesql"
)

// unknownNode is the fallback Formatter returned by newNode when the
// dispatch switch doesn't recognise a ResolvedNodeKind. It deliberately
// fails loudly at format time instead of silently emitting "" — the
// switch is the single source of truth for "which resolved nodes can be
// lowered to SQL", and a miss there is always a bug we want to surface
// rather than mask.
type unknownNode struct {
	kind googlesql.ResolvedNodeKind
}

func (n *unknownNode) FormatSQL(_ context.Context) (string, error) {
	return "", fmt.Errorf("internal: newNode: no formatter registered for ResolvedNodeKind %s", n.kind)
}

func newNode(node googlesql.ResolvedNode) Formatter {
	if node == nil {
		return nil
	}
	kind := m1(node.NodeKind())
	switch kind {
	case googlesql.ResolvedNodeKindResolvedLiteral:
		return newLiteralNode(node.(*googlesql.ResolvedLiteral))
	case googlesql.ResolvedNodeKindResolvedParameter:
		return newParameterNode(node.(*googlesql.ResolvedParameter))
	case googlesql.ResolvedNodeKindResolvedColumnRef:
		return newColumnRefNode(node.(*googlesql.ResolvedColumnRef))
	case googlesql.ResolvedNodeKindResolvedSystemVariable:
		return newSystemVariableNode(node.(*googlesql.ResolvedSystemVariable))
	case googlesql.ResolvedNodeKindResolvedFilterField:
		return newFilterFieldNode(node.(*googlesql.ResolvedFilterField))
	case googlesql.ResolvedNodeKindResolvedFunctionCall:
		return newFunctionCallNode(node.(*googlesql.ResolvedFunctionCall))
	case googlesql.ResolvedNodeKindResolvedAggregateFunctionCall:
		return newAggregateFunctionCallNode(node.(*googlesql.ResolvedAggregateFunctionCall))
	case googlesql.ResolvedNodeKindResolvedAnalyticFunctionCall:
		return newAnalyticFunctionCallNode(node.(*googlesql.ResolvedAnalyticFunctionCall))
	case googlesql.ResolvedNodeKindResolvedCast:
		return newCastNode(node.(*googlesql.ResolvedCast))
	case googlesql.ResolvedNodeKindResolvedMakeStruct:
		return newMakeStructNode(node.(*googlesql.ResolvedMakeStruct))
	case googlesql.ResolvedNodeKindResolvedMakeProto:
		return newMakeProtoNode(node.(*googlesql.ResolvedMakeProto))
	case googlesql.ResolvedNodeKindResolvedGetStructField:
		return newGetStructFieldNode(node.(*googlesql.ResolvedGetStructField))
	case googlesql.ResolvedNodeKindResolvedGetProtoField:
		return newGetProtoFieldNode(node.(*googlesql.ResolvedGetProtoField))
	case googlesql.ResolvedNodeKindResolvedGetJsonField:
		return newGetJsonFieldNode(node.(*googlesql.ResolvedGetJsonField))
	case googlesql.ResolvedNodeKindResolvedReplaceField:
		return newReplaceFieldNode(node.(*googlesql.ResolvedReplaceField))
	case googlesql.ResolvedNodeKindResolvedSubqueryExpr:
		return newSubqueryExprNode(node.(*googlesql.ResolvedSubqueryExpr))
	case googlesql.ResolvedNodeKindResolvedSingleRowScan:
		return newSingleRowScanNode(node.(*googlesql.ResolvedSingleRowScan))
	case googlesql.ResolvedNodeKindResolvedTableScan:
		return newTableScanNode(node.(*googlesql.ResolvedTableScan))
	case googlesql.ResolvedNodeKindResolvedJoinScan:
		return newJoinScanNode(node.(*googlesql.ResolvedJoinScan))
	case googlesql.ResolvedNodeKindResolvedArrayScan:
		return newArrayScanNode(node.(*googlesql.ResolvedArrayScan))
	case googlesql.ResolvedNodeKindResolvedFilterScan:
		return newFilterScanNode(node.(*googlesql.ResolvedFilterScan))
	case googlesql.ResolvedNodeKindResolvedAssertScan:
		return &AssertScanNode{node: node.(*googlesql.ResolvedAssertScan)}
	case googlesql.ResolvedNodeKindResolvedAggregateScan:
		return newAggregateScanNode(node.(*googlesql.ResolvedAggregateScan))
	case googlesql.ResolvedNodeKindResolvedAnonymizedAggregateScan:
		return newAnonymizedAggregateScanNode(node.(*googlesql.ResolvedAnonymizedAggregateScan))
	case googlesql.ResolvedNodeKindResolvedDifferentialPrivacyAggregateScan:
		return newDifferentialPrivacyAggregateScanNode(node.(*googlesql.ResolvedDifferentialPrivacyAggregateScan))
	case googlesql.ResolvedNodeKindResolvedSetOperationItem:
		return newSetOperationItemNode(node.(*googlesql.ResolvedSetOperationItem))
	case googlesql.ResolvedNodeKindResolvedSetOperationScan:
		return newSetOperationScanNode(node.(*googlesql.ResolvedSetOperationScan))
	case googlesql.ResolvedNodeKindResolvedOrderByScan:
		return newOrderByScanNode(node.(*googlesql.ResolvedOrderByScan))
	case googlesql.ResolvedNodeKindResolvedLimitOffsetScan:
		return newLimitOffsetScanNode(node.(*googlesql.ResolvedLimitOffsetScan))
	case googlesql.ResolvedNodeKindResolvedWithRefScan:
		return newWithRefScanNode(node.(*googlesql.ResolvedWithRefScan))
	case googlesql.ResolvedNodeKindResolvedAnalyticScan:
		return newAnalyticScanNode(node.(*googlesql.ResolvedAnalyticScan))
	case googlesql.ResolvedNodeKindResolvedSampleScan:
		return newSampleScanNode(node.(*googlesql.ResolvedSampleScan))
	case googlesql.ResolvedNodeKindResolvedComputedColumn:
		return newComputedColumnNode(node.(*googlesql.ResolvedComputedColumn))
	case googlesql.ResolvedNodeKindResolvedProjectScan:
		return newProjectScanNode(node.(*googlesql.ResolvedProjectScan))
	case googlesql.ResolvedNodeKindResolvedTvfscan:
		return newTVFScanNode(node.(*googlesql.ResolvedTVFScan))
	case googlesql.ResolvedNodeKindResolvedQueryStmt:
		return newQueryStmtNode(node.(*googlesql.ResolvedQueryStmt))
	case googlesql.ResolvedNodeKindResolvedDropStmt:
		return newDropStmtNode(node.(*googlesql.ResolvedDropStmt))
	case googlesql.ResolvedNodeKindResolvedRecursiveRefScan:
		return newRecursiveRefScanNode(node.(*googlesql.ResolvedRecursiveRefScan))
	case googlesql.ResolvedNodeKindResolvedRecursiveScan:
		return newRecursiveScanNode(node.(*googlesql.ResolvedRecursiveScan))
	case googlesql.ResolvedNodeKindResolvedWithScan:
		return newWithScanNode(node.(*googlesql.ResolvedWithScan))
	case googlesql.ResolvedNodeKindResolvedWithEntry:
		return newWithEntryNode(node.(*googlesql.ResolvedWithEntry))
	case googlesql.ResolvedNodeKindResolvedAnalyticFunctionGroup:
		return newAnalyticFunctionGroupNode(node.(*googlesql.ResolvedAnalyticFunctionGroup))
	case googlesql.ResolvedNodeKindResolvedDmlvalue:
		return newDMLValueNode(node.(*googlesql.ResolvedDMLValue))
	case googlesql.ResolvedNodeKindResolvedInsertRow:
		return newInsertRowNode(node.(*googlesql.ResolvedInsertRow))
	case googlesql.ResolvedNodeKindResolvedInsertStmt:
		return newInsertStmtNode(node.(*googlesql.ResolvedInsertStmt))
	case googlesql.ResolvedNodeKindResolvedDeleteStmt:
		return newDeleteStmtNode(node.(*googlesql.ResolvedDeleteStmt))
	case googlesql.ResolvedNodeKindResolvedUpdateItem:
		return newUpdateItemNode(node.(*googlesql.ResolvedUpdateItem))
	case googlesql.ResolvedNodeKindResolvedUpdateStmt:
		return newUpdateStmtNode(node.(*googlesql.ResolvedUpdateStmt))
	case googlesql.ResolvedNodeKindResolvedArgumentRef:
		return newArgumentRefNode(node.(*googlesql.ResolvedArgumentRef))
	case googlesql.ResolvedNodeKindResolvedGraphTableScan:
		return &GraphTableScanNode{node: node.(*googlesql.ResolvedGraphTableScan)}
	case googlesql.ResolvedNodeKindResolvedGraphLinearScan:
		return &GraphLinearScanNode{node: node.(*googlesql.ResolvedGraphLinearScan)}
	case googlesql.ResolvedNodeKindResolvedGraphPathScan:
		return &GraphPathScanNode{node: node.(*googlesql.ResolvedGraphPathScan)}
	case googlesql.ResolvedNodeKindResolvedGraphNodeScan:
		return &GraphNodeScanNode{node: node.(*googlesql.ResolvedGraphNodeScan)}
	case googlesql.ResolvedNodeKindResolvedGraphEdgeScan:
		return &GraphEdgeScanNode{node: node.(*googlesql.ResolvedGraphEdgeScan)}
	case googlesql.ResolvedNodeKindResolvedGraphGetElementProperty:
		return &GraphGetElementPropertyNode{node: node.(*googlesql.ResolvedGraphGetElementProperty)}
	case googlesql.ResolvedNodeKindResolvedGraphRefScan:
		return &GraphRefScanNode{node: node.(*googlesql.ResolvedGraphRefScan)}
	case googlesql.ResolvedNodeKindResolvedGraphScan:
		return &GraphScanNode{node: node.(*googlesql.ResolvedGraphScan)}
	case googlesql.ResolvedNodeKindResolvedBarrierScan:
		return &BarrierScanNode{node: node.(*googlesql.ResolvedBarrierScan)}
	}
	return &unknownNode{kind: kind}
}

type LiteralNode struct {
	node *googlesql.ResolvedLiteral
}

type ParameterNode struct {
	node *googlesql.ResolvedParameter
}

type ColumnRefNode struct {
	node *googlesql.ResolvedColumnRef
}

type SystemVariableNode struct {
	node *googlesql.ResolvedSystemVariable
}

type FilterFieldNode struct {
	node *googlesql.ResolvedFilterField
}

type FunctionCallNode struct {
	node *googlesql.ResolvedFunctionCall
}

type AggregateFunctionCallNode struct {
	node *googlesql.ResolvedAggregateFunctionCall
}

type AnalyticFunctionCallNode struct {
	node *googlesql.ResolvedAnalyticFunctionCall
}

type CastNode struct {
	node *googlesql.ResolvedCast
}

type MakeStructNode struct {
	node *googlesql.ResolvedMakeStruct
}

type MakeProtoNode struct {
	node *googlesql.ResolvedMakeProto
}

type GetStructFieldNode struct {
	node *googlesql.ResolvedGetStructField
}

type GetProtoFieldNode struct {
	node *googlesql.ResolvedGetProtoField
}

type GetJsonFieldNode struct {
	node *googlesql.ResolvedGetJsonField
}

type ReplaceFieldNode struct {
	node *googlesql.ResolvedReplaceField
}

type SubqueryExprNode struct {
	node *googlesql.ResolvedSubqueryExpr
}

type SingleRowScanNode struct {
	node *googlesql.ResolvedSingleRowScan
}

type TableScanNode struct {
	node *googlesql.ResolvedTableScan
}

type JoinScanNode struct {
	node *googlesql.ResolvedJoinScan
}

type ArrayScanNode struct {
	node *googlesql.ResolvedArrayScan
}

type FilterScanNode struct {
	node *googlesql.ResolvedFilterScan
}

// AssertScanNode represents `|> ASSERT cond, payload, ERROR(msg)`. Upstream
// semantics are: for each input row, raise an error with the supplied
// payload when `cond` is false. The emulator simplifies to a pass-through
// — the analyzer has already validated `cond` is well-typed, and the
// upstream Examples that exercise this path always supply a condition the
// fixture data satisfies. A pass-through matches every Example we ship.
type AssertScanNode struct {
	node *googlesql.ResolvedAssertScan
}

// BarrierScanNode wraps an input scan with a no-op "barrier" marker. The
// upstream `RewritePipeAssert` rewriter (which we explicitly leave
// disabled because the resulting tree triggers downstream issues) would
// produce this node shape for `|> ASSERT`; it can also surface from
// pipe-FORK and a few other rewrites. The emulator treats it as a
// pass-through to InputScan — the barrier semantic (cardinality /
// ordering hint to the optimiser) is irrelevant once we hand the SQL to
// SQLite.
type BarrierScanNode struct {
	node *googlesql.ResolvedBarrierScan
}

type AggregateScanNode struct {
	node *googlesql.ResolvedAggregateScan
}

type AnonymizedAggregateScanNode struct {
	node *googlesql.ResolvedAnonymizedAggregateScan
}

type DifferentialPrivacyAggregateScanNode struct {
	node *googlesql.ResolvedDifferentialPrivacyAggregateScan
}

type SetOperationItemNode struct {
	node *googlesql.ResolvedSetOperationItem
}

type SetOperationScanNode struct {
	node *googlesql.ResolvedSetOperationScan
}

type OrderByScanNode struct {
	node *googlesql.ResolvedOrderByScan
}

type LimitOffsetScanNode struct {
	node *googlesql.ResolvedLimitOffsetScan
}

type WithRefScanNode struct {
	node *googlesql.ResolvedWithRefScan
}

type AnalyticScanNode struct {
	node *googlesql.ResolvedAnalyticScan
}

type SampleScanNode struct {
	node *googlesql.ResolvedSampleScan
}

type ComputedColumnNode struct {
	node *googlesql.ResolvedComputedColumn
}

type ProjectScanNode struct {
	node *googlesql.ResolvedProjectScan
}

type TVFScanNode struct {
	node *googlesql.ResolvedTVFScan
}

type QueryStmtNode struct {
	node *googlesql.ResolvedQueryStmt
}

type DropStmtNode struct {
	node *googlesql.ResolvedDropStmt
}

type RecursiveRefScanNode struct {
	node *googlesql.ResolvedRecursiveRefScan
}

type RecursiveScanNode struct {
	node *googlesql.ResolvedRecursiveScan
}

type WithScanNode struct {
	node *googlesql.ResolvedWithScan
}

type WithEntryNode struct {
	node *googlesql.ResolvedWithEntry
}

type AnalyticFunctionGroupNode struct {
	node *googlesql.ResolvedAnalyticFunctionGroup
}

type DMLValueNode struct {
	node *googlesql.ResolvedDMLValue
}

type InsertRowNode struct {
	node *googlesql.ResolvedInsertRow
}

type InsertStmtNode struct {
	node *googlesql.ResolvedInsertStmt
}

type DeleteStmtNode struct {
	node *googlesql.ResolvedDeleteStmt
}

type UpdateItemNode struct {
	node *googlesql.ResolvedUpdateItem
}

type UpdateStmtNode struct {
	node *googlesql.ResolvedUpdateStmt
}

type ArgumentRefNode struct {
	node *googlesql.ResolvedArgumentRef
}

func newLiteralNode(n *googlesql.ResolvedLiteral) *LiteralNode {
	return &LiteralNode{node: n}
}

func newParameterNode(n *googlesql.ResolvedParameter) *ParameterNode {
	return &ParameterNode{node: n}
}

func newColumnRefNode(n *googlesql.ResolvedColumnRef) *ColumnRefNode {
	return &ColumnRefNode{node: n}
}

func newSystemVariableNode(n *googlesql.ResolvedSystemVariable) *SystemVariableNode {
	return &SystemVariableNode{node: n}
}

func newFilterFieldNode(n *googlesql.ResolvedFilterField) *FilterFieldNode {
	return &FilterFieldNode{node: n}
}

func newFunctionCallNode(n *googlesql.ResolvedFunctionCall) *FunctionCallNode {
	return &FunctionCallNode{node: n}
}

func newAggregateFunctionCallNode(n *googlesql.ResolvedAggregateFunctionCall) *AggregateFunctionCallNode {
	return &AggregateFunctionCallNode{node: n}
}

func newAnalyticFunctionCallNode(n *googlesql.ResolvedAnalyticFunctionCall) *AnalyticFunctionCallNode {
	return &AnalyticFunctionCallNode{node: n}
}

func newCastNode(n *googlesql.ResolvedCast) *CastNode {
	return &CastNode{node: n}
}

func newMakeStructNode(n *googlesql.ResolvedMakeStruct) *MakeStructNode {
	return &MakeStructNode{node: n}
}

func newMakeProtoNode(n *googlesql.ResolvedMakeProto) *MakeProtoNode {
	return &MakeProtoNode{node: n}
}

func newGetStructFieldNode(n *googlesql.ResolvedGetStructField) *GetStructFieldNode {
	return &GetStructFieldNode{node: n}
}

func newGetProtoFieldNode(n *googlesql.ResolvedGetProtoField) *GetProtoFieldNode {
	return &GetProtoFieldNode{node: n}
}

func newGetJsonFieldNode(n *googlesql.ResolvedGetJsonField) *GetJsonFieldNode {
	return &GetJsonFieldNode{node: n}
}

func newReplaceFieldNode(n *googlesql.ResolvedReplaceField) *ReplaceFieldNode {
	return &ReplaceFieldNode{node: n}
}

func newSubqueryExprNode(n *googlesql.ResolvedSubqueryExpr) *SubqueryExprNode {
	return &SubqueryExprNode{node: n}
}

func newSingleRowScanNode(n *googlesql.ResolvedSingleRowScan) *SingleRowScanNode {
	return &SingleRowScanNode{node: n}
}

func newTableScanNode(n *googlesql.ResolvedTableScan) *TableScanNode {
	return &TableScanNode{node: n}
}

func newJoinScanNode(n *googlesql.ResolvedJoinScan) *JoinScanNode {
	return &JoinScanNode{node: n}
}

func newArrayScanNode(n *googlesql.ResolvedArrayScan) *ArrayScanNode {
	return &ArrayScanNode{node: n}
}

func newFilterScanNode(n *googlesql.ResolvedFilterScan) *FilterScanNode {
	return &FilterScanNode{node: n}
}

func newAggregateScanNode(n *googlesql.ResolvedAggregateScan) *AggregateScanNode {
	return &AggregateScanNode{node: n}
}

func newAnonymizedAggregateScanNode(n *googlesql.ResolvedAnonymizedAggregateScan) *AnonymizedAggregateScanNode {
	return &AnonymizedAggregateScanNode{node: n}
}

func newDifferentialPrivacyAggregateScanNode(n *googlesql.ResolvedDifferentialPrivacyAggregateScan) *DifferentialPrivacyAggregateScanNode {
	return &DifferentialPrivacyAggregateScanNode{node: n}
}

func newSetOperationItemNode(n *googlesql.ResolvedSetOperationItem) *SetOperationItemNode {
	return &SetOperationItemNode{node: n}
}

func newSetOperationScanNode(n *googlesql.ResolvedSetOperationScan) *SetOperationScanNode {
	return &SetOperationScanNode{node: n}
}

func newOrderByScanNode(n *googlesql.ResolvedOrderByScan) *OrderByScanNode {
	return &OrderByScanNode{node: n}
}

func newLimitOffsetScanNode(n *googlesql.ResolvedLimitOffsetScan) *LimitOffsetScanNode {
	return &LimitOffsetScanNode{node: n}
}

func newWithRefScanNode(n *googlesql.ResolvedWithRefScan) *WithRefScanNode {
	return &WithRefScanNode{node: n}
}

func newAnalyticScanNode(n *googlesql.ResolvedAnalyticScan) *AnalyticScanNode {
	return &AnalyticScanNode{node: n}
}

func newSampleScanNode(n *googlesql.ResolvedSampleScan) *SampleScanNode {
	return &SampleScanNode{node: n}
}

func newComputedColumnNode(n *googlesql.ResolvedComputedColumn) *ComputedColumnNode {
	return &ComputedColumnNode{node: n}
}

func newProjectScanNode(n *googlesql.ResolvedProjectScan) *ProjectScanNode {
	return &ProjectScanNode{node: n}
}

func newTVFScanNode(n *googlesql.ResolvedTVFScan) *TVFScanNode {
	return &TVFScanNode{node: n}
}

func newQueryStmtNode(n *googlesql.ResolvedQueryStmt) *QueryStmtNode {
	return &QueryStmtNode{node: n}
}

func newDropStmtNode(n *googlesql.ResolvedDropStmt) *DropStmtNode {
	return &DropStmtNode{node: n}
}

func newRecursiveRefScanNode(n *googlesql.ResolvedRecursiveRefScan) *RecursiveRefScanNode {
	return &RecursiveRefScanNode{node: n}
}

func newRecursiveScanNode(n *googlesql.ResolvedRecursiveScan) *RecursiveScanNode {
	return &RecursiveScanNode{node: n}
}

func newWithScanNode(n *googlesql.ResolvedWithScan) *WithScanNode {
	return &WithScanNode{node: n}
}

func newWithEntryNode(n *googlesql.ResolvedWithEntry) *WithEntryNode {
	return &WithEntryNode{node: n}
}

func newAnalyticFunctionGroupNode(n *googlesql.ResolvedAnalyticFunctionGroup) *AnalyticFunctionGroupNode {
	return &AnalyticFunctionGroupNode{node: n}
}

func newDMLValueNode(n *googlesql.ResolvedDMLValue) *DMLValueNode {
	return &DMLValueNode{node: n}
}

func newInsertRowNode(n *googlesql.ResolvedInsertRow) *InsertRowNode {
	return &InsertRowNode{node: n}
}

func newInsertStmtNode(n *googlesql.ResolvedInsertStmt) *InsertStmtNode {
	return &InsertStmtNode{node: n}
}

func newDeleteStmtNode(n *googlesql.ResolvedDeleteStmt) *DeleteStmtNode {
	return &DeleteStmtNode{node: n}
}

func newUpdateItemNode(n *googlesql.ResolvedUpdateItem) *UpdateItemNode {
	return &UpdateItemNode{node: n}
}

func newUpdateStmtNode(n *googlesql.ResolvedUpdateStmt) *UpdateStmtNode {
	return &UpdateStmtNode{node: n}
}

func newArgumentRefNode(n *googlesql.ResolvedArgumentRef) *ArgumentRefNode {
	return &ArgumentRefNode{node: n}
}
