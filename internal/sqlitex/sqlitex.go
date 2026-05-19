// Package sqlitex provides a thin compatibility shim that lets us
// register Go-typed scalar, aggregate, and window functions on an
// ncruces/go-sqlite3 connection using the same shapes that the
// predecessor exposed through its SQLite binding.
//
// The shim does the reflection that the predecessor's binding handled
// implicitly so that the bulk of the engine code (function tables,
// catalog, statement execution) can be ported without changing the
// per-function signatures.
package sqlitex

import (
	"errors"
	"fmt"
	"reflect"

	sqlite3 "github.com/ncruces/go-sqlite3"
)

// SQLiteFunc is the shape of a scalar function registered through
// RegisterFunc when callers want full control over arguments.
//
// It mirrors the predecessor's SQLiteFunction type:
//
//	func(args ...interface{}) (interface{}, error)
//
// Engine code passes such functions wholesale; the shim handles
// argument unmarshalling and result encoding into the ncruces
// runtime.
type SQLiteFunc = func(args ...any) (any, error)

// Aggregator-shaped types must implement these methods, mirroring the
// predecessor's mattn-driven shape:
//
//	Step(args ...interface{}) error
//	Done() (interface{}, error)
//
// RegisterAggregator accepts any function value of shape
// `func() T` where T satisfies that interface.

// CollatingFunc is the shape of a string collation callback.
type CollatingFunc = func(a, b string) int

// FunctionFlags describes optional traits of a registered function.
type FunctionFlags struct {
	Deterministic bool
}

// nArgFor returns the SQLite "nArg" value for the given function.
// SQLite uses -1 to mean "any number of arguments" (variadic).
func nArgFor(ft reflect.Type) int {
	if ft.IsVariadic() {
		return -1
	}
	return ft.NumIn()
}

func flagBits(opts FunctionFlags) sqlite3.FunctionFlag {
	var f sqlite3.FunctionFlag
	if opts.Deterministic {
		f |= sqlite3.DETERMINISTIC
	}
	return f
}

// RegisterFunc registers fn under name on conn. fn must be a Go
// function whose final return value is `error` (or whose only return
// is the result value). Arguments are decoded from sqlite3.Value into
// the Go types declared by fn's signature; the return value is encoded
// back through ctx.Result*.
func RegisterFunc(conn *sqlite3.Conn, name string, fn any, opts FunctionFlags) error {
	fv := reflect.ValueOf(fn)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		return fmt.Errorf("sqlitex: RegisterFunc(%s): not a function (%s)", name, ft.Kind())
	}
	if err := validateReturn(ft); err != nil {
		return fmt.Errorf("sqlitex: RegisterFunc(%s): %w", name, err)
	}

	nArg := nArgFor(ft)
	flag := flagBits(opts)

	return conn.CreateFunction(name, nArg, flag, func(ctx sqlite3.Context, vargs ...sqlite3.Value) {
		in, err := buildArgs(ft, vargs)
		if err != nil {
			ctx.ResultError(err)
			return
		}
		out := fv.Call(in)
		if err := callError(out); err != nil {
			ctx.ResultError(err)
			return
		}
		applyResult(ctx, callValue(out))
	})
}

