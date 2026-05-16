// Unit tests for the Bind* surface in the array package.
// Expected outputs follow the upstream GoogleSQL / BigQuery array
// functions reference at
// docs/third_party/googlesql-docs/array_functions.md.
package array_test

import (
	"math/big"
	"reflect"
	"testing"

	arrayfn "github.com/goccy/googlesqlite/internal/functions/array"
	"github.com/goccy/googlesqlite/internal/value"
)

// arr is a small helper for building an *value.ArrayValue from a
// slice of typed cells.
func arr(vs ...value.Value) *value.ArrayValue {
	return &value.ArrayValue{Values: vs}
}

// asArr asserts the value is an *ArrayValue and returns its slice
// of cells for further inspection.
func asArr(t *testing.T, v value.Value) []value.Value {
	t.Helper()
	a, ok := v.(*value.ArrayValue)
	if !ok {
		t.Fatalf("expected *ArrayValue, got %T", v)
	}
	return a.Values
}

// ------------------------------------------------------------------
// Arity / error matrix
// ------------------------------------------------------------------

func TestArrayBind_Arity(t *testing.T) {
	t.Parallel()

	if _, err := arrayfn.BindArrayConcat(); err == nil {
		t.Errorf("ARRAY_CONCAT: expected error on zero args")
	}
	if _, err := arrayfn.BindArrayLength(); err == nil {
		t.Errorf("ARRAY_LENGTH: expected error on zero args")
	}
	if _, err := arrayfn.BindArrayLength(arr(), arr()); err == nil {
		t.Errorf("ARRAY_LENGTH: expected error on >1 args")
	}
	if _, err := arrayfn.BindArrayReverse(); err == nil {
		t.Errorf("ARRAY_REVERSE: expected error on zero args")
	}
	if _, err := arrayfn.BindArrayReverse(arr(), arr()); err == nil {
		t.Errorf("ARRAY_REVERSE: expected error on >1 args")
	}
	if _, err := arrayfn.BindArrayToString(arr()); err == nil {
		t.Errorf("ARRAY_TO_STRING: expected error on <2 args")
	}
	if _, err := arrayfn.BindGenerateArray(value.IntValue(1)); err == nil {
		t.Errorf("GENERATE_ARRAY: expected error on 1 arg")
	}
	if _, err := arrayfn.BindGenerateArray(value.IntValue(1), value.IntValue(2), value.IntValue(3), value.IntValue(4)); err == nil {
		t.Errorf("GENERATE_ARRAY: expected error on 4 args")
	}
	if _, err := arrayfn.ARRAY_IS_DISTINCT(); err == nil {
		t.Errorf("ARRAY_IS_DISTINCT: expected error on 0 args")
	}
	if _, err := arrayfn.ARRAY_IS_DISTINCT(arr(), arr()); err == nil {
		t.Errorf("ARRAY_IS_DISTINCT: expected error on 2 args")
	}
}

// ------------------------------------------------------------------
// ARRAY_CONCAT
// ------------------------------------------------------------------

func TestArrayConcat(t *testing.T) {
	t.Parallel()

	// ARRAY_CONCAT([1, 2], [3, 4]) -> [1, 2, 3, 4]
	got, err := arrayfn.BindArrayConcat(
		arr(value.IntValue(1), value.IntValue(2)),
		arr(value.IntValue(3), value.IntValue(4)),
	)
	if err != nil {
		t.Fatalf("BindArrayConcat: %v", err)
	}
	cells := asArr(t, got)
	if len(cells) != 4 {
		t.Fatalf("len = %d; want 4", len(cells))
	}
	want := []int64{1, 2, 3, 4}
	for i, w := range want {
		v, _ := cells[i].ToInt64()
		if v != w {
			t.Errorf("cell %d = %d; want %d", i, v, w)
		}
	}
}

// ------------------------------------------------------------------
// ARRAY_IN
// ------------------------------------------------------------------

func TestArrayIn(t *testing.T) {
	t.Parallel()

	// 2 IN UNNEST([1, 2, 3]) -> true
	got, err := arrayfn.BindInArray(
		value.IntValue(2),
		arr(value.IntValue(1), value.IntValue(2), value.IntValue(3)),
	)
	if err != nil {
		t.Fatalf("BindInArray hit: %v", err)
	}
	if got != value.BoolValue(true) {
		t.Errorf("hit = %v; want true", got)
	}
	// 4 IN UNNEST([1, 2, 3]) -> false
	got, err = arrayfn.BindInArray(
		value.IntValue(4),
		arr(value.IntValue(1), value.IntValue(2), value.IntValue(3)),
	)
	if err != nil {
		t.Fatalf("BindInArray miss: %v", err)
	}
	if got != value.BoolValue(false) {
		t.Errorf("miss = %v; want false", got)
	}
	// NULL IN UNNEST([1]) -> false (helper.ExistsNull short-circuit)
	got, err = arrayfn.BindInArray(nil, arr(value.IntValue(1)))
	if err != nil {
		t.Fatalf("BindInArray null: %v", err)
	}
	if got != value.BoolValue(false) {
		t.Errorf("null = %v; want false", got)
	}
}

