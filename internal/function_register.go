package internal

import (
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	sqlite3 "github.com/ncruces/go-sqlite3"

	"github.com/goccy/googlesqlite/internal/functions/window"
	"github.com/goccy/googlesqlite/internal/sqlitex"
	"github.com/goccy/googlesqlite/internal/value"
)

type nameAndFunc struct {
	Name string
	Func any
	// NonDeterministic, when true, drops SQLITE_DETERMINISTIC at
	// registration so SQLite re-evaluates the call per row.
	NonDeterministic bool
}

var (
	funcMapMu          sync.RWMutex
	registerFuncOnce   sync.Once
	normalFuncMap      = map[string][]*nameAndFunc{}
	aggregateFuncMap   = map[string][]*nameAndFunc{}
	windowFuncMap      = map[string][]*nameAndFunc{}
	currentTimeFuncMap = map[string]struct{}{
		"current_date":      {},
		"current_datetime":  {},
		"current_time":      {},
		"current_timestamp": {},
	}
)

func RegisterFunctions(conn *sqlite3.Conn) error {
	funcMapMu.RLock()
	defer funcMapMu.RUnlock()

	var onceErr error
	registerFuncOnce.Do(func() {
		for _, info := range normalFuncs {
			setupNormalFuncMap(info)
		}
		for _, info := range aggregateFuncs {
			setupAggregateFuncMap(info)
		}
		for _, info := range windowFuncs {
			setupWindowFuncMap(info)
		}
		// Replace the predecessor's per-row full-scan emulation for
		// the window forms of these aggregators with native
		// incremental implementations. The formatter emits these
		// through OVER syntax in customNativeWindowFuncMap, so SQLite
		// drives Step/Done over the active frame instead of the
		// O(N²) correlated subquery the predecessor required.
		windowFuncMap["array_agg"] = []*nameAndFunc{
			{Name: "googlesqlite_window_array_agg", Func: window.NewArrayAggWindowNative()},
		}
		windowFuncMap["string_agg"] = []*nameAndFunc{
			{Name: "googlesqlite_window_string_agg", Func: window.NewStringAggWindowNative()},
		}
		windowFuncMap["countif"] = []*nameAndFunc{
			{Name: "googlesqlite_window_countif", Func: window.NewCountifWindowNative()},
		}
		windowFuncMap["count_star"] = []*nameAndFunc{
			{Name: "googlesqlite_window_count_star", Func: window.NewCountStarWindowNative()},
		}
		windowFuncMap["any_value"] = []*nameAndFunc{
			{Name: "googlesqlite_window_any_value", Func: window.NewAnyValueWindowNative()},
		}
		windowFuncMap["corr"] = []*nameAndFunc{
			{Name: "googlesqlite_window_corr", Func: window.NewCorrWindowNative()},
		}
		windowFuncMap["covar_pop"] = []*nameAndFunc{
			{Name: "googlesqlite_window_covar_pop", Func: window.NewCovarPopWindowNative()},
		}
		windowFuncMap["covar_samp"] = []*nameAndFunc{
			{Name: "googlesqlite_window_covar_samp", Func: window.NewCovarSampWindowNative()},
		}
		windowFuncMap["stddev_pop"] = []*nameAndFunc{
			{Name: "googlesqlite_window_stddev_pop", Func: window.NewStddevPopWindowNative()},
		}
		windowFuncMap["stddev_samp"] = []*nameAndFunc{
			{Name: "googlesqlite_window_stddev_samp", Func: window.NewStddevSampWindowNative()},
		}
		windowFuncMap["var_pop"] = []*nameAndFunc{
			{Name: "googlesqlite_window_var_pop", Func: window.NewVarPopWindowNative()},
		}
		windowFuncMap["var_samp"] = []*nameAndFunc{
			{Name: "googlesqlite_window_var_samp", Func: window.NewVarSampWindowNative()},
		}
		// DISTINCT-aware window aggregators (SUM/COUNT/AVG). These
		// don't replace existing entries — they're in addition,
		// keyed under their custom registration names so they're
		// looked up by the formatter's distinctAwareNativeWindowFuncs
		// path rather than via predecessor name.
		windowFuncMap["sum_distinct"] = []*nameAndFunc{
			{Name: "googlesqlite_window_sum_distinct", Func: window.NewSumDistinctWindowNative()},
		}
		windowFuncMap["count_distinct"] = []*nameAndFunc{
			{Name: "googlesqlite_window_count_distinct", Func: window.NewCountDistinctWindowNative()},
		}
		windowFuncMap["avg_distinct"] = []*nameAndFunc{
			{Name: "googlesqlite_window_avg_distinct", Func: window.NewAvgDistinctWindowNative()},
		}
		windowFuncMap["percentile_cont"] = []*nameAndFunc{
			{Name: "googlesqlite_window_percentile_cont", Func: window.NewPercentileContWindowNative()},
		}
		windowFuncMap["percentile_disc"] = []*nameAndFunc{
			{Name: "googlesqlite_window_percentile_disc", Func: window.NewPercentileDiscWindowNative()},
		}
		// Native window forms for the boolean / bitwise aggregators
		// that previously fell through to the predecessor's
		// correlated-subquery emulation.
		windowFuncMap["logical_or"] = []*nameAndFunc{
			{Name: "googlesqlite_window_logical_or", Func: window.NewLogicalOrWindowNative()},
		}
		windowFuncMap["logical_and"] = []*nameAndFunc{
			{Name: "googlesqlite_window_logical_and", Func: window.NewLogicalAndWindowNative()},
		}
		windowFuncMap["bit_and"] = []*nameAndFunc{
			{Name: "googlesqlite_window_bit_and", Func: window.NewBitAndAggWindowNative()},
		}
		windowFuncMap["bit_or"] = []*nameAndFunc{
			{Name: "googlesqlite_window_bit_or", Func: window.NewBitOrAggWindowNative()},
		}
		windowFuncMap["bit_xor"] = []*nameAndFunc{
			{Name: "googlesqlite_window_bit_xor", Func: window.NewBitXorAggWindowNative()},
		}
		windowFuncMap["array_concat_agg"] = []*nameAndFunc{
			{Name: "googlesqlite_window_array_concat_agg", Func: window.NewArrayConcatAggWindowNative()},
		}
	})
	if onceErr != nil {
		return onceErr
	}

	deterministic := sqlitex.FunctionFlags{Deterministic: true}

	if err := sqlitex.RegisterFunc(conn, "googlesqlite_decode_array", func(v any) (string, error) {
		decoded, err := DecodeValue(v)
		if err != nil {
			return "", err
		}
		if decoded == nil {
			return "[]", nil
		}
		array, err := decoded.ToArray()
		if err != nil {
			return "", err
		}
		encodedValues := make([]any, 0, len(array.Values))
		for _, value := range array.Values {
			v, err := EncodeValue(value)
			if err != nil {
				return "", err
			}
			encodedValues = append(encodedValues, v)
		}
		b, err := json.Marshal(encodedValues)
		if err != nil {
			return "", err
		}
		return string(b), err
	}, deterministic); err != nil {
		return fmt.Errorf("failed to register decode_array function: %w", err)
	}

	if err := sqlitex.RegisterFunc(conn, "googlesqlite_group_by", func(v any) (any, error) {
		decoded, err := DecodeValue(v)
		if err != nil {
			return "", err
		}
		if decoded == nil {
			return nil, nil
		}
		return decoded.Interface(), nil
	}, deterministic); err != nil {
		return fmt.Errorf("failed to register group_by function: %w", err)
	}

	if err := sqlitex.RegisterCollation(conn, "googlesqlite_collate", func(a, b string) int {
		va, _ := DecodeValue(a)
		vb, _ := DecodeValue(b)
		eq, _ := va.EQ(vb)
		if eq {
			return 0
		}
		cond, _ := va.GT(vb)
		if cond {
			return 1
		}
		return -1
	}); err != nil {
		return fmt.Errorf("failed to register collate function: %w", err)
	}

	for _, values := range normalFuncMap {
		for _, v := range values {
			flags := deterministic
			if v.NonDeterministic {
				flags = sqlitex.FunctionFlags{}
			}
			if err := sqlitex.RegisterFunc(conn, v.Name, v.Func, flags); err != nil {
				return fmt.Errorf("failed to register function %s: %w", v.Name, err)
			}
		}
	}
	for _, values := range aggregateFuncMap {
		for _, v := range values {
			if err := sqlitex.RegisterAggregator(conn, v.Name, v.Func, deterministic); err != nil {
				return fmt.Errorf("failed to register aggregate function %s: %w", v.Name, err)
			}
		}
	}
	// Window-aggregate functions register through CreateWindowFunction
	// so they participate in native OVER frame iteration. Functions
	// that the formatter still routes via the predecessor's
	// correlated-subquery emulation (DISTINCT, IGNORE NULLS,
	// non-native ARRAY_AGG/STDDEV/etc.) keep working because
	// CreateWindowFunction also accepts plain aggregate-shaped
	// instances.
	for _, values := range windowFuncMap {
		for _, v := range values {
			if err := sqlitex.RegisterWindow(conn, v.Name, v.Func, deterministic); err != nil {
				return fmt.Errorf("failed to register window function %s: %w", v.Name, err)
			}
		}
	}
	return nil
}