// RegisterAggregator registers a plain (non-window) aggregate
// function. ctor must be a Go function with signature `func() T` where
// T has Step and Done methods matching the predecessor's shape
// (variadic interface{} args; interface{} result on Done).
//
// Registration goes through ncruces' CreateWindowFunction rather than
// CreateAggregateFunction. CreateAggregateFunction's iter.Seq wrapper
// drives the user function from a coroutine that is started lazily on
// the first Step: when a group has zero input rows SQLite invokes only
// the final callback, which stops the never-started coroutine without
// ever running the user function. The aggregate then never calls
// ctx.Result*, so SQLite reports NULL — e.g. COUNT(*) over an empty
// table yields NULL instead of 0.
//
// CreateWindowFunction has no such gap: ncruces constructs a fresh
// aggregate instance and calls Value even when Step never ran, so an
// empty group still produces each aggregate's zero value (COUNT -> 0,
// SUM -> NULL, ...). windowAggSimpleAdapter forwards Step directly to
// the instance (O(1) memory, unlike the buffered window adapter) and
// calls Done on Value. It implements no Inverse method, so it is
// registered as a plain aggregate; functions registered here are only
// ever emitted by the formatter for non-OVER aggregates (window use
// routes to the separately registered googlesqlite_window_* variants),
// so SQLite never drives the inverse callback on them.
func RegisterAggregator(conn *sqlite3.Conn, name string, ctor any, opts FunctionFlags) error {
	cv := reflect.ValueOf(ctor)
	if !cv.IsValid() || cv.Kind() != reflect.Func {
		return fmt.Errorf("sqlitex: RegisterAggregator(%s): ctor is not a function", name)
	}
	ct := cv.Type()
	if ct.NumIn() != 0 || ct.NumOut() != 1 {
		return fmt.Errorf("sqlitex: RegisterAggregator(%s): ctor must be func() T", name)
	}
	// Probe one instance to learn Step's argument count.
	probe := cv.Call(nil)[0].Interface()
	stepMethod, _, err := lookupStepDone(probe)
	if err != nil {
		return fmt.Errorf("sqlitex: RegisterAggregator(%s): %w", name, err)
	}
	stepNArg := stepArgCount(stepMethod.Type())

	flag := flagBits(opts)
	return conn.CreateWindowFunction(name, stepNArg, flag, sqlite3.AggregateConstructor(func() sqlite3.AggregateFunction {
		inst := cv.Call(nil)[0].Interface()
		return &windowAggSimpleAdapter{inst: inst}
	}))
}

// RegisterWindow registers a window-aggregate function backed by
// ncruces' CreateWindowFunction. Unlike RegisterAggregator (which
// runs a function once per partition with the entire scan), this lets
// the SQLite engine drive Step/Value per row, eliminating the
// per-output-row full-scan emulation the predecessor needed.
//
// ctor must be a Go function `func() T` where T satisfies the
// predecessor's aggregator shape:
//
//	Step(args ...interface{}) error
//	Done() (interface{}, error)
//
// If T additionally has an Inverse(args ...interface{}) error method,
// the registration becomes a true window function and SQLite can use
// frame-relative incremental updates for moving frames.
func RegisterWindow(conn *sqlite3.Conn, name string, ctor any, opts FunctionFlags) error {
	cv := reflect.ValueOf(ctor)
	if !cv.IsValid() || cv.Kind() != reflect.Func {
		return fmt.Errorf("sqlitex: RegisterWindow(%s): ctor is not a function", name)
	}
	ct := cv.Type()
	if ct.NumIn() != 0 || ct.NumOut() != 1 {
		return fmt.Errorf("sqlitex: RegisterWindow(%s): ctor must be func() T", name)
	}
	probe := cv.Call(nil)[0].Interface()
	stepMethod, _, err := lookupStepDone(probe)
	if err != nil {
		return fmt.Errorf("sqlitex: RegisterWindow(%s): %w", name, err)
	}
	hasInverse := reflect.ValueOf(probe).MethodByName("Inverse").IsValid()
	stepNArg := stepArgCount(stepMethod.Type())

	flag := flagBits(opts)
	return conn.CreateWindowFunction(name, stepNArg, flag, sqlite3.AggregateConstructor(func() sqlite3.AggregateFunction {
		if hasInverse {
			inst := cv.Call(nil)[0].Interface()
			return &windowAggAdapter{windowAggSimpleAdapter: windowAggSimpleAdapter{inst: inst}}
		}
		return newBufferedWindowAdapter(cv)
	}))
}

// windowAggSimpleAdapter wraps an aggregator with Step + Done,
// forwarding each Step straight to the instance and producing the
// result from Done on Value. It carries no Inverse method, so ncruces
// registers it as a plain aggregate.
//
// It backs two registration paths:
//
//   - RegisterAggregator, for plain (non-OVER) aggregates. These never
//     receive an inverse callback from SQLite.
//   - RegisterWindow, but only when the wrapped instance ALSO has an
//     Inverse method (then windowAggAdapter embeds this and adds
//     Inverse). A window aggregator without Inverse instead goes
//     through newBufferedWindowAdapter, which replays buffered Step
//     args, because SQLite drives the inverse callback for moving
//     frames and a plain Step+Done aggregator cannot undo a Step.
type windowAggSimpleAdapter struct {
	inst    any
	stepErr error
}