// ------------------------------------------------------------------
// ARRAY_IS_DISTINCT
// ------------------------------------------------------------------

func TestArrayIsDistinct(t *testing.T) {
	t.Parallel()

	// [1, 2, 3] -> true
	got, err := arrayfn.ARRAY_IS_DISTINCT(arr(value.IntValue(1), value.IntValue(2), value.IntValue(3)))
	if err != nil {
		t.Fatalf("distinct: %v", err)
	}
	if got != value.BoolValue(true) {
		t.Errorf("distinct = %v; want true", got)
	}
	// [1, 2, 1] -> false
	got, err = arrayfn.ARRAY_IS_DISTINCT(arr(value.IntValue(1), value.IntValue(2), value.IntValue(1)))
	if err != nil {
		t.Fatalf("dup: %v", err)
	}
	if got != value.BoolValue(false) {
		t.Errorf("dup = %v; want false", got)
	}
	// [NULL] -> true (single NULL is its own distinct value)
	got, err = arrayfn.ARRAY_IS_DISTINCT(arr(nil))
	if err != nil {
		t.Fatalf("one null: %v", err)
	}
	if got != value.BoolValue(true) {
		t.Errorf("one null = %v; want true", got)
	}
	// [NULL, NULL] -> false
	got, err = arrayfn.ARRAY_IS_DISTINCT(arr(nil, nil))
	if err != nil {
		t.Fatalf("two nulls: %v", err)
	}
	if got != value.BoolValue(false) {
		t.Errorf("two nulls = %v; want false", got)
	}
	// NULL input -> NULL
	got, err = arrayfn.ARRAY_IS_DISTINCT(value.Value(nil))
	if err != nil {
		t.Fatalf("null arg: %v", err)
	}
	if got != nil {
		t.Errorf("null arg = %v; want nil", got)
	}
}

// ------------------------------------------------------------------
// ARRAY_LENGTH
// ------------------------------------------------------------------

func TestArrayLength(t *testing.T) {
	t.Parallel()

	got, err := arrayfn.BindArrayLength(arr(value.IntValue(1), value.IntValue(2), value.IntValue(3)))
	if err != nil {
		t.Fatalf("BindArrayLength: %v", err)
	}
	if got != value.IntValue(3) {
		t.Errorf("got %v; want 3", got)
	}
	// NULL -> NULL
	got, err = arrayfn.BindArrayLength(value.Value(nil))
	if err != nil {
		t.Fatalf("BindArrayLength NULL: %v", err)
	}
	if got != nil {
		t.Errorf("ARRAY_LENGTH(NULL) = %v; want nil", got)
	}
}

// ------------------------------------------------------------------
// ARRAY_OFFSET / ARRAY_ORDINAL and SAFE variants
// ------------------------------------------------------------------

func TestArrayAtOffsetAndOrdinal(t *testing.T) {
	t.Parallel()
	a := arr(value.StringValue("a"), value.StringValue("b"), value.StringValue("c"))

	// OFFSET(0) -> "a"
	got, err := arrayfn.BindArrayAtOffset(a, value.IntValue(0))
	if err != nil {
		t.Fatalf("offset 0: %v", err)
	}
	if got != value.StringValue("a") {
		t.Errorf("got %v; want a", got)
	}
	// OFFSET out of range -> error
	if _, err := arrayfn.BindArrayAtOffset(a, value.IntValue(99)); err == nil {
		t.Errorf("offset 99: expected error")
	}
	// SAFE_OFFSET out of range -> nil
	got, err = arrayfn.BindSafeArrayAtOffset(a, value.IntValue(99))
	if err != nil {
		t.Fatalf("safe offset 99: %v", err)
	}
	if got != nil {
		t.Errorf("safe offset 99: %v; want nil", got)
	}
	// SAFE_OFFSET negative -> nil
	got, err = arrayfn.BindSafeArrayAtOffset(a, value.IntValue(-1))
	if err != nil {
		t.Fatalf("safe offset -1: %v", err)
	}
	if got != nil {
		t.Errorf("safe offset -1: %v; want nil", got)
	}
	// ORDINAL(1) -> "a"
	got, err = arrayfn.BindArrayAtOrdinal(a, value.IntValue(1))
	if err != nil {
		t.Fatalf("ordinal 1: %v", err)
	}
	if got != value.StringValue("a") {
		t.Errorf("got %v; want a", got)
	}
	// ORDINAL(0) -> error (must be >= 1)
	if _, err := arrayfn.BindArrayAtOrdinal(a, value.IntValue(0)); err == nil {
		t.Errorf("ordinal 0: expected error")
	}
	// SAFE_ORDINAL(0) -> nil
	got, err = arrayfn.BindSafeArrayAtOrdinal(a, value.IntValue(0))
	if err != nil {
		t.Fatalf("safe ordinal 0: %v", err)
	}
	if got != nil {
		t.Errorf("safe ordinal 0: %v; want nil", got)
	}
	// NULL second arg -> nil
	got, err = arrayfn.BindArrayAtOffset(a, nil)
	if err != nil {
		t.Fatalf("offset NULL: %v", err)
	}
	if got != nil {
		t.Errorf("offset NULL: %v; want nil", got)
	}
}

