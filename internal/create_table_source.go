package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	googlesql "github.com/goccy/go-googlesql"
)

// This file implements the source-table half of the `CREATE TABLE`
// family that builds a table out of another table rather than out of a
// column list:
//
//	CREATE TABLE <name> LIKE <source>
//	CREATE TABLE <name> {COPY | CLONE} <source> [WHERE <expr>]
//	CREATE SNAPSHOT TABLE <name> CLONE <source>
//
// The grammar and the resolved-AST shape are recorded upstream in
// docs/third_party/googlesql-docs/resolved_ast.md under
// ResolvedCreateTableStmt and ResolvedCreateSnapshotTableStmt.
//
// Two upstream statements drive the design here:
//
//   - "By default, all fields (column names, types, constraints,
//     partition, clustering, options etc.) will be inherited from the
//     source table." Inheritance is specified behaviour, not a nicety.
//     The analyzer does not do it for us: its materialised column list
//     for LIKE carries names and types only, dropping NOT NULL,
//     DEFAULT and PRIMARY KEY. We recover those from the source
//     table's stored TableSpec.
//
//   - "The 'clone_from.column_list' field may be set, but should be
//     ignored." So the created table's schema comes from the source
//     spec. The scan's column list is still what the formatter names
//     its output columns after, so it stays the basis for the
//     data-copy projection — but never for the schema.

// createTableCloneScan returns the copy_from / clone_from scan of a
// CREATE TABLE statement, or nil when the statement is neither a COPY
// nor a CLONE. Upstream allows at most one of the two to be set.
func createTableCloneScan(node *googlesql.ResolvedCreateTableStmt) (googlesql.ResolvedScanNode, error) {
	copyFrom, err := node.CopyFrom()
	if err != nil {
		return nil, fmt.Errorf("failed to read the CREATE TABLE COPY source: %w", err)
	}
	if copyFrom != nil {
		return copyFrom, nil
	}
	cloneFrom, err := node.CloneFrom()
	if err != nil {
		return nil, fmt.Errorf("failed to read the CREATE TABLE CLONE source: %w", err)
	}
	return cloneFrom, nil
}

// resolveSourceTableScan unwraps the scan that upstream places under
// copy_from / clone_from and returns the source TableScan.
//
// Upstream pins the shape: the source is a ResolvedTableScan, wrapped
// in exactly one ResolvedFilterScan when the statement carries a WHERE
// clause, and "no other Scan types are allowed here".
func resolveSourceTableScan(scan googlesql.ResolvedScanNode) (*googlesql.ResolvedTableScan, error) {
	switch s := scan.(type) {
	case *googlesql.ResolvedTableScan:
		return s, nil
	case *googlesql.ResolvedFilterScan:
		input, err := s.InputScan()
		if err != nil {
			return nil, fmt.Errorf("failed to read the source scan of a WHERE-filtered clone: %w", err)
		}
		table, ok := input.(*googlesql.ResolvedTableScan)
		if !ok {
			return nil, fmt.Errorf("unsupported clone source: expected a table scan under the WHERE clause, got %T", input)
		}
		return table, nil
	default:
		return nil, fmt.Errorf("unsupported clone source scan %T", scan)
	}
}

// checkNoSystemTime rejects `FOR SYSTEM_TIME AS OF`. The analyzer keeps
// the expression on the source TableScan, and the formatter has no way
// to honour it: googlesqlite stores only the current state of a table,
// so a point-in-time read would silently return present-day rows.
// Plain queries already reject the clause (its own LanguageFeature is
// left off), and this keeps the CREATE TABLE path consistent with them
// instead of quietly answering the wrong question.
func checkNoSystemTime(scan *googlesql.ResolvedTableScan) error {
	expr, err := scan.ForSystemTimeExpr()
	if err != nil {
		return fmt.Errorf("failed to read FOR SYSTEM_TIME AS OF: %w", err)
	}
	if expr != nil {
		return fmt.Errorf("FOR SYSTEM_TIME AS OF is not supported")
	}
	return nil
}

