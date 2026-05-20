package internal

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"

	googlesql "github.com/goccy/go-googlesql"

	"github.com/goccy/googlesqlite/internal/exportdata"
	"github.com/goccy/googlesqlite/internal/value"
)

type StmtAction interface {
	Prepare(context.Context, *Conn) (driver.Stmt, error)
	ExecContext(context.Context, *Conn) (driver.Result, error)
	QueryContext(context.Context, *Conn) (*Rows, error)
	Cleanup(context.Context, *Conn) error
	Args() []any
}

type CreateTableStmtAction struct {
	query           string
	args            []any
	spec            *TableSpec
	catalog         *Catalog
	isAutoIndexMode bool
}

func (a *CreateTableStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	if a.spec.CreateMode == googlesql.ResolvedCreateStatementEnums_CreateModeCreateOrReplace {
		if _, err := conn.ExecContext(
			ctx,
			fmt.Sprintf("DROP TABLE IF EXISTS `%s`", a.spec.TableName()),
		); err != nil {
			return nil, err
		}
	}
	stmt, err := conn.PrepareContext(ctx, a.spec.SQLiteSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare %s: %w", a.query, err)
	}
	return newCreateTableStmt(stmt, conn, a.catalog, a.spec), nil
}

func (a *CreateTableStmtAction) createIndexAutomatically(ctx context.Context, conn *Conn) error {
	for _, col := range a.spec.Columns {
		if !col.Type.AvailableAutoIndex() {
			continue
		}
		indexName := fmt.Sprintf("googlesqlite_autoindex_%s_%s", col.Name, strings.Join(a.spec.NamePath, "_"))
		createIndexQuery := fmt.Sprintf(
			"CREATE INDEX IF NOT EXISTS %s ON `%s`(`%s`)",
			indexName,
			a.spec.TableName(),
			col.Name,
		)
		if _, err := conn.ExecContext(ctx, createIndexQuery); err != nil {
			return fmt.Errorf("failed to create index automatically %s: %w", createIndexQuery, err)
		}
	}
	return nil
}

func (a *CreateTableStmtAction) exec(ctx context.Context, conn *Conn) error {
	// TEMP tables have "replace-on-recreate" semantics: re-running a
	// `CREATE TEMP TABLE foo` is expected to overwrite the previous
	// temp definition, not error out. The analyzer treats CREATE OR
	// REPLACE differently from CREATE TEMP, so we need the explicit
	// drop here as well.
	if a.spec.CreateMode == googlesql.ResolvedCreateStatementEnums_CreateModeCreateOrReplace || a.spec.IsTemp {
		if _, err := conn.ExecContext(
			ctx,
			fmt.Sprintf("DROP TABLE IF EXISTS `%s`", a.spec.TableName()),
		); err != nil {
			return err
		}
	}
	if _, err := conn.ExecContext(ctx, a.spec.SQLiteSchema(), a.args...); err != nil {
		return fmt.Errorf("failed to exec %s: %w", a.query, err)
	}
	if a.isAutoIndexMode {
		if err := a.createIndexAutomatically(ctx, conn); err != nil {
			return err
		}
	}
	if err := a.catalog.AddNewTableSpec(ctx, conn, a.spec); err != nil {
		return fmt.Errorf("failed to add new table spec: %w", err)
	}
	if !a.spec.IsTemp {
		conn.addTable(a.spec)
	}
	return nil
}

func (a *CreateTableStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *CreateTableStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *CreateTableStmtAction) Args() []any {
	return a.args
}

func (a *CreateTableStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	if !a.spec.IsTemp {
		return nil
	}

	if _, err := conn.ExecContext(
		ctx,
		fmt.Sprintf("DROP TABLE IF EXISTS `%s`", a.spec.TableName()),
	); err != nil {
		return fmt.Errorf("failed to cleanup table %s: %w", a.spec.TableName(), err)
	}
	if err := a.catalog.DeleteTableSpec(ctx, conn, a.spec.TableName()); err != nil {
		return fmt.Errorf("failed to delete table spec: %w", err)
	}
	return nil
}

