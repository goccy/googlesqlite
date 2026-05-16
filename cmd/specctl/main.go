// Command specctl is the operational CLI for the spec → test pipeline.
//
// Run `specctl` (no args) for the list of subcommands.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
)

// command is one specctl subcommand.
type command struct {
	name    string
	summary string
	run     func(ctx context.Context, args []string) error
}

var registry = map[string]command{}

func register(c command) {
	if _, exists := registry[c.name]; exists {
		panic("specctl: duplicate subcommand " + c.name)
	}
	registry[c.name] = c
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
		os.Exit(2)
	}

	name := flag.Arg(0)
	cmd, ok := registry[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "specctl: unknown subcommand %q\n", name)
		usage()
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.run(ctx, flag.Args()[1:]); err != nil {
		if errors.Is(err, errNotImplemented) {
			fmt.Fprintf(os.Stderr, "specctl: %s is not yet implemented\n", name)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "specctl %s: %v\n", name, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: specctl <subcommand> [flags...]")
	fmt.Fprintln(os.Stderr, "subcommands:")

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(os.Stderr, "  %-16s %s\n", name, registry[name].summary)
	}
}

var errNotImplemented = errors.New("not implemented")

// projectRoot walks upward from the working directory to the repository
// root, identified by the go.mod file.
func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for range 8 {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found upward from %s", dir)
}

// All subcommands are implemented in their own files
// (upstream_sync.go, normalize.go, validate.go, check.go, coverage.go).
