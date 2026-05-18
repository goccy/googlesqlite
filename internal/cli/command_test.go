package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestSessionRunInputWithDotCommands(t *testing.T) {
	r := newTestRunner(t)
	var buf bytes.Buffer
	s := NewSession(r, &buf)

	script := `
-- create some schema
CREATE TABLE users (id INT64, name STRING);
INSERT INTO users VALUES (1, 'a'), (2, 'b');
.tables
SELECT COUNT(*) AS n FROM users;
`
	results := s.RunInput(context.Background(), script, true, nil)
	for _, res := range results {
		if res.Err != nil {
			t.Fatalf("statement %q failed: %v", res.Statement, res.Err)
		}
	}
	out := buf.String()
	if !strings.Contains(out, "users") {
		t.Errorf(".tables did not list the users table:\n%s", out)
	}
}

func TestSessionDebugDotCommand(t *testing.T) {
	r := newTestRunner(t)
	var buf bytes.Buffer
	s := NewSession(r, &buf)
	ctx := context.Background()

	if dot := s.HandleDot(ctx, ".debug on"); !dot.Handled {
		t.Fatalf(".debug not handled")
	}
	if !s.Debug {
		t.Fatalf(".debug on did not enable debug mode")
	}
	buf.Reset()
	s.RunStatement(ctx, "SELECT 1 AS x")
	if !strings.Contains(buf.String(), "-- sqlite:") {
		t.Errorf("debug mode did not print the translated query:\n%s", buf.String())
	}
}

func TestSessionMultiStatementScript(t *testing.T) {
	r := newTestRunner(t)
	var buf bytes.Buffer
	s := NewSession(r, &buf)

	// A semicolon inside a string literal must not split the INSERT.
	script := `CREATE TABLE t (v STRING);
INSERT INTO t VALUES ('a;b;c');
SELECT v FROM t`
	results := s.RunInput(context.Background(), script, true, nil)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	for _, res := range results {
		if res.Err != nil {
			t.Fatalf("statement %q failed: %v", res.Statement, res.Err)
		}
	}
	last := results[2]
	if len(last.Rows) != 1 || last.Rows[0][0] != "a;b;c" {
		t.Errorf("final SELECT = %v, want [[a;b;c]]", last.Rows)
	}
}
