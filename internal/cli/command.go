package cli

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Session couples a Runner with the mutable CLI state (debug mode,
// colour) and an output sink. The native REPL and the wasm Playground
// entrypoint both drive a Session.
type Session struct {
	runner *Runner
	out    io.Writer
	// Debug, when true, prints the translated SQLite query above each
	// result.
	Debug bool
	// Color enables ANSI colourisation of rendered output.
	Color bool
}

// NewSession creates a Session writing to out.
func NewSession(runner *Runner, out io.Writer) *Session {
	return &Session{runner: runner, out: out}
}

// Runner returns the underlying Runner.
func (s *Session) Runner() *Runner { return s.runner }

// TableNames returns the names of every table and view in the session,
// sorted. It is the structured form of the .tables dot-command, for
// callers (the wasm Playground) that render a schema view.
func (s *Session) TableNames(ctx context.Context) ([]string, error) {
	return s.runner.catalogNames(ctx, catalogKindTable, catalogKindView)
}

func (s *Session) renderOptions() RenderOptions {
	return RenderOptions{Color: s.Color, Debug: s.Debug}
}

// RunStatement executes a single statement and renders its result to
// the session output. The Result is also returned so callers (the
// wasm Playground) can record it in history.
func (s *Session) RunStatement(ctx context.Context, stmt string) Result {
	res := s.runner.Exec(ctx, stmt)
	RenderResult(s.out, res, s.renderOptions())
	return res
}

// ReadFileFunc resolves a path referenced by a `.read` dot-command to
// the file's contents. The core package stays filesystem-free so it
// can build for js/wasm; the native CLI supplies an os.ReadFile-backed
// implementation, and the wasm Playground supplies one that reports
// `.read` as unsupported in the browser.
type ReadFileFunc func(path string) (string, error)

// RunInput processes a block of input — a SQL file or piped stdin.
// Lines that begin with `.` are dot-commands; everything else is
// accumulated into `;`-terminated (or trailing-\G) statements and
// executed in order. Each result is rendered to the session output
// and returned. With stopOnError set, processing halts at the first
// failing statement.
func (s *Session) RunInput(ctx context.Context, text string, stopOnError bool, readFile ReadFileFunc) []Result {
	var (
		results []Result
		buf     strings.Builder
	)
	runStmt := func(stmt string) bool {
		res := s.RunStatement(ctx, stmt)
		results = append(results, res)
		return res.Err == nil || !stopOnError
	}
	for _, line := range strings.Split(text, "\n") {
		if buf.Len() == 0 && strings.HasPrefix(strings.TrimSpace(line), ".") {
			dot := s.HandleDot(ctx, line)
			if dot.Quit {
				return results
			}
			if dot.ReadPath != "" {
				if !s.sourceFile(ctx, dot.ReadPath, stopOnError, readFile, &results) && stopOnError {
					return results
				}
			}
			continue
		}
		buf.WriteString(line)
		buf.WriteString("\n")
		complete, remainder := SplitComplete(buf.String())
		for _, stmt := range complete {
			if !runStmt(stmt) {
				return results
			}
		}
		buf.Reset()
		buf.WriteString(remainder)
	}
	// Run any trailing statement that was not terminated by a semicolon.
	if tail := strings.TrimSpace(buf.String()); tail != "" {
		for _, stmt := range SplitStatements(tail) {
			if !runStmt(stmt) {
				return results
			}
		}
	}
	return results
}

// sourceFile handles a `.read` reference encountered inside RunInput.
// It returns false when reading fails or a nested statement fails
// under stop-on-error.
func (s *Session) sourceFile(ctx context.Context, path string, stopOnError bool, readFile ReadFileFunc, results *[]Result) bool {
	if readFile == nil {
		fmt.Fprintln(s.out, ".read is not supported in this environment")
		return false
	}
	content, err := readFile(path)
	if err != nil {
		fmt.Fprintf(s.out, "ERROR: %v\n", err)
		return false
	}
	nested := s.RunInput(ctx, content, stopOnError, readFile)
	*results = append(*results, nested...)
	for _, res := range nested {
		if res.Err != nil {
			return false
		}
	}
	return true
}

// DotResult reports the outcome of dispatching a line that began with
// a dot-command.
type DotResult struct {
	// Handled is true when the line was recognised as a dot-command.
	Handled bool
	// Quit is true for .quit / .exit.
	Quit bool
	// ReadPath is set by `.read <path>`; the caller is responsible for
	// reading the file and feeding it to RunScript (the core stays
	// filesystem-free so it can build for js/wasm).
	ReadPath string
}

// HandleDot interprets line as a dot-command. It returns Handled=false
// when line is not a dot-command, leaving the caller to treat it as
// SQL. Recognised commands write their own output to the session.
func (s *Session) HandleDot(ctx context.Context, line string) DotResult {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ".") {
		return DotResult{}
	}
	fields := strings.Fields(trimmed)
	cmd := fields[0]
	args := fields[1:]
	switch cmd {
	case ".quit", ".exit":
		return DotResult{Handled: true, Quit: true}
	case ".help":
		s.printHelp()
		return DotResult{Handled: true}
	case ".debug":
		s.setToggle(args, &s.Debug, "debug")
		return DotResult{Handled: true}
	case ".tables":
		s.printCatalog(ctx, "tables", string(catalogKindTable), string(catalogKindView))
		return DotResult{Handled: true}
	case ".functions":
		s.printCatalog(ctx, "functions", string(catalogKindFunction), string(catalogKindTVF))
		return DotResult{Handled: true}
	case ".read":
		if len(args) == 0 {
			fmt.Fprintln(s.out, ".read requires a file path")
			return DotResult{Handled: true}
		}
		return DotResult{Handled: true, ReadPath: strings.Join(args, " ")}
	default:
		fmt.Fprintf(s.out, "unknown command: %s (try .help)\n", cmd)
		return DotResult{Handled: true}
	}
}