// ------------------------------------------------------------------
// ARRAY_REVERSE
// ------------------------------------------------------------------

func TestArrayReverse(t *testing.T) {
	t.Parallel()
	// ARRAY_REVERSE([1, 2, 3]) -> [3, 2, 1]
	got, err := arrayfn.BindArrayReverse(arr(value.IntValue(1), value.IntValue(2), value.IntValue(3)))
	if err != nil {
		t.Fatalf("BindArrayReverse: %v", err)
	}
	cells := asArr(t, got)
	want := []int64{3, 2, 1}
	for i, w := range want {
		v, _ := cells[i].ToInt64()
		if v != w {
			t.Errorf("cell %d = %d; want %d", i, v, w)
		}
	}
}

// ------------------------------------------------------------------
// ARRAY_TO_STRING
// ------------------------------------------------------------------

func TestArrayToString(t *testing.T) {
	t.Parallel()
	a := arr(value.StringValue("a"), value.StringValue("b"), value.StringValue("c"))
	// ARRAY_TO_STRING(["a","b","c"], "--") -> "a--b--c"
	got, err := arrayfn.BindArrayToString(a, value.StringValue("--"))
	if err != nil {
		t.Fatalf("BindArrayToString: %v", err)
	}
	if got != value.StringValue("a--b--c") {
		t.Errorf("got %v; want a--b--c", got)
	}
	// With NULL element and no null_text -> omitted.
	an := arr(value.StringValue("a"), nil, value.StringValue("c"))
	got, err = arrayfn.BindArrayToString(an, value.StringValue("--"))
	if err != nil {
		t.Fatalf("BindArrayToString null: %v", err)
	}
	if got != value.StringValue("a--c") {
		t.Errorf("got %v; want a--c", got)
	}
	// With null_text -> NULL becomes that string.
	got, err = arrayfn.BindArrayToString(an, value.StringValue("--"), value.StringValue("N"))
	if err != nil {
		t.Fatalf("BindArrayToString null_text: %v", err)
	}
	if got != value.StringValue("a--N--c") {
		t.Errorf("got %v; want a--N--c", got)
	}
}

// ------------------------------------------------------------------
// MAKE_ARRAY
// ------------------------------------------------------------------

func TestMakeArray(t *testing.T) {
	t.Parallel()
	got, err := arrayfn.BindMakeArray(value.IntValue(1), value.IntValue(2))
	if err != nil {
		t.Fatalf("BindMakeArray: %v", err)
	}
	cells := asArr(t, got)
	if len(cells) != 2 {
		t.Fatalf("len = %d; want 2", len(cells))
	}
}

// ------------------------------------------------------------------
// GENERATE_ARRAY
// ------------------------------------------------------------------