// sourceTableSpec resolves a catalog table handle to the stored spec
// for that table. The second result is false when the table has no
// stored spec — INFORMATION_SCHEMA views and wildcard tables are
// registered directly in the analyzer catalog and never reach
// tableMap.
//
// The two callers treat that miss differently, deliberately. LIKE still
// has the analyzer's materialised column list to fall back on, so it
// creates the table without the inherited constraints. COPY / CLONE
// have nothing to fall back on — the whole schema comes from the spec —
// so they reject the statement.
func sourceTableSpec(catalog *Catalog, table googlesql.TableNode) (*TableSpec, bool) {
	if catalog == nil || table == nil {
		return nil, false
	}
	name := catalog.StorageNameForTable(table)
	if name == "" {
		return nil, false
	}
	return catalog.TableSpec(name)
}

// inheritColumnConstraints copies NOT NULL and DEFAULT from the source
// table's columns onto columns of the same name in dst.
//
// dst holds what the analyzer materialised, which is authoritative for
// which columns exist and what type each one has; src is authoritative
// for the constraints the analyzer dropped. Columns absent from src are
// left as-is.
//
// DefaultExpr is copied verbatim and deliberately not interpreted: it
// holds the driver's internal encoded value form, not SQL text.
func inheritColumnConstraints(dst []*ColumnSpec, src *TableSpec) {
	if src == nil {
		return
	}
	for _, col := range dst {
		srcCol := src.Column(col.Name)
		if srcCol == nil {
			continue
		}
		col.IsNotNull = srcCol.IsNotNull
		col.DefaultExpr = srcCol.DefaultExpr
	}
}

// copyColumnSpecs deep-copies a column list so an inherited spec never
// aliases the source table's own ColumnSpec pointers — the catalog
// hands out live pointers, and a later ALTER on either table would
// otherwise mutate both.
func copyColumnSpecs(cols []*ColumnSpec) []*ColumnSpec {
	out := make([]*ColumnSpec, 0, len(cols))
	for _, col := range cols {
		dup := *col
		out = append(out, &dup)
	}
	return out
}

// mergeTableOptions applies upstream's option-inheritance rule for
// COPY / CLONE: "If the OPTIONS clause is explicitly specified, the
// option values are intended to be used for the created or replaced
// table. If any OPTION is unspecified, the corresponding option from
// the source table will be used instead."
//
// Explicit options win by name; everything else falls through from the
// source. Source order is preserved so INFORMATION_SCHEMA.TABLE_OPTIONS
// reads back stably.
func mergeTableOptions(src, explicit []*tableOptionSpec) []*tableOptionSpec {
	if len(src) == 0 && len(explicit) == 0 {
		return nil
	}
	byName := make(map[string]*tableOptionSpec, len(explicit))
	for _, opt := range explicit {
		byName[opt.Name] = opt
	}
	out := make([]*tableOptionSpec, 0, len(src)+len(explicit))
	seen := make(map[string]bool, len(src))
	for _, opt := range src {
		seen[opt.Name] = true
		if override, ok := byName[opt.Name]; ok {
			out = append(out, override)
			continue
		}
		dup := *opt
		out = append(out, &dup)
	}
	for _, opt := range explicit {
		if !seen[opt.Name] {
			out = append(out, opt)
		}
	}
	return out
}

// newClonedTableSpec builds the target TableSpec for COPY / CLONE and
// for CREATE SNAPSHOT TABLE. Unlike LIKE, these statements give the
// analyzer no column list at all — `column_definition_list` comes back
// empty — so the whole schema is inherited from the source spec.
func newClonedTableSpec(
	src *TableSpec,
	namePath []string,
	isTemp bool,
	createMode googlesql.ResolvedCreateStatementEnums_CreateMode,
	explicitOptions []*tableOptionSpec,
) *TableSpec {
	now := time.Now()
	return &TableSpec{
		IsTemp:     isTemp,
		NamePath:   namePath,
		Columns:    copyColumnSpecs(src.Columns),
		PrimaryKey: append([]string(nil), src.PrimaryKey...),
		CreateMode: createMode,
		Options:    mergeTableOptions(src.Options, explicitOptions),
		UpdatedAt:  now,
		CreatedAt:  now,
	}
}

