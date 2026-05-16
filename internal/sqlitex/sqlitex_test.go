package sqlitex_test

import (
	"errors"
	"strings"
	"testing"

	sqlite3 "github.com/ncruces/go-sqlite3"

	"github.com/goccy/googlesqlite/internal/sqlitex"
)

// openConn opens an in-memory ncruces sqlite3 connection. Cleanup
// closes the handle on test end.
func openConn(t *testing.T) *sqlite3.Conn {
	t.Helper()
	c, err := sqlite3.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// scalarResult executes a one-row SELECT and returns the first cell
// in canonical Go form (int64/float64/string/[]byte/nil).
func scalarResult(t *testing.T, c *sqlite3.Conn, sql string) any {
	t.Helper()
	stmt, _, err := c.Prepare(sql)
	if err != nil {
		t.Fatalf("prepare %q: %v", sql, err)
	}
	defer stmt.Close()
	if !stmt.Step() {
		if err := stmt.Err(); err != nil {
			t.Fatalf("step %q: %v", sql, err)
		}
		t.Fatalf("no row for %q", sql)
	}
	switch stmt.ColumnType(0) {
	case sqlite3.INTEGER:
		return stmt.ColumnInt64(0)
	case sqlite3.FLOAT:
		return stmt.ColumnFloat(0)
	case sqlite3.TEXT:
		return stmt.ColumnText(0)
	case sqlite3.BLOB:
		return append([]byte(nil), stmt.ColumnRawBlob(0)...)
	case sqlite3.NULL:
		return nil
	}
	return nil
}

// stepErr returns the first non-nil error from preparing or stepping
// the SQL — used by the negative-path tests.
func stepErr(t *testing.T, c *sqlite3.Conn, sql string) error {
	t.Helper()
	stmt, _, err := c.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	stmt.Step()
	return stmt.Err()
}

// TestRegisterFuncScalar registers a one-input scalar that adds 1 and
// asserts SQL invocations round-trip the value.
func TestRegisterFuncScalar(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterFunc(c, "add_one", func(x int64) int64 { return x + 1 }, sqlitex.FunctionFlags{Deterministic: true}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if got := scalarResult(t, c, "SELECT add_one(41)"); got != int64(42) {
		t.Fatalf("got %v (%T)", got, got)
	}
}

// TestRegisterFuncTwoReturn drives the `(T, error)` signature branch
// — both the happy path and the surfaced-error path.
func TestRegisterFuncTwoReturn(t *testing.T) {
	t.Parallel()
	c := openConn(t)

	register := func(name string, fn any) {
		if err := sqlitex.RegisterFunc(c, name, fn, sqlitex.FunctionFlags{}); err != nil {
			t.Fatalf("register %s: %v", name, err)
		}
	}
	register("ok_two_ret", func(x int64) (int64, error) { return x * 2, nil })
	register("err_two_ret", func(_ int64) (int64, error) { return 0, errors.New("boom") })

	if got := scalarResult(t, c, "SELECT ok_two_ret(7)"); got != int64(14) {
		t.Fatalf("ok_two_ret: %v", got)
	}
	if err := stepErr(t, c, "SELECT err_two_ret(0)"); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected boom, got %v", err)
	}
}

// TestRegisterFuncVariadic exercises the variadic argument branch in
// buildArgs.
func TestRegisterFuncVariadic(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterFunc(c, "var_sum", func(args ...any) (int64, error) {
		var sum int64
		for _, a := range args {
			if v, ok := a.(int64); ok {
				sum += v
			}
		}
		return sum, nil
	}, sqlitex.FunctionFlags{}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if got := scalarResult(t, c, "SELECT var_sum(1, 2, 3, 4)"); got != int64(10) {
		t.Fatalf("got %v", got)
	}
}

// TestRegisterFuncFloatAndStringAndBlob round-trips the float, string
// and []byte cases of applyResult / decodeValue.
func TestRegisterFuncFloatStringBlob(t *testing.T) {
	t.Parallel()
	c := openConn(t)

	if err := sqlitex.RegisterFunc(c, "echo_float", func(x float64) float64 { return x }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatal(err)
	}
	if err := sqlitex.RegisterFunc(c, "echo_text", func(x string) string { return x + "!" }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatal(err)
	}
	if err := sqlitex.RegisterFunc(c, "echo_blob", func(x []byte) []byte { return append([]byte{0xff}, x...) }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatal(err)
	}
	if err := sqlitex.RegisterFunc(c, "echo_bool", func(x int64) bool { return x != 0 }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatal(err)
	}

	if got := scalarResult(t, c, "SELECT echo_float(2.5)"); got != float64(2.5) {
		t.Fatalf("float: %v", got)
	}
	if got := scalarResult(t, c, "SELECT echo_text('hi')"); got != "hi!" {
		t.Fatalf("text: %v", got)
	}
	got := scalarResult(t, c, "SELECT echo_blob(x'01')")
	gotBlob, _ := got.([]byte)
	if len(gotBlob) != 2 || gotBlob[0] != 0xff || gotBlob[1] != 0x01 {
		t.Fatalf("blob: %v", got)
	}
	// SQLite encodes bool as int.
	if got := scalarResult(t, c, "SELECT echo_bool(1)"); got != int64(1) {
		t.Fatalf("bool: %v", got)
	}
}

// TestRegisterFuncNullArgIsZeroValue makes sure a NULL argument is
// passed as the target type's zero value to a typed signature.
func TestRegisterFuncNullArgIsZeroValue(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterFunc(c, "len_or_zero", func(s string) int64 { return int64(len(s)) }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatal(err)
	}
	if got := scalarResult(t, c, "SELECT len_or_zero(NULL)"); got != int64(0) {
		t.Fatalf("got %v", got)
	}
}

// TestRegisterFuncInvalidShapes feeds RegisterFunc each rejected
// callable shape.
func TestRegisterFuncInvalidShapes(t *testing.T) {
	t.Parallel()
	c := openConn(t)

	// Not a function at all.
	if err := sqlitex.RegisterFunc(c, "not_fn", 42, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error for non-function")
	}
	// Zero return values.
	if err := sqlitex.RegisterFunc(c, "no_ret", func() {}, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error for zero returns")
	}
	// Two returns but second is not error.
	if err := sqlitex.RegisterFunc(c, "two_int", func() (int64, int64) { return 0, 0 }, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error for non-error second return")
	}
}

// TestRegisterFuncArgCountMismatch surfaces the
// "expected N args, got M" error path inside buildArgs.
func TestRegisterFuncArgCountMismatch(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	// Register a 2-arg function and call it with 1 arg via ABI hack:
	// SQLite lets us issue the call with the wrong count because
	// CreateFunction allowed nArg=2. We use a variadic SQL wrapper.
	//
	// Easiest path is to register a 0-arg function and call it with
	// extra args — `add` expects 0 but SQL passes 1.
	if err := sqlitex.RegisterFunc(c, "zero_args", func() int64 { return 1 }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatal(err)
	}
	if err := stepErr(t, c, "SELECT zero_args(1)"); err == nil {
		t.Fatal("expected wrong-number-of-args error")
	}
}

// Aggregator-shaped helper used by both RegisterAggregator and
// RegisterWindow tests.
type sumInt struct{ total int64 }

func (s *sumInt) Step(args ...any) error {
	for _, a := range args {
		if v, ok := a.(int64); ok {
			s.total += v
		}
	}
	return nil
}

func (s *sumInt) Done() (any, error) { return s.total, nil }

// sumIntWithInverse adds an Inverse method, promoting it to a true
// window function for ncruces.
type sumIntWithInverse struct{ sumInt }

func (s *sumIntWithInverse) Inverse(args ...any) error {
	for _, a := range args {
		if v, ok := a.(int64); ok {
			s.total -= v
		}
	}
	return nil
}

// TestRegisterAggregator drives the standard aggregate path.
func TestRegisterAggregator(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterAggregator(c, "sum_int", func() *sumInt { return &sumInt{} }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := c.Exec("CREATE TABLE t(x); INSERT INTO t VALUES(1),(2),(3)"); err != nil {
		t.Fatal(err)
	}
	if got := scalarResult(t, c, "SELECT sum_int(x) FROM t"); got != int64(6) {
		t.Fatalf("got %v", got)
	}
}

// TestRegisterAggregatorInvalidCtor surfaces ctor-shape errors.
func TestRegisterAggregatorInvalidCtor(t *testing.T) {
	t.Parallel()
	c := openConn(t)

	if err := sqlitex.RegisterAggregator(c, "agg_not_fn", 42, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: not a function")
	}
	// Wrong signature: takes args.
	if err := sqlitex.RegisterAggregator(c, "agg_args", func(_ int) *sumInt { return nil }, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: ctor takes args")
	}
	// Wrong signature: returns two values.
	if err := sqlitex.RegisterAggregator(c, "agg_twoout", func() (*sumInt, error) { return nil, nil }, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: ctor returns two values")
	}
}

// missingStep deliberately omits the Step method.
type missingStep struct{}

func (m *missingStep) Done() (any, error) { return nil, nil }

// missingDone deliberately omits the Done method.
type missingDone struct{}

func (m *missingDone) Step(_ ...any) error { return nil }

func TestRegisterAggregatorMissingMethods(t *testing.T) {
	t.Parallel()
	c := openConn(t)

	if err := sqlitex.RegisterAggregator(c, "no_step", func() *missingStep { return &missingStep{} }, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: missing Step")
	}
	if err := sqlitex.RegisterAggregator(c, "no_done", func() *missingDone { return &missingDone{} }, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: missing Done")
	}
}

// TestRegisterWindowBuffered drives the buffered-window adapter (the
// aggregator has Step+Done but no Inverse).
func TestRegisterWindowBuffered(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterWindow(c, "win_sum_b", func() *sumInt { return &sumInt{} }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := c.Exec("CREATE TABLE t(x); INSERT INTO t VALUES(1),(2),(3),(4)"); err != nil {
		t.Fatal(err)
	}
	// Running total via a moving frame — exercises Step + Inverse +
	// Value flow.
	stmt, _, err := c.Prepare("SELECT win_sum_b(x) OVER (ORDER BY x ROWS BETWEEN 1 PRECEDING AND CURRENT ROW) FROM t")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	want := []int64{1, 3, 5, 7}
	for i := 0; stmt.Step(); i++ {
		if got := stmt.ColumnInt64(0); got != want[i] {
			t.Fatalf("row %d: got %d, want %d", i, got, want[i])
		}
	}
	if err := stmt.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestRegisterWindowInverse drives the windowAggAdapter path: the
// wrapped aggregator exposes Inverse so ncruces uses it for true
// frame-incremental updates.
func TestRegisterWindowInverse(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterWindow(c, "win_sum_i", func() *sumIntWithInverse { return &sumIntWithInverse{} }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := c.Exec("CREATE TABLE t(x); INSERT INTO t VALUES(1),(2),(3),(4)"); err != nil {
		t.Fatal(err)
	}
	stmt, _, err := c.Prepare("SELECT win_sum_i(x) OVER (ORDER BY x ROWS BETWEEN 1 PRECEDING AND CURRENT ROW) FROM t")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	want := []int64{1, 3, 5, 7}
	for i := 0; stmt.Step(); i++ {
		if got := stmt.ColumnInt64(0); got != want[i] {
			t.Fatalf("row %d: got %d, want %d", i, got, want[i])
		}
	}
	if err := stmt.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestRegisterWindowInvalidCtor(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterWindow(c, "win_not_fn", 42, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: not a function")
	}
	if err := sqlitex.RegisterWindow(c, "win_args", func(_ int) *sumInt { return nil }, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: ctor takes args")
	}
	if err := sqlitex.RegisterWindow(c, "win_no_step", func() *missingStep { return &missingStep{} }, sqlitex.FunctionFlags{}); err == nil {
		t.Fatal("expected error: missing Step")
	}
}

// TestRegisterCollation registers a simple length-based collation
// and asserts SQLite uses it for ORDER BY.
func TestRegisterCollation(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterCollation(c, "by_len", func(a, b string) int {
		switch {
		case len(a) < len(b):
			return -1
		case len(a) > len(b):
			return 1
		}
		return 0
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := c.Exec("CREATE TABLE t(x TEXT); INSERT INTO t VALUES('xxxx'),('xx'),('xxx')"); err != nil {
		t.Fatal(err)
	}
	stmt, _, err := c.Prepare("SELECT x FROM t ORDER BY x COLLATE by_len")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	var order []string
	for stmt.Step() {
		order = append(order, stmt.ColumnText(0))
	}
	if err := stmt.Err(); err != nil {
		t.Fatal(err)
	}
	if len(order) != 3 || order[0] != "xx" || order[1] != "xxx" || order[2] != "xxxx" {
		t.Fatalf("unexpected order: %v", order)
	}
}

// TestSetVariableNumberLimit only verifies the helper runs without
// panicking — its observable effect (parameter-count cap) is
// difficult to assert directly without binding thousands of params.
func TestSetVariableNumberLimit(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	sqlitex.SetVariableNumberLimit(c, -1)
	sqlitex.SetVariableNumberLimit(c, 9999)
}

// TestRegisterFuncReturnsError drives the `func(...) error` single-
// return shape (callError surfaces the err).
func TestRegisterFuncReturnsError(t *testing.T) {
	t.Parallel()
	c := openConn(t)
	if err := sqlitex.RegisterFunc(c, "always_err", func() error { return errors.New("nope") }, sqlitex.FunctionFlags{}); err != nil {
		t.Fatal(err)
	}
	if err := stepErr(t, c, "SELECT always_err()"); err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected nope, got %v", err)
	}
}