func TestGenerateArray(t *testing.T) {
	t.Parallel()
	// GENERATE_ARRAY(1, 5) -> [1, 2, 3, 4, 5]
	got, err := arrayfn.BindGenerateArray(value.IntValue(1), value.IntValue(5))
	if err != nil {
		t.Fatalf("GENERATE_ARRAY(1, 5): %v", err)
	}
	cells := asArr(t, got)
	want := []int64{1, 2, 3, 4, 5}
	if len(cells) != len(want) {
		t.Fatalf("len = %d; want %d", len(cells), len(want))
	}
	for i, w := range want {
		v, _ := cells[i].ToInt64()
		if v != w {
			t.Errorf("cell %d = %d; want %d", i, v, w)
		}
	}
	// GENERATE_ARRAY(0, 10, 3) -> [0, 3, 6, 9]
	got, err = arrayfn.BindGenerateArray(value.IntValue(0), value.IntValue(10), value.IntValue(3))
	if err != nil {
		t.Fatalf("GENERATE_ARRAY(0,10,3): %v", err)
	}
	cells = asArr(t, got)
	want = []int64{0, 3, 6, 9}
	if len(cells) != len(want) {
		t.Fatalf("len = %d; want %d", len(cells), len(want))
	}
	for i, w := range want {
		v, _ := cells[i].ToInt64()
		if v != w {
			t.Errorf("cell %d = %d; want %d", i, v, w)
		}
	}
	// GENERATE_ARRAY(5, 1, -2) -> [5, 3, 1]
	got, err = arrayfn.BindGenerateArray(value.IntValue(5), value.IntValue(1), value.IntValue(-2))
	if err != nil {
		t.Fatalf("GENERATE_ARRAY(5,1,-2): %v", err)
	}
	cells = asArr(t, got)
	want = []int64{5, 3, 1}
	if len(cells) != len(want) {
		t.Fatalf("descending len = %d; want %d", len(cells), len(want))
	}
	for i, w := range want {
		v, _ := cells[i].ToInt64()
		if v != w {
			t.Errorf("descending cell %d = %d; want %d", i, v, w)
		}
	}
	// Mismatched direction: GENERATE_ARRAY(1, 5, -1) -> [] (start<end, neg step)
	got, err = arrayfn.BindGenerateArray(value.IntValue(1), value.IntValue(5), value.IntValue(-1))
	if err != nil {
		t.Fatalf("mismatch: %v", err)
	}
	cells = asArr(t, got)
	if len(cells) != 0 {
		t.Errorf("mismatch len = %d; want 0", len(cells))
	}
	// Mismatched direction: GENERATE_ARRAY(5, 1, 1) -> []
	got, err = arrayfn.BindGenerateArray(value.IntValue(5), value.IntValue(1), value.IntValue(1))
	if err != nil {
		t.Fatalf("mismatch reverse: %v", err)
	}
	cells = asArr(t, got)
	if len(cells) != 0 {
		t.Errorf("mismatch reverse len = %d; want 0", len(cells))
	}
	// NULL inputs -> NULL output.
	got, err = arrayfn.BindGenerateArray(nil, value.IntValue(5))
	if err != nil {
		t.Fatalf("GENERATE_ARRAY(NULL, 5): %v", err)
	}
	if got != nil {
		t.Errorf("NULL input: got %v; want nil", got)
	}
}

// ------------------------------------------------------------------
// RANGE_BUCKET
// ------------------------------------------------------------------

func TestRangeBucket(t *testing.T) {
	t.Parallel()
	bounds := arr(value.IntValue(10), value.IntValue(20), value.IntValue(30))
	// RANGE_BUCKET(5,  [10,20,30]) -> 0
	// RANGE_BUCKET(10, [10,20,30]) -> 1
	// RANGE_BUCKET(15, [10,20,30]) -> 1
	// RANGE_BUCKET(25, [10,20,30]) -> 2
	// RANGE_BUCKET(40, [10,20,30]) -> 3
	cases := []struct {
		point int64
		want  int64
	}{
		{5, 0},
		{10, 1},
		{15, 1},
		{25, 2},
		{40, 3},
	}
	for _, c := range cases {
		got, err := arrayfn.BindRangeBucket(value.IntValue(c.point), bounds)
		if err != nil {
			t.Fatalf("RANGE_BUCKET(%d): %v", c.point, err)
		}
		if got != value.IntValue(c.want) {
			t.Errorf("RANGE_BUCKET(%d) = %v; want %d", c.point, got, c.want)
		}
	}

	// NULL point -> NULL.
	got, err := arrayfn.BindRangeBucket(nil, bounds)
	if err != nil {
		t.Fatalf("RANGE_BUCKET(NULL): %v", err)
	}
	if got != nil {
		t.Errorf("NULL point: got %v; want nil", got)
	}

	// NULL inside the bound array -> error (per body when point passes nil).
	withNull := arr(value.IntValue(10), nil, value.IntValue(30))
	if _, err := arrayfn.BindRangeBucket(value.IntValue(15), withNull); err == nil {
		t.Errorf("expected error for NULL inside bound array")
	}
}

// ------------------------------------------------------------------
// Numeric variants: ensure ARRAY_CONCAT keeps NumericValue ordering.
// ------------------------------------------------------------------

func TestArrayConcatNumeric(t *testing.T) {
	t.Parallel()
	one := &value.NumericValue{Rat: big.NewRat(1, 1)}
	two := &value.NumericValue{Rat: big.NewRat(2, 1)}
	three := &value.NumericValue{Rat: big.NewRat(3, 1)}
	got, err := arrayfn.BindArrayConcat(arr(one, two), arr(three))
	if err != nil {
		t.Fatalf("BindArrayConcat numeric: %v", err)
	}
	cells := asArr(t, got)
	if len(cells) != 3 {
		t.Fatalf("len = %d; want 3", len(cells))
	}
	if !reflect.DeepEqual(cells[2], three) {
		t.Errorf("cell 2 = %v; want %v", cells[2], three)
	}
}