// skipCreateIntoExistingTable reports whether a CREATE TABLE statement
// must do nothing because its target is already there.
//
// Upstream: "`IF NOT EXISTS`: If any table exists with the same name,
// the `CREATE` statement will have no effect."
//
// Handing the clause straight to SQLite covers the CREATE but not the
// rest of the statement. The catalog write that follows would still
// overwrite the existing table's stored schema, so INFORMATION_SCHEMA
// would start describing columns the table does not have; and for
// COPY / CLONE the row copy is a separate INSERT that SQLite's own
// IF NOT EXISTS cannot guard at all, appending a second set of rows on
// every re-run.
//
// TEMP tables are exempt: they carry deliberate replace-on-recreate
// semantics (see CreateTableStmtAction.exec), which the caller applies
// before this check matters. CREATE OR REPLACE needs no guard either —
// it drops the target first.
func skipCreateIntoExistingTable(ctx context.Context, conn *Conn, spec *TableSpec) (bool, error) {
	if spec.CreateMode != googlesql.ResolvedCreateStatementEnums_CreateModeCreateIfNotExists || spec.IsTemp {
		return false, nil
	}
	var name string
	err := conn.QueryRowContext(
		ctx,
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?",
		spec.TableName(),
	).Scan(&name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("failed to check whether %s already exists: %w", spec.TableName(), err)
	}
	return true, nil
}

// execCloneData runs the INSERT that fills a freshly created COPY /
// CLONE target, and undoes the CREATE if that INSERT fails.
//
// The two statements are not atomic, and the statement may already be
// inside a caller-owned transaction that we must not roll back. Without
// the cleanup a failed copy would leave a table that SQLite knows about
// but the catalog does not: `SELECT` from it reports "Table not found"
// while re-creating it reports "table already exists", wedging the name.
// By the time this runs the table is always one we just created —
// IF NOT EXISTS returns earlier, OR REPLACE has dropped the old one, and
// a plain CREATE would have failed against an existing name — so
// dropping it on failure cannot destroy a pre-existing table.
func execCloneData(ctx context.Context, conn *Conn, spec *TableSpec, cloneData string, args []any, query string) error {
	insert := fmt.Sprintf("INSERT INTO `%s` %s", spec.TableName(), cloneData)
	if _, err := conn.ExecContext(ctx, insert, args...); err != nil {
		if _, dropErr := conn.ExecContext(
			ctx,
			fmt.Sprintf("DROP TABLE IF EXISTS `%s`", spec.TableName()),
		); dropErr != nil {
			return fmt.Errorf(
				"failed to copy rows for %s: %w (and rolling back the created table failed: %v)",
				query, err, dropErr,
			)
		}
		return fmt.Errorf("failed to copy rows for %s: %w", query, err)
	}
	return nil
}

// newCloneDataQuery renders the SELECT that materialises a COPY /
// CLONE source into the freshly created table, plus the INSERT column
// list it feeds.
//
// The scan is formatted with useColumnID set, matching every other
// multi-column statement path, so the subquery exposes `name#id`
// columns. The outer projection maps those back onto the source column
// names, which are also the target's — building the INSERT column list
// from the scan (rather than from the target spec) keeps the two sides
// aligned by construction.
func newCloneDataQuery(ctx context.Context, scan googlesql.ResolvedScanNode) (string, []*googlesql.ResolvedParameter, error) {
	ctx = withUseColumnID(ctx)
	scanSQL, params, err := collectFormatParams(ctx, scan)
	if err != nil {
		return "", nil, err
	}
	input, err := formatInput(scanSQL)
	if err != nil {
		return "", nil, err
	}
	cols, err := scan.MutableColumnList()
	if err != nil {
		return "", nil, fmt.Errorf("failed to read the clone source columns: %w", err)
	}
	targets := make([]string, 0, len(cols))
	projection := make([]string, 0, len(cols))
	for _, col := range cols {
		name, err := col.Name()
		if err != nil {
			return "", nil, fmt.Errorf("failed to read a clone source column name: %w", err)
		}
		targets = append(targets, fmt.Sprintf("`%s`", name))
		projection = append(projection, fmt.Sprintf("`%s`", uniqueColumnName(ctx, col)))
	}
	if len(targets) == 0 {
		return "", nil, fmt.Errorf("clone source exposes no columns")
	}
	return fmt.Sprintf(
		"(%s) SELECT %s %s",
		strings.Join(targets, ","),
		strings.Join(projection, ","),
		input,
	), params, nil
}