type CreateViewStmtAction struct {
	query   string
	spec    *TableSpec
	catalog *Catalog
}

func (a *CreateViewStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	if a.spec.CreateMode == googlesql.ResolvedCreateStatementEnums_CreateModeCreateOrReplace {
		if _, err := conn.ExecContext(
			ctx,
			fmt.Sprintf("DROP VIEW IF EXISTS `%s`", a.spec.TableName()),
		); err != nil {
			return nil, err
		}
	}
	stmt, err := conn.PrepareContext(ctx, a.spec.SQLiteSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare %s: %w", a.query, err)
	}
	return newCreateViewStmt(stmt, conn, a.catalog, a.spec), nil
}

func (a *CreateViewStmtAction) exec(ctx context.Context, conn *Conn) error {
	if a.spec.CreateMode == googlesql.ResolvedCreateStatementEnums_CreateModeCreateOrReplace {
		if _, err := conn.ExecContext(
			ctx,
			fmt.Sprintf("DROP VIEW IF EXISTS `%s`", a.spec.TableName()),
		); err != nil {
			return err
		}
	}
	if _, err := conn.ExecContext(ctx, a.spec.SQLiteSchema()); err != nil {
		return fmt.Errorf("failed to exec %s: %w", a.query, err)
	}

	if err := a.catalog.AddNewTableSpec(ctx, conn, a.spec); err != nil {
		return fmt.Errorf("failed to add new view spec: %w", err)
	}
	return nil
}

func (a *CreateViewStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *CreateViewStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *CreateViewStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	if !a.spec.IsTemp {
		conn.addTable(a.spec)
		return nil
	}
	if _, err := conn.ExecContext(
		ctx,
		fmt.Sprintf("DROP VIEW IF EXISTS `%s`", a.spec.TableName()),
	); err != nil {
		return fmt.Errorf("failed to cleanup view %s: %w", a.spec.TableName(), err)
	}
	if err := a.catalog.DeleteTableSpec(ctx, conn, a.spec.TableName()); err != nil {
		return fmt.Errorf("failed to delete table spec: %w", err)
	}
	return nil
}

func (a *CreateViewStmtAction) Args() []any {
	return nil
}

type CreateFunctionStmtAction struct {
	spec    *FunctionSpec
	catalog *Catalog
	funcMap map[string]*FunctionSpec
}

func (a *CreateFunctionStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return newCreateFunctionStmt(conn, a.catalog, a.spec), nil
}

func (a *CreateFunctionStmtAction) exec(ctx context.Context, conn *Conn) error {
	if err := a.catalog.AddNewFunctionSpec(ctx, conn, a.spec); err != nil {
		return fmt.Errorf("failed to add new function spec: %w", err)
	}
	a.funcMap[a.spec.FuncName()] = a.spec
	conn.addFunction(a.spec)
	return nil
}

func (a *CreateFunctionStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *CreateFunctionStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *CreateFunctionStmtAction) Args() []any {
	return nil
}

func (a *CreateFunctionStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	if !a.spec.IsTemp {
		return nil
	}
	funcName := a.spec.FuncName()
	if err := a.catalog.DeleteFunctionSpec(ctx, conn, funcName); err != nil {
		return fmt.Errorf("failed to delete function spec: %w", err)
	}
	delete(a.funcMap, funcName)
	return nil
}

type CreateTableFunctionStmtAction struct {
	spec    *TVFSpec
	catalog *Catalog
	tvfMap  map[string]*TVFSpec
}

func (a *CreateTableFunctionStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return newCreateTableFunctionStmt(conn, a.catalog, a.spec), nil
}

func (a *CreateTableFunctionStmtAction) exec(ctx context.Context, conn *Conn) error {
	if err := a.catalog.AddNewTVFSpec(ctx, conn, a.spec); err != nil {
		return fmt.Errorf("failed to add new TVF spec: %w", err)
	}
	a.tvfMap[a.spec.TVFName()] = a.spec
	return nil
}

