package internal

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	googlesql "github.com/goccy/go-googlesql"

	"github.com/goccy/googlesqlite/internal/exportdata"
	"github.com/goccy/googlesqlite/internal/value"
)

type Analyzer struct {
	namePath        *NamePath
	isAutoIndexMode bool
	isExplainMode   bool
	catalog         *Catalog
	opt             *googlesql.AnalyzerOptions
}

// sharedAnalyzerOpt is the process-wide singleton AnalyzerOptions
// every Analyzer reuses. The wasm runtime keeps each AnalyzerOptions
// alive in its scratch memory and the FFI calls that build it (~50
// LanguageFeature toggles, ProductMode, system-variable declarations,
// rewrite flags) are expensive. Allocating a fresh one per Conn used
// to push the wasm pool past its capacity under heavy parallel test
// load, surfacing as `NewLanguageOptions returned nil` or
// `<Invalid simple type's value>` from downstream FFI calls.
//
// All mutations to this options object — ParameterMode,
// ClearQueryParameters, AddQueryParameter, and AnalyzeStatement-
// FromParserAST itself — happen under wasmAnalyzeMu, so a single
// goroutine owns the wasm runtime for one Analyze call at a time.
var (
	sharedAnalyzerOptOnce sync.Once
	sharedAnalyzerOpt     *googlesql.AnalyzerOptions
	sharedAnalyzerOptErr  error
	wasmAnalyzeMu         sync.Mutex
)

func getSharedAnalyzerOptions() (*googlesql.AnalyzerOptions, error) {
	sharedAnalyzerOptOnce.Do(func() {
		sharedAnalyzerOpt, sharedAnalyzerOptErr = newAnalyzerOptions()
	})
	return sharedAnalyzerOpt, sharedAnalyzerOptErr
}

func NewAnalyzer(catalog *Catalog) (*Analyzer, error) {
	opt, err := getSharedAnalyzerOptions()
	if err != nil {
		return nil, err
	}
	return &Analyzer{
		catalog:  catalog,
		opt:      opt,
		namePath: &NamePath{},
	}, nil
}

// enabledLanguageFeatures lists every googlesql.LanguageFeature the
// analyzer turns on. It is pure data consumed by newAnalyzerOptions;
// the order is preserved from the historical inline literal.
var enabledLanguageFeatures = []googlesql.LanguageFeature{
	googlesql.LanguageFeatureFeatureAnalyticFunctions,
	googlesql.LanguageFeatureFeatureNamedArguments,
	googlesql.LanguageFeatureFeatureNumericType,
	googlesql.LanguageFeatureFeatureBignumericType,
	googlesql.LanguageFeatureFeatureV13DecimalAlias,
	googlesql.LanguageFeatureFeatureCreateTableNotNull,
	googlesql.LanguageFeatureFeatureParameterizedTypes,
	googlesql.LanguageFeatureFeatureTablesample,
	googlesql.LanguageFeatureFeatureTimestampNanos,
	googlesql.LanguageFeatureFeatureV11HavingInAggregate,
	googlesql.LanguageFeatureFeatureV11NullHandlingModifierInAggregate,
	googlesql.LanguageFeatureFeatureV11NullHandlingModifierInAnalytic,
	googlesql.LanguageFeatureFeatureV11OrderByCollate,
	googlesql.LanguageFeatureFeatureV11SelectStarExceptReplace,
	googlesql.LanguageFeatureFeatureV12SafeFunctionCall,
	googlesql.LanguageFeatureFeatureJsonType,
	googlesql.LanguageFeatureFeatureJsonArrayFunctions,
	googlesql.LanguageFeatureFeatureJsonStrictNumberParsing,
	googlesql.LanguageFeatureFeatureV13IsDistinct,
	googlesql.LanguageFeatureFeatureV13FormatInCast,
	googlesql.LanguageFeatureFeatureV13DateArithmetics,
	googlesql.LanguageFeatureFeatureV11OrderByInAggregate,
	googlesql.LanguageFeatureFeatureV11LimitInAggregate,
	googlesql.LanguageFeatureFeatureV13DateTimeConstructors,
	googlesql.LanguageFeatureFeatureV13ExtendedDateTimeSignatures,
	googlesql.LanguageFeatureFeatureV12CivilTime,
	googlesql.LanguageFeatureFeatureV12WeekWithWeekday,
	googlesql.LanguageFeatureFeatureIntervalType,
	googlesql.LanguageFeatureFeatureGroupByRollup,
	googlesql.LanguageFeatureFeatureV13NullsFirstLastInOrderBy,
	googlesql.LanguageFeatureFeatureV13Qualify,
	googlesql.LanguageFeatureFeatureV13AllowDashesInTableName,
	googlesql.LanguageFeatureFeatureGeography,
	googlesql.LanguageFeatureFeatureV13ExtendedGeographyParsers,
	googlesql.LanguageFeatureFeatureTemplateFunctions,
	googlesql.LanguageFeatureFeatureV11WithOnSubquery,
	googlesql.LanguageFeatureFeatureV13Pivot,
	googlesql.LanguageFeatureFeatureV13Unpivot,
	googlesql.LanguageFeatureFeatureCreateTableAsSelectColumnList,
	googlesql.LanguageFeatureFeatureCreateTablePartitionBy,
	googlesql.LanguageFeatureFeatureCreateTableClusterBy,
	googlesql.LanguageFeatureFeatureColumnDefaultValue,
	// Function-set feature flags. Enable the families we have
	// runtime support for so the analyzer recognises the new
	// builtins (LAX_*, EDIT_DISTANCE, MAX_BY / MIN_BY, etc.)
	// without us having to rewrite them at the SQL pre-process
	// stage.
	googlesql.LanguageFeatureFeatureJsonLaxValueExtractionFunctions,
	googlesql.LanguageFeatureFeatureV14JsonArrayValueExtractionFunctions,
	googlesql.LanguageFeatureFeatureJsonValueExtractionFunctions,
	googlesql.LanguageFeatureFeatureJsonMoreValueExtractionFunctions,
	googlesql.LanguageFeatureFeatureAdditionalStringFunctions,
	googlesql.LanguageFeatureFeatureArrayAggregationFunctions,
	googlesql.LanguageFeatureFeatureEnableEditDistanceBytes,
	googlesql.LanguageFeatureFeatureV14EnableEditDistanceBytes,
	// Items previously listed as "upstream asks" — turn out to
	// be feature flags we had not enabled. Enable them here so
	// the analyzer accepts the syntax / function name; runtime
	// support is still required separately for some.
	googlesql.LanguageFeatureFeatureJsonMutatorFunctions,
	googlesql.LanguageFeatureFeatureRangeType,
	googlesql.LanguageFeatureFeatureWithRecursive,
	googlesql.LanguageFeatureFeatureTableValuedFunctions,
	googlesql.LanguageFeatureFeatureCreateTableFunction,
	googlesql.LanguageFeatureFeatureOmitInsertColumnList,
	googlesql.LanguageFeatureFeatureTokenizedSearch,
	// Permits ResolvedArgumentRef inside JSON_VALUE / JSON_QUERY
	// path arguments. Without this, the analyzer rejects bodies
	// like `JSON_VALUE(json_arg, $arg_path)` inside CREATE
	// FUNCTION because the path is required to be a literal.
	googlesql.LanguageFeatureFeatureEnableConstantExpressionInJsonPath,
	// JSON_KEYS — the analyzer recognises the function via the
	// FEATURE_JSON_KEYS_FUNCTION gate (otherwise it emits
	// "Function not found"). Runtime registration sits in
	// internal/functions/json/json_keys.go.
	googlesql.LanguageFeatureFeatureJsonKeysFunction,
	// DATE_BUCKET / DATETIME_BUCKET / TIMESTAMP_BUCKET — analyzer
	// gate; runtime impl lives in
	// internal/functions/{date,datetime,timestamp}/*_bucket.go.
	googlesql.LanguageFeatureFeatureTimeBucketFunctions,
	// Math long-tail. Each flag opens up one family of
	// analyzer-recognised builtins; runtime sits in
	// internal/functions/math.
	googlesql.LanguageFeatureFeatureCbrtFunctions,
	googlesql.LanguageFeatureFeatureInverseTrigFunctions,
	googlesql.LanguageFeatureFeaturePiFunctions,
	googlesql.LanguageFeatureFeatureRadiansDegreesFunctions,
	googlesql.LanguageFeatureFeatureManhattanDistance,
	googlesql.LanguageFeatureFeatureL1Norm,
	googlesql.LanguageFeatureFeatureL2Norm,
	googlesql.LanguageFeatureFeatureEnableFloatDistanceFunctions,
	googlesql.LanguageFeatureFeatureV14DotProduct,
	// JSON long-tail family flags.
	googlesql.LanguageFeatureFeatureJsonConstructorFunctions,
	googlesql.LanguageFeatureFeatureJsonContainsFunction,
	googlesql.LanguageFeatureFeatureJsonFlattenFunction,
	// PropertyGraph + AGG / MEASURE. Measure rewriting emits
	// multi-level aggregation expressions, so the
	// MultilevelAggregation feature is required for the rewriter
	// output to type-check.
	googlesql.LanguageFeatureFeatureSqlGraph,
	googlesql.LanguageFeatureFeatureEnableMeasures,
	googlesql.LanguageFeatureFeatureSqlGraphExposeGraphElement,
	googlesql.LanguageFeatureFeatureUnenforcedPrimaryKeys,
	googlesql.LanguageFeatureFeatureSqlGraphMeasureDdl,
	googlesql.LanguageFeatureFeatureSqlGraphAdvancedQuery,
	googlesql.LanguageFeatureFeatureMultilevelAggregation,
	googlesql.LanguageFeatureFeatureV14SqlGraphPathType,
	googlesql.LanguageFeatureFeatureV14SqlGraphPathMode,
	googlesql.LanguageFeatureFeatureV14SqlGraphAdvancedQuery,
	googlesql.LanguageFeatureFeatureV14SqlGraphBoundedPathQuantification,
	googlesql.LanguageFeatureFeatureSqlGraphUnboundedPathQuantification,
	googlesql.LanguageFeatureFeatureSqlGraphReturnExtensions,
	googlesql.LanguageFeatureFeatureGroupByGraphPath,
	googlesql.LanguageFeatureFeatureGroupByStruct,
	// Features that the rewriters above transform into ordinary
	// expressions. Each is gated upstream so enabling them is
	// safe — they only fire when the corresponding construct
	// appears in a query.
	googlesql.LanguageFeatureFeatureInlineLambdaArgument,
	googlesql.LanguageFeatureFeatureLikeAnySomeAll,
	// Anonymization / Differential Privacy syntax gates. The
	// analyzer-side rewriter
	// (ResolvedASTRewriteRewriteAnonymization) is enabled above;
	// these flags make the SELECT WITH ANONYMIZATION /
	// DIFFERENTIAL_PRIVACY surface parse. Runtime support for
	// the rewriter output ($differential_privacy_sum etc.) is
	// not in scope — the upstream rewriter emits internal
	// aggregate names whose semantics depend on Laplace /
	// Gaussian noise injection that the spec notes flag as
	// out-of-scope for an eager-evaluation runtime. Enabling
	// just the parser features lets us return a clear error
	// at analysis-or-rewrite time instead of "Unexpected
	// keyword WITH". See docs/specs/googlesql/functions/
	// aggregate/differential_privacy/*.md for the recorded
	// rationale.
	googlesql.LanguageFeatureFeatureAnonymization,
	googlesql.LanguageFeatureFeatureAnonymizationThresholding,
	googlesql.LanguageFeatureFeatureAnonymizationCaseInsensitiveOptions,
	googlesql.LanguageFeatureFeatureDifferentialPrivacy,
	googlesql.LanguageFeatureFeatureDifferentialPrivacyReportFunctions,
	googlesql.LanguageFeatureFeatureDifferentialPrivacyThresholding,
	// BigQuery AEAD encryption family (KEYS.* / AEAD.* /
	// DETERMINISTIC_*). The catalog-side registration uses the
	// same flag; the analyzer's resolver also gates the
	// function call on this LanguageOption at analyze time.
	googlesql.LanguageFeatureFeatureEncryption,
	// GoogleSQL proto-reflection family. Each function carries
	// its own feature flag; enable the lot so Conn.RegisterProto
	// consumers can reach PROTO_DEFAULT_IF_NULL / EXTRACT /
	// FILTER_FIELDS / REPLACE_FIELDS / PROTO_MAP_* /
	// ENUM_VALUE_DESCRIPTOR_PROTO / FROM_PROTO / TO_PROTO.
	googlesql.LanguageFeatureFeatureProtoBase,
	googlesql.LanguageFeatureFeatureProtoDefaultIfNull,
	googlesql.LanguageFeatureFeatureExtractFromProto,
	googlesql.LanguageFeatureFeatureReplaceFields,
	googlesql.LanguageFeatureFeatureFilterFields,
	googlesql.LanguageFeatureFeatureProtoMaps,
	googlesql.LanguageFeatureFeatureProtoExtensionsWithNew,
	googlesql.LanguageFeatureFeatureProtoExtensionsWithSet,
	googlesql.LanguageFeatureFeatureEnumValueDescriptorProto,
	// Array helpers shipped behind feature flags upstream:
	// ARRAY_ZIP / ARRAY_FIND / FLATTEN with chained-path semantics.
	// Our runtime already binds the matching UDFs; flipping the
	// analyzer flag lets the resolver dispatch by signature.
	googlesql.LanguageFeatureFeatureArrayZip,
	googlesql.LanguageFeatureFeatureUnnestAndFlattenArrays,
	googlesql.LanguageFeatureFeatureGroupingSets,
	googlesql.LanguageFeatureFeatureGroupingBuiltin,
	googlesql.LanguageFeatureFeatureV14UuidType,
	// Pipe-syntax queries (`FROM t |> EXTEND ... |> SELECT ...`).
	// Upstream gates the parser behind FeaturePipes; the resolver
	// then handles individual pipe operators via their own feature
	// flags. REGEXP_EXTRACT_GROUPS' Pipe-syntax upstream Example
	// exercises this path.
	googlesql.LanguageFeatureFeaturePipes,
	// `|> IF ... ELSEIF ... ELSE ...` and `|> ASSERT ...` pipe
	// operators. The Debug ERROR / IFERROR upstream Examples
	// exercise the pipe-IF cascade; without these toggles the
	// analyzer rejects them with "Pipe IF not supported" /
	// "Pipe ASSERT not supported".
	googlesql.LanguageFeatureFeaturePipeIf,
	googlesql.LanguageFeatureFeaturePipeAssert,
	// Procedural language: `FOR row IN (...) DO ... END FOR`,
	// `CASE WHEN ... THEN ... END CASE`, `REPEAT ... UNTIL ...`,
	// and labelled script blocks. Toggling these on lets
	// AnalyzeNextScriptStatement succeed on the corresponding
	// upstream constructs.
	googlesql.LanguageFeatureFeatureForIn,
	googlesql.LanguageFeatureFeatureCaseStmt,
	googlesql.LanguageFeatureFeatureRepeat,
	googlesql.LanguageFeatureFeatureScriptLabel,
	// STRUCT positional accessors `s[OFFSET(i)]` / `s[ORDINAL(n)]`
	// for the compliance fixtures under types/struct.
	googlesql.LanguageFeatureFeatureV14StructPositionalAccessor,
}