// catalog kind strings as stored in the googlesqlite_catalog table.
const (
	catalogKindTable    = "table"
	catalogKindView     = "view"
	catalogKindFunction = "function"
	catalogKindTVF      = "tvf"
)

// setToggle applies an on/off argument to a boolean session flag. With
// no argument it flips the current value.
func (s *Session) setToggle(args []string, flag *bool, name string) {
	if len(args) == 0 {
		*flag = !*flag
	} else {
		switch strings.ToLower(args[0]) {
		case "on", "true", "1":
			*flag = true
		case "off", "false", "0":
			*flag = false
		default:
			fmt.Fprintf(s.out, ".%s requires on/off\n", name)
			return
		}
	}
	state := "off"
	if *flag {
		state = "on"
	}
	fmt.Fprintf(s.out, "%s mode %s\n", name, state)
}

// printCatalog lists the names of catalog entries of the given kinds.
func (s *Session) printCatalog(ctx context.Context, label string, kinds ...string) {
	names, err := s.runner.catalogNames(ctx, kinds...)
	if err != nil {
		fmt.Fprintf(s.out, "ERROR: %v\n", err)
		return
	}
	if len(names) == 0 {
		fmt.Fprintf(s.out, "no %s\n", label)
		return
	}
	for _, name := range names {
		fmt.Fprintln(s.out, name)
	}
}

func (s *Session) printHelp() {
	fmt.Fprint(s.out, `Commands:
  .help              show this help
  .quit, .exit       leave the CLI
  .debug [on|off]    show the translated SQLite query for each statement
  .tables            list tables and views
  .functions         list functions
  .read <path>       run the statements in a SQL file

Append \G to a query to print it in vertical (row-per-stanza) form.
`)
}

// catalogNames reads entry names of the given kinds from the
// googlesqlite_catalog bookkeeping table. It opens a short-lived
// connection to the raw SQLite engine because that table is internal
// to the engine and not visible to the GoogleSQL analyzer. A missing
// catalog table (nothing created yet) is reported as an empty list.
func (r *Runner) catalogNames(ctx context.Context, kinds ...string) ([]string, error) {
	if len(kinds) == 0 {
		return nil, nil
	}
	db, err := sql.Open(rawDriverName, r.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(kinds)), ",")
	args := make([]any, len(kinds))
	for i, k := range kinds {
		args[i] = k
	}
	query := "SELECT name FROM googlesqlite_catalog WHERE kind IN (" + placeholders + ")"
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		// No catalog table yet: the session has not created anything.
		if strings.Contains(err.Error(), "no such table") {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}