func (a *CreateTableFunctionStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *CreateTableFunctionStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *CreateTableFunctionStmtAction) Args() []any {
	return nil
}

func (a *CreateTableFunctionStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	if !a.spec.IsTemp {
		return nil
	}
	tvfName := a.spec.TVFName()
	if err := a.catalog.DeleteTVFSpec(ctx, conn, tvfName); err != nil {
		return fmt.Errorf("failed to delete TVF spec: %w", err)
	}
	delete(a.tvfMap, tvfName)
	return nil
}

type DropStmtAction struct {
	name           string
	objectType     string
	funcMap        map[string]*FunctionSpec
	catalog        *Catalog
	query          string
	formattedQuery string
	args           []any
}

func (a *DropStmtAction) exec(ctx context.Context, conn *Conn) error {
	switch a.objectType {
	case "TABLE", "VIEW":
		if _, err := conn.ExecContext(ctx, a.formattedQuery, a.args...); err != nil {
			return fmt.Errorf("failed to exec %s: %w", a.query, err)
		}
		spec := a.catalog.tableMap[a.name]
		if err := a.catalog.DeleteTableSpec(ctx, conn, a.name); err != nil {
			return fmt.Errorf("failed to delete table spec: %w", err)
		}
		conn.deleteTable(spec)
	case "FUNCTION":
		if err := a.catalog.DeleteFunctionSpec(ctx, conn, a.name); err != nil {
			return fmt.Errorf("failed to delete function spec: %w", err)
		}
		conn.deleteFunction(a.funcMap[a.name])
		delete(a.funcMap, a.name)
	default:
		return fmt.Errorf("currently unsupported DROP %s statement", a.objectType)
	}
	return nil
}

func (a *DropStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}

func (a *DropStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *DropStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *DropStmtAction) Args() []any {
	return nil
}

func (a *DropStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}

type DMLStmtAction struct {
	query          string
	params         []*googlesql.ResolvedParameter
	args           []any
	colTypes       []googlesql.Googlesql_TypeNode
	formattedQuery string
}

func (a *DMLStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	s, err := conn.PrepareContext(ctx, a.formattedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare %s: %w", a.query, err)
	}
	return newDMLStmt(s, a.params, a.colTypes, a.formattedQuery), nil
}

func (a *DMLStmtAction) exec(ctx context.Context, conn *Conn) (driver.Result, error) {
	result, err := conn.ExecContext(ctx, a.formattedQuery, a.args...)
	if err != nil {
		return nil, fmt.Errorf("failed to exec %s: %w", a.formattedQuery, err)
	}
	return result, nil
}

func (a *DMLStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	result, err := a.exec(ctx, conn)
	if err != nil {
		return nil, err
	}
	return &Result{conn: conn, result: result}, nil
}

func (a *DMLStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if _, err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *DMLStmtAction) Args() []any {
	return nil
}

func (a *DMLStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}

type QueryStmtAction struct {
	query          string
	params         []*googlesql.ResolvedParameter
	args           []any
	formattedQuery string
	outputColumns  []*ColumnSpec
	isExplainMode  bool
}

func (a *QueryStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	s, err := conn.PrepareContext(ctx, a.formattedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare %s: %w", a.query, err)
	}
	return newQueryStmt(s, a.params, a.formattedQuery, a.outputColumns), nil
}

func (a *QueryStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if _, err := conn.ExecContext(ctx, a.formattedQuery, a.args...); err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", a.query, err)
	}
	return &Result{conn: conn}, nil
}

func (a *QueryStmtAction) ExplainQueryPlan(ctx context.Context, conn *Conn) error {
	rows, err := conn.QueryContext(ctx, fmt.Sprintf("EXPLAIN QUERY PLAN %s", a.formattedQuery), a.args...)
	if err != nil {
		return fmt.Errorf("failed to explain query plan: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to explain query plan: %w", err)
	}
	defer rows.Close()
	fmt.Println("|selectid|order|from|detail|")
	fmt.Println("----------------------------")
	for rows.Next() {
		var (
			selectID, order, from int
			detail                string
		)
		if err := rows.Scan(&selectID, &order, &from, &detail); err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}
		fmt.Printf("|%d|%d|%d|%s|\n", selectID, order, from, detail)
	}
	return nil
}