func (a *windowAggSimpleAdapter) Step(_ sqlite3.Context, args ...sqlite3.Value) {
	if a.stepErr != nil {
		return
	}
	step := reflect.ValueOf(a.inst).MethodByName("Step")
	rets := callMethod(step, decodeArgs(args))
	if err := lastError(rets); err != nil {
		a.stepErr = err
	}
}

func (a *windowAggSimpleAdapter) Value(ctx sqlite3.Context) {
	if a.stepErr != nil {
		ctx.ResultError(a.stepErr)
		return
	}
	done := reflect.ValueOf(a.inst).MethodByName("Done")
	out := done.Call(nil)
	if err := lastError(out); err != nil {
		ctx.ResultError(err)
		return
	}
	if len(out) == 0 {
		ctx.ResultNull()
		return
	}
	applyResult(ctx, out[0].Interface())
}

// windowAggAdapter wraps a Step/Inverse/Done aggregator. ncruces
// promotes such an instance to a true window function, letting SQLite
// drive incremental frame updates.
type windowAggAdapter struct {
	windowAggSimpleAdapter
}

func (a *windowAggAdapter) Inverse(_ sqlite3.Context, args ...sqlite3.Value) {
	if a.stepErr != nil {
		return
	}
	inv := reflect.ValueOf(a.inst).MethodByName("Inverse")
	rets := callMethod(inv, decodeArgs(args))
	if err := lastError(rets); err != nil {
		a.stepErr = err
	}
}

// bufferedWindowAdapter satisfies sqlite3.WindowFunction even when the
// underlying Go aggregator only knows how to do Step + Done. It
// records every Step's arguments verbatim, drops the matching args
// on Inverse, and rebuilds the aggregator from scratch on each
// Value() call by replaying the buffered Step args.
//
// The cost vs a true Inverse-supporting aggregator is O(frame_size)
// per Value() instead of O(1). Compared with the predecessor's
// correlated-subquery emulation that scanned the entire partition
// per output row, this is still a substantial win for typical
// frame sizes.
type bufferedWindowAdapter struct {
	ctor    reflect.Value
	frame   [][]any
	stepErr error
}

func newBufferedWindowAdapter(ctor reflect.Value) *bufferedWindowAdapter {
	return &bufferedWindowAdapter{ctor: ctor}
}

func (a *bufferedWindowAdapter) Step(_ sqlite3.Context, args ...sqlite3.Value) {
	if a.stepErr != nil {
		return
	}
	// Snapshot the args; ncruces invalidates the underlying slice
	// after Step returns.
	snapshot := make([]any, len(args))
	for i, v := range args {
		snapshot[i] = decodeValue(v)
	}
	a.frame = append(a.frame, snapshot)
}

func (a *bufferedWindowAdapter) Inverse(_ sqlite3.Context, _ ...sqlite3.Value) {
	if a.stepErr != nil {
		return
	}
	if len(a.frame) == 0 {
		return
	}
	// SQLite always pops the oldest row when sliding the frame
	// forward. The args passed to Inverse are the same as the
	// matching Step (verified by SQLite); we can therefore drop the
	// front entry without comparing.
	a.frame = a.frame[1:]
}

func (a *bufferedWindowAdapter) Value(ctx sqlite3.Context) {
	if a.stepErr != nil {
		ctx.ResultError(a.stepErr)
		return
	}
	inst := a.ctor.Call(nil)[0].Interface()
	step := reflect.ValueOf(inst).MethodByName("Step")
	for _, row := range a.frame {
		rv := make([]reflect.Value, len(row))
		for i, v := range row {
			if v == nil {
				rv[i] = reflect.Zero(emptyIfaceType)
			} else {
				rv[i] = reflect.ValueOf(v)
			}
		}
		rets := callMethod(step, rv)
		if err := lastError(rets); err != nil {
			ctx.ResultError(err)
			return
		}
	}
	done := reflect.ValueOf(inst).MethodByName("Done")
	out := done.Call(nil)
	if err := lastError(out); err != nil {
		ctx.ResultError(err)
		return
	}
	if len(out) == 0 {
		ctx.ResultNull()
		return
	}
	applyResult(ctx, out[0].Interface())
}

