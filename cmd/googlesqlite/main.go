//go:build !js

// Command googlesqlite is an interactive console for running
// GoogleSQL queries against the SQLite-backed googlesqlite engine,
// in the spirit of the sqlite3 and spanner-cli REPLs.
//
// It can also run SQL scripts: pass --file to execute a file before
// the REPL starts, pipe SQL on stdin, or use the .read dot-command
// mid-session. In every case the REPL opens afterwards in the
// resulting session state.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"

	"github.com/goccy/googlesqlite/internal/cli"
)

const (
	primaryPrompt      = "googlesql> "
	continuationPrompt = "      ...> "
)

// stringSliceFlag collects a repeatable string flag.
type stringSliceFlag []string

func (s *stringSliceFlag) String() string { return strings.Join(*s, ",") }

func (s *stringSliceFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	var (
		dsnFlag   string
		files     stringSliceFlag
		debug     bool
		noColor   bool
		histFile  string
		contOnErr bool
	)
	flag.StringVar(&dsnFlag, "dsn", "", "data source name (overrides the positional database argument)")
	flag.Var(&files, "file", "run statements from a SQL file before the REPL starts (repeatable)")
	flag.BoolVar(&debug, "debug", false, "show the translated SQLite query for each statement")
	flag.BoolVar(&noColor, "no-color", false, "disable coloured output")
	flag.StringVar(&histFile, "history", defaultHistoryFile(), "REPL history file")
	flag.BoolVar(&contOnErr, "continue-on-error", false, "keep running a script after a statement fails")
	flag.Parse()

	ctx := context.Background()
	dsn := resolveDSN(dsnFlag, flag.Args())

	runner, err := cli.NewRunner(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "googlesqlite:", err)
		return 1
	}
	defer runner.Close()

	useColor := !noColor && isatty.IsTerminal(os.Stdout.Fd())
	color.NoColor = !useColor

	session := cli.NewSession(runner, os.Stdout)
	session.Debug = debug
	session.Color = useColor
	stopOnErr := !contOnErr

	// 1. --file scripts, in order.
	for _, f := range files {
		if err := runFile(ctx, session, f, stopOnErr); err != nil {
			fmt.Fprintf(os.Stderr, "googlesqlite: %v\n", err)
			return 1
		}
	}

	// 2. Piped stdin, run as a script.
	stdinPiped := !isatty.IsTerminal(os.Stdin.Fd())
	if stdinPiped {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "googlesqlite: failed to read stdin: %v\n", err)
			return 1
		}
		session.RunInput(ctx, string(data), stopOnErr, osReadFile)
	}

	// 3. Interactive REPL, in the state left by steps 1 and 2. When
	// stdin was piped its terminal is gone, so reopen the controlling
	// terminal; if there is none, the script has already run — stop.
	in, ok := interactiveInput(stdinPiped)
	if !ok {
		return 0
	}
	if closer, isCloser := in.(io.Closer); isCloser && in != nil {
		defer closer.Close()
	}
	return repl(ctx, session, in, histFile, stopOnErr)
}

// resolveDSN turns the --dsn flag and positional arguments into a
// googlesqlite DSN. A positional database name becomes a shared-cache
// file DSN; with neither, a shared-cache in-memory database is used.
func resolveDSN(flagDSN string, args []string) string {
	if flagDSN != "" {
		return flagDSN
	}
	if len(args) > 0 {
		return fmt.Sprintf("file:%s?cache=shared", args[0])
	}
	return cli.DefaultMemoryDSN
}

// osReadFile is the filesystem-backed ReadFileFunc the native CLI
// supplies for --file and `.read`.
func osReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

// runFile reads a SQL file and runs its statements (and any
// dot-commands it contains) through the session.
func runFile(ctx context.Context, session *cli.Session, path string, stopOnErr bool) error {
	content, err := osReadFile(path)
	if err != nil {
		return err
	}
	session.RunInput(ctx, content, stopOnErr, osReadFile)
	return nil
}

// repl runs the interactive read-eval-print loop. stdin is the
// terminal to read from; a nil stdin tells readline to use os.Stdin.
func repl(ctx context.Context, session *cli.Session, stdin io.Reader, histFile string, stopOnErr bool) int {
	cfg := &readline.Config{
		Prompt:      primaryPrompt,
		HistoryFile: histFile,
	}
	if rc, ok := stdin.(io.ReadCloser); ok && stdin != nil {
		cfg.Stdin = rc
	}
	rl, err := readline.NewEx(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "googlesqlite:", err)
		return 1
	}
	defer rl.Close()

	var buf strings.Builder
	for {
		if buf.Len() == 0 {
			rl.SetPrompt(primaryPrompt)
		} else {
			rl.SetPrompt(continuationPrompt)
		}
		line, err := rl.Readline()
		switch err {
		case readline.ErrInterrupt:
			// Ctrl-C discards a half-typed statement.
			buf.Reset()
			continue
		case io.EOF:
			return 0
		case nil:
			// proceed
		default:
			fmt.Fprintln(os.Stderr, "googlesqlite:", err)
			return 1
		}

		// A dot-command is recognised only at the start of a statement.
		if buf.Len() == 0 && strings.HasPrefix(strings.TrimSpace(line), ".") {
			dot := session.HandleDot(ctx, line)
			if dot.Quit {
				return 0
			}
			if dot.ReadPath != "" {
				if err := runFile(ctx, session, dot.ReadPath, stopOnErr); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				}
			}
			continue
		}

		buf.WriteString(line)
		buf.WriteString("\n")
		complete, remainder := cli.SplitComplete(buf.String())
		for _, stmt := range complete {
			session.RunStatement(ctx, stmt)
		}
		buf.Reset()
		buf.WriteString(remainder)
	}
}

// defaultHistoryFile returns the REPL history file path, preferring
// the user's home directory.
func defaultHistoryFile() string {
	const name = ".googlesqlite_history"
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, name)
	}
	return name
}