func (a *QueryStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if a.isExplainMode {
		if err := a.ExplainQueryPlan(ctx, conn); err != nil {
			return nil, err
		}
		return &Rows{}, nil
	}
	rows, err := conn.QueryContext(ctx, a.formattedQuery, a.args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", a.query, err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", a.query, err)
	}
	return &Rows{conn: conn, rows: rows, columns: a.outputColumns}, nil
}

func (a *QueryStmtAction) Args() []any {
	return nil
}

func (a *QueryStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}

// ExportDataStmtAction materializes the rows of an
// `EXPORT DATA OPTIONS(...) AS <query>` statement to the destination URI.
// The inner query is run through the standard QueryStmtAction code path;
// the rows it produces are streamed through the format encoder, an
// optional compression wrapper and the writer obtained from the URI's
// registered scheme (see exportdata.RegisterURIWriter — `gs://` is
// registered by default).
//
// EXPORT DATA returns no rows to the caller — neither Exec nor Query
// yields any — so the result is always an empty driver.Result / *Rows
// after the destination object has been written. Real BigQuery has the
// same shape.
type ExportDataStmtAction struct {
	query          string
	params         []*googlesql.ResolvedParameter
	args           []any
	formattedQuery string
	outputColumns  []*ColumnSpec
	opts           *ResolvedExportDataOptions
}

// NewExportDataStmtAction constructs the action from a parsed inner query
// and the OPTIONS extracted from the EXPORT DATA statement's resolved AST.
// Exported so the analyzer (which lives in the same package today but may
// move) can construct it without reaching into private fields.
func NewExportDataStmtAction(
	query, formattedQuery string,
	params []*googlesql.ResolvedParameter,
	args []any,
	outputColumns []*ColumnSpec,
	opts *ResolvedExportDataOptions,
) *ExportDataStmtAction {
	return &ExportDataStmtAction{
		query:          query,
		params:         params,
		args:           args,
		formattedQuery: formattedQuery,
		outputColumns:  outputColumns,
		opts:           opts,
	}
}

func (a *ExportDataStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	// EXPORT DATA is fundamentally an Exec — there is nothing to prepare
	// in advance for the destination write. Fall back to preparing the
	// inner query so a Stmt wrapper has something to bind against; the
	// actual export happens on ExecContext / QueryContext below.
	s, err := conn.PrepareContext(ctx, a.formattedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare %s: %w", a.query, err)
	}
	return newQueryStmt(s, a.params, a.formattedQuery, a.outputColumns), nil
}