// supportedStatementKinds lists the ResolvedStatement kinds the
// analyzer is explicitly opened up to. It is pure data consumed by
// newAnalyzerOptions; the order is preserved from the historical
// inline literal.
var supportedStatementKinds = []googlesql.ResolvedNodeKind{
	googlesql.ResolvedNodeKindResolvedBeginStmt,
	googlesql.ResolvedNodeKindResolvedCommitStmt,
	googlesql.ResolvedNodeKindResolvedMergeStmt,
	googlesql.ResolvedNodeKindResolvedQueryStmt,
	googlesql.ResolvedNodeKindResolvedInsertStmt,
	googlesql.ResolvedNodeKindResolvedUpdateStmt,
	googlesql.ResolvedNodeKindResolvedDeleteStmt,
	googlesql.ResolvedNodeKindResolvedDropStmt,
	googlesql.ResolvedNodeKindResolvedTruncateStmt,
	googlesql.ResolvedNodeKindResolvedCreateTableStmt,
	googlesql.ResolvedNodeKindResolvedCreateTableAsSelectStmt,
	googlesql.ResolvedNodeKindResolvedCreateProcedureStmt,
	googlesql.ResolvedNodeKindResolvedCreateFunctionStmt,
	googlesql.ResolvedNodeKindResolvedCreateTableFunctionStmt,
	googlesql.ResolvedNodeKindResolvedCreateViewStmt,
	googlesql.ResolvedNodeKindResolvedDropFunctionStmt,
	googlesql.ResolvedNodeKindResolvedAssignmentStmt,
	googlesql.ResolvedNodeKindResolvedExportDataStmt,
	// PropertyGraph DDL.
	googlesql.ResolvedNodeKindResolvedCreatePropertyGraphStmt,
	// DDL: ALTER TABLE / CREATE SCHEMA. We accept these so the
	// analyzer surfaces a resolved AST; runtime treats them as
	// metadata-only no-ops via NoopStmtAction.
	googlesql.ResolvedNodeKindResolvedAlterTableStmt,
	googlesql.ResolvedNodeKindResolvedCreateSchemaStmt,
	// DCL: GRANT / REVOKE. googlesqlite does not enforce
	// permissions, but accepting them lets DCL-laced scripts
	// flow through.
	googlesql.ResolvedNodeKindResolvedGrantStmt,
	googlesql.ResolvedNodeKindResolvedRevokeStmt,
	// Procedural / scripting: ASSERT (runs the predicate and
	// fails if false) and EXECUTE IMMEDIATE (re-analyzes its
	// string argument). Both are no-op for the moment;
	// future-work will hook them up.
	googlesql.ResolvedNodeKindResolvedAssertStmt,
	googlesql.ResolvedNodeKindResolvedExecuteImmediateStmt,
	// LOAD DATA: surface the resolved AST so external tools can
	// inspect the file list; the driver no-ops on execution.
	googlesql.ResolvedNodeKindResolvedAuxLoadDataStmt,
}

func newAnalyzerOptions() (*googlesql.AnalyzerOptions, error) {
	langOpt := m1(googlesql.NewLanguageOptions())
	if langOpt == nil {
		return nil, fmt.Errorf("failed to initialize SQL language options")
	}
	langOpt.SetNameResolutionMode(googlesql.NameResolutionModeNameResolutionDefault)
	langOpt.SetProductMode(googlesql.ProductModeProductInternal)
	for _, f := range enabledLanguageFeatures {
		_ = langOpt.EnableLanguageFeature(f)
	}
	// Open the analyzer up to every ResolvedStatement kind. The
	// manual AddSupportedStatementKind list below is still useful as
	// documentation, but starting from "all supported" means freshly
	// added kinds (IF / WHILE / FOR / RAISE / CALL ... and any
	// future additions in go-googlesql) don't have to be enumerated
	// twice.
	_ = langOpt.SetSupportsAllStatementKinds()
	for _, k := range supportedStatementKinds {
		_ = langOpt.AddSupportedStatementKind(k)
	}
	// Enable QUALIFY without WHERE
	// https://github.com/google/googlesql/issues/124
	if err := langOpt.EnableReservableKeyword("QUALIFY", true); err != nil {
		return nil, err
	}
	opt, optErr := googlesql.NewAnalyzerOptions2()
	if opt == nil {
		return nil, fmt.Errorf("failed to initialize analyzer options: %w", optErr)
	}
	opt.SetAllowUndeclaredParameters(true)
	opt.SetLanguage(langOpt)
	opt.SetParseLocationRecordType(googlesql.ParseLocationRecordTypeParseLocationRecordFullNodeScope)
	if err := applyAnalyzerRewrites(opt); err != nil {
		return nil, err
	}
	if err := registerSystemVariables(opt); err != nil {
		return nil, fmt.Errorf("failed to register system variables: %w", err)
	}
	return opt, nil
}

// applyAnalyzerRewrites toggles the ResolvedAST rewriters the driver
// relies on. It is the self-contained rewrite-configuration block
// extracted from newAnalyzerOptions; the set of enabled / disabled
// rewriters is unchanged.
func applyAnalyzerRewrites(opt *googlesql.AnalyzerOptions) error {
	// Disable the upstream MEASURE-type rewriter: it requires AGG
	// to wrap a direct ColumnRef, but graph queries produce
	// AGG(GraphGetElementProperty(...)) which the rewriter refuses
	// to handle. Our driver-side formatter (see tryFormatMeasureAGG
	// in internal/graph_scan_node.go) lowers measure AGGs to the
	// `googlesqlite_agg` aggregate manually, so the rewriter would
	// only get in the way.
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteMeasureType, false)
	// Lazy-evaluation conditional family.
	// NULLIFERROR / IFERROR / IS_ERROR are expanded by the analyzer
	// into a SAFE(...) + IFNULL pattern, so they reach the
	// formatter as a plain expression tree without needing a
	// runtime UDF. Similarly IS_FIRST / IS_LAST get rewritten into
	// numbering-window equivalents.
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteNulliferrorFunction, true)
	// BuiltinFunctionInliner lowers a small set of polymorphic builtins
	// (IFERROR / ISERROR / NULLIFERROR / NULLIFZERO / ZEROIFNULL) at
	// analyzer time, propagating the outer expression's expected type
	// into ERROR()'s placeholder so nested IFERROR(IFERROR(ERROR(...),
	// ERROR(...)), 'string') no longer fails the "common supertype"
	// check.
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteBuiltinFunctionInliner, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteIsFirstIsLastFunction, true)
	// ANON_* and DIFFERENTIAL_PRIVACY family. The analyzer
	// rewriter turns these into ordinary aggregates over noisy
	// aggregates — fine as long as our runtime honours the basic
	// noise injection, which the standard aggregate runtimes
	// already approximate (no DP guarantee, but
	// signature-compatible).
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteAnonymization, true)
	// PROTO_MAP_CONTAINS_KEY / PROTO_MODIFY_MAP have an upstream rewrite
	// that expands them into ARRAY_AGG-driven SQL. PROTO_MODIFY_MAP in
	// particular feeds NULL deletion sentinels into ARRAY_AGG, which
	// our runtime aggregator rejects (parity with historical BigQuery
	// behavior). Disabling RewriteProtoMapFns leaves both functions
	// unrewritten so the formatter / runtime can implement them
	// directly via googlesqlite_proto_map_contains_key /
	// googlesqlite_proto_modify_map.
	if err := opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteProtoMapFns, false); err != nil {
		return fmt.Errorf("disable RewriteProtoMapFns: %w", err)
	}
	// Generally-useful rewriters that turn rare / awkward syntax
	// into the standard form. Each is gated upstream so enabling
	// them is safe — they fire only when the corresponding
	// construct appears in the query.
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteFlatten, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteTypeofFunction, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteUnpivot, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewritePivot, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteLikeAnyAll, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteWithExpr, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteArrayFilterTransform, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteArrayIncludes, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteOrderByAndLimitInAggregate, true)
	// Templated builtin inliner (RewriteBuiltinFunctionInliner above)
	// is gated by this meta-rewriter — without it the inliner does
	// not visit templated function calls. Empirically this flag does
	// not rescue the deeply-nested IFERROR Example 6 (the resolver
	// rejects at type-check before any rewriter runs), but enabling
	// it costs nothing for the cases that already pass.
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteApplyEnabledRewritesToTemplatedFunctionCalls, true)
	// Pipe-syntax IF / ASSERT rewriters: with the BarrierScan
	// formatter handler in place (see internal/node.go), the rewriters
	// produce trees the formatter can traverse end-to-end.
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewritePipeIf, true)
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewritePipeAssert, true)
	// GROUP BY GROUPING SETS / ROLLUP / CUBE: the upstream rewriter
	// expands these into a UNION ALL of per-grouping-set aggregations
	// and rewrites each GROUPING(col) call into a constant 0 / 1 inside
	// each branch, so the resulting tree never references
	// `$grouping_call*` as a column.
	_ = opt.EnableRewrite(googlesql.ResolvedASTRewriteRewriteGroupingSet, true)
	return nil
}

// debugStream returns os.Stderr when the GOOGLESQLITE_DEBUG env
// var is set to anything truthy; otherwise io.Discard. Used to
// trace graph-query SQL rewriting without polluting normal output.
func debugStream() io.Writer {
	if os.Getenv("GOOGLESQLITE_DEBUG") != "" {
		return os.Stderr
	}
	return io.Discard
}

// systemVariableTypes lists the @@system_variable names the analyzer
// must recognise. They are typed STRING because the only consumers
// (consumer session settings) treat them as opaque strings;
// the runtime side stores and reads back through Conn.systemVars.
var systemVariableTypes = []struct {
	path []string
	kind googlesql.TypeKind
}{
	{[]string{"time_zone"}, googlesql.TypeKindTypeString},
	{[]string{"dataset_project_id"}, googlesql.TypeKindTypeString},
	{[]string{"dataset_id"}, googlesql.TypeKindTypeString},
	{[]string{"project_id"}, googlesql.TypeKindTypeString},
}

