package internal

import (
	"context"

	googlesql "github.com/goccy/go-googlesql"
)

// cteInlineBudget bounds how much SQL text SQLite may copy when it
// inlines single-reference CTEs. SQLite's query flattener substitutes
// a subquery's result-column expression into every place the outer
// query references that column, so a CTE whose columns are large
// expressions (a long CASE, say) and whose consumer reads them many
// times turns into body size x reference count of parse tree. Chains
// of such CTEs multiply. The dbt snowplow page_views/sessions models
// (one ~700 KB CTE body feeding a channel CASE that reads mkt_source
// hundreds of times) needed more than 2 GiB of SQLite heap to prepare
// and failed with "sqlite3: out of memory" on the 256 MiB wasm heap.
//
// Above the budget the CTE is emitted as MATERIALIZED, which SQLite
// never flattens. 4 MiB is about 1/64 of the default heap; CTEs that
// are referenced once and are small or lightly used stay inlined so
// SQLite can still push predicates into them.
const cteInlineBudget = 4 << 20

// cteInlineCost estimates the parse-tree cost of inlining the entries
// of one WITH clause. refsIn[consumer][cte] counts the column
// references that the consumer (an entry name, or "" for the main
// query) makes to the cte's output columns. expanded holds, for each
// entry already decided to stay inlined, its formatted size plus the
// copies of earlier inlined entries it pulls in.
type cteInlineCost struct {
	refsIn   map[string]map[string]int
	expanded map[string]int
}

func newCteInlineCost(with *googlesql.ResolvedWithScan) *cteInlineCost {
	c := &cteInlineCost{
		refsIn:   map[string]map[string]int{},
		expanded: map[string]int{},
	}
	owner := map[int32]string{}
	query, _ := with.Query()
	collectWithRefColumns(query, owner)
	entries, _ := with.WithEntryList()
	for _, entry := range entries {
		sub, _ := entry.WithSubquery()
		collectWithRefColumns(sub, owner)
	}
	if len(owner) == 0 {
		return c
	}
	c.refsIn[""] = map[string]int{}
	countWithRefColumnRefs(query, owner, c.refsIn[""])
	for _, entry := range entries {
		name, _ := entry.WithQueryName()
		sub, _ := entry.WithSubquery()
		counts := map[string]int{}
		countWithRefColumnRefs(sub, owner, counts)
		c.refsIn[name] = counts
	}
	return c
}

// materialize reports whether the entry `name`, whose body formats to
// bodyLen bytes, would cost more than cteInlineBudget to inline. An
// entry that stays inlined is remembered so that later entries reading
// it account for the copies they pull in. Entries must be decided in
// WITH order.
func (c *cteInlineCost) materialize(name string, bodyLen int) bool {
	if c == nil {
		return false
	}
	expanded := bodyLen
	for dep, n := range c.refsIn[name] {
		if e, ok := c.expanded[dep]; ok {
			expanded += e * n
		}
		if expanded > cteInlineBudget {
			return true
		}
	}
	refs := 0
	for _, counts := range c.refsIn {
		refs += counts[name]
	}
	if refs > 0 && expanded > cteInlineBudget/refs {
		return true
	}
	c.expanded[name] = expanded
	return false
}

// collectWithRefColumns maps every column produced by a
// ResolvedWithRefScan in the subtree to the CTE it reads.
func collectWithRefColumns(node googlesql.ResolvedNode, owner map[int32]string) {
	if node == nil {
		return
	}
	if ref, ok := node.(*googlesql.ResolvedWithRefScan); ok {
		name, _ := ref.WithQueryName()
		cols, _ := ref.ColumnList()
		for _, col := range cols {
			if id, err := col.ColumnId(); err == nil {
				owner[id] = name
			}
		}
	}
	children, _ := node.GetChildNodes()
	for _, child := range children {
		collectWithRefColumns(child, owner)
	}
}

// countWithRefColumnRefs counts, per CTE, the column references in the
// subtree that read a column produced by a reference to that CTE.
func countWithRefColumnRefs(node googlesql.ResolvedNode, owner map[int32]string, counts map[string]int) {
	if node == nil {
		return
	}
	if ref, ok := node.(*googlesql.ResolvedColumnRef); ok {
		if col, _ := ref.Column(); col != nil {
			if id, err := col.ColumnId(); err == nil {
				if name, ok := owner[id]; ok {
					counts[name]++
				}
			}
		}
	}
	children, _ := node.GetChildNodes()
	for _, child := range children {
		countWithRefColumnRefs(child, owner, counts)
	}
}

type cteInlineCostKey struct{}

func withCteInlineCost(ctx context.Context, c *cteInlineCost) context.Context {
	return context.WithValue(ctx, cteInlineCostKey{}, c)
}

func cteInlineCostFrom(ctx context.Context) *cteInlineCost {
	c, _ := ctx.Value(cteInlineCostKey{}).(*cteInlineCost)
	return c
}