func (a *ExportDataStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.export(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *ExportDataStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.export(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *ExportDataStmtAction) Args() []any { return nil }

func (a *ExportDataStmtAction) Cleanup(ctx context.Context, conn *Conn) error { return nil }

// export resolves the URI's scheme writer, opens the destination
// (honouring `overwrite`), runs the inner query, wraps the writer in the
// requested compression codec, and streams each row through the format
// encoder. Every wrapper is closed in reverse order before the function
// returns — that is where the bytes commit for stores like GCS, and where
// the compression trailer is flushed.
func (a *ExportDataStmtAction) export(ctx context.Context, conn *Conn) error {
	u, err := url.Parse(a.opts.URI)
	if err != nil {
		return fmt.Errorf("EXPORT DATA: invalid uri %q: %w", a.opts.URI, err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("EXPORT DATA: uri %q has no scheme", a.opts.URI)
	}
	writer := exportdata.LookupURIWriter(u.Scheme)
	if writer == nil {
		return fmt.Errorf("EXPORT DATA: no writer registered for %q scheme — call googlesqlite.RegisterExportURIWriter to install one", u.Scheme)
	}

	rows, err := conn.QueryContext(ctx, a.formattedQuery, a.args...)
	if err != nil {
		return fmt.Errorf("EXPORT DATA: run inner query: %w", err)
	}
	defer rows.Close()

	// Prefer the resolved output column names; fall back to whatever the
	// driver reports if the analyzer did not surface a column list (an
	// inner SELECT that bypasses OutputColumnList — unlikely, but
	// defensive).
	columns := make([]string, len(a.outputColumns))
	for i, c := range a.outputColumns {
		columns[i] = c.Name
	}
	if len(columns) == 0 {
		columns, err = rows.Columns()
		if err != nil {
			return fmt.Errorf("EXPORT DATA: read column names: %w", err)
		}
	}

	dest, err := writer(ctx, a.opts.URI, exportdata.WriterOpts{Overwrite: a.opts.Overwrite})
	if err != nil {
		return fmt.Errorf("EXPORT DATA: open destination %q: %w", a.opts.URI, err)
	}
	// Wrap with the compression codec; both Closers must run so the
	// compression trailer flushes AND the underlying object commits.
	encStream, err := exportdata.WrapCompressor(dest, a.opts.Compression)
	if err != nil {
		_ = dest.Close()
		return err
	}
	encodeErr := exportdata.EncodeRows(encStream, a.opts.Format, columns, a.opts.CSV, sqlRowsSource(rows, len(columns)))
	if closeErr := encStream.Close(); closeErr != nil && encodeErr == nil {
		encodeErr = fmt.Errorf("EXPORT DATA: close destination %q: %w", a.opts.URI, closeErr)
	}
	return encodeErr
}

// sqlRowsSource adapts a database/sql.Rows iterator to exportdata.RowSource.
// Each invocation advances one row, scans the raw driver values, and decodes
// them out of googlesqlite's envelope into Go-native primitives so the
// format encoders see real strings / ints / floats rather than the base64
// `{header, body}` layout the driver round-trips internally. The (false,
// nil) terminator surfaces both natural end-of-stream and an underlying
// scan error (via rows.Err()).
func sqlRowsSource(rows *sql.Rows, ncols int) exportdata.RowSource {
	return func() ([]any, bool, error) {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return nil, false, fmt.Errorf("EXPORT DATA: iterate inner query: %w", err)
			}
			return nil, false, nil
		}
		raw := make([]any, ncols)
		ptrs := make([]any, ncols)
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, false, fmt.Errorf("EXPORT DATA: scan row: %w", err)
		}
		out := make([]any, ncols)
		for i, v := range raw {
			decoded, err := decodeExportValue(v)
			if err != nil {
				return nil, false, fmt.Errorf("EXPORT DATA: decode column %d: %w", i, err)
			}
			out[i] = decoded
		}
		return out, true, nil
	}
}

// decodeExportValue turns a single scanned driver value (still wrapped in
// googlesqlite's `{header, body}` envelope) into the Go-native value the
// EXPORT DATA encoders work in. Primitives (string, int64, float64, bool,
// []byte) pass through to keep the encoder's type-driven branches happy;
// richer types (Date, Timestamp, Numeric, Array, Struct ...) collapse to
// their canonical string form via the value's ToString — that string is
// what BigQuery itself uses for the same column in CSV / JSON output.
func decodeExportValue(v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	decoded, err := DecodeValue(v)
	if err != nil {
		return nil, err
	}
	if decoded == nil {
		return nil, nil
	}
	switch tv := decoded.(type) {
	case value.StringValue:
		return string(tv), nil
	case value.IntValue:
		return int64(tv), nil
	case value.FloatValue:
		return float64(tv), nil
	case value.BoolValue:
		return bool(tv), nil
	case value.BytesValue:
		return []byte(tv), nil
	}
	return decoded.ToString()
}

// NoopStmtAction handles statements that the analyzer accepts but
// whose side effects fall outside the driver's scope (GRANT, REVOKE,
// CREATE SCHEMA, ALTER TABLE, ASSERT, AUX LOAD DATA, etc.). Reporting
// success makes scripts that include them executable end-to-end so
// downstream tools can drive the analyzer through every statement
// without dropping back to error paths.
type NoopStmtAction struct{}

