package internal

import (
	"github.com/goccy/googlesqlite/internal/functions/aggregate"
	approx "github.com/goccy/googlesqlite/internal/functions/approx_aggregate"
	geofn "github.com/goccy/googlesqlite/internal/functions/geography"
	"github.com/goccy/googlesqlite/internal/functions/hll"
	"github.com/goccy/googlesqlite/internal/functions/kll"
	stat "github.com/goccy/googlesqlite/internal/functions/stat_aggregate"
)

// aggregateFuncs lists every aggregate function the driver registers
// through sqlitex.RegisterAggregator. The functions are grouped by
// their GoogleSQL reference category page:
//
//   - aggregate_functions.md
//   - statistical_aggregate_functions.md
//   - approximate_aggregate_functions.md
//   - hll_functions.md
//
// Each category lives in its own internal/function/<category> package
// so the boundary on disk matches the boundary in the docs.

var aggregateFuncs = []*aggregateFuncInfo{
	{Name: "array", BindFunc: aggregate.BindArray},

	// aggregate_functions
	{Name: "any_value", BindFunc: aggregate.BindAnyValue},
	{Name: "agg", BindFunc: aggregate.BindAGG},
	// having_any_value emulates the HAVING { MAX | MIN } modifier on
	// ANY_VALUE: the formatter rewrites
	//   ANY_VALUE(value HAVING MAX/MIN having_expr)
	// to
	//   googlesqlite_having_any_value(value, having_expr, 'MAX'|'MIN')
	// because go-googlesql's resolved tree exposes the modifier but
	// SQLite has no native equivalent. The setupAggregateFuncMap
	// prefix turns the Name below into googlesqlite_having_any_value
	// at registration time.
	{Name: "having_any_value", BindFunc: aggregate.BindHavingAnyValue},
	// Differential Privacy / Anonymization internal builtins. The
	// analyzer's Anonymization rewriter lowers SELECT WITH
	// DIFFERENTIAL_PRIVACY / WITH ANONYMIZATION queries to
	// $differential_privacy_<op> / $anon_<op> aggregate calls.
	// Our formatter strips the leading `$` and prefixes with
	// `googlesqlite_` (see internal/formatter.go), so these
	// runtime UDFs back both surfaces.
	{Name: "differential_privacy_sum", BindFunc: aggregate.BindDifferentialPrivacySum},
	{Name: "differential_privacy_count", BindFunc: aggregate.BindDifferentialPrivacyCount},
	{Name: "differential_privacy_count_star", BindFunc: aggregate.BindDifferentialPrivacyCountStar},
	{Name: "differential_privacy_avg", BindFunc: aggregate.BindDifferentialPrivacyAvg},
	{Name: "differential_privacy_var_pop", BindFunc: aggregate.BindDifferentialPrivacyVarPop},
	{Name: "differential_privacy_stddev_pop", BindFunc: aggregate.BindDifferentialPrivacyStddevPop},
	{Name: "differential_privacy_approx_count_distinct", BindFunc: aggregate.BindDifferentialPrivacyApproxCountDistinct},
	{Name: "anon_sum", BindFunc: aggregate.BindDifferentialPrivacySum},
	{Name: "anon_count", BindFunc: aggregate.BindDifferentialPrivacyCount},
	{Name: "anon_count_star", BindFunc: aggregate.BindDifferentialPrivacyCountStar},
	{Name: "anon_avg", BindFunc: aggregate.BindDifferentialPrivacyAvg},
	{Name: "anon_var_pop", BindFunc: aggregate.BindDifferentialPrivacyVarPop},
	{Name: "anon_stddev_pop", BindFunc: aggregate.BindDifferentialPrivacyStddevPop},
	{Name: "anon_percentile_cont", BindFunc: aggregate.BindDifferentialPrivacyPercentileCont},
	{Name: "anon_quantiles", BindFunc: aggregate.BindDifferentialPrivacyQuantiles},
	{Name: "differential_privacy_percentile_cont", BindFunc: aggregate.BindDifferentialPrivacyPercentileCont},
	{Name: "differential_privacy_quantiles", BindFunc: aggregate.BindDifferentialPrivacyQuantiles},
	{Name: "array_agg", BindFunc: aggregate.BindArrayAgg},
	{Name: "array_concat_agg", BindFunc: aggregate.BindArrayConcatAgg},
	{Name: "avg", BindFunc: aggregate.BindAvg},
	{Name: "count", BindFunc: aggregate.BindCount},
	{Name: "count_star", BindFunc: aggregate.BindCountStar},
	{Name: "bit_and", BindFunc: aggregate.BindBitAndAgg},
	{Name: "bit_or", BindFunc: aggregate.BindBitOrAgg},
	{Name: "bit_xor", BindFunc: aggregate.BindBitXorAgg},
	{Name: "countif", BindFunc: aggregate.BindCountIf},
	{Name: "logical_and", BindFunc: aggregate.BindLogicalAnd},
	{Name: "logical_or", BindFunc: aggregate.BindLogicalOr},
	{Name: "max", BindFunc: aggregate.BindMax},
	{Name: "max_by", BindFunc: aggregate.BindMaxBy},
	{Name: "min", BindFunc: aggregate.BindMin},
	{Name: "min_by", BindFunc: aggregate.BindMinBy},
	{Name: "string_agg", BindFunc: aggregate.BindStringAgg},
	{Name: "sum", BindFunc: aggregate.BindSum},

	// statistical_aggregate_functions
	{Name: "corr", BindFunc: stat.BindCorr},
	{Name: "covar_pop", BindFunc: stat.BindCovarPop},
	{Name: "covar_samp", BindFunc: stat.BindCovarSamp},
	{Name: "stddev_pop", BindFunc: stat.BindStddevPop},
	{Name: "stddev_samp", BindFunc: stat.BindStddevSamp},
	{Name: "stddev", BindFunc: stat.BindStddev},
	{Name: "var_pop", BindFunc: stat.BindVarPop},
	{Name: "var_samp", BindFunc: stat.BindVarSamp},
	{Name: "variance", BindFunc: stat.BindVariance},

	// approximate_aggregate_functions
	{Name: "approx_count_distinct", BindFunc: approx.BindApproxCountDistinct},
	{Name: "approx_quantiles", BindFunc: approx.BindApproxQuantiles},
	{Name: "approx_top_count", BindFunc: approx.BindApproxTopCount},
	{Name: "approx_top_sum", BindFunc: approx.BindApproxTopSum},

	// geography aggregates
	{Name: "st_centroid_agg", BindFunc: geofn.BindStCentroidAgg},
	{Name: "st_union_agg", BindFunc: geofn.BindStUnionAgg},
	{Name: "st_extent", BindFunc: geofn.BindStExtentAgg},

	// longtail HLL aggregates: HLL_COUNT.MERGE / .MERGE_PARTIAL.
	// The hll package's BindHllCountMerge variants do the heavy lift
	// (HyperLogLog state merge); these aliases reuse them under
	// alternate names the catalog exposes after analyzer rewrite.

	// KLL quantile sketches (BigQuery KLL_QUANTILES.*).
	// Names are the analyzer-canonicalised forms (lowercased,
	// dotted namespace flattened).
	{Name: "kll_quantiles_init_int64", BindFunc: kll.BindInit},
	{Name: "kll_quantiles_init_float64", BindFunc: kll.BindInit},
	{Name: "kll_quantiles_merge_int64", BindFunc: kll.BindMergeInt64},
	{Name: "kll_quantiles_merge_float64", BindFunc: kll.BindMergeFloat64},
	{Name: "kll_quantiles_merge_partial", BindFunc: kll.BindMergePartial},
	{Name: "kll_quantiles_merge_point_int64", BindFunc: kll.BindMergePointInt64},
	{Name: "kll_quantiles_merge_point_float64", BindFunc: kll.BindMergePointFloat64},

	// hll_functions
	{Name: "hll_count_init", BindFunc: hll.BindHllCountInit},
	{Name: "hll_count_merge", BindFunc: hll.BindHllCountMerge},
	{Name: "hll_count_merge_partial", BindFunc: hll.BindHllCountMergePartial},
}
