package internal

import (
	geofn "github.com/goccy/googlesqlite/internal/functions/geography"
	textfn "github.com/goccy/googlesqlite/internal/functions/text_analysis"
	"github.com/goccy/googlesqlite/internal/functions/window"
)

// windowFuncs lists every analytic / window function the driver
// registers via Conn.CreateWindowFunction. The native registration
// path replaces the predecessor going-through-emulation hot loop;
// see the bench writeup for the impact.

var windowFuncs = []*windowFuncInfo{
	// aggregate functions
	{Name: "any_value", BindFunc: window.BindWindowAnyValue},
	{Name: "array_agg", BindFunc: window.BindWindowArrayAgg},
	{Name: "avg", BindFunc: window.BindWindowAvg},
	{Name: "count", BindFunc: window.BindWindowCount},
	{Name: "count_star", BindFunc: window.BindWindowCountStar},
	{Name: "countif", BindFunc: window.BindWindowCountIf},
	{Name: "max", BindFunc: window.BindWindowMax},
	{Name: "min", BindFunc: window.BindWindowMin},
	{Name: "logical_or", BindFunc: window.BindWindowLogicalOr},
	{Name: "logical_and", BindFunc: window.BindWindowLogicalAnd},
	{Name: "string_agg", BindFunc: window.BindWindowStringAgg},
	{Name: "sum", BindFunc: window.BindWindowSum},

	// statistical aggregate functions
	{Name: "corr", BindFunc: window.BindWindowCorr},
	{Name: "covar_pop", BindFunc: window.BindWindowCovarPop},
	{Name: "covar_samp", BindFunc: window.BindWindowCovarSamp},
	{Name: "stddev_pop", BindFunc: window.BindWindowStddevPop},
	{Name: "stddev_samp", BindFunc: window.BindWindowStddevSamp},
	{Name: "stddev", BindFunc: window.BindWindowStddev},
	{Name: "var_pop", BindFunc: window.BindWindowVarPop},
	{Name: "var_samp", BindFunc: window.BindWindowVarSamp},
	{Name: "variance", BindFunc: window.BindWindowVariance},

	// navigation functions
	{Name: "first_value", BindFunc: window.BindWindowFirstValue},
	{Name: "last_value", BindFunc: window.BindWindowLastValue},
	{Name: "nth_value", BindFunc: window.BindWindowNthValue},
	{Name: "lead", BindFunc: window.BindWindowLead},
	{Name: "lag", BindFunc: window.BindWindowLag},
	{Name: "percentile_cont", BindFunc: window.BindWindowPercentileCont},
	{Name: "percentile_disc", BindFunc: window.BindWindowPercentileDisc},

	// text_analysis (BigQuery-only window aggregator)
	{Name: "tf_idf", BindFunc: textfn.BindTfIdf},

	// geography
	{Name: "st_clusterdbscan", BindFunc: geofn.BindWindowStClusterDBSCAN},

	// numbering functions
	{Name: "rank", BindFunc: window.BindWindowRank},
	{Name: "dense_rank", BindFunc: window.BindWindowDenseRank},
	{Name: "percent_rank", BindFunc: window.BindWindowPercentRank},
	{Name: "cume_dist", BindFunc: window.BindWindowCumeDist},
	{Name: "ntile", BindFunc: window.BindWindowNtile},
	{Name: "row_number", BindFunc: window.BindWindowRowNumber},
}