func (a *NoopStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}

func (a *NoopStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	return &Result{conn: conn}, nil
}

func (a *NoopStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	return &Rows{conn: conn}, nil
}

func (a *NoopStmtAction) Args() []any { return nil }

func (a *NoopStmtAction) Cleanup(ctx context.Context, conn *Conn) error { return nil }

type BeginStmtAction struct{}

func (a *BeginStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}

func (a *BeginStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	return &Result{conn: conn}, nil
}

func (a *BeginStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	return &Rows{conn: conn}, nil
}

func (a *BeginStmtAction) Args() []any {
	return nil
}

func (a *BeginStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}

// AssignmentStmtAction implements `SET @@var = expr` for
// session-scope system variables. The expression has already been
// formatted to a SQL string by the analyzer; we evaluate it through
// the underlying SQLite connection so simple literals, casts, and
// string concatenations all work without re-implementing the
// expression evaluator here. The result is stored on Conn.systemVars
// keyed by the dotted variable name (e.g. "time_zone").
type AssignmentStmtAction struct {
	name    string
	exprSQL string
}

func (a *AssignmentStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}

func (a *AssignmentStmtAction) exec(ctx context.Context, conn *Conn) error {
	if a.exprSQL == "" {
		conn.SetSystemVariable(a.name, "")
		return nil
	}
	row := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT %s", a.exprSQL))
	var raw any
	if err := row.Scan(&raw); err != nil {
		return fmt.Errorf("failed to evaluate SET @@%s expression: %w", a.name, err)
	}
	if raw == nil {
		conn.SetSystemVariable(a.name, "")
		return nil
	}
	val, err := DecodeValue(raw)
	if err != nil {
		// Primitive coming back un-enveloped (numbers, raw strings):
		// fall back to fmt.Sprint.
		conn.SetSystemVariable(a.name, fmt.Sprint(raw))
		return nil
	}
	if val == nil {
		conn.SetSystemVariable(a.name, "")
		return nil
	}
	s, err := val.ToString()
	if err != nil {
		return fmt.Errorf("SET @@%s: cannot stringify result: %w", a.name, err)
	}
	conn.SetSystemVariable(a.name, s)
	return nil
}

func (a *AssignmentStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *AssignmentStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *AssignmentStmtAction) Args() []any {
	return nil
}

func (a *AssignmentStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}

type CommitStmtAction struct{}

func (a *CommitStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}

func (a *CommitStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	return &Result{conn: conn}, nil
}

func (a *CommitStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	return &Rows{conn: conn}, nil
}

func (a *CommitStmtAction) Args() []any {
	return nil
}

func (a *CommitStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}

type TruncateStmtAction struct {
	query string
}

func (a *TruncateStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}

func (a *TruncateStmtAction) exec(ctx context.Context, conn *Conn) error {
	if _, err := conn.ExecContext(ctx, a.query); err != nil {
		return fmt.Errorf("failed to truncate %s: %w", a.query, err)
	}
	return nil
}

func (a *TruncateStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *TruncateStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *TruncateStmtAction) Args() []any {
	return nil
}

func (a *TruncateStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}

type MergeStmtAction struct {
	stmts []string
}

func (a *MergeStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}

func (a *MergeStmtAction) exec(ctx context.Context, conn *Conn) error {
	for _, stmt := range a.stmts {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to exec merge statement %s: %w", stmt, err)
		}
	}
	return nil
}

func (a *MergeStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Result{conn: conn}, nil
}

func (a *MergeStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	if err := a.exec(ctx, conn); err != nil {
		return nil, err
	}
	return &Rows{conn: conn}, nil
}

func (a *MergeStmtAction) Args() []any {
	return nil
}

func (a *MergeStmtAction) Cleanup(ctx context.Context, conn *Conn) error {
	return nil
}
