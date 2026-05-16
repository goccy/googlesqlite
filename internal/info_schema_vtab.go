package internal

import (
	"fmt"
	"strings"

	sqlite3 "github.com/ncruces/go-sqlite3"

	"github.com/goccy/googlesqlite/internal/value"
)

// RegisterInfoSchemaModules registers one SQLite virtual-table
// module per BigQuery INFORMATION_SCHEMA view on the given
// connection. Each module is bound to the supplied catalog so the
// vtab serves rows from whatever the catalog currently contains —
// crucial for file-backed catalogs where state survives across
// process restarts.
func RegisterInfoSchemaModules(conn *sqlite3.Conn, catalog *Catalog) error {
	for _, view := range infoSchemaViews {
		v := view
		if err := sqlite3.CreateModule[*infoSchemaTable](
			conn, v.Module,
			nil, // create: nil => eponymous-only, no CREATE VIRTUAL TABLE needed
			func(_ *sqlite3.Conn, _, _, _ string, _ ...string) (*infoSchemaTable, error) {
				return newInfoSchemaTable(conn, v, catalog)
			},
		); err != nil {
			return fmt.Errorf("googlesqlite: register info-schema module %s: %w", v.Module, err)
		}
	}
	return nil
}

// infoSchemaTable is the per-connection virtual table for one
// INFORMATION_SCHEMA view. The trailing hidden `__schema` column
// carries the dataset filter that the formatter pushes down via
// `WHERE __schema = '<dataset>'`.
type infoSchemaTable struct {
	view    *infoSchemaView
	catalog *Catalog
}

func newInfoSchemaTable(conn *sqlite3.Conn, view *infoSchemaView, catalog *Catalog) (*infoSchemaTable, error) {
	colDecls := make([]string, 0, len(view.Columns)+1)
	for _, c := range view.Columns {
		colDecls = append(colDecls, fmt.Sprintf("`%s` %s", c.Name, c.Type))
	}
	colDecls = append(colDecls, "`__schema` TEXT HIDDEN")
	declSQL := fmt.Sprintf("CREATE TABLE x (%s)", strings.Join(colDecls, ", "))
	if err := conn.DeclareVTab(declSQL); err != nil {
		return nil, err
	}
	return &infoSchemaTable{view: view, catalog: catalog}, nil
}

// schemaConstraintColumn is the index (0-based) of the hidden
// `__schema` column within the declared vtab schema.
func (t *infoSchemaTable) schemaConstraintColumn() int {
	return len(t.view.Columns)
}

// BestIndex tells SQLite that an `__schema = ?` equality constraint
// is the cheapest path. The Filter call receives that value as
// arg[0]; xBestIndex is also where we set the cost so SQLite picks
// the constrained scan over a full scan.
func (t *infoSchemaTable) BestIndex(idx *sqlite3.IndexInfo) error {
	schemaCol := t.schemaConstraintColumn()
	idx.EstimatedCost = 1e6 // unconstrained scan: pessimistic
	idx.EstimatedRows = 1000
	for i, c := range idx.Constraint {
		if !c.Usable {
			continue
		}
		if c.Column == schemaCol && c.Op == sqlite3.INDEX_CONSTRAINT_EQ {
			idx.ConstraintUsage[i].ArgvIndex = 1
			idx.ConstraintUsage[i].Omit = true
			idx.IdxNum = 1
			idx.EstimatedCost = 1
			idx.EstimatedRows = 100
		}
	}
	return nil
}

func (t *infoSchemaTable) Open() (sqlite3.VTabCursor, error) {
	return &infoSchemaCursor{table: t}, nil
}

// infoSchemaCursor walks a per-Filter snapshot of rows. The slice
// is built fresh on every Filter() call so subsequent table
// mutations are visible on the next scan; we never cache rows
// across cursors.
type infoSchemaCursor struct {
	table *infoSchemaTable
	rows  []map[string]value.Value
	pos   int
}

func (c *infoSchemaCursor) Filter(_ int, _ string, args ...sqlite3.Value) error {
	c.pos = 0
	schema := ""
	if len(args) > 0 && args[0].Type() != sqlite3.NULL {
		schema = args[0].Text()
	}
	c.table.catalog.mu.Lock()
	rows := c.table.view.BuildRows(c.table.catalog, schema)
	c.table.catalog.mu.Unlock()
	c.rows = rows
	return nil
}

func (c *infoSchemaCursor) Next() error {
	c.pos++
	return nil
}

func (c *infoSchemaCursor) EOF() bool {
	return c.pos >= len(c.rows)
}

func (c *infoSchemaCursor) Column(ctx sqlite3.Context, n int) error {
	if c.pos >= len(c.rows) {
		ctx.ResultNull()
		return nil
	}
	if n >= len(c.table.view.Columns) {
		// Hidden __schema column: SQLite never asks for it (Omit
		// flag in BestIndex), but return NULL defensively.
		ctx.ResultNull()
		return nil
	}
	colName := c.table.view.Columns[n].Name
	v, ok := c.rows[c.pos][colName]
	if !ok || v == nil {
		ctx.ResultNull()
		return nil
	}
	// IntValue / FloatValue / BoolValue ride SQLite-native because
	// the rest of the pipeline (collation, scanner) treats those
	// primitives as raw — see internal/encoder.go:LiteralFromValue.
	// Everything else MUST be envelope-encoded so the
	// googlesqlite_collate collation and the row scanner can decode
	// it back the same way they handle values from regular tables.
	switch vv := v.(type) {
	case value.IntValue:
		ctx.ResultInt64(int64(vv))
		return nil
	case value.FloatValue:
		ctx.ResultFloat(float64(vv))
		return nil
	case value.BoolValue:
		if vv {
			ctx.ResultInt64(1)
		} else {
			ctx.ResultInt64(0)
		}
		return nil
	}
	encoded, err := EncodeValue(v)
	if err != nil {
		return err
	}
	switch enc := encoded.(type) {
	case string:
		ctx.ResultText(enc)
	case nil:
		ctx.ResultNull()
	default:
		// EncodeValue passes through bool/int64/float64 unchanged
		// for primitive value types, but those are handled in the
		// switch above. Anything that lands here is unexpected.
		return fmt.Errorf("googlesqlite info-schema: unexpected encoded type %T", enc)
	}
	return nil
}

func (c *infoSchemaCursor) RowID() (int64, error) {
	return int64(c.pos), nil
}