// RegisterCollation installs a string collation under name.
func RegisterCollation(conn *sqlite3.Conn, name string, fn CollatingFunc) error {
	return conn.CreateCollation(name, func(a, b []byte) int {
		return fn(string(a), string(b))
	})
}

// SetVariableNumberLimit removes the parameter-count cap on the given
// connection. The predecessor relied on SQLite's
// SQLITE_LIMIT_VARIABLE_NUMBER being effectively unbounded for some of
// its codegen paths. Negative values request the engine maximum.
func SetVariableNumberLimit(conn *sqlite3.Conn, value int) {
	conn.Limit(sqlite3.LIMIT_VARIABLE_NUMBER, value)
}

// --- internal helpers -------------------------------------------------

func validateReturn(ft reflect.Type) error {
	switch ft.NumOut() {
	case 1:
		// (T) — fine. error counts as T here too.
		return nil
	case 2:
		// Must be (T, error).
		if !ft.Out(1).Implements(errorType) {
			return fmt.Errorf("two-return form requires last return to be error, got %s", ft.Out(1))
		}
		return nil
	}
	return fmt.Errorf("must return 1 or 2 values, got %d", ft.NumOut())
}

func buildArgs(ft reflect.Type, vargs []sqlite3.Value) ([]reflect.Value, error) {
	if ft.IsVariadic() {
		// Variadic param is the last input; the inner element type is
		// the variadic slice's element. We pass each value as that
		// element type.
		variadicElem := ft.In(ft.NumIn() - 1).Elem()
		fixed := ft.NumIn() - 1
		if len(vargs) < fixed {
			return nil, fmt.Errorf("expected at least %d args, got %d", fixed, len(vargs))
		}
		in := make([]reflect.Value, 0, len(vargs))
		for i := range fixed {
			rv, err := decodeValueAs(vargs[i], ft.In(i))
			if err != nil {
				return nil, err
			}
			in = append(in, rv)
		}
		for i := fixed; i < len(vargs); i++ {
			rv, err := decodeValueAs(vargs[i], variadicElem)
			if err != nil {
				return nil, err
			}
			in = append(in, rv)
		}
		return in, nil
	}
	if len(vargs) != ft.NumIn() {
		return nil, fmt.Errorf("expected %d args, got %d", ft.NumIn(), len(vargs))
	}
	in := make([]reflect.Value, ft.NumIn())
	for i := 0; i < ft.NumIn(); i++ {
		rv, err := decodeValueAs(vargs[i], ft.In(i))
		if err != nil {
			return nil, err
		}
		in[i] = rv
	}
	return in, nil
}

func decodeValueAs(v sqlite3.Value, target reflect.Type) (reflect.Value, error) {
	val := decodeValue(v)
	if val == nil {
		// Pass nil — only valid for interface targets and pointer types.
		switch target.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Slice, reflect.Map:
			return reflect.Zero(target), nil
		}
		// For non-nilable targets, pass the zero value. Callers that
		// care about NULL must take interface{} arguments (the
		// predecessor's universal pattern).
		return reflect.Zero(target), nil
	}
	rv := reflect.ValueOf(val)
	if rv.Type().AssignableTo(target) {
		return rv, nil
	}
	if rv.Type().ConvertibleTo(target) {
		return rv.Convert(target), nil
	}
	if target.Kind() == reflect.Interface && target.NumMethod() == 0 {
		return rv, nil
	}
	return reflect.Value{}, fmt.Errorf("cannot pass %s as %s", rv.Type(), target)
}

// decodeValue mirrors mattn's behavior: decode to nil / int64 / float64
// / string / []byte. The predecessor's bind functions immediately call
// DecodeValue on each interface{} argument, so we only need to deliver
// values that DecodeValue can interpret.
func decodeValue(v sqlite3.Value) any {
	switch v.Type() {
	case sqlite3.NULL:
		return nil
	case sqlite3.INTEGER:
		return v.Int64()
	case sqlite3.FLOAT:
		return v.Float()
	case sqlite3.TEXT:
		return v.Text()
	case sqlite3.BLOB:
		// Make a defensive copy: ncruces docs warn that the slice
		// returned by RawBlob/Blob is invalidated when the connection
		// returns to user code.
		return append([]byte(nil), v.RawBlob()...)
	}
	return nil
}

