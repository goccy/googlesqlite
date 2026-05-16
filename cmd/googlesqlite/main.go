// Command googlesqlite is a CLI for running GoogleSQL queries against a
// SQLite-backed engine.
//
// This is a bootstrap skeleton. A REPL and script-runner UI will be added
// once the analyzer and execution layers land.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/goccy/googlesqlite"
)

func main() {
	dsn := flag.String("dsn", ":memory:", "data source name passed to sql.Open")
	flag.Parse()

	if err := run(*dsn); err != nil {
		fmt.Fprintln(os.Stderr, "googlesqlite:", err)
		os.Exit(1)
	}
}

func run(dsn string) error {
	db, err := sql.Open("googlesqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}
	fmt.Printf("connected to %s\n", dsn)
	return nil
}
