package math

import (
	gomath "math"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

func mustFloat(t *testing.T, v value.Value) float64 {
	t.Helper()
	f, err := v.ToFloat64()
	if err != nil {
		t.Fatalf("ToFloat64: %v", err)
	}
	return f
}

func mustInt(t *testing.T, v value.Value) int64 {
	t.Helper()
	n, err := v.ToInt64()
	if err != nil {
		t.Fatalf("ToInt64: %v", err)
	}
	return n
}

func mustBool(t *testing.T, v value.Value) bool {
	t.Helper()
	b, err := v.ToBool()
	if err != nil {
		t.Fatalf("ToBool: %v", err)
	}
	return b
}

// --- simple unary fns ---

func TestUnaryMath(t *testing.T) {
	cases := []struct {
		name string
		fn   func(value.Value) (value.Value, error)
		in   float64
		want float64
	}{
		{"ABS positive", ABS, 3, 3},
		{"ABS negative", ABS, -3, 3},
		{"ABS zero", ABS, 0, 0},
		{"SQRT", SQRT, 9, 3},
		{"CEIL up", CEIL, 1.5, 2},
		{"CEIL int", CEIL, 1, 1},
		{"FLOOR down", FLOOR, 1.5, 1},
		{"FLOOR neg", FLOOR, -1.5, -2},
		{"LN e", LN, gomath.E, 1},
		{"LOG10 100", LOG10, 100, 2},
		{"EXP 0", EXP, 0, 1},
		{"TRUNC 3.7", TRUNC, 3.7, 3},
		{"CBRT 27", CBRT, 27, 3},
	}
	for _, tc := range cases {
		got, err := tc.fn(value.FloatValue(tc.in))
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		f := mustFloat(t, got)
		if gomath.Abs(f-tc.want) > 1e-9 {
			t.Errorf("%s(%v) = %v, want %v", tc.name, tc.in, f, tc.want)
		}
	}
}

func TestUnaryTrig(t *testing.T) {
	// COS(0)=1, SIN(0)=0, TAN(0)=0
	got, _ := COS(value.FloatValue(0))
	if mustFloat(t, got) != 1 {
		t.Fatalf("COS(0) = %v, want 1", mustFloat(t, got))
	}
	got, _ = SIN(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("SIN(0) = %v, want 0", mustFloat(t, got))
	}
	got, _ = TAN(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("TAN(0) = %v, want 0", mustFloat(t, got))
	}

	// Inverse trig: ACOS(1)=0, ASIN(0)=0, ATAN(0)=0
	got, _ = ACOS(value.FloatValue(1))
	if mustFloat(t, got) != 0 {
		t.Fatalf("ACOS(1) = %v, want 0", mustFloat(t, got))
	}
	got, _ = ASIN(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("ASIN(0) = %v", mustFloat(t, got))
	}
	got, _ = ATAN(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("ATAN(0) = %v", mustFloat(t, got))
	}

	// Hyperbolic: COSH(0)=1, SINH(0)=0, TANH(0)=0
	got, _ = COSH(value.FloatValue(0))
	if mustFloat(t, got) != 1 {
		t.Fatalf("COSH(0)")
	}
	got, _ = SINH(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("SINH(0)")
	}
	got, _ = TANH(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("TANH(0)")
	}

	// Inverse hyperbolic: ACOSH(1)=0
	got, _ = ACOSH(value.FloatValue(1))
	if mustFloat(t, got) != 0 {
		t.Fatalf("ACOSH(1)")
	}
	got, _ = ASINH(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("ASINH(0)")
	}
	got, _ = ATANH(value.FloatValue(0))
	if mustFloat(t, got) != 0 {
		t.Fatalf("ATANH(0)")
	}

	// Reciprocal trig: CSC(pi/2)=1, SEC(0)=1, COT(pi/4)=1
	got, _ = CSC(value.FloatValue(gomath.Pi / 2))
	if gomath.Abs(mustFloat(t, got)-1) > 1e-9 {
		t.Fatalf("CSC(pi/2) = %v", mustFloat(t, got))
	}
	got, _ = SEC(value.FloatValue(0))
	if mustFloat(t, got) != 1 {
		t.Fatalf("SEC(0) = %v", mustFloat(t, got))
	}
	got, _ = COT(value.FloatValue(gomath.Pi / 4))
	if gomath.Abs(mustFloat(t, got)-1) > 1e-9 {
		t.Fatalf("COT(pi/4) = %v", mustFloat(t, got))
	}
	got, _ = CSCH(value.FloatValue(1))
	if got == nil {
		t.Fatalf("CSCH(1) nil")
	}
	got, _ = SECH(value.FloatValue(0))
	if mustFloat(t, got) != 1 {
		t.Fatalf("SECH(0)")
	}
	got, _ = COTH(value.FloatValue(1))
	if got == nil {
		t.Fatalf("COTH(1) nil")
	}
}

// --- binary ops ---

func TestBinaryMath(t *testing.T) {
	got, _ := ATAN2(value.FloatValue(1), value.FloatValue(1))
	if gomath.Abs(mustFloat(t, got)-gomath.Pi/4) > 1e-9 {
		t.Fatalf("ATAN2(1,1) = %v, want pi/4", mustFloat(t, got))
	}

	got, _ = POW(value.FloatValue(2), value.FloatValue(8))
	if mustFloat(t, got) != 256 {
		t.Fatalf("POW(2,8) = %v, want 256", mustFloat(t, got))
	}

	got, _ = MOD(value.FloatValue(10), value.FloatValue(3))
	if gomath.Abs(mustFloat(t, got)-1) > 1e-9 {
		t.Fatalf("MOD(10,3) = %v, want 1", mustFloat(t, got))
	}
	if _, err := MOD(value.FloatValue(1), value.FloatValue(0)); err == nil {
		t.Fatalf("MOD division by zero should error")
	}
}

// --- DIV (INT64 integer division) ---

func TestDiv(t *testing.T) {
	got, _ := DIV(value.IntValue(7), value.IntValue(2))
	if mustInt(t, got) != 3 {
		t.Fatalf("DIV(7,2) = %d, want 3", mustInt(t, got))
	}
	if _, err := DIV(value.IntValue(1), value.IntValue(0)); err == nil {
		t.Fatalf("DIV zero divisor should error")
	}
}

// --- IS_INF / IS_NAN / SIGN ---

func TestIsInfNanSign(t *testing.T) {
	got, _ := IS_INF(value.FloatValue(gomath.Inf(1)))
	if !mustBool(t, got) {
		t.Fatalf("IS_INF(+Inf) = false")
	}
	got, _ = IS_INF(value.FloatValue(1))
	if mustBool(t, got) {
		t.Fatalf("IS_INF(1) = true")
	}

	got, _ = IS_NAN(value.FloatValue(gomath.NaN()))
	if !mustBool(t, got) {
		t.Fatalf("IS_NAN(NaN) = false")
	}
	got, _ = IS_NAN(value.FloatValue(1))
	if mustBool(t, got) {
		t.Fatalf("IS_NAN(1) = true")
	}

	got, _ = SIGN(value.FloatValue(3))
	if mustInt(t, got) != 1 {
		t.Fatalf("SIGN(3) want 1")
	}
	got, _ = SIGN(value.FloatValue(-2))
	if mustInt(t, got) != -1 {
		t.Fatalf("SIGN(-2) want -1")
	}
	got, _ = SIGN(value.FloatValue(0))
	if mustInt(t, got) != 0 {
		t.Fatalf("SIGN(0) want 0")
	}
}

// --- ROUND ---

func TestRoundDefault(t *testing.T) {
	got, err := BindRound(value.FloatValue(2.5))
	if err != nil {
		t.Fatalf("BindRound: %v", err)
	}
	if mustFloat(t, got) != 3 {
		t.Fatalf("ROUND(2.5) = %v, want 3", mustFloat(t, got))
	}
}

func TestRoundPrecision(t *testing.T) {
	got, _ := BindRound(value.FloatValue(3.14159), value.IntValue(2))
	if gomath.Abs(mustFloat(t, got)-3.14) > 1e-9 {
		t.Fatalf("ROUND(3.14159, 2) = %v, want 3.14", mustFloat(t, got))
	}
}

func TestBindRoundNullAndArity(t *testing.T) {
	got, _ := BindRound(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindRound(); err == nil {
		t.Fatalf("arity error expected (0)")
	}
}

// --- GREATEST / LEAST ---

func TestGreatestAndLeast(t *testing.T) {
	got, _ := BindGreatest(value.IntValue(1), value.IntValue(3), value.IntValue(2))
	if mustInt(t, got) != 3 {
		t.Fatalf("GREATEST = %d, want 3", mustInt(t, got))
	}
	got, _ = BindLeast(value.IntValue(1), value.IntValue(3), value.IntValue(2))
	if mustInt(t, got) != 1 {
		t.Fatalf("LEAST = %d, want 1", mustInt(t, got))
	}

	// Single arg.
	got, _ = BindGreatest(value.IntValue(5))
	if mustInt(t, got) != 5 {
		t.Fatalf("GREATEST single = %d, want 5", mustInt(t, got))
	}

	// NULL propagation.
	got, _ = BindGreatest(value.IntValue(1), nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	got, _ = BindLeast(value.IntValue(1), nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

// --- IEEE_DIVIDE ---

func TestIeeeDivide(t *testing.T) {
	// 0/0 = NaN
	got, _ := IEEE_DIVIDE(value.FloatValue(0), value.FloatValue(0))
	if !gomath.IsNaN(mustFloat(t, got)) {
		t.Fatalf("IEEE 0/0 want NaN")
	}
	// 0/x = 0
	got, _ = IEEE_DIVIDE(value.FloatValue(0), value.FloatValue(3))
	if mustFloat(t, got) != 0 {
		t.Fatalf("IEEE 0/3 want 0")
	}
	// 0/NaN = NaN
	got, _ = IEEE_DIVIDE(value.FloatValue(0), value.FloatValue(gomath.NaN()))
	if !gomath.IsNaN(mustFloat(t, got)) {
		t.Fatalf("IEEE 0/NaN want NaN")
	}
	// NaN/0 = NaN
	got, _ = IEEE_DIVIDE(value.FloatValue(gomath.NaN()), value.FloatValue(0))
	if !gomath.IsNaN(mustFloat(t, got)) {
		t.Fatalf("IEEE NaN/0 want NaN")
	}
	// Inf/Inf = NaN
	got, _ = IEEE_DIVIDE(value.FloatValue(gomath.Inf(1)), value.FloatValue(gomath.Inf(1)))
	if !gomath.IsNaN(mustFloat(t, got)) {
		t.Fatalf("IEEE Inf/Inf want NaN")
	}
	// 1/0 = +Inf
	got, _ = IEEE_DIVIDE(value.FloatValue(1), value.FloatValue(0))
	if !gomath.IsInf(mustFloat(t, got), 1) {
		t.Fatalf("IEEE 1/0 want +Inf")
	}
	// -1/0 = -Inf
	got, _ = IEEE_DIVIDE(value.FloatValue(-1), value.FloatValue(0))
	if !gomath.IsInf(mustFloat(t, got), -1) {
		t.Fatalf("IEEE -1/0 want -Inf")
	}
	// normal 6/3 = 2
	got, _ = IEEE_DIVIDE(value.FloatValue(6), value.FloatValue(3))
	if mustFloat(t, got) != 2 {
		t.Fatalf("IEEE 6/3 = %v", mustFloat(t, got))
	}
}

// --- SAFE_* ---

func TestSafeOps(t *testing.T) {
	got, _ := SAFE_ADD(value.FloatValue(1), value.FloatValue(2))
	if mustFloat(t, got) != 3 {
		t.Fatalf("SAFE_ADD(1,2) = %v", mustFloat(t, got))
	}
	got, _ = SAFE_SUBTRACT(value.FloatValue(5), value.FloatValue(3))
	if mustFloat(t, got) != 2 {
		t.Fatalf("SAFE_SUBTRACT(5,3) = %v", mustFloat(t, got))
	}
	got, _ = SAFE_MULTIPLY(value.FloatValue(2), value.FloatValue(3))
	if mustFloat(t, got) != 6 {
		t.Fatalf("SAFE_MULTIPLY(2,3) = %v", mustFloat(t, got))
	}
	got, _ = SAFE_DIVIDE(value.FloatValue(6), value.FloatValue(3))
	if mustFloat(t, got) != 2 {
		t.Fatalf("SAFE_DIVIDE(6,3) = %v", mustFloat(t, got))
	}
	// Division by zero returns NULL.
	got, _ = SAFE_DIVIDE(value.FloatValue(1), value.FloatValue(0))
	if got != nil {
		t.Fatalf("SAFE_DIVIDE(1,0) want NULL")
	}
	got, _ = SAFE_NEGATE(value.FloatValue(3))
	if mustFloat(t, got) != -3 {
		t.Fatalf("SAFE_NEGATE(3) = %v", mustFloat(t, got))
	}
}

// --- COSINE_DISTANCE / EUCLIDEAN_DISTANCE ---

func TestCosineDistance(t *testing.T) {
	a := &value.ArrayValue{Values: []value.Value{value.FloatValue(1), value.FloatValue(0)}}
	b := &value.ArrayValue{Values: []value.Value{value.FloatValue(1), value.FloatValue(0)}}
	got, err := COSINE_DISTANCE(a, b)
	if err != nil {
		t.Fatalf("COSINE_DISTANCE: %v", err)
	}
	if gomath.Abs(mustFloat(t, got)) > 1e-9 {
		t.Fatalf("same-vector distance = %v, want 0", mustFloat(t, got))
	}

	// Orthogonal
	a = &value.ArrayValue{Values: []value.Value{value.FloatValue(1), value.FloatValue(0)}}
	b = &value.ArrayValue{Values: []value.Value{value.FloatValue(0), value.FloatValue(1)}}
	got, _ = COSINE_DISTANCE(a, b)
	if gomath.Abs(mustFloat(t, got)-1) > 1e-9 {
		t.Fatalf("orthogonal distance = %v, want 1", mustFloat(t, got))
	}

	// Mismatched lengths.
	a = &value.ArrayValue{Values: []value.Value{value.FloatValue(1)}}
	b = &value.ArrayValue{Values: []value.Value{value.FloatValue(1), value.FloatValue(0)}}
	if _, err := COSINE_DISTANCE(a, b); err == nil {
		t.Fatalf("mismatched lengths should error")
	}

	// Zero magnitude.
	a = &value.ArrayValue{Values: []value.Value{value.FloatValue(0), value.FloatValue(0)}}
	if _, err := COSINE_DISTANCE(a, a); err == nil {
		t.Fatalf("zero-magnitude should error")
	}

	got, _ = COSINE_DISTANCE(nil, &value.ArrayValue{})
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}

	if _, err := COSINE_DISTANCE(a); err == nil {
		t.Fatalf("arity error expected")
	}
}

func TestEuclideanDistance(t *testing.T) {
	a := &value.ArrayValue{Values: []value.Value{value.FloatValue(0), value.FloatValue(0)}}
	b := &value.ArrayValue{Values: []value.Value{value.FloatValue(3), value.FloatValue(4)}}
	got, _ := EUCLIDEAN_DISTANCE(a, b)
	if mustFloat(t, got) != 5 {
		t.Fatalf("EUCLIDEAN_DISTANCE(0,3-4) = %v, want 5", mustFloat(t, got))
	}

	a = &value.ArrayValue{Values: []value.Value{value.FloatValue(1)}}
	b = &value.ArrayValue{Values: []value.Value{value.FloatValue(1), value.FloatValue(0)}}
	if _, err := EUCLIDEAN_DISTANCE(a, b); err == nil {
		t.Fatalf("mismatched lengths should error")
	}

	got, _ = EUCLIDEAN_DISTANCE(nil, &value.ArrayValue{})
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := EUCLIDEAN_DISTANCE(a); err == nil {
		t.Fatalf("arity error expected")
	}

	// Element error: NULL element.
	a = &value.ArrayValue{Values: []value.Value{nil}}
	b = &value.ArrayValue{Values: []value.Value{value.FloatValue(0)}}
	if _, err := EUCLIDEAN_DISTANCE(a, b); err == nil {
		t.Fatalf("NULL element should error")
	}
}

// --- BindRand: just confirm it returns a FLOAT64 in [0,1) ---

func TestBindRand(t *testing.T) {
	got, err := BindRand()
	if err != nil {
		t.Fatalf("BindRand: %v", err)
	}
	f := mustFloat(t, got)
	if f < 0 || f >= 1 {
		t.Fatalf("RAND() = %v, expected in [0,1)", f)
	}
}

// --- OP_DIV ---

func TestOpDiv(t *testing.T) {
	got, _ := OP_DIV(value.FloatValue(6), value.FloatValue(3))
	if mustFloat(t, got) != 2 {
		t.Fatalf("OP_DIV(6,3) = %v, want 2", mustFloat(t, got))
	}
}

// --- ADD / SUB / MUL ---

func TestArith(t *testing.T) {
	got, _ := ADD(value.IntValue(2), value.IntValue(3))
	if mustInt(t, got) != 5 {
		t.Fatalf("ADD = %d", mustInt(t, got))
	}
	got, _ = SUB(value.IntValue(5), value.IntValue(2))
	if mustInt(t, got) != 3 {
		t.Fatalf("SUB = %d", mustInt(t, got))
	}
	got, _ = MUL(value.IntValue(4), value.IntValue(3))
	if mustInt(t, got) != 12 {
		t.Fatalf("MUL = %d", mustInt(t, got))
	}
}

// --- error-path coverage ---

// newBad returns a value whose ToFloat64 / ToInt64 always fails, so
// we can exercise the err-return branches of every kernel without
// modifying production code.
func newBad() value.Value {
	return &value.ArrayValue{Values: []value.Value{value.IntValue(1)}}
}

func TestUnaryErrorPaths(t *testing.T) {
	bad := newBad()
	funcs := map[string]func(value.Value) (value.Value, error){
		"ABS":         ABS,
		"SQRT":        SQRT,
		"CEIL":        CEIL,
		"FLOOR":       FLOOR,
		"LN":          LN,
		"LOG10":       LOG10,
		"EXP":         EXP,
		"TRUNC":       TRUNC,
		"CBRT":        CBRT,
		"COS":         COS,
		"SIN":         SIN,
		"TAN":         TAN,
		"ACOS":        ACOS,
		"ASIN":        ASIN,
		"ATAN":        ATAN,
		"COSH":        COSH,
		"SINH":        SINH,
		"TANH":        TANH,
		"ACOSH":       ACOSH,
		"ASINH":       ASINH,
		"ATANH":       ATANH,
		"CSC":         CSC,
		"SEC":         SEC,
		"COT":         COT,
		"CSCH":        CSCH,
		"SECH":        SECH,
		"COTH":        COTH,
		"IS_INF":      IS_INF,
		"IS_NAN":      IS_NAN,
		"SIGN":        SIGN,
		"SAFE_NEGATE": SAFE_NEGATE,
	}
	for name, fn := range funcs {
		if _, err := fn(bad); err == nil {
			t.Errorf("%s: expected ToFloat64 error to propagate", name)
		}
	}
}

func TestBinaryErrorPaths(t *testing.T) {
	bad := newBad()
	good := value.FloatValue(1)
	binaryFloat := map[string]func(value.Value, value.Value) (value.Value, error){
		"ATAN2":         ATAN2,
		"POW":           POW,
		"MOD":           MOD,
		"IEEE_DIVIDE":   IEEE_DIVIDE,
		"SAFE_ADD":      SAFE_ADD,
		"SAFE_SUBTRACT": SAFE_SUBTRACT,
		"SAFE_MULTIPLY": SAFE_MULTIPLY,
		"SAFE_DIVIDE":   SAFE_DIVIDE,
	}
	for name, fn := range binaryFloat {
		if _, err := fn(bad, good); err == nil {
			t.Errorf("%s: x error not propagated", name)
		}
		if _, err := fn(good, bad); err == nil {
			t.Errorf("%s: y error not propagated", name)
		}
	}

	if _, err := DIV(bad, value.IntValue(1)); err == nil {
		t.Errorf("DIV x error not propagated")
	}
	if _, err := DIV(value.IntValue(1), bad); err == nil {
		t.Errorf("DIV y error not propagated")
	}
}

// TestRoundErrorPath covers BindRound's ToFloat64 error path.
func TestRoundErrorPath(t *testing.T) {
	bad := newBad()
	if _, err := BindRound(bad); err == nil {
		t.Errorf("BindRound: x error not propagated")
	}
	if _, err := BindRound(value.FloatValue(1), bad); err == nil {
		t.Errorf("BindRound: precision error not propagated")
	}
}

// TestBindGreatestErrorPath drives a value type that errors on GT.
func TestBindGreatestErrorPath(t *testing.T) {
	if _, err := BindGreatest(value.BoolValue(true), value.BoolValue(false)); err == nil {
		t.Errorf("BindGreatest: GT error not propagated for BOOL")
	}
}

func TestBindLeastErrorPath(t *testing.T) {
	if _, err := BindLeast(value.BoolValue(true), value.BoolValue(false)); err == nil {
		t.Errorf("BindLeast: LT error not propagated for BOOL")
	}
}

// --- BindLog ---

func TestBindLog(t *testing.T) {
	// 1-arg form = LN.
	got, err := BindLog(value.FloatValue(gomath.E))
	if err != nil {
		t.Fatalf("BindLog 1-arg: %v", err)
	}
	if gomath.Abs(mustFloat(t, got)-1) > 1e-9 {
		t.Fatalf("BindLog(e) = %v, want 1", mustFloat(t, got))
	}

	// 2-arg form delegates to LOG (which uses Ldexp via the
	// implementation). Just confirm no error.
	if _, err := BindLog(value.FloatValue(2), value.FloatValue(4)); err != nil {
		t.Fatalf("BindLog 2-arg: %v", err)
	}

	got, _ = BindLog(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}