func registerSystemVariables(opt *googlesql.AnalyzerOptions) error {
	factory := tf()
	for _, sv := range systemVariableTypes {
		typ, err := factory.MakeSimpleType(sv.kind)
		if err != nil {
			return err
		}
		if err := opt.AddSystemVariable(sv.path, typ); err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) SetAutoIndexMode(enabled bool) {
	a.isAutoIndexMode = enabled
}

func (a *Analyzer) SetExplainMode(enabled bool) {
	a.isExplainMode = enabled
}

func (a *Analyzer) NamePath() []string {
	return a.namePath.path
}

func (a *Analyzer) SetNamePath(path []string) error {
	return a.namePath.setPath(path)
}

func (a *Analyzer) SetMaxNamePath(num int) {
	a.namePath.setMaxNum(num)
}

func (a *Analyzer) MaxNamePath() int {
	return a.namePath.maxNum
}

func (a *Analyzer) AddNamePath(path string) error {
	return a.namePath.addPath(path)
}

// applyIngestionPartitionRewrite makes ingestion-time partition
// references (`_PARTITIONTIME`, `_PARTITIONDATE`) acceptable to the
// upstream analyzer, which would otherwise reject them as
// "Unrecognized name". BigQuery exposes these as pseudo columns on
// tables created with `PARTITION BY DATE(_PARTITIONTIME)`.
//
// Two-pass rewrite:
//
//  1. Strip `PARTITION BY <expr>` from CREATE TABLE statements.
//     googlesqlite does not physically partition (SQLite is the
//     storage layer); the clause is metadata-only and the upstream
//     analyzer rejects `PARTITION BY DATE(CURRENT_TIMESTAMP())` as
//     "PARTITION BY expression must not be constant" if we tried to
//     rewrite the inner reference.
//  2. Replace bare `_PARTITIONTIME` / `_PARTITIONDATE` references
//     elsewhere (SELECT, WHERE, ...) with CURRENT_TIMESTAMP() /
//     CURRENT_DATE(). Driver-side semantics: every row looks like
//     it was ingested "now"; consumers that need per-row ingestion
//     time should track it themselves.
//
// String literals and backtick-quoted identifiers are skipped so
// embedded tokens (`'... _PARTITIONTIME ...'`) remain untouched.
func applyIngestionPartitionRewrite(query string) string {
	if strings.Contains(query, "PARTITION") || strings.Contains(query, "partition") {
		query = stripPartitionByClause(query)
	}
	if !strings.Contains(query, "_PARTITION") {
		return query
	}
	var b strings.Builder
	b.Grow(len(query))
	i := 0
	for i < len(query) {
		c := query[i]
		if c == '\'' || c == '"' {
			end := scanStringLiteral(query, i)
			b.WriteString(query[i:end])
			i = end
			continue
		}
		if c == '`' {
			end := i + 1
			for end < len(query) && query[end] != '`' {
				end++
			}
			if end < len(query) {
				end++
			}
			b.WriteString(query[i:end])
			i = end
			continue
		}
		if isIdentStart(c) {
			j := i + 1
			for j < len(query) && isIdentCont(query[j]) {
				j++
			}
			word := query[i:j]
			switch strings.ToUpper(word) {
			case "_PARTITIONTIME":
				b.WriteString("CURRENT_TIMESTAMP()")
			case "_PARTITIONDATE":
				b.WriteString("CURRENT_DATE()")
			default:
				b.WriteString(word)
			}
			i = j
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// stripPartitionByClause removes a top-level `PARTITION BY <expr>`
// clause from the SQL. The clause is metadata-only for googlesqlite
// (no physical partitioning), and the inner expression often
// contains pseudo-references (`_PARTITIONTIME`) that the upstream
// analyzer would otherwise reject.
//
// Detects the clause by an outer-paren depth scan: matches PARTITION
// BY only at depth 0 so a `WITH t(x) AS (... PARTITION BY ...)`
// inside an OVER clause is left alone.
func stripPartitionByClause(query string) string {
	upper := strings.ToUpper(query)
	var b strings.Builder
	b.Grow(len(query))
	i := 0
	depth := 0
	for i < len(query) {
		c := query[i]
		if c == '\'' || c == '"' || c == '`' {
			end := i + 1
			if c == '`' {
				for end < len(query) && query[end] != '`' {
					end++
				}
				if end < len(query) {
					end++
				}
			} else {
				end = scanStringLiteral(query, i)
			}
			b.WriteString(query[i:end])
			i = end
			continue
		}
		if c == '(' {
			depth++
			b.WriteByte(c)
			i++
			continue
		}
		if c == ')' {
			depth--
			b.WriteByte(c)
			i++
			continue
		}
		if depth == 0 && i+9 <= len(query) && upper[i:i+9] == "PARTITION" &&
			(i == 0 || !isIdentCont(query[i-1])) &&
			i+9 < len(query) && (query[i+9] == ' ' || query[i+9] == '\t' || query[i+9] == '\n') {
			// Must be followed by BY (skip whitespace).
			k := i + 9
			for k < len(query) && (query[k] == ' ' || query[k] == '\t' || query[k] == '\n') {
				k++
			}
			if k+2 <= len(query) && upper[k:k+2] == "BY" &&
				(k+2 == len(query) || !isIdentCont(query[k+2])) {
				// Skip past `PARTITION BY <expr>`. The expression
				// runs until the next top-level keyword (CLUSTER,
				// OPTIONS, AS, ;) or end of statement.
				k += 2
				exprDepth := 0
				for k < len(query) {
					ch := query[k]
					if ch == '\'' || ch == '"' {
						k = scanStringLiteral(query, k)
						continue
					}
					if ch == '(' {
						exprDepth++
						k++
						continue
					}
					if ch == ')' {
						if exprDepth == 0 {
							break
						}
						exprDepth--
						k++
						continue
					}
					if exprDepth == 0 && isIdentStart(ch) {
						kEnd := k + 1
						for kEnd < len(query) && isIdentCont(query[kEnd]) {
							kEnd++
						}
						w := strings.ToUpper(query[k:kEnd])
						if w == "CLUSTER" || w == "OPTIONS" || w == "AS" {
							break
						}
						k = kEnd
						continue
					}
					if ch == ';' && exprDepth == 0 {
						break
					}
					k++
				}
				i = k
				continue
			}
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// applyMixedParameterRewrite supports queries that combine positional
// `?` placeholders with named `@name` parameters in a single
// statement. The googlesql analyzer rejects this mix because its
// ParameterMode is a single enum (positional XOR named); we sidestep
// the limitation by re-naming every `?` to a synthetic
// `@gsl_pos_<N>` (1-based) and re-mapping any positional
// driver.NamedValue (Name == "") to that synthetic name in order of
// appearance. The transform is no-op for the common single-mode
// callers.
//
// Quote-aware: `?` inside string / backtick literals is left alone.
func applyMixedParameterRewrite(query string, args []driver.NamedValue) (string, []driver.NamedValue) {
	if !strings.ContainsRune(query, '?') {
		return query, args
	}
	hasNamed := false
	hasPositional := false
	for _, a := range args {
		if a.Name == "" {
			hasPositional = true
		} else {
			hasNamed = true
		}
	}
	// Args-less Prepare: detect both syntactic forms in the query and
	// trigger the rewrite so the analyzer accepts the prepared SQL
	// (deferred binding will assign synthetic names at Exec time).
	if !(hasNamed && hasPositional) {
		if len(args) > 0 || !queryHasNamedRef(query) {
			return query, args
		}
	}
	var b strings.Builder
	b.Grow(len(query) + 16)
	posCount := 0
	i := 0
	for i < len(query) {
		c := query[i]
		if c == '\'' || c == '"' {
			end := scanStringLiteral(query, i)
			b.WriteString(query[i:end])
			i = end
			continue
		}
		if c == '`' {
			end := i + 1
			for end < len(query) && query[end] != '`' {
				end++
			}
			if end < len(query) {
				end++
			}
			b.WriteString(query[i:end])
			i = end
			continue
		}
		if c == '?' {
			posCount++
			fmt.Fprintf(&b, "@gsl_pos_%d", posCount)
			i++
			continue
		}
		b.WriteByte(c)
		i++
	}
	if posCount == 0 {
		return query, args
	}
	newArgs := make([]driver.NamedValue, len(args))
	copy(newArgs, args)
	assigned := 0
	for k := range newArgs {
		if newArgs[k].Name != "" {
			continue
		}
		assigned++
		newArgs[k].Name = fmt.Sprintf("gsl_pos_%d", assigned)
	}
	return b.String(), newArgs
}

// queryHasNamedRef reports whether the query contains a `@name`
// reference outside of string / backtick literals. Used by
// applyMixedParameterRewrite to recognise mixed-mode syntax at
// Prepare time, when no args have been bound yet.
func queryHasNamedRef(query string) bool {
	i := 0
	for i < len(query) {
		c := query[i]
		if c == '\'' || c == '"' {
			end := scanStringLiteral(query, i)
			i = end
			continue
		}
		if c == '`' {
			end := i + 1
			for end < len(query) && query[end] != '`' {
				end++
			}
			if end < len(query) {
				end++
			}
			i = end
			continue
		}
		if c == '@' && i+1 < len(query) {
			nc := query[i+1]
			if (nc >= 'a' && nc <= 'z') || (nc >= 'A' && nc <= 'Z') || nc == '_' {
				return true
			}
		}
		i++
	}
	return false
}

// applyTypeNameAliases rewrites a small set of type-name aliases that
// the upstream parser does not accept natively. The transform is
// string-literal-aware: characters inside `'...'`, `"..."`, or
// triple-quoted strings (raw / byte literals included) are not
// touched. Currently:
//
//	BIG_NUMERIC                                  -> BIGNUMERIC
//	INT / INTEGER / SMALLINT / BIGINT / TINYINT  -> INT64
//
// BigQuery documents INT, SMALLINT, INTEGER, BIGINT, TINYINT, BYTEINT
// as aliases for INT64; the upstream analyzer rejects them.
//
// More aliases can be added as new compatibility gaps surface.
func applyTypeNameAliases(query string) string {
	if !typeAliasNeedle.MatchString(query) {
		return query
	}
	var b strings.Builder
	b.Grow(len(query))
	i := 0
	for i < len(query) {
		c := query[i]
		// Skip past a string literal entirely, copying it verbatim.
		// Handle 'x' / "x" / '''x''' / """x""" plus optional `r` / `b`
		// prefix per GoogleSQL lexical rules.
		if c == '\'' || c == '"' {
			end := scanStringLiteral(query, i)
			b.WriteString(query[i:end])
			i = end
			continue
		}
		// Backtick-quoted identifiers are user-chosen names — never
		// touch their contents even if they happen to spell `Int`.
		if c == '`' {
			end := i + 1
			for end < len(query) && query[end] != '`' {
				end++
			}
			if end < len(query) {
				end++
			}
			b.WriteString(query[i:end])
			i = end
			continue
		}
		if (c == 'r' || c == 'R' || c == 'b' || c == 'B') && i+1 < len(query) {
			n := query[i+1]
			if n == '\'' || n == '"' {
				b.WriteByte(c)
				end := scanStringLiteral(query, i+1)
				b.WriteString(query[i+1 : end])
				i = end
				continue
			}
		}
		// Only attempt the alias replacement on word boundaries.
		if isIdentStart(c) {
			j := i + 1
			for j < len(query) && isIdentCont(query[j]) {
				j++
			}
			word := query[i:j]
			if replacement, ok := typeAliasReplacements[strings.ToUpper(word)]; ok {
				if isAliasInTypePosition(query, i, j) {
					b.WriteString(replacement)
				} else {
					b.WriteString(word)
				}
			} else {
				b.WriteString(word)
			}
			i = j
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// isAliasInTypePosition returns true when the identifier at
// query[start:end] sits in a position where a TYPE name is expected
// rather than a column / parameter name. The check is purely lexical:
//
//   - Preceded by `(` or `,` → column-list / argument-list NAME slot.
//     INSERT INTO t (Int, Str) and CALL fn(Int, Str) both fit.
//     Rewriting INT here would break valid identifiers spelled `Int`.
//   - Followed by another identifier or backticked name → also a NAME
//     slot (column declaration `(col_name INT64)`).
//
// Otherwise the alias is in a type position: CAST(x AS INT),
// RETURNS INT, ARRAY<INT>, STRUCT<a INT, b INT>, column type after the
// column name in a CREATE TABLE list, and so on.
func isAliasInTypePosition(query string, start, end int) bool {
	for k := start - 1; k >= 0; k-- {
		switch query[k] {
		case ' ', '\t', '\n', '\r':
			continue
		case '(', ',':
			return false
		}
		break
	}
	for k := end; k < len(query); k++ {
		switch query[k] {
		case ' ', '\t', '\n', '\r':
			continue
		}
		if isIdentStart(query[k]) || query[k] == '`' {
			return false
		}
		break
	}
	return true
}

// applyNaiveTimestampUTC rewrites TIMESTAMP literals that lack a
// timezone marker by appending `+00:00`, so the analyzer parses them
// as UTC instead of using its built-in default (America/Los_Angeles
// upstream). BigQuery documents naive TIMESTAMP literals as UTC.
//
// Touches only `TIMESTAMP '<value>'` and `TIMESTAMP "<value>"` forms
// where <value> contains neither a timezone offset (`[+-]\d{2}:?\d{2}`),
// the literal `Z` / `UTC` suffix, nor any other timezone marker.
// `TIMESTAMP_ADD` and other identifier prefixes are unaffected because
// the scanner consumes them as a single word.
func applyNaiveTimestampUTC(query string) string {
	if !timestampNeedle.MatchString(query) {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 8)
	i := 0
	for i < len(query) {
		c := query[i]
		// Pass through string literals verbatim unless we have just
		// emitted a TIMESTAMP keyword (handled below).
		if c == '\'' || c == '"' {
			end := scanStringLiteral(query, i)
			b.WriteString(query[i:end])
			i = end
			continue
		}
		if (c == 'r' || c == 'R' || c == 'b' || c == 'B') && i+1 < len(query) {
			n := query[i+1]
			if n == '\'' || n == '"' {
				b.WriteByte(c)
				end := scanStringLiteral(query, i+1)
				b.WriteString(query[i+1 : end])
				i = end
				continue
			}
		}
		if isIdentStart(c) {
			j := i + 1
			for j < len(query) && isIdentCont(query[j]) {
				j++
			}
			word := query[i:j]
			b.WriteString(word)
			i = j
			if strings.EqualFold(word, "TIMESTAMP") {
				// Skip whitespace, then if the next token is a string
				// literal without a TZ marker, rewrite it to UTC.
				k := i
				for k < len(query) && (query[k] == ' ' || query[k] == '\t' || query[k] == '\n' || query[k] == '\r') {
					k++
				}
				if k < len(query) && (query[k] == '\'' || query[k] == '"') {
					end := scanStringLiteral(query, k)
					if end > k+1 && end <= len(query) {
						b.WriteString(query[i:k]) // whitespace
						i = end
						quote := query[k]
						content := query[k+1 : end-1]
						if !timestampHasTZ(content) {
							b.WriteByte(quote)
							b.WriteString(content)
							b.WriteString("+00:00")
							b.WriteByte(quote)
						} else {
							b.WriteString(query[k:end])
						}
						continue
					}
				}
			}
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// timestampHasTZ reports whether a TIMESTAMP literal body already
// carries a timezone marker (offset, `Z`, or an IANA / abbreviation
// timezone name).
func timestampHasTZ(s string) bool {
	if len(s) == 0 {
		return false
	}
	if strings.HasSuffix(s, "Z") || strings.HasSuffix(s, "z") {
		return true
	}
	if tzOffsetTail.MatchString(s) {
		return true
	}
	// Named timezones appear after a space at the tail, e.g.
	// `2015-09-01 12:34:56 America/Los_Angeles` or
	// `2024-01-01 10:00:00 UTC`. The HH:MM time component contains a
	// colon but no space, so a trailing word after a space is a TZ
	// name. Reject if the trailing token is purely numeric (i.e. the
	// whole string is just a date with no TZ).
	return tzNameTail.MatchString(s)
}

var (
	timestampNeedle = regexp.MustCompile(`(?i)\bTIMESTAMP\s*['"]`)
	// Match `[+-]HH`, `[+-]HHMM`, or `[+-]HH:MM` at the tail.
	tzOffsetTail = regexp.MustCompile(`[+-]\d{2}(?::?\d{2})?\s*$`)
	// Match a trailing alphabetic timezone identifier separated from
	// the time component by a space (e.g. `UTC`, `GMT`,
	// `America/Los_Angeles`, `EST`, `PST`).
	tzNameTail = regexp.MustCompile(`\s[A-Za-z][A-Za-z_/+-]*\s*$`)
)

var typeAliasReplacements = map[string]string{
	"BIG_NUMERIC": "BIGNUMERIC",
	"INT":         "INT64",
	"INTEGER":     "INT64",
	"SMALLINT":    "INT64",
	"BIGINT":      "INT64",
	"TINYINT":     "INT64",
	"BYTEINT":     "INT64",
}

var typeAliasNeedle = regexp.MustCompile(`(?i)\b(BIG_NUMERIC|INT|INTEGER|SMALLINT|BIGINT|TINYINT|BYTEINT)\b`)

func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isIdentCont(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

// scanStringLiteral returns the index just past the end of the
// string literal that begins at start. Handles single-quote,
// double-quote, and triple-quoted forms with a minimal
// backslash-escape walk. On malformed input it returns len(s) so
// the caller stops scanning.
func scanStringLiteral(s string, start int) int {
	if start >= len(s) {
		return len(s)
	}
	q := s[start]
	if q != '\'' && q != '"' {
		return start + 1
	}
	triple := false
	if start+2 < len(s) && s[start+1] == q && s[start+2] == q {
		triple = true
	}
	i := start + 1
	if triple {
		i = start + 3
	}
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			i += 2
			continue
		}
		if triple {
			if s[i] == q && i+2 < len(s) && s[i+1] == q && s[i+2] == q {
				return i + 3
			}
			i++
			continue
		}
		if s[i] == q {
			return i + 1
		}
		if s[i] == '\n' {
			// Unterminated single-line string: bail.
			return i
		}
		i++
	}
	return len(s)
}

// parsedScript bundles parsed statements with the handles they point
// into. The ParserOutput handles own the AST subtrees referenced by
// stmts; keeping them live here lets callers iterate AST nodes without
// them being freed mid-iteration.
type parsedScript struct {
	outputs []*googlesql.ParserOutput
	stmts   []googlesql.ASTStatementNode
}

// isScriptControlStmt reports whether stmt is one of the procedural-
// language statements that AnalyzeStatementFromParserAST refuses
// because they are not subclasses of ResolvedStatement (they are
// handled by the upstream Script executor instead). The driver
// accepts them as no-ops so the procedural-language reference at
// docs.cloud.google.com/bigquery/docs/reference/standard-sql/
// procedural-language flows through end-to-end.
func isScriptControlStmt(stmt googlesql.ASTStatementNode) bool {
	switch stmt.(type) {
	case *googlesql.ASTIfStatement,
		*googlesql.ASTWhileStatement,
		*googlesql.ASTRepeatStatement,
		*googlesql.ASTForInStatement,
		*googlesql.ASTCaseStatement,
		*googlesql.ASTRaiseStatement,
		*googlesql.ASTBreakStatement,
		*googlesql.ASTContinueStatement,
		*googlesql.ASTReturnStatement,
		*googlesql.ASTCallStatement:
		return true
	}
	return false
}

func (a *Analyzer) parseScript(query string) (*parsedScript, error) {
	loc, locErr := googlesql.NewParseResumeLocationFromString(query)
	if loc == nil {
		return nil, fmt.Errorf("failed to create parse resume location for %q: %w", query, locErr)
	}
	parserOpts, err := a.opt.GetParserOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get parser options: %w", err)
	}
	result := &parsedScript{}
	input, _ := loc.Input()
	inputLen := int32(len(input))
	for {
		out, err := googlesql.ParseNextScriptStatement(loc, parserOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse statement: %w", err)
		}
		if out == nil {
			break
		}
		result.outputs = append(result.outputs, out)
		stmt, err := out.Statement()
		if err != nil {
			return nil, fmt.Errorf("failed to get statement from parser output: %w", err)
		}
		// ASTBeginEndBlock has a StatementListNode child — expand it.
		if block, ok := stmt.(*googlesql.ASTBeginEndBlock); ok {
			list, err := block.StatementListNode()
			if err == nil && list != nil {
				result.stmts = append(result.stmts, collectStatementsFromList(list)...)
			}
		} else {
			result.stmts = append(result.stmts, stmt)
		}
		pos, _ := loc.BytePosition()
		if pos >= inputLen {
			break
		}
		// Skip any trailing whitespace / semicolon noise that follows
		// the last statement. If only whitespace remains we're done.
		if strings.TrimSpace(input[pos:]) == "" {
			break
		}
	}
	return result, nil
}

// collectStatementsFromList walks an ASTStatementList gathering its
// member statements. StatementList(i) does not bounds-check on the C++
// side, so we cap the loop with NumChildren() inherited from ASTNodeBase.
func collectStatementsFromList(list *googlesql.ASTStatementList) []googlesql.ASTStatementNode {
	n, err := list.NumChildren()
	if err != nil || n <= 0 {
		return nil
	}
	if n > 10000 {
		// Sanity cap — a single script with >10k top-level statements
		// is almost certainly a bogus value returned by a corrupted bridge.
		n = 10000
	}
	out := make([]googlesql.ASTStatementNode, 0, n)
	for i := int32(0); i < n; i++ {
		s, err := list.StatementList(i)
		if err != nil || s == nil {
			break
		}
		out = append(out, s)
	}
	return out
}

// validateLiteralCasts walks the AST looking for CAST(<string-literal>
// AS <numeric-type>) expressions where the literal would fail at
// runtime. googlesql's analyzer currently does not fold such casts at
// analysis time; rejecting them here produces the
// "Could not cast literal" message users expect, and matches the
// historical googlesql behavior.
func validateLiteralCasts(stmt googlesql.ASTStatementNode, query string) error {
	var firstErr error
	_ = astWalk(stmt, func(node googlesql.ASTNode) error {
		if firstErr != nil {
			return nil
		}
		cast, ok := node.(*googlesql.ASTCastExpression)
		if !ok {
			return nil
		}
		exprNode, _ := cast.Expr()
		strLit, ok := exprNode.(*googlesql.ASTStringLiteral)
		if !ok {
			return nil
		}
		strVal, _ := strLit.StringValue()
		typeNode, _ := cast.Type()
		simple, ok := typeNode.(*googlesql.ASTSimpleType)
		if !ok {
			return nil
		}
		path, _ := simple.TypeName()
		ids, _ := path.ToIdentifierVector()
		if len(ids) != 1 {
			return nil
		}
		typeName := strings.ToUpper(ids[0])
		switch typeName {
		case "INT32", "INT64", "UINT32", "UINT64":
			// Mirror StringValue.ToInt64: base 0 for hex/0x-prefixed.
			base := 10
			if strings.Contains(strings.ToLower(strVal), "0x") {
				base = 0
			}
			if _, perr := strconv.ParseInt(strVal, base, 64); perr == nil {
				return nil
			}
			if _, perr := strconv.ParseInt(strVal, 0, 64); perr == nil {
				return nil
			}
		case "FLOAT", "FLOAT64", "DOUBLE":
			if _, perr := strconv.ParseFloat(strVal, 64); perr == nil {
				return nil
			}
		default:
			return nil
		}
		if m1(cast.IsSafeCast()) {
			return nil
		}
		line, col := parseLocationLineCol(cast, query)
		firstErr = fmt.Errorf(
			"failed to analyze: INVALID_ARGUMENT: Could not cast literal %q to type %s [at %d:%d]",
			strVal, typeName, line, col,
		)
		return nil
	})
	return firstErr
}

// parseLocationLineCol converts the byte offset of an AST node's start
// location into a 1-indexed (line, column) pair relative to the query
// string. Falls back to (1, 1) if the location isn't available.
func parseLocationLineCol(node googlesql.ASTNode, query string) (int, int) {
	sp, err := node.StartLocation()
	if err != nil || sp == nil {
		return 1, 1
	}
	off, err := sp.GetByteOffset()
	if err != nil {
		return 1, 1
	}
	line, col := 1, 1
	for i := int32(0); i < off && int(i) < len(query); i++ {
		if query[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// preRegisterWildcardTables walks the AST looking for table references
// whose last path segment ends with `*` (e.g. `project.dataset.table_*`).
// For each such reference, build the corresponding wildcard table and
// add it to the wasm-side catalog so the analyzer resolves the
// reference. The googlesql SimpleCatalog has no native wildcard
// support, so we pre-register an equivalent SimpleTable here.
func (a *Analyzer) preRegisterWildcardTables(stmt googlesql.ASTStatementNode) {
	_ = astWalk(stmt, func(node googlesql.ASTNode) error {
		tpath, ok := node.(*googlesql.ASTTablePathExpression)
		if !ok {
			return nil
		}
		pe, _ := tpath.PathExpr()
		if pe == nil {
			return nil
		}
		ids, _ := pe.ToIdentifierVector()
		if len(ids) == 0 {
			return nil
		}
		last := ids[len(ids)-1]
		if last == "" || last[len(last)-1] != '*' {
			return nil
		}
		a.catalog.registerWildcardTableByPath(ids)
		return nil
	})
}

// countPositionalParams walks the AST counting `?` occurrences. Used to
// split positional arguments across statements in a multi-statement
// script, where the resolved-tree walker's limited descent otherwise
// mis-reports per-statement parameter counts.
func countPositionalParams(stmt googlesql.ASTStatementNode) int {
	var n int
	_ = astWalk(stmt, func(node googlesql.ASTNode) error {
		if p, ok := node.(*googlesql.ASTParameterExpr); ok {
			if m1(p.Position()) > 0 {
				n++
			}
		}
		return nil
	})
	return n
}

// collectNamedParamNames walks the AST and returns the lowercased names
// of every distinct `@name` reference. Used to validate that the
// caller-supplied named args cover every name the query references —
// the resolved-tree walker can't surface this directly today.
func collectNamedParamNames(stmt googlesql.ASTStatementNode) []string {
	seen := map[string]struct{}{}
	var out []string
	_ = astWalk(stmt, func(node googlesql.ASTNode) error {
		p, ok := node.(*googlesql.ASTParameterExpr)
		if !ok {
			return nil
		}
		nameID := m1(p.Name())
		if nameID == nil {
			return nil
		}
		raw, _ := nameID.GetAsString()
		name := strings.ToLower(raw)
		if name == "" {
			return nil
		}
		if _, ok := seen[name]; ok {
			return nil
		}
		seen[name] = struct{}{}
		out = append(out, name)
		return nil
	})
	return out
}

func (a *Analyzer) getParameterMode(stmt googlesql.ASTStatementNode) (googlesql.ParameterMode, error) {
	var (
		enabledNamedParameter      bool
		enabledPositionalParameter bool
	)
	_ = astWalk(stmt, func(node googlesql.ASTNode) error {
		switch n := node.(type) {
		case *googlesql.ASTParameterExpr:
			if m1(n.Position()) > 0 {
				enabledPositionalParameter = true
			}
			if m1(n.Name()) != nil {
				enabledNamedParameter = true
			}
		}
		return nil
	})
	if enabledNamedParameter && enabledPositionalParameter {
		return googlesql.ParameterModeParameterNone, fmt.Errorf("named parameter and positional parameter cannot be used together")
	}
	if enabledPositionalParameter {
		return googlesql.ParameterModeParameterPositional, nil
	}
	return googlesql.ParameterModeParameterNamed, nil
}

type stmtActionFunc func() (StmtAction, error)

// sqlRewriteStep is one named pre-processing rewrite with the uniform
// func(string) string signature.
type sqlRewriteStep struct {
	name string
	fn   func(string) string
}

// sqlRewritePipelinePreScriptVars holds the uniform-signature rewrites
// that run before applyScriptVariables. ORDER IS LOAD-BEARING — these
// must run in exactly this sequence.
var sqlRewritePipelinePreScriptVars = []sqlRewriteStep{
	{"typeNameAliases", applyTypeNameAliases},
	{"naiveTimestampUTC", applyNaiveTimestampUTC},
	{"ingestionPartitionRewrite", applyIngestionPartitionRewrite},
}

// sqlRewritePipelinePostScriptVars holds the uniform-signature rewrites
// that run after applyScriptVariables. ORDER IS LOAD-BEARING.
var sqlRewritePipelinePostScriptVars = []sqlRewriteStep{
	{"gapFillRewrite", applyGapFillRewrite},
	{"rangeSessionizeRewrite", applyRangeSessionizeRewrite},
	{"iferrorTypePropagation", applyIferrorTypePropagation},
	{"bigQueryTvfStubs", applyBigQueryTvfStubs},
}

// applySQLRewritePipeline runs each rewrite step in order over query.
func applySQLRewritePipeline(query string, steps []sqlRewriteStep) string {
	for _, step := range steps {
		query = step.fn(query)
	}
	return query
}

// sliceArgsPerStatement pre-slices positional args per statement up
// front. The resolved-tree walker cannot descend into per-statement
// parameter nodes, so we count `?` occurrences directly on the parser
// AST here and hand each statement only the args it owns. Without this,
// a multi-statement script with positional placeholders would hand all
// args to every statement, and the underlying SQLite stmt would reject
// the over-supply.
func (a *Analyzer) sliceArgsPerStatement(parsed *parsedScript, args []driver.NamedValue) ([][]driver.NamedValue, error) {
	stmtArgsList := make([][]driver.NamedValue, len(parsed.stmts))
	remaining := args
	for i, stmt := range parsed.stmts {
		stmtMode, err := a.getParameterMode(stmt)
		if err != nil {
			return nil, err
		}
		switch stmtMode {
		case googlesql.ParameterModeParameterPositional:
			consumed := min(countPositionalParams(stmt), len(remaining))
			stmtArgsList[i] = remaining[:consumed]
			remaining = remaining[consumed:]
		case googlesql.ParameterModeParameterNamed:
			// Validate parameter coverage. The resolved-tree walker
			// can't yet surface the @name list, so we count it
			// directly off the parser AST. We accept either a
			// matching named arg or, for callers that pass plain
			// positional args (database/sql wraps these as
			// NamedArg with empty Name), as long as the total
			// arg count covers every @name reference.
			names := collectNamedParamNames(stmt)
			providedNamed := map[string]struct{}{}
			for _, v := range args {
				if v.Name != "" {
					providedNamed[strings.ToLower(v.Name)] = struct{}{}
				}
			}
			allByName := true
			for _, name := range names {
				if _, ok := providedNamed[name]; !ok {
					allByName = false
					break
				}
			}
			// Defer the coverage check when the caller is in the
			// Prepare path (args is empty). The lower-level Stmt
			// binds at Exec/Query time and will re-validate then.
			if len(args) > 0 && !allByName && len(args) < len(names) {
				return nil, fmt.Errorf("not enough query arguments")
			}
			stmtArgsList[i] = args
		default:
			stmtArgsList[i] = args
		}
	}
	return stmtArgsList, nil
}

// analyzeStatementLocked runs the wasm analyzer for one statement
// under wasmAnalyzeMu. SetParameterMode and declareParameterTypes
// mutate the process-wide shared AnalyzerOptions, and
// AnalyzeStatementFromParserAST reads it; the three steps must form a
// single atomic critical section, otherwise a parallel caller's
// SetParameterMode races in between — observed under load as spurious
// "Positional parameters are not supported" failures.
//
// The lock is deliberately narrow. The action closures returned by
// Analyze are invoked lazily by the caller, and the surrounding work
// they do — wildcard-table pre-registration, StmtAction construction
// — can itself re-enter Analyze; keeping that outside this critical
// section is what prevents a self-deadlock.
func (a *Analyzer) analyzeStatementLocked(stmt googlesql.ASTStatementNode, mode googlesql.ParameterMode, args []driver.NamedValue, query string) (googlesql.ResolvedStatementNode, error) {
	wasmAnalyzeMu.Lock()
	defer wasmAnalyzeMu.Unlock()
	a.opt.SetParameterMode(mode)
	if err := a.declareParameterTypes(mode, args); err != nil {
		return nil, fmt.Errorf("failed to declare parameter types: %w", err)
	}
	out, err := googlesql.AnalyzeStatementFromParserAST(stmt, a.opt, query, a.catalog.catalog, tf())
	if err != nil {
		return nil, fmt.Errorf("failed to analyze: %w", err)
	}
	stmtNode, _ := out.ResolvedStatement()
	return stmtNode, nil
}

func (a *Analyzer) Analyze(ctx context.Context, conn *Conn, query string, args []driver.NamedValue) ([]stmtActionFunc, error) {
	// One goroutine at a time owns the wasm runtime / shared
	// AnalyzerOptions. See the doc on sharedAnalyzerOpt above.
	wasmAnalyzeMu.Lock()
	defer wasmAnalyzeMu.Unlock()
	if err := a.catalog.Sync(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to sync catalog: %w", err)
	}
	// SQL-rewrite pipeline. ORDER IS LOAD-BEARING. The two rewrites
	// with non-uniform signatures (applyMixedParameterRewrite needs the
	// arg slice, applyScriptVariables needs ctx/conn) stay as explicit
	// calls in their exact positions relative to the uniform pipeline
	// steps.
	query, args = applyMixedParameterRewrite(query, args)
	query = applySQLRewritePipeline(query, sqlRewritePipelinePreScriptVars)
	query = applyScriptVariables(ctx, query, conn)
	query = applySQLRewritePipeline(query, sqlRewritePipelinePostScriptVars)
	parsed, err := a.parseScript(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse statements: %w", err)
	}
	funcMap := map[string]*FunctionSpec{}
	for _, spec := range a.catalog.getFunctions(a.namePath) {
		funcMap[spec.FuncName()] = spec
	}
	tvfMap := map[string]*TVFSpec{}
	for _, spec := range a.catalog.getTVFs(a.namePath) {
		tvfMap[spec.TVFName()] = spec
	}
	actionFuncs := make([]stmtActionFunc, 0, len(parsed.stmts))
	stmtArgsList, err := a.sliceArgsPerStatement(parsed, args)
	if err != nil {
		return nil, err
	}
	for i, stmt := range parsed.stmts {
		stmtArgs := stmtArgsList[i]
		// Script-level control-flow AST nodes (IF / WHILE / LOOP /
		// REPEAT / FOR IN / CASE / RAISE / BREAK / CONTINUE / RETURN
		// / CALL) are not subclasses of ResolvedStatement, so the
		// analyzer's AnalyzeStatementFromParserAST entry point
		// refuses them with "Statement not supported". Recognise
		// them at the parser-AST level and accept as no-ops so the
		// driver's surface stays consistent with BigQuery's
		// procedural-language reference — the body still has to
		// parse, just nothing happens at execution.
		if isScriptControlStmt(stmt) {
			actionFuncs = append(actionFuncs, func() (StmtAction, error) {
				return &NoopStmtAction{}, nil
			})
			continue
		}
		actionFuncs = append(actionFuncs, func() (StmtAction, error) {
			mode, err := a.getParameterMode(stmt)
			if err != nil {
				return nil, err
			}
			if err := validateLiteralCasts(stmt, query); err != nil {
				return nil, err
			}
			a.preRegisterWildcardTables(stmt)
			stmtNode, err := a.analyzeStatementLocked(stmt, mode, stmtArgs, query)
			if err != nil {
				return nil, err
			}
			ctx = a.context(ctx, funcMap, tvfMap)
			ctx = withSystemVars(ctx, conn.systemVars)
			ctx = withConn(ctx, conn)
			action, err := a.newStmtAction(ctx, query, stmtArgs, stmtNode)
			if err != nil {
				return nil, err
			}
			return action, nil
		})
	}
	return actionFuncs, nil
}

func (a *Analyzer) context(
	ctx context.Context,
	funcMap map[string]*FunctionSpec,
	tvfMap map[string]*TVFSpec,
) context.Context {
	ctx = withAnalyzer(ctx, a)
	ctx = withNamePath(ctx, a.namePath)
	ctx = withColumnRefMap(ctx, map[string]string{})
	ctx = withTableNameToColumnListMap(ctx, map[string][]*googlesql.ResolvedColumn{})
	ctx = withFuncMap(ctx, funcMap)
	ctx = withTVFMap(ctx, tvfMap)
	ctx = withAnalyticOrderColumnNames(ctx, &analyticOrderColumnNames{})
	return ctx
}

// declareParameterTypes calls AddQueryParameter on the analyzer options
// for each named value, deducing the GoogleSQL type from the Go runtime
// type. Without this, the analyzer falls back to default type inference
// (typically INT64) when a parameter is used in a context whose
// required type cannot be inferred from neighbouring expressions —
// e.g. UNNEST(@arr) which expects an ARRAY but cannot infer the
// element type.
//
// Positional parameters are deliberately left undeclared. go-googlesql
// disallows mixing declared positional parameters with
// SetAllowUndeclaredParameters(true), and the typing inference path
// the latter enables is the canonical way callers pass `?` placeholder
// values without an explicit schema.
//
// Existing declared parameters are cleared each call so the same
// options instance can be reused across queries.
func (a *Analyzer) declareParameterTypes(mode googlesql.ParameterMode, args []driver.NamedValue) error {
	if err := a.opt.ClearQueryParameters(); err != nil {
		return err
	}
	if mode != googlesql.ParameterModeParameterNamed {
		return nil
	}
	for _, v := range args {
		if v.Name == "" {
			continue
		}
		typ, ok := googleSQLTypeForValue(v.Value)
		if !ok {
			continue
		}
		if err := a.opt.AddQueryParameter(strings.ToLower(v.Name), typ); err != nil {
			return err
		}
	}
	return nil
}

// googleSQLTypeForValue returns the GoogleSQL type for a value, when
// it can be deduced reliably. Conn.CheckNamedValue runs ahead of
// analysis and base64-encodes complex shapes; for those, we decode
// just enough to recover the original layout type. Unknown shapes
// return ok=false and the analyzer falls back to its undeclared-
// parameter inference.
func googleSQLTypeForValue(v any) (googlesql.Googlesql_TypeNode, bool) {
	if v == nil {
		return nil, false
	}
	factory := tf()
	switch x := v.(type) {
	case bool:
		return makeSimpleType(factory, googlesql.TypeKindTypeBool)
	case int, int8, int16, int32, int64:
		return makeSimpleType(factory, googlesql.TypeKindTypeInt64)
	case uint, uint8, uint16, uint32, uint64:
		return makeSimpleType(factory, googlesql.TypeKindTypeInt64)
	case float32, float64:
		return makeSimpleType(factory, googlesql.TypeKindTypeDouble)
	case string:
		// Try to decode as a googlesqlite layout first — Conn.CheckNamedValue
		// has already encoded most non-primitive values into a base64
		// string. If decoding succeeds, deduce the type from the layout.
		if t, ok := googleSQLTypeForEncodedString(factory, x); ok {
			return t, true
		}
		return makeSimpleType(factory, googlesql.TypeKindTypeString)
	case []byte:
		return makeSimpleType(factory, googlesql.TypeKindTypeBytes)
	case []bool:
		return makeArrayType(factory, googlesql.TypeKindTypeBool)
	case []int, []int8, []int16, []int32, []int64:
		return makeArrayType(factory, googlesql.TypeKindTypeInt64)
	case []uint, []uint16, []uint32, []uint64:
		return makeArrayType(factory, googlesql.TypeKindTypeInt64)
	case []float32, []float64:
		return makeArrayType(factory, googlesql.TypeKindTypeDouble)
	case []string:
		return makeArrayType(factory, googlesql.TypeKindTypeString)
	default:
		_ = x
		return nil, false
	}
}

func googleSQLTypeForEncodedString(factory *googlesql.TypeFactory, s string) (googlesql.Googlesql_TypeNode, bool) {
	if len(s) == 0 {
		return nil, false
	}
	decoded, err := DecodeValue(s)
	if err != nil || decoded == nil {
		return nil, false
	}
	switch dv := decoded.(type) {
	case value.IntValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeInt64)
	case value.FloatValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeDouble)
	case value.BoolValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeBool)
	case value.StringValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeString)
	case value.BytesValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeBytes)
	case value.DateValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeDate)
	case value.TimestampValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeTimestamp)
	case value.DatetimeValue:
		return makeSimpleType(factory, googlesql.TypeKindTypeDatetime)
	case *value.ArrayValue:
		// element type from the first non-nil element
		var elemKind googlesql.TypeKind = googlesql.TypeKindTypeString
		for _, elem := range dv.Values {
			if elem == nil {
				continue
			}
			switch elem.(type) {
			case value.IntValue:
				elemKind = googlesql.TypeKindTypeInt64
			case value.FloatValue:
				elemKind = googlesql.TypeKindTypeDouble
			case value.BoolValue:
				elemKind = googlesql.TypeKindTypeBool
			case value.StringValue:
				elemKind = googlesql.TypeKindTypeString
			case value.BytesValue:
				elemKind = googlesql.TypeKindTypeBytes
			case value.DateValue:
				elemKind = googlesql.TypeKindTypeDate
			case value.TimestampValue:
				elemKind = googlesql.TypeKindTypeTimestamp
			case value.DatetimeValue:
				elemKind = googlesql.TypeKindTypeDatetime
			default:
				return nil, false
			}
			break
		}
		return makeArrayType(factory, elemKind)
	}
	return nil, false
}

func makeSimpleType(factory *googlesql.TypeFactory, kind googlesql.TypeKind) (googlesql.Googlesql_TypeNode, bool) {
	t, err := factory.MakeSimpleType(kind)
	if err != nil || t == nil {
		return nil, false
	}
	return t, true
}

func makeArrayType(factory *googlesql.TypeFactory, kind googlesql.TypeKind) (googlesql.Googlesql_TypeNode, bool) {
	elem, err := factory.MakeSimpleType(kind)
	if err != nil || elem == nil {
		return nil, false
	}
	arr, err := factory.MakeArrayType2(elem)
	if err != nil || arr == nil {
		return nil, false
	}
	return arr, true
}

func (a *Analyzer) analyzeTemplatedFunctionWithRuntimeArgument(ctx context.Context, query string) (*FunctionSpec, error) {
	out, err := googlesql.AnalyzeStatement(query, a.opt, a.catalog.catalog, tf())
	if err != nil {
		return nil, fmt.Errorf("failed to analyze: %w", err)
	}
	node, _ := out.ResolvedStatement()
	stmt, ok := node.(*googlesql.ResolvedCreateFunctionStmt)
	if !ok {
		return nil, fmt.Errorf("unexpected create function query %s", query)
	}
	spec, err := newFunctionSpec(ctx, a.namePath, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create function spec: %w", err)
	}
	return spec, nil
}

func (a *Analyzer) newStmtAction(ctx context.Context, query string, args []driver.NamedValue, node googlesql.ResolvedStatementNode) (StmtAction, error) {
	kind, _ := node.NodeKind()
	switch kind {
	case googlesql.ResolvedNodeKindResolvedCreateTableStmt:
		return a.newCreateTableStmtAction(ctx, query, args, node.(*googlesql.ResolvedCreateTableStmt))
	case googlesql.ResolvedNodeKindResolvedCreateTableAsSelectStmt:
		ctx = withUseColumnID(ctx)
		return a.newCreateTableAsSelectStmtAction(ctx, query, args, node.(*googlesql.ResolvedCreateTableAsSelectStmt))
	case googlesql.ResolvedNodeKindResolvedCreateFunctionStmt:
		return a.newCreateFunctionStmtAction(ctx, query, args, node.(*googlesql.ResolvedCreateFunctionStmt))
	case googlesql.ResolvedNodeKindResolvedCreateViewStmt:
		ctx = withUseColumnID(ctx)
		return a.newCreateViewStmtAction(ctx, query, args, node.(*googlesql.ResolvedCreateViewStmt))
	case googlesql.ResolvedNodeKindResolvedDropStmt:
		return a.newDropStmtAction(ctx, query, args, node.(*googlesql.ResolvedDropStmt))
	case googlesql.ResolvedNodeKindResolvedDropFunctionStmt:
		return a.newDropFunctionStmtAction(ctx, query, args, node.(*googlesql.ResolvedDropFunctionStmt))
	case googlesql.ResolvedNodeKindResolvedInsertStmt, googlesql.ResolvedNodeKindResolvedUpdateStmt, googlesql.ResolvedNodeKindResolvedDeleteStmt:
		return a.newDMLStmtAction(ctx, query, args, node)
	case googlesql.ResolvedNodeKindResolvedTruncateStmt:
		return a.newTruncateStmtAction(ctx, query, args, node.(*googlesql.ResolvedTruncateStmt))
	case googlesql.ResolvedNodeKindResolvedMergeStmt:
		ctx = withUseColumnID(ctx)
		return a.newMergeStmtAction(ctx, query, args, node.(*googlesql.ResolvedMergeStmt))
	case googlesql.ResolvedNodeKindResolvedQueryStmt:
		ctx = withUseColumnID(ctx)
		return a.newQueryStmtAction(ctx, query, args, node.(*googlesql.ResolvedQueryStmt))
	case googlesql.ResolvedNodeKindResolvedCreateTableFunctionStmt:
		ctx = withUseColumnID(ctx)
		return a.newCreateTableFunctionStmtAction(ctx, query, args, node.(*googlesql.ResolvedCreateTableFunctionStmt))
	case googlesql.ResolvedNodeKindResolvedExportDataStmt:
		// EXPORT DATA OPTIONS(...) AS SELECT ... — driver-side
		// behaviour is to evaluate the inner SELECT and return its
		// rows. The emulator (or whatever sits on top of the driver)
		// handles the actual write to the URI; the OPTION list is
		// surfaced for that layer through the resolved AST.
		ctx = withUseColumnID(ctx)
		return a.newExportDataStmtAction(ctx, query, args, node.(*googlesql.ResolvedExportDataStmt))
	case googlesql.ResolvedNodeKindResolvedBeginStmt:
		return a.newBeginStmtAction(ctx, query, args, node)
	case googlesql.ResolvedNodeKindResolvedCommitStmt:
		return a.newCommitStmtAction(ctx, query, args, node)
	case googlesql.ResolvedNodeKindResolvedAssignmentStmt:
		return a.newAssignmentStmtAction(ctx, query, args, node.(*googlesql.ResolvedAssignmentStmt))
	case googlesql.ResolvedNodeKindResolvedCreatePropertyGraphStmt:
		return a.newCreatePropertyGraphStmtAction(ctx, node.(*googlesql.ResolvedCreatePropertyGraphStmt))
	case
		googlesql.ResolvedNodeKindResolvedAlterTableStmt,
		googlesql.ResolvedNodeKindResolvedCreateSchemaStmt,
		googlesql.ResolvedNodeKindResolvedGrantStmt,
		googlesql.ResolvedNodeKindResolvedRevokeStmt,
		googlesql.ResolvedNodeKindResolvedAssertStmt,
		googlesql.ResolvedNodeKindResolvedExecuteImmediateStmt,
		googlesql.ResolvedNodeKindResolvedAuxLoadDataStmt,
		googlesql.ResolvedNodeKindResolvedCreateProcedureStmt:
		// Metadata-only or out-of-scope statements: succeed without
		// observable side effects. The analyzer has already type-
		// checked the body, so the call is at least syntactically
		// validated for downstream tooling.
		return &NoopStmtAction{}, nil
	}
	dbg, _ := node.DebugString()
	return nil, fmt.Errorf("unsupported stmt %s", dbg)
}

// newCreatePropertyGraphStmtAction handles CREATE [OR REPLACE]
// PROPERTY GRAPH ... by recording the spec in the catalog AND
// registering a SimplePropertyGraph in the go-googlesql
// SimpleCatalog so subsequent graph queries (`GRAPH X MATCH ...`)
// can resolve. The statement has no SQLite-side effect.
func (a *Analyzer) newCreatePropertyGraphStmtAction(ctx context.Context, node *googlesql.ResolvedCreatePropertyGraphStmt) (StmtAction, error) {
	namePath, _ := node.NamePath()
	name := strings.Join(namePath, ".")
	spec := &propertyGraphSpec{Name: name}
	if nodeTables, err := node.NodeTableList(); err == nil {
		for _, nt := range nodeTables {
			if alias, err := nt.Alias(); err == nil && alias != "" {
				spec.NodeTables = append(spec.NodeTables, alias)
			}
		}
	}
	if edgeTables, err := node.EdgeTableList(); err == nil {
		for _, et := range edgeTables {
			if alias, err := et.Alias(); err == nil && alias != "" {
				spec.EdgeTables = append(spec.EdgeTables, alias)
			}
		}
	}
	a.catalog.AddPropertyGraph(spec)
	// Register a SimplePropertyGraph in the analyzer catalog so
	// `GRAPH <name> MATCH ...` queries resolve. Errors surface
	// to the caller so we don't silently end up with a graph
	// that exists in our Go map but not the analyzer catalog.
	pg, err := buildSimplePropertyGraph(node)
	if err != nil {
		return nil, fmt.Errorf("build property graph %q: %w", name, err)
	}
	if pg != nil {
		// A previous registration under the same name is fine for
		// CREATE OR REPLACE; ignore the duplicate-add error and let
		// the new spec coexist for metadata purposes.
		_ = a.catalog.catalog.AddPropertyGraph(pg)
	}
	return &noopStmtAction{}, nil
}

// noopStmtAction is a do-nothing StmtAction used by metadata-only
// DDL statements (currently CREATE PROPERTY GRAPH).
type noopStmtAction struct{}

func (a *noopStmtAction) Prepare(ctx context.Context, conn *Conn) (driver.Stmt, error) {
	return nil, nil
}
func (a *noopStmtAction) ExecContext(ctx context.Context, conn *Conn) (driver.Result, error) {
	return &Result{conn: conn}, nil
}
func (a *noopStmtAction) QueryContext(ctx context.Context, conn *Conn) (*Rows, error) {
	return &Rows{conn: conn}, nil
}
func (a *noopStmtAction) Args() []any                                   { return nil }
func (a *noopStmtAction) Cleanup(ctx context.Context, conn *Conn) error { return nil }

func (a *Analyzer) newCreateTableStmtAction(ctx context.Context, query string, args []driver.NamedValue, node *googlesql.ResolvedCreateTableStmt) (*CreateTableStmtAction, error) {
	spec := newTableSpecWithQuery(ctx, a.namePath, query, node)
	queryArgs, err := getArgsFromParams(args, nil)
	if err != nil {
		return nil, err
	}
	return &CreateTableStmtAction{
		query:           query,
		spec:            spec,
		args:            queryArgs,
		catalog:         a.catalog,
		isAutoIndexMode: a.isAutoIndexMode,
	}, nil
}

func (a *Analyzer) newCreateTableAsSelectStmtAction(ctx context.Context, _ string, args []driver.NamedValue, node *googlesql.ResolvedCreateTableAsSelectStmt) (*CreateTableStmtAction, error) {
	query, params, err := collectFormatParams(ctx, m1(node.Query()))
	if err != nil {
		return nil, err
	}
	spec := newTableAsSelectSpec(ctx, a.namePath, query, node)
	queryArgs, err := getArgsFromParams(args, params)
	if err != nil {
		return nil, err
	}
	return &CreateTableStmtAction{
		query:           query,
		spec:            spec,
		args:            queryArgs,
		catalog:         a.catalog,
		isAutoIndexMode: a.isAutoIndexMode,
	}, nil
}

func (a *Analyzer) newCreateFunctionStmtAction(ctx context.Context, query string, _ []driver.NamedValue, node *googlesql.ResolvedCreateFunctionStmt) (*CreateFunctionStmtAction, error) {
	var spec *FunctionSpec
	if a.resultTypeIsTemplatedType(m1(node.Signature())) {
		realStmts, err := a.inferTemplatedTypeByRealType(query, node)
		if err != nil {
			return nil, err
		}
		templatedFuncSpec, err := newTemplatedFunctionSpec(ctx, a.namePath, node, realStmts)
		if err != nil {
			return nil, err
		}
		spec = templatedFuncSpec
	} else {
		funcSpec, err := newFunctionSpec(ctx, a.namePath, node)
		if err != nil {
			return nil, fmt.Errorf("failed to create function spec: %w", err)
		}
		spec = funcSpec
	}
	return &CreateFunctionStmtAction{
		spec:    spec,
		catalog: a.catalog,
		funcMap: funcMapFromContext(ctx),
	}, nil
}

func (a *Analyzer) newCreateViewStmtAction(ctx context.Context, _ string, _ []driver.NamedValue, node *googlesql.ResolvedCreateViewStmt) (*CreateViewStmtAction, error) {
	query, err := newNode(m1(node.Query())).FormatSQL(ctx)
	if err != nil {
		return nil, err
	}
	spec := newTableAsViewSpec(a.namePath, query, node)
	return &CreateViewStmtAction{
		query:   query,
		spec:    spec,
		catalog: a.catalog,
	}, nil
}

func (a *Analyzer) resultTypeIsTemplatedType(sig *googlesql.FunctionSignature) bool {
	if !m1(sig.IsTemplated()) {
		return false
	}
	rt := m1(sig.ResultType())
	if rt == nil {
		return false
	}
	return m1(rt.IsTemplated())
}

var inferTypes = []string{
	"INT64", "DOUBLE", "BOOL", "STRING", "BYTES",
	"JSON", "DATE", "DATETIME", "TIME", "TIMESTAMP",
	"INTERVAL", "GEOGRAPHY",
	"STRUCT<>",
}

func (a *Analyzer) inferTemplatedTypeByRealType(query string, node *googlesql.ResolvedCreateFunctionStmt) ([]*googlesql.ResolvedCreateFunctionStmt, error) {
	var stmts []*googlesql.ResolvedCreateFunctionStmt
	for _, typ := range inferTypes {
		if out, err := googlesql.AnalyzeStatement(a.buildScalarTypeFuncFromTemplatedFunc(node, typ), a.opt, a.catalog.catalog, tf()); err == nil {
			stmts = append(stmts, m1(out.ResolvedStatement()).(*googlesql.ResolvedCreateFunctionStmt))
		}
	}
	if len(stmts) != 0 {
		return stmts, nil
	}
	for _, typ := range inferTypes {
		if out, err := googlesql.AnalyzeStatement(a.buildArrayTypeFuncFromTemplatedFunc(node, typ), a.opt, a.catalog.catalog, tf()); err == nil {
			stmts = append(stmts, m1(out.ResolvedStatement()).(*googlesql.ResolvedCreateFunctionStmt))
		}
	}
	if len(stmts) != 0 {
		return stmts, nil
	}
	return nil, fmt.Errorf("failed to infer templated function result type for %s", query)
}

func (a *Analyzer) buildScalarTypeFuncFromTemplatedFunc(node *googlesql.ResolvedCreateFunctionStmt, realType string) string {
	signature, _ := node.Signature()
	var args []string
	for _, arg := range m1(signature.Arguments()) {
		typ := realType
		if !m1(arg.IsTemplated()) {
			typ = newType(m1(arg.Type())).FormatType()
		}
		args = append(args, fmt.Sprintf("%s %s", m1(arg.ArgumentName()), typ))
	}
	return fmt.Sprintf(
		"CREATE TEMP FUNCTION __googlesqlite_func__(%s) as (%s)",
		strings.Join(args, ","),
		m1(node.Code()),
	)
}

func (a *Analyzer) buildArrayTypeFuncFromTemplatedFunc(node *googlesql.ResolvedCreateFunctionStmt, realType string) string {
	signature, _ := node.Signature()
	var args []string
	for _, arg := range m1(signature.Arguments()) {
		typ := fmt.Sprintf("ARRAY<%s>", realType)
		if !m1(arg.IsTemplated()) {
			typ = newType(m1(arg.Type())).FormatType()
		}
		args = append(args, fmt.Sprintf("%s %s", m1(arg.ArgumentName()), typ))
	}
	return fmt.Sprintf(
		"CREATE TEMP FUNCTION __googlesqlite_func__(%s) as (%s)",
		strings.Join(args, ","),
		m1(node.Code()),
	)
}

func (a *Analyzer) newDropStmtAction(ctx context.Context, query string, args []driver.NamedValue, node *googlesql.ResolvedDropStmt) (*DropStmtAction, error) {
	formattedQuery, params, err := collectFormatParams(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("failed to format query %s: %w", query, err)
	}
	if formattedQuery == "" {
		return nil, fmt.Errorf("failed to format query %s", query)
	}
	queryArgs, err := getArgsFromParams(args, params)
	if err != nil {
		return nil, err
	}
	objectType, _ := node.ObjectType()
	name := a.namePath.format2(node.NamePath())
	return &DropStmtAction{
		name:           name,
		objectType:     objectType,
		funcMap:        funcMapFromContext(ctx),
		catalog:        a.catalog,
		query:          query,
		formattedQuery: formattedQuery,
		args:           queryArgs,
	}, nil
}

func (a *Analyzer) newDropFunctionStmtAction(ctx context.Context, query string, args []driver.NamedValue, node *googlesql.ResolvedDropFunctionStmt) (*DropStmtAction, error) {
	queryArgs, err := getArgsFromParams(args, nil)
	if err != nil {
		return nil, err
	}
	name := a.namePath.format2(node.NamePath())
	return &DropStmtAction{
		name:       name,
		objectType: "FUNCTION",
		funcMap:    funcMapFromContext(ctx),
		catalog:    a.catalog,
		query:      query,
		args:       queryArgs,
	}, nil
}

func (a *Analyzer) newDMLStmtAction(ctx context.Context, query string, args []driver.NamedValue, node googlesql.ResolvedNode) (*DMLStmtAction, error) {
	// For INSERT statements, reshape struct-valued args against the
	// target InsertColumnList types so sparse Go maps expand to match
	// the declared STRUCT field order. See reshapeInsertArgs.
	args = reshapeInsertArgs(args, node)
	formattedQuery, params, err := collectFormatParams(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("failed to format query %s: %w", query, err)
	}
	if formattedQuery == "" {
		return nil, fmt.Errorf("failed to format query %s", query)
	}
	queryArgs, err := getArgsFromParams(args, params)
	if err != nil {
		return nil, err
	}
	return &DMLStmtAction{
		query:          query,
		params:         params,
		args:           queryArgs,
		colTypes:       insertColumnTypes(node),
		formattedQuery: formattedQuery,
	}, nil
}

// insertColumnTypes returns the per-position destination column types
// for a ResolvedInsertStmt, or nil for any other resolved node. The
// types feed DMLStmt.Exec's reshape pass so positional `?` placeholders
// can be matched against the schema even when the resolved-tree
// walker does not surface a typed parameter list.
func insertColumnTypes(node googlesql.ResolvedNode) []googlesql.Googlesql_TypeNode {
	insert, ok := node.(*googlesql.ResolvedInsertStmt)
	if !ok {
		return nil
	}
	cols, err := insert.InsertColumnList()
	if err != nil || len(cols) == 0 {
		return nil
	}
	out := make([]googlesql.Googlesql_TypeNode, len(cols))
	for i, col := range cols {
		t, err := col.Type()
		if err != nil {
			continue
		}
		out[i] = t
	}
	return out
}

// reshapeInsertArgs adjusts args for ResolvedInsertStmt so struct-typed
// columns receive fully-populated struct values (missing fields filled
// with nil, fields in declaration order). Go callers sometimes pass
// sparse maps that don't enumerate every declared field; without this,
// downstream STRUCT_FIELD(value, index) reads land out-of-range.
func reshapeInsertArgs(args []driver.NamedValue, node googlesql.ResolvedNode) []driver.NamedValue {
	if len(args) == 0 {
		return args
	}
	insert, ok := node.(*googlesql.ResolvedInsertStmt)
	if !ok {
		return args
	}
	cols, err := insert.InsertColumnList()
	if err != nil || len(cols) == 0 {
		return args
	}
	// Align args to columns positionally. Fewer args than columns is
	// fine (extra columns may receive defaults); extra args are passed
	// through unchanged (they may belong to WHERE clauses).
	out := make([]driver.NamedValue, len(args))
	copy(out, args)
	n := min(len(cols), len(out))
	for i := 0; i < n; i++ {
		colType, err := cols[i].Type()
		if err != nil || colType == nil {
			continue
		}
		out[i].Value = reshapeArgToType(out[i].Value, colType)
	}
	return out
}

// reshapeArgToType takes a single arg value (possibly a
// googlesqlite-encoded base64 string or a raw Go value) and a declared
// googlesql type, and returns a value whose underlying structure
// matches the declared type. Only STRUCT reshape is interesting here;
// other shapes pass through unchanged.
func reshapeArgToType(v any, t googlesql.Googlesql_TypeNode) any {
	if v == nil || t == nil {
		return v
	}
	kind, _ := t.Kind()
	switch kind {
	case googlesql.TypeKindTypeStruct,
		googlesql.TypeKindTypeArray,
		googlesql.TypeKindTypeDate,
		googlesql.TypeKindTypeTimestamp:
		// Reshape candidates — fall through.
	default:
		return v
	}
	val, err := DecodeValue(v)
	if err != nil || val == nil {
		return v
	}
	val = reshapeBQEStructShape(val, t)
	reshaped, err := CastValue(t, val)
	if err != nil || reshaped == nil {
		return v
	}
	out, err := EncodeValue(reshaped)
	if err != nil {
		return v
	}
	return out
}

// reshapeBQEStructShape walks a decoded value against the declared
// type and rewrites the BigQuery wire formats that valueFromGoReflectValue
// cannot disambiguate without a schema hint. Currently:
//
//   - STRUCT-as-list: the BigQuery JSON wire encodes a struct value as
//     an ordered list of single-key objects (e.g.
//     `[{"A":4},{"B":"5"}]`); merge them back into a single StructValue.
//   - DATE-as-int: BigQuery sends DATE values as epoch days; convert
//     IntValue to DateValue.
//   - TIMESTAMP-as-int: BigQuery sends TIMESTAMP values as epoch
//     microseconds; convert IntValue to TimestampValue.
//
// The cast / format / scan pipeline downstream sees the canonical shape.
func reshapeBQEStructShape(v value.Value, t googlesql.Googlesql_TypeNode) value.Value {
	if v == nil || t == nil {
		return v
	}
	kind, _ := t.Kind()
	switch kind {
	case googlesql.TypeKindTypeStruct:
		if av, ok := v.(*value.ArrayValue); ok {
			if merged, ok := mergeSingleKeyStructArray(av); ok {
				return merged
			}
		}
	case googlesql.TypeKindTypeArray:
		av, ok := v.(*value.ArrayValue)
		if !ok {
			return v
		}
		at, ok := t.(*googlesql.ArrayType)
		if !ok {
			return v
		}
		elemType, err := at.ElementType()
		if err != nil || elemType == nil {
			return v
		}
		out := &value.ArrayValue{}
		for _, elem := range av.Values {
			out.Values = append(out.Values, reshapeBQEStructShape(elem, elemType))
		}
		return out
	case googlesql.TypeKindTypeDate:
		if iv, ok := v.(value.IntValue); ok {
			return value.DateValue(time.Unix(int64(iv)*86400, 0).UTC())
		}
	case googlesql.TypeKindTypeTimestamp:
		if iv, ok := v.(value.IntValue); ok {
			micros := int64(iv)
			return value.TimestampValue(time.Unix(micros/1_000_000, (micros%1_000_000)*1_000).UTC())
		}
	}
	return v
}

// mergeSingleKeyStructArray collapses an ArrayValue whose elements are
// each StructValues of length 1 into a single StructValue. Returns the
// merged StructValue and true on success, or (nil, false) if any
// element is not a single-key struct or if a key repeats.
func mergeSingleKeyStructArray(av *value.ArrayValue) (*value.StructValue, bool) {
	merged := &value.StructValue{M: map[string]value.Value{}}
	for _, elem := range av.Values {
		sv, ok := elem.(*value.StructValue)
		if !ok || len(sv.Keys) != 1 {
			return nil, false
		}
		key := sv.Keys[0]
		if _, dup := merged.M[key]; dup {
			return nil, false
		}
		merged.Keys = append(merged.Keys, key)
		merged.Values = append(merged.Values, sv.Values[0])
		merged.M[key] = sv.Values[0]
	}
	if len(merged.Keys) == 0 {
		return nil, false
	}
	return merged, true
}

func (a *Analyzer) newQueryStmtAction(ctx context.Context, query string, args []driver.NamedValue, node *googlesql.ResolvedQueryStmt) (*QueryStmtAction, error) {
	outputColumns := []*ColumnSpec{}
	for _, col := range m1(node.OutputColumnList()) {
		outputColumns = append(outputColumns, &ColumnSpec{
			Name: m1(col.Name()),
			Type: newType(m1(m1(col.Column()).Type())),
		})
	}
	formattedQuery, params, err := collectFormatParams(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("failed to format query %s: %w", query, err)
	}
	if formattedQuery == "" {
		return nil, fmt.Errorf("failed to format query %s", query)
	}
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "GRAPH ") {
		fmt.Fprintf(debugStream(), "[googlesqlite][graph] in : %s\n[googlesqlite][graph] out: %s\n", query, formattedQuery)
	}
	queryArgs, err := getArgsFromParams(args, params)
	if err != nil {
		return nil, err
	}
	return &QueryStmtAction{
		query:          query,
		params:         params,
		args:           queryArgs,
		formattedQuery: formattedQuery,
		outputColumns:  outputColumns,
		isExplainMode:  a.isExplainMode,
	}, nil
}

// newCreateTableFunctionStmtAction handles `CREATE TABLE FUNCTION`
// DDL. The TVF body is captured into a TVFSpec, the spec is
// registered on the catalog so `SELECT * FROM <tvf>(...)` resolves at
// the analyzer, and the spec is persisted via the standard catalog
// table.
func (a *Analyzer) newCreateTableFunctionStmtAction(ctx context.Context, _ string, _ []driver.NamedValue, node *googlesql.ResolvedCreateTableFunctionStmt) (*CreateTableFunctionStmtAction, error) {
	spec, err := newTVFSpec(ctx, a.namePath, node)
	if err != nil {
		return nil, fmt.Errorf("failed to create TVF spec: %w", err)
	}
	return &CreateTableFunctionStmtAction{
		spec:    spec,
		catalog: a.catalog,
		tvfMap:  tvfMapFromContext(ctx),
	}, nil
}

// newExportDataStmtAction lowers an
// `EXPORT DATA OPTIONS(uri = '...', format = '...') AS <query>` statement
// into an ExportDataStmtAction. The OPTIONS clause is parsed off the
// resolved AST (uri is required, format defaults to CSV — matching real
// BigQuery), the inner query is formatted in the usual way, and the action
// at execute time runs the inner query and streams its rows through the
// format encoder into the writer registered for the URI's scheme (see
// exportdata.RegisterURIWriter).
func (a *Analyzer) newExportDataStmtAction(ctx context.Context, query string, args []driver.NamedValue, node *googlesql.ResolvedExportDataStmt) (StmtAction, error) {
	uri, formatStr, err := readExportDataOptions(ctx, m1(node.OptionList()))
	if err != nil {
		return nil, err
	}
	if uri == "" {
		return nil, fmt.Errorf("EXPORT DATA: required option `uri` is missing")
	}
	format, err := exportdata.ParseFormat(formatStr)
	if err != nil {
		return nil, err
	}

	outputColumns := []*ColumnSpec{}
	for _, col := range m1(node.OutputColumnList()) {
		outputColumns = append(outputColumns, &ColumnSpec{
			Name: m1(col.Name()),
			Type: newType(m1(m1(col.Column()).Type())),
		})
	}
	formattedQuery, params, err := collectFormatParams(ctx, m1(node.Query()))
	if err != nil {
		return nil, fmt.Errorf("failed to format export-data inner query: %w", err)
	}
	if formattedQuery == "" {
		return nil, fmt.Errorf("failed to format export-data inner query")
	}
	queryArgs, err := getArgsFromParams(args, params)
	if err != nil {
		return nil, err
	}
	return NewExportDataStmtAction(query, formattedQuery, params, queryArgs, outputColumns, uri, format), nil
}

// readExportDataOptions extracts the string-valued options off a ResolvedOption
// list. The option's value is read directly from its resolved literal — no
// SQL re-format / unquote dance — so escapes and quoting handled by the
// analyzer are honoured by construction. Unknown options are tolerated for
// forward compatibility (real BigQuery accepts options like `overwrite`,
// `header`, `compression` etc. that this engine has not implemented yet, and
// silently ignoring them on parse keeps callers' SQL portable until they
// are added).
func readExportDataOptions(_ context.Context, opts []*googlesql.ResolvedOption) (uri string, format string, err error) {
	for _, opt := range opts {
		name, _ := opt.Name()
		val, ok, verr := exportDataStringOption(opt)
		if verr != nil {
			return "", "", fmt.Errorf("EXPORT DATA: read option %q: %w", name, verr)
		}
		if !ok {
			// Non-literal or non-string options are not relevant to
			// the two we currently consume; skip silently.
			continue
		}
		switch strings.ToLower(name) {
		case "uri":
			uri = val
		case "format":
			format = val
		}
	}
	return uri, format, nil
}

// exportDataStringOption reads a string value from a ResolvedOption whose
// value is a ResolvedLiteral. Returns (value, true, nil) on success,
// (_, false, nil) for non-literal / non-string options, or an error if the
// underlying value extraction fails.
func exportDataStringOption(opt *googlesql.ResolvedOption) (string, bool, error) {
	expr, _ := opt.Value()
	lit, ok := expr.(*googlesql.ResolvedLiteral)
	if !ok {
		return "", false, nil
	}
	v, err := lit.Value()
	if err != nil {
		return "", false, err
	}
	kind, err := v.TypeKind()
	if err != nil {
		return "", false, err
	}
	if kind != googlesql.TypeKindTypeString {
		return "", false, nil
	}
	s, err := v.StringValue()
	if err != nil {
		return "", false, err
	}
	return s, true, nil
}

func (a *Analyzer) newBeginStmtAction(ctx context.Context, query string, args []driver.NamedValue, node googlesql.ResolvedNode) (*BeginStmtAction, error) {
	return &BeginStmtAction{}, nil
}

func (a *Analyzer) newCommitStmtAction(ctx context.Context, query string, args []driver.NamedValue, node googlesql.ResolvedNode) (*CommitStmtAction, error) {
	return &CommitStmtAction{}, nil
}

// newAssignmentStmtAction handles `SET @@var = expr`. The Target
// must be a SystemVariable; other lvalues (script DECLARE'd
// variables) are not yet supported and would arrive here under
// future Resolve.AssignmentStmt extensions.
func (a *Analyzer) newAssignmentStmtAction(ctx context.Context, _ string, _ []driver.NamedValue, node *googlesql.ResolvedAssignmentStmt) (*AssignmentStmtAction, error) {
	target, _ := node.Target()
	sysVar, ok := target.(*googlesql.ResolvedSystemVariable)
	if !ok {
		return nil, fmt.Errorf("SET assignment target must be @@system_variable; got %T", target)
	}
	pathParts, _ := sysVar.NamePath()
	expr, _ := node.Expr()
	exprSQL, err := newNode(expr).FormatSQL(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to format assignment expression: %w", err)
	}
	return &AssignmentStmtAction{
		name:    strings.Join(pathParts, "."),
		exprSQL: exprSQL,
	}, nil
}

//nolint:unparam
func (a *Analyzer) newTruncateStmtAction(_ context.Context, _ string, _ []driver.NamedValue, node *googlesql.ResolvedTruncateStmt) (*TruncateStmtAction, error) {
	table, _ := m1(m1(node.TableScan()).Table()).Name()
	return &TruncateStmtAction{query: fmt.Sprintf("DELETE FROM `%s`", table)}, nil
}

func (a *Analyzer) newMergeStmtAction(ctx context.Context, _ string, args []driver.NamedValue, node *googlesql.ResolvedMergeStmt) (*MergeStmtAction, error) {
	targetTable, err := newNode(m1(node.TableScan())).FormatSQL(ctx)
	if err != nil {
		return nil, err
	}
	sourceTable, err := newNode(m1(node.FromScan())).FormatSQL(ctx)
	if err != nil {
		return nil, err
	}
	expr, err := newNode(m1(node.MergeExpr())).FormatSQL(ctx)
	if err != nil {
		return nil, err
	}
	fn, ok := m1(node.MergeExpr()).(*googlesql.ResolvedFunctionCall)
	if !ok {
		return nil, fmt.Errorf("currently MERGE expression is supported equal expression only")
	}
	if m1(m1(fn.Function()).FullName(false)) != "$equal" {
		return nil, fmt.Errorf("currently MERGE expression is supported equal expression only")
	}
	argList := m1(fn.ArgumentList())
	if len(argList) != 2 {
		return nil, fmt.Errorf("unexpected MERGE expression column num. expected 2 column but specified %d column", len(args))
	}
	colA, ok := argList[0].(*googlesql.ResolvedColumnRef)
	if !ok {
		return nil, fmt.Errorf("unexpected MERGE expression. expected column reference but got %T", argList[0])
	}
	colB, ok := argList[1].(*googlesql.ResolvedColumnRef)
	if !ok {
		return nil, fmt.Errorf("unexpected MERGE expression. expected column reference but got %T", argList[1])
	}
	var (
		sourceColumn *googlesql.ResolvedColumn
		targetColumn *googlesql.ResolvedColumn
	)
	if strings.Contains(sourceTable, m1(m1(colA.Column()).TableName())) {
		sourceColumn = m1(colA.Column())
		targetColumn = m1(colB.Column())
	} else {
		sourceColumn = m1(colB.Column())
		targetColumn = m1(colA.Column())
	}
	mergedTableSourceColumnName := fmt.Sprintf("`%s`", uniqueColumnName(ctx, sourceColumn))
	mergedTableTargetColumnName := fmt.Sprintf("`%s`", uniqueColumnName(ctx, targetColumn))
	mergedTableOutputColumns := []string{
		mergedTableTargetColumnName,
		mergedTableSourceColumnName,
	}
	var stmts []string
	stmts = append(stmts, fmt.Sprintf(
		"CREATE TABLE googlesqlite_merged_table AS SELECT DISTINCT * FROM (SELECT * FROM %[1]s LEFT JOIN %[2]s ON %[3]s UNION ALL SELECT * FROM %[2]s LEFT JOIN %[1]s ON %[3]s)",
		sourceTable, targetTable, expr,
	))

	// exists target table and source table
	matchedFromStmt := fmt.Sprintf(
		"FROM googlesqlite_merged_table WHERE %[2]s = %[1]s AND %[3]s = %[1]s",
		m1(targetColumn.Name()),
		mergedTableSourceColumnName,
		mergedTableTargetColumnName,
	)

	// exists target table but not exists source table
	notMatchedBySourceFromStmt := fmt.Sprintf(
		"FROM googlesqlite_merged_table WHERE %[2]s = `%[1]s` AND %[3]s IS NULL",
		m1(targetColumn.Name()),
		mergedTableTargetColumnName,
		mergedTableSourceColumnName,
	)

	// exists source table but not exists target table
	notMatchedByTargetFromStmt := fmt.Sprintf(
		"FROM googlesqlite_merged_table WHERE %[2]s = `%[1]s` AND %[3]s IS NULL",
		m1(sourceColumn.Name()),
		mergedTableSourceColumnName,
		mergedTableTargetColumnName,
	)
	for _, when := range m1(node.WhenClauseList()) {
		var fromStmt string
		switch m1(when.MatchType()) {
		case googlesql.ResolvedMergeWhenEnums_MatchTypeMatched:
			fromStmt = matchedFromStmt
		case googlesql.ResolvedMergeWhenEnums_MatchTypeNotMatchedBySource:
			fromStmt = notMatchedBySourceFromStmt
		case googlesql.ResolvedMergeWhenEnums_MatchTypeNotMatchedByTarget:
			fromStmt = notMatchedByTargetFromStmt
		}
		whereStmt := fmt.Sprintf(
			"WHERE EXISTS(SELECT %s %s)",
			strings.Join(mergedTableOutputColumns, ","),
			fromStmt,
		)
		switch m1(when.ActionType()) {
		case googlesql.ResolvedMergeWhenEnums_ActionTypeInsert:
			var columns []string
			for _, col := range m1(when.InsertColumnList()) {
				columns = append(columns, fmt.Sprintf("`%s`", m1(col.Name())))
			}
			row, err := newNode(m1(when.InsertRow())).FormatSQL(unuseColumnID(ctx))
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, fmt.Sprintf(
				"INSERT INTO `%[1]s`(%[2]s) SELECT %[3]s FROM (SELECT * FROM `%[4]s` %[5]s)",
				m1(targetColumn.TableName()),
				strings.Join(columns, ","),
				row,
				m1(sourceColumn.TableName()),
				whereStmt,
			))
		case googlesql.ResolvedMergeWhenEnums_ActionTypeUpdate:
			var items []string
			for _, item := range m1(when.UpdateItemList()) {
				sql, err := newNode(item).FormatSQL(ctx)
				if err != nil {
					return nil, err
				}
				items = append(items, sql)
			}
			stmts = append(stmts, fmt.Sprintf(
				"UPDATE `%s` SET %s %s",
				m1(targetColumn.TableName()),
				strings.Join(items, ","),
				fromStmt,
			))
		case googlesql.ResolvedMergeWhenEnums_ActionTypeDelete:
			stmts = append(stmts, fmt.Sprintf(
				"DELETE FROM `%s` %s",
				m1(targetColumn.TableName()),
				whereStmt,
			))
		}
	}
	stmts = append(stmts, "DROP TABLE googlesqlite_merged_table")
	return &MergeStmtAction{stmts: stmts}, nil
}

// getParamsFromNode returns the deduplicated `*ResolvedParameter`
// list collected as a side-effect of formatting the resolved tree.
// The wasm bridge in v0.1.0 does not expose a descendants walker on
// ResolvedNode, so the collector populated by ParameterNode.FormatSQL
// is the only source. Callers must install a paramCollector on ctx
// before invoking FormatSQL — newQueryStmtAction / newDMLStmtAction
// do this through collectFormatParams.
func getParamsFromNode(c *paramCollector) []*googlesql.ResolvedParameter {
	if c == nil {
		return nil
	}
	var (
		out  = make([]*googlesql.ResolvedParameter, 0, len(c.params))
		seen = map[string]struct{}{}
	)
	for _, p := range c.params {
		name, _ := p.Name()
		if name != "" {
			if _, dup := seen[name]; dup {
				continue
			}
			seen[name] = struct{}{}
		}
		out = append(out, p)
	}
	return out
}

// collectFormatParams runs FormatSQL with a fresh paramCollector
// attached to ctx and returns (formattedQuery, params). Splitting it
// out keeps newQueryStmtAction / newDMLStmtAction symmetric.
func collectFormatParams(ctx context.Context, n googlesql.ResolvedNode) (string, []*googlesql.ResolvedParameter, error) {
	c := &paramCollector{}
	ctx = withParamCollector(ctx, c)
	formatted, err := newNode(n).FormatSQL(ctx)
	if err != nil {
		return "", nil, err
	}
	// Surface the translated SQLite text to the CLI debug hook, if one
	// is installed on the context. This is the single chokepoint every
	// statement kind (query / DML / DDL / graph) routes through.
	if collector := sqlCollectorFromContext(ctx); collector != nil {
		collector.Add(formatted)
	}
	return formatted, getParamsFromNode(c), nil
}

func getArgsFromParams(values []driver.NamedValue, params []*googlesql.ResolvedParameter) ([]any, error) {
	if values == nil {
		return nil, nil
	}
	// The resolved-tree walker is a placeholder while ResolvedNode
	// child iteration isn't yet bridged, so params may be empty even
	// when the user supplied real parameters. In that case pass the
	// named values straight through and let SQLite handle native
	// @name / ? binding.
	if len(params) == 0 && len(values) > 0 {
		out := make([]any, 0, len(values))
		for _, v := range values {
			// Conn.CheckNamedValue has already encoded each value
			// into the canonical googlesqlite layout; we just need
			// to wrap named args so SQLite can match them against
			// the analyzer's lower-cased placeholder names.
			if v.Name != "" {
				out = append(out, sql.Named(strings.ToLower(v.Name), v.Value))
			} else {
				out = append(out, v.Value)
			}
		}
		return out, nil
	}
	argNum := len(params)
	if len(values) < argNum {
		return nil, fmt.Errorf("not enough query arguments")
	}
	namedValuesMap := map[string]driver.NamedValue{}
	for _, value := range values {
		// Name() value of *googlesql.ResolvedParameter always returns lowercase name.
		namedValuesMap[strings.ToLower(value.Name)] = value
	}
	var namedValues []driver.NamedValue
	for idx, param := range params {
		name, _ := param.Name()
		if name != "" {
			value, exists := namedValuesMap[name]
			if exists {
				namedValues = append(namedValues, value)
			} else {
				namedValues = append(namedValues, values[idx])
			}
		} else {
			namedValues = append(namedValues, values[idx])
		}
	}
	newNamedValues, err := encodeNamedValues(namedValues, params)
	if err != nil {
		return nil, err
	}
	args := make([]any, 0, argNum)
	for _, newNamedValue := range newNamedValues {
		args = append(args, newNamedValue)
	}
	return args, nil
}
