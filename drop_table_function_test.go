package googlesqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestDropTableFunction covers DROP TABLE FUNCTION [IF EXISTS], per
// https://cloud.google.com/bigquery/docs/reference/standard-sql/data-definition-language#drop_table_function
func TestDropTableFunction(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=drop_table_function")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	exec := func(stmt string) error {
		_, err := conn.ExecContext(ctx, stmt)
		return err
	}
	call := func() (int64, error) {
		var n int64
		err := conn.QueryRowContext(ctx, "SELECT n FROM PlusOne(1)").Scan(&n)
		return n, err
	}

	if err := exec("CREATE TABLE FUNCTION PlusOne(k INT64) AS (SELECT k + 1 AS n)"); err != nil {
		t.Fatalf("CREATE TABLE FUNCTION: %v", err)
	}
	if n, err := call(); err != nil || n != 2 {
		t.Fatalf("call before drop: n=%d err=%v", n, err)
	}
	if err := exec("DROP TABLE FUNCTION PlusOne"); err != nil {
		t.Fatalf("DROP TABLE FUNCTION: %v", err)
	}
	if _, err := call(); err == nil {
		t.Fatal("call after DROP TABLE FUNCTION succeeded; want an error")
	}
	if err := exec("DROP TABLE FUNCTION IF EXISTS PlusOne"); err != nil {
		t.Fatalf("DROP TABLE FUNCTION IF EXISTS on a dropped function: %v", err)
	}
	if err := exec("DROP TABLE FUNCTION PlusOne"); err == nil {
		t.Fatal("DROP TABLE FUNCTION on a missing function succeeded; want an error")
	}
	// The name is free again after the drop.
	if err := exec("CREATE TABLE FUNCTION PlusOne(k INT64) AS (SELECT k + 10 AS n)"); err != nil {
		t.Fatalf("re-CREATE after drop: %v", err)
	}
	if n, err := call(); err != nil || n != 11 {
		t.Fatalf("call after re-create: n=%d err=%v", n, err)
	}
}