func decodeArgs(vargs []sqlite3.Value) []reflect.Value {
	out := make([]reflect.Value, len(vargs))
	for i, v := range vargs {
		out[i] = reflect.ValueOf(decodeValue(v))
		// reflect.ValueOf(nil) yields the zero Value, which Call
		// rejects. Substitute a typed nil interface{}.
		if !out[i].IsValid() {
			out[i] = reflect.Zero(emptyIfaceType)
		}
	}
	return out
}

func callMethod(method reflect.Value, args []reflect.Value) []reflect.Value {
	mt := method.Type()
	if !mt.IsVariadic() && len(args) != mt.NumIn() {
		// Allow trailing zero-fill or truncation: the predecessor's
		// Step is variadic-interface so this branch is rarely hit.
		// Treat mismatch as a programming error.
		return []reflect.Value{reflect.ValueOf(fmt.Errorf(
			"sqlitex: aggregate Step expected %d args, got %d", mt.NumIn(), len(args),
		))}
	}
	return method.Call(args)
}

func applyResult(ctx sqlite3.Context, v any) {
	switch x := v.(type) {
	case nil:
		ctx.ResultNull()
	case bool:
		ctx.ResultBool(x)
	case int:
		ctx.ResultInt64(int64(x))
	case int8:
		ctx.ResultInt64(int64(x))
	case int16:
		ctx.ResultInt64(int64(x))
	case int32:
		ctx.ResultInt64(int64(x))
	case int64:
		ctx.ResultInt64(x)
	case uint:
		ctx.ResultInt64(int64(x))
	case uint8:
		ctx.ResultInt64(int64(x))
	case uint16:
		ctx.ResultInt64(int64(x))
	case uint32:
		ctx.ResultInt64(int64(x))
	case uint64:
		ctx.ResultInt64(int64(x))
	case float32:
		ctx.ResultFloat(float64(x))
	case float64:
		ctx.ResultFloat(x)
	case string:
		ctx.ResultText(x)
	case []byte:
		ctx.ResultBlob(x)
	case error:
		ctx.ResultError(x)
	default:
		ctx.ResultError(fmt.Errorf("sqlitex: cannot encode result of type %T", v))
	}
}

func callError(out []reflect.Value) error {
	if len(out) == 0 {
		return nil
	}
	last := out[len(out)-1]
	if !last.Type().Implements(errorType) {
		return nil
	}
	if last.IsNil() {
		return nil
	}
	return last.Interface().(error)
}

func callValue(out []reflect.Value) any {
	if len(out) == 0 {
		return nil
	}
	first := out[0]
	if first.Type().Implements(errorType) {
		// Single error return: nothing to encode.
		return nil
	}
	return first.Interface()
}

func lastError(out []reflect.Value) error {
	if len(out) == 0 {
		return nil
	}
	last := out[len(out)-1]
	if !last.Type().Implements(errorType) {
		return nil
	}
	if last.IsNil() {
		return nil
	}
	return last.Interface().(error)
}

func lookupStepDone(instance any) (step, done reflect.Value, err error) {
	v := reflect.ValueOf(instance)
	if !v.IsValid() {
		return reflect.Value{}, reflect.Value{}, errors.New("aggregator instance is nil")
	}
	step = v.MethodByName("Step")
	done = v.MethodByName("Done")
	if !step.IsValid() {
		return reflect.Value{}, reflect.Value{}, errors.New("aggregator missing Step method")
	}
	if !done.IsValid() {
		return reflect.Value{}, reflect.Value{}, errors.New("aggregator missing Done method")
	}
	return step, done, nil
}

func stepArgCount(stepType reflect.Type) int {
	if stepType.IsVariadic() {
		return -1
	}
	return stepType.NumIn()
}

var (
	errorType      = reflect.TypeFor[error]()
	emptyIfaceType = reflect.TypeFor[any]()
)