func setupNormalFuncMap(info *funcInfo) {
	normalFuncMap[info.Name] = append(normalFuncMap[info.Name], &nameAndFunc{
		Name: fmt.Sprintf("googlesqlite_%s", info.Name),
		Func: func(args ...any) (any, error) {
			values, err := value.ConvertArgs(args...)
			if err != nil {
				return nil, err
			}
			ret, err := info.BindFunc(values...)
			if err != nil {
				return nil, err
			}
			return EncodeValue(ret)
		},
		NonDeterministic: info.NonDeterministic,
	}, &nameAndFunc{
		Name: fmt.Sprintf("googlesqlite_safe_%s", info.Name),
		Func: func(args ...any) (any, error) {
			values, err := value.ConvertArgs(args...)
			if err != nil {
				return nil, err
			}
			ret, err := info.BindFunc(values...)
			if err != nil {
				// Note, this should only suppress semantic errors based on the
				// input data. See
				// https://github.com/google/googlesql/blob/master/docs/resolved_ast.md#resolvedfunctioncallbase
				return nil, nil
			}
			return EncodeValue(ret)
		},
		NonDeterministic: info.NonDeterministic,
	})
}

func setupAggregateFuncMap(info *aggregateFuncInfo) {
	aggregateFuncMap[info.Name] = append(aggregateFuncMap[info.Name], &nameAndFunc{
		Name: fmt.Sprintf("googlesqlite_%s", info.Name),
		Func: info.BindFunc(),
	})
}

func setupWindowFuncMap(info *windowFuncInfo) {
	windowFuncMap[info.Name] = append(windowFuncMap[info.Name], &nameAndFunc{
		Name: fmt.Sprintf("googlesqlite_window_%s", info.Name),
		Func: info.BindFunc(),
	})
}
