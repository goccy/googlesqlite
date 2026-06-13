package internal

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
	"time"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/goccy/go-json"
)

var (
	createCatalogTableQuery = `
CREATE TABLE IF NOT EXISTS googlesqlite_catalog(
  name STRING NOT NULL PRIMARY KEY,
  kind STRING NOT NULL,
  spec STRING NOT NULL,
  updatedAt TIMESTAMP NOT NULL,
  createdAt TIMESTAMP NOT NULL
)
`
	upsertCatalogQuery = `
INSERT INTO googlesqlite_catalog (
  name,
  kind,
  spec,
  updatedAt,
  createdAt
) VALUES (
  @name,
  @kind,
  @spec,
  @updatedAt,
  @createdAt
) ON CONFLICT(name) DO UPDATE SET
  spec = @spec,
  updatedAt = @updatedAt
`
	deleteCatalogQuery = `
DELETE FROM googlesqlite_catalog WHERE name = @name
`
)

type catalogSpecKind string

const (
	TableSpecKind    catalogSpecKind = "table"
	ViewSpecKind     catalogSpecKind = "view"
	FunctionSpecKind catalogSpecKind = "function"
	TVFSpecKind      catalogSpecKind = "tvf"
	catalogName                      = "googlesqlite"
)

type Catalog struct {
	db           *sql.DB
	lastSyncedAt time.Time
	mu           sync.Mutex
	tables       []*TableSpec
	functions    []*FunctionSpec
	tvfs         []*TVFSpec
	catalog      *googlesql.SimpleCatalog
	// descriptorPool is the protobuf DescriptorPool attached to the
	// SimpleCatalog at construction time. Consumers register FileDescriptorProto
	// payloads through Catalog.RegisterProto (Conn.RegisterProto on the
	// public surface). Once a Descriptor lives in the pool, callers can
	// promote a specific message into a SQL-visible type via
	// RegisterProtoMessage, which builds the *ProtoType through
	// TypeFactory.MakeProtoType and pins it on the catalog under its
	// fully-qualified name.
	descriptorPool   *googlesql.DescriptorPool
	registeredProtos map[string]*googlesql.ProtoType
	registeredEnums  map[string]*googlesql.EnumType
	tableMap         map[string]*TableSpec
	funcMap          map[string]*FunctionSpec
	tvfMap           map[string]*TVFSpec
	// infoSchemaRegistered tracks (project, dataset) pairs that have
	// already had their INFORMATION_SCHEMA views auto-registered, so
	// subsequent table additions in the same dataset don't re-add
	// duplicates.
	infoSchemaRegistered map[string]struct{}
	// infoSchemaTables maps an info-schema SimpleTable handle key
	// (returned by handleRawPtr) to its dataset filter, so the
	// formatter can detect a TableScan on an INFORMATION_SCHEMA
	// view and emit the right vtab call.
	infoSchemaTables map[string]*infoSchemaTableMeta
	// subCatalogs maps a parent SimpleCatalog (Go pointer) to its
	// name→sub-catalog table, so every AddCatalog call happens at
	// most once per (parent, name) — the wasm side traps if you try
	// to add the same name twice. The Go pointer is the right key
	// here: each call to getOrCreateSubCatalog hands back a stable
	// cached *SimpleCatalog so re-encountering the same parent
	// always lands on the same map slot. SimpleCatalog.FullName()
	// is NOT usable because it returns only the local name and
	// collides across hierarchy levels (e.g. two sub-catalogs both
	// locally named "INFORMATION_SCHEMA").
	subCatalogs map[*googlesql.SimpleCatalog]map[string]*googlesql.SimpleCatalog
	// propertyGraphs records property-graph definitions registered
	// through CREATE PROPERTY GRAPH. Currently a metadata-only
	// store; full GRAPH_EXPAND / AGG integration is not yet wired.
	propertyGraphs map[string]*propertyGraphSpec
}

// PropertyGraphSpec captures the essentials of a CREATE PROPERTY
// GRAPH statement: the graph's name and the underlying node /
// edge table aliases. We don't currently surface MEASURE
// properties or label hierarchies through this struct — they
// are read directly off the resolved AST when needed.
type propertyGraphSpec struct {
	Name       string
	NodeTables []string
	EdgeTables []string
}

func newSimpleCatalog(name string) *googlesql.SimpleCatalog {
	catalog, err := googlesql.NewSimpleCatalog(name, tf())
	if err != nil {
		return nil
	}
	// Register all GoogleSQL builtins with every language feature
	// enabled, so conditional families (Anonymization / Differential
	// Privacy, PropertyGraph, SqlGraph, Measures, etc.) are available
	// for callers that turn the matching analyzer feature on. Having
	// extras in the catalog is harmless — the analyzer's own
	// LanguageOptions gates whether each call site can use them.
	if opts, err := googlesql.NewLanguageOptionsMaximumFeatures(); err == nil && opts != nil {
		// Explicitly enable the Anonymization / Differential
		// Privacy language features. NewLanguageOptionsMaximumFeatures
		// does NOT include them by default (DP/Anon are gated as
		// "in development"), so without these the catalog wouldn't
		// even contain the DP signature ids we ask for via
		// IncludeFunctionIds.
		for _, f := range []googlesql.LanguageFeature{
			googlesql.LanguageFeatureFeatureDifferentialPrivacy,
			googlesql.LanguageFeatureFeatureDifferentialPrivacyThresholding,
			googlesql.LanguageFeatureFeatureDifferentialPrivacyReportFunctions,
			googlesql.LanguageFeatureFeatureAnonymization,
			googlesql.LanguageFeatureFeatureAnonymizationThresholding,
			// AEAD / KEYS encryption family is gated behind the
			// Encryption feature flag — not on by default in
			// NewLanguageOptionsMaximumFeatures.
			googlesql.LanguageFeatureFeatureEncryption,
			// GoogleSQL proto-reflection feature flags. Enabling
			// these on the catalog-construction options ensures the
			// proto family is included in the builtin function set
			// when Conn.RegisterProto plugs a DescriptorPool into
			// the catalog.
			googlesql.LanguageFeatureFeatureProtoBase,
			googlesql.LanguageFeatureFeatureProtoDefaultIfNull,
			googlesql.LanguageFeatureFeatureExtractFromProto,
			googlesql.LanguageFeatureFeatureReplaceFields,
			googlesql.LanguageFeatureFeatureFilterFields,
			googlesql.LanguageFeatureFeatureProtoMaps,
			googlesql.LanguageFeatureFeatureProtoExtensionsWithNew,
			googlesql.LanguageFeatureFeatureProtoExtensionsWithSet,
			googlesql.LanguageFeatureFeatureEnumValueDescriptorProto,
		} {
			_ = opts.EnableLanguageFeature(f)
		}
		// Combine the public-builtin registration and the DP/Anon
		// IncludeFunctionIds into a single AddGoogleSQLFunctions2
		// call. Calling AddGoogleSQLFunctions3 first and then
		// AddGoogleSQLFunctions2 results in a duplicate-registration
		// crash inside the wasm catalog.
		bf := newDPBuiltinFunctionOptions(opts)
		if bf == nil {
			bf = &googlesql.BuiltinFunctionOptions{LanguageOptions: opts}
		}
		if err := catalog.AddGoogleSQLFunctions2(bf); err != nil {
			_ = catalog.AddGoogleSQLFunctions()
		}
	} else {
		_ = catalog.AddGoogleSQLFunctions()
	}
	registerBigQueryExtensionFunctions(catalog)
	registerSpannerExtensionFunctions(catalog)
	return catalog
}

// newDPBuiltinFunctionOptions builds a BuiltinFunctionOptions
// instance whose IncludeFunctionIds covers every Anonymization /
// Differential Privacy internal builtin. AddGoogleSQLFunctions2
// reads the protobuf-encoded varint list as the set of extra
// signature IDs to register — without this, the rewriter can
// emit `$differential_privacy_sum` references the catalog
// cannot resolve.
func newDPBuiltinFunctionOptions(opts *googlesql.LanguageOptions) *googlesql.BuiltinFunctionOptions {
	ids := []googlesql.FunctionSignatureId{
		// Differential Privacy.
		googlesql.FunctionSignatureIdFnDifferentialPrivacyCount,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyCountStar,
		googlesql.FunctionSignatureIdFnDifferentialPrivacySumInt64,
		googlesql.FunctionSignatureIdFnDifferentialPrivacySumUint64,
		googlesql.FunctionSignatureIdFnDifferentialPrivacySumDouble,
		googlesql.FunctionSignatureIdFnDifferentialPrivacySumNumeric,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyAvgDouble,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyAvgNumeric,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyVarPopDouble,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyVarPopDoubleArray,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyStddevPopDouble,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyStddevPopDoubleArray,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyQuantilesDouble,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyQuantilesDoubleArray,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyPercentileContDouble,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyPercentileContDoubleArray,
		googlesql.FunctionSignatureIdFnDifferentialPrivacyApproxCountDistinct,
		// Anonymization (legacy ANON_*).
		googlesql.FunctionSignatureIdFnAnonCount,
		googlesql.FunctionSignatureIdFnAnonCountStar,
		googlesql.FunctionSignatureIdFnAnonSumDouble,
		googlesql.FunctionSignatureIdFnAnonSumInt64,
		googlesql.FunctionSignatureIdFnAnonSumNumeric,
		googlesql.FunctionSignatureIdFnAnonSumUint64,
		googlesql.FunctionSignatureIdFnAnonAvgDouble,
		googlesql.FunctionSignatureIdFnAnonAvgNumeric,
		googlesql.FunctionSignatureIdFnAnonVarPopDouble,
		googlesql.FunctionSignatureIdFnAnonVarPopDoubleArray,
		googlesql.FunctionSignatureIdFnAnonStddevPopDouble,
		googlesql.FunctionSignatureIdFnAnonStddevPopDoubleArray,
		googlesql.FunctionSignatureIdFnAnonQuantilesDouble,
		googlesql.FunctionSignatureIdFnAnonQuantilesDoubleArray,
		googlesql.FunctionSignatureIdFnAnonPercentileContDouble,
		googlesql.FunctionSignatureIdFnAnonPercentileContDoubleArray,
		// Longtail GoogleSQL builtins gated as "in development"
		// upstream: their signature IDs exist in builtin_function.proto
		// but NewLanguageOptionsMaximumFeatures does not enable them.
		// We opt in here so the analyzer can resolve the call sites;
		// the runtime UDFs live in internal/functions/longtail and
		// dispatch via the lowercased name.
		googlesql.FunctionSignatureIdFnIferror,
		googlesql.FunctionSignatureIdFnNulliferror,
		googlesql.FunctionSignatureIdFnIserror,
		googlesql.FunctionSignatureIdFnRegexpMatchString,
		googlesql.FunctionSignatureIdFnRegexpMatchBytes,
		googlesql.FunctionSignatureIdFnCollate,
		googlesql.FunctionSignatureIdFnSplitSubstr,
		googlesql.FunctionSignatureIdFnRegexpExtractGroupsString,
		googlesql.FunctionSignatureIdFnRegexpExtractGroupsBytes,
		googlesql.FunctionSignatureIdFnGrouping,
		googlesql.FunctionSignatureIdFnSafeToJson,
		googlesql.FunctionSignatureIdFnFlatten,
		googlesql.FunctionSignatureIdFnArrayZipTwoArray,
		googlesql.FunctionSignatureIdFnArrayZipTwoArrayLambda,
		googlesql.FunctionSignatureIdFnArrayZipThreeArray,
		googlesql.FunctionSignatureIdFnArrayZipThreeArrayLambda,
		googlesql.FunctionSignatureIdFnArrayZipFourArray,
		googlesql.FunctionSignatureIdFnArrayZipFourArrayLambda,
		googlesql.FunctionSignatureIdFnJsonArrayInsert,
		googlesql.FunctionSignatureIdFnJsonArrayAppend,
		googlesql.FunctionSignatureIdFnHllCountMerge,
		googlesql.FunctionSignatureIdFnHllCountMergePartial,
		googlesql.FunctionSignatureIdFnNetFormatIp,
		googlesql.FunctionSignatureIdFnNetParseIp,
		googlesql.FunctionSignatureIdFnNetFormatPackedIp,
		googlesql.FunctionSignatureIdFnNetParsePackedIp,
		googlesql.FunctionSignatureIdFnNetIpInNet,
		googlesql.FunctionSignatureIdFnNetMakeNet,
		googlesql.FunctionSignatureIdFnTimestampFromUnixSecondsInt64,
		googlesql.FunctionSignatureIdFnTimestampFromUnixSecondsTimestamp,
		googlesql.FunctionSignatureIdFnTimestampFromUnixMillisInt64,
		// DOT_PRODUCT / APPROX_DOT_PRODUCT / APPROX_EUCLIDEAN_DISTANCE
		// / APPROX_COSINE_DISTANCE. Used by Spanner; spec lives under
		// docs/specs/spanner/functions/math/.
		googlesql.FunctionSignatureIdFnDotProductInt64,
		googlesql.FunctionSignatureIdFnDotProductFloat,
		googlesql.FunctionSignatureIdFnDotProductDouble,
		googlesql.FunctionSignatureIdFnApproxDotProductInt64,
		googlesql.FunctionSignatureIdFnApproxDotProductFloat,
		googlesql.FunctionSignatureIdFnApproxEuclideanDistanceDouble,
		googlesql.FunctionSignatureIdFnApproxEuclideanDistanceFloat,
		googlesql.FunctionSignatureIdFnApproxCosineDistanceDouble,
		googlesql.FunctionSignatureIdFnApproxCosineDistanceFloat,
		// ZSTD_COMPRESS / DECOMPRESS_TO_*. Spanner string_functions.
		googlesql.FunctionSignatureIdFnZstdCompressFromBytes,
		googlesql.FunctionSignatureIdFnZstdCompressFromString,
		googlesql.FunctionSignatureIdFnZstdDecompressToBytes,
		googlesql.FunctionSignatureIdFnZstdDecompressToString,
		// RANGE predicates and ops. The Maximum features enables the
		// RANGE type but not all predicates.
		googlesql.FunctionSignatureIdFnRangeOverlaps,
		googlesql.FunctionSignatureIdFnRangeIntersect,
		googlesql.FunctionSignatureIdFnRangeContainsRange,
		googlesql.FunctionSignatureIdFnRangeContainsElement,
		// ST_UNION_AGG geography aggregate.
		googlesql.FunctionSignatureIdFnStUnionAgg,
		// BigQuery AEAD encryption family. Pure-Go AES-GCM bodies
		// live in internal/functions/aead.
		googlesql.FunctionSignatureIdFnKeysNewKeyset,
		googlesql.FunctionSignatureIdFnKeysAddKeyFromRawBytes,
		googlesql.FunctionSignatureIdFnKeysRotateKeyset,
		googlesql.FunctionSignatureIdFnKeysKeysetLength,
		googlesql.FunctionSignatureIdFnKeysKeysetToJson,
		googlesql.FunctionSignatureIdFnKeysKeysetFromJson,
		googlesql.FunctionSignatureIdFnAeadEncryptString,
		googlesql.FunctionSignatureIdFnAeadEncryptBytes,
		googlesql.FunctionSignatureIdFnAeadDecryptString,
		googlesql.FunctionSignatureIdFnAeadDecryptBytes,
		googlesql.FunctionSignatureIdFnKeysKeysetChainStringBytesBytes,
		googlesql.FunctionSignatureIdFnKeysKeysetChainStringBytes,
		googlesql.FunctionSignatureIdFnDeterministicEncryptString,
		googlesql.FunctionSignatureIdFnDeterministicEncryptBytes,
		googlesql.FunctionSignatureIdFnDeterministicDecryptString,
		googlesql.FunctionSignatureIdFnDeterministicDecryptBytes,
	}
	bf := &googlesql.BuiltinFunctionOptions{LanguageOptions: opts}
	for _, id := range ids {
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(buf, uint64(id))
		bf.IncludeFunctionIds = append(bf.IncludeFunctionIds, buf[:n])
	}
	return bf
}

// newSimpleArgType builds a *googlesql.FunctionArgumentType wrapping a
// simple type of the given kind. It returns nil on any construction
// error so callers can fail soft. The factory is passed explicitly
// because the helper is shared by several extension-registration
// functions that each obtain their own TypeFactory via tf().
func newSimpleArgType(factory *googlesql.TypeFactory, kind googlesql.TypeKind) *googlesql.FunctionArgumentType {
	t, err := factory.MakeSimpleType(kind)
	if err != nil || t == nil {
		return nil
	}
	opt, err := googlesql.NewFunctionArgumentTypeOptions()
	if err != nil || opt == nil {
		return nil
	}
	arg, err := googlesql.NewFunctionArgumentType(t, opt, -1)
	if err != nil {
		return nil
	}
	return arg
}

// newAnyArgType builds a *googlesql.FunctionArgumentType for a templated
// signature-argument kind (ArgTypeAny1, ArgArrayTypeAny1, ...). It
// returns nil on any construction error so callers can fail soft.
func newAnyArgType(kind googlesql.SignatureArgumentKind) *googlesql.FunctionArgumentType {
	opt, err := googlesql.NewFunctionArgumentTypeOptions()
	if err != nil || opt == nil {
		return nil
	}
	arg, err := googlesql.NewFunctionArgumentType5(kind, opt, -1)
	if err != nil {
		return nil
	}
	return arg
}

// registerSpannerExtensionFunctions exposes the Spanner-specific
// function surface as a `mysql.*` sub-catalog of aliases that
// route to GoogleSQL builtins. The Spanner dialect's MySQL
// compatibility namespace is mostly a renaming of standard
// functions, so we add one Function per supported alias whose
// signature mirrors the target GoogleSQL function and whose
// runtime simply delegates to it.
//
// The aliases are registered unconditionally on every catalog
// (regardless of dialect) because the analyzer side has no
// per-dialect catalog switch — keeping the aliases always-on is
// harmless (callers who don't use `mysql.*` never see them) and
// removes the need for dialect plumbing at the Conn level.
func registerSpannerExtensionFunctions(catalog *googlesql.SimpleCatalog) {
	factory := tf()
	// Build a `mysql` sub-catalog the analyzer will descend into
	// for `mysql.<func>` references. AddCatalog must run exactly
	// once per parent catalog name; subsequent registrations
	// reuse the same handle.
	mysqlCat, err := googlesql.NewSimpleCatalog("mysql", factory)
	if err != nil || mysqlCat == nil {
		return
	}
	if err := catalog.AddCatalog(mysqlCat); err != nil {
		return
	}
	// The Spanner extension surface is split into focused per-family
	// sub-functions. They are invoked in the SAME ORDER as the
	// original flat registration sequence; catalog registration is
	// order-sensitive, so the order must not change.
	registerSpannerMysqlAliasFuncs(mysqlCat, factory)
	registerSpannerSearchFuncs(catalog, factory)
	registerSpannerStandaloneScalarFuncs(catalog, mysqlCat, factory)
	registerSpannerDLPFuncs(catalog, factory)
	registerSpannerNamespaceFuncs(catalog, factory)
}

// registerSpannerMysqlAliasFuncs registers the Spanner `mysql.*`
// compatibility aliases (string / numeric / datetime / utility /
// json / encryption families) into the supplied `mysql` sub-catalog.
func registerSpannerMysqlAliasFuncs(mysqlCat *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	mkArg := func(kind googlesql.TypeKind) *googlesql.FunctionArgumentType {
		return newSimpleArgType(factory, kind)
	}
	mkAny := newAnyArgType
	register := func(name string, mode googlesql.FunctionEnums_Mode, retArg *googlesql.FunctionArgumentType, argTypes []*googlesql.FunctionArgumentType) {
		if retArg == nil {
			return
		}
		for _, a := range argTypes {
			if a == nil {
				return
			}
		}
		sig, err := googlesql.NewFunctionSignature3(retArg, argTypes, 0)
		if err != nil || sig == nil {
			return
		}
		// Storage name is the dotted form so the formatter routes
		// the call through `mysql_<name>` at SQLite UDF dispatch
		// time (formatter.go flattens "." to "_" on the funcName).
		fn, err := googlesql.NewFunction([]string{"mysql_" + name}, "", mode, []*googlesql.FunctionSignature{sig}, nil)
		if err != nil || fn == nil {
			return
		}
		_ = mysqlCat.AddFunction2(name, fn)
	}

	// mysql_datetime: aliases for date / time helpers. Each
	// signature mirrors the GoogleSQL counterpart; the runtime
	// dispatch table maps "mysql_<name>" to the same UDF as the
	// bare name.
	register("curdate", googlesql.FunctionEnums_ModeScalar, mkArg(googlesql.TypeKindTypeDate), nil)
	register("current_date", googlesql.FunctionEnums_ModeScalar, mkArg(googlesql.TypeKindTypeDate), nil)
	register("now", googlesql.FunctionEnums_ModeScalar, mkArg(googlesql.TypeKindTypeTimestamp), nil)
	register("current_timestamp", googlesql.FunctionEnums_ModeScalar, mkArg(googlesql.TypeKindTypeTimestamp), nil)
	register("unix_timestamp", googlesql.FunctionEnums_ModeScalar, mkArg(googlesql.TypeKindTypeInt64), nil)

	// mysql_string: classic MySQL string builtins. Most map 1:1
	// to GoogleSQL function names; we register only the alias.
	register("length", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("char_length", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("lower", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("upper", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("trim", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("ltrim", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("rtrim", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("concat", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})
	register("reverse", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("md5", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("sha1", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})

	// mysql_numeric: integer / float helpers.
	register("abs", googlesql.FunctionEnums_ModeScalar,
		mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
		[]*googlesql.FunctionArgumentType{mkAny(googlesql.SignatureArgumentKindArgTypeAny1)})
	register("floor", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})
	register("ceil", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})
	register("ceiling", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})
	register("round", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})

	// mysql_utility: misc helpers.
	register("coalesce", googlesql.FunctionEnums_ModeScalar,
		mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
		[]*googlesql.FunctionArgumentType{
			mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
			mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
		})

	// mysql_datetime extras — EXTRACT-equivalent helpers. Each
	// alias accepts DATE, DATETIME, or TIMESTAMP; the runtime
	// (internal/functions/spanner/mysql_helpers.go) normalises
	// every shape through ToTime().
	registerMulti := func(name string, argKinds []googlesql.TypeKind) {
		var sigs []*googlesql.FunctionSignature
		for _, k := range argKinds {
			arg := mkArg(k)
			ret := mkArg(googlesql.TypeKindTypeInt64)
			if arg == nil || ret == nil {
				return
			}
			s, err := googlesql.NewFunctionSignature3(ret, []*googlesql.FunctionArgumentType{arg}, 0)
			if err != nil || s == nil {
				return
			}
			sigs = append(sigs, s)
		}
		fn, err := googlesql.NewFunction([]string{"mysql_" + name}, "", googlesql.FunctionEnums_ModeScalar, sigs, nil)
		if err != nil || fn == nil {
			return
		}
		_ = mysqlCat.AddFunction2(name, fn)
	}
	for _, fn := range []string{"day", "month", "year", "hour", "minute", "second", "microsecond", "quarter", "dayofweek", "dayofyear", "dayofmonth", "week", "weekofyear", "weekday"} {
		registerMulti(fn, []googlesql.TypeKind{
			googlesql.TypeKindTypeDate,
			googlesql.TypeKindTypeDatetime,
			googlesql.TypeKindTypeTimestamp,
		})
	}

	// mysql_string extras.
	register("bit_length", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("hex", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes)})
	register("space", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64)})
	register("position", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})
	// More mysql_string aliases. char + concat_ws are variadic;
	// register a small fixed-arity ladder of signatures (1..6 args)
	// so common call patterns parse without us reaching for
	// FunctionArgumentTypeOptions.SetCardinality. 6 is more than
	// enough for typical "CHAR-encode 1..3 code points" /
	// "CONCAT_WS join 2..5 fields" usage.
	registerArityLadder := func(name string, argKind googlesql.TypeKind, retKind googlesql.TypeKind, arities []int) {
		var sigs []*googlesql.FunctionSignature
		for _, arity := range arities {
			argsList := make([]*googlesql.FunctionArgumentType, arity)
			ok := true
			for i := range argsList {
				argsList[i] = mkArg(argKind)
				if argsList[i] == nil {
					ok = false
				}
			}
			if !ok {
				continue
			}
			ret := mkArg(retKind)
			if ret == nil {
				continue
			}
			s, err := googlesql.NewFunctionSignature3(ret, argsList, 0)
			if err == nil && s != nil {
				sigs = append(sigs, s)
			}
		}
		if len(sigs) == 0 {
			return
		}
		fn, err := googlesql.NewFunction([]string{"mysql_" + name}, "", googlesql.FunctionEnums_ModeScalar, sigs, nil)
		if err == nil && fn != nil {
			_ = mysqlCat.AddFunction2(name, fn)
		}
	}
	registerArityLadder("char", googlesql.TypeKindTypeInt64, googlesql.TypeKindTypeString, []int{1, 2, 3, 4, 5, 6})
	registerArityLadder("concat_ws", googlesql.TypeKindTypeString, googlesql.TypeKindTypeString, []int{2, 3, 4, 5, 6})
	register("insert", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeInt64), mkArg(googlesql.TypeKindTypeInt64), mkArg(googlesql.TypeKindTypeString)})
	register("locate", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})
	register("mid", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeInt64), mkArg(googlesql.TypeKindTypeInt64)})
	register("oct", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64)})
	register("ord", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("quote", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("regexp_like", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})
	register("strcmp", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})
	register("substring_index", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeInt64)})
	register("unhex", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})

	// More mysql_datetime aliases — STRING return.
	for _, fn := range []string{"dayname", "monthname"} {
		var sigs []*googlesql.FunctionSignature
		for _, k := range []googlesql.TypeKind{googlesql.TypeKindTypeDate, googlesql.TypeKindTypeDatetime, googlesql.TypeKindTypeTimestamp} {
			arg := mkArg(k)
			ret := mkArg(googlesql.TypeKindTypeString)
			if arg == nil || ret == nil {
				continue
			}
			s, err := googlesql.NewFunctionSignature3(ret, []*googlesql.FunctionArgumentType{arg}, 0)
			if err == nil && s != nil {
				sigs = append(sigs, s)
			}
		}
		if len(sigs) == 0 {
			continue
		}
		fnObj, err := googlesql.NewFunction([]string{"mysql_" + fn}, "", googlesql.FunctionEnums_ModeScalar, sigs, nil)
		if err == nil && fnObj != nil {
			_ = mysqlCat.AddFunction2(fn, fnObj)
		}
	}
	register("from_days", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDate),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64)})
	// to_days / to_seconds accept DATE / DATETIME / TIMESTAMP.
	registerMulti("to_days", []googlesql.TypeKind{
		googlesql.TypeKindTypeDate,
		googlesql.TypeKindTypeDatetime,
		googlesql.TypeKindTypeTimestamp,
	})
	registerMulti("to_seconds", []googlesql.TypeKind{
		googlesql.TypeKindTypeDate,
		googlesql.TypeKindTypeDatetime,
		googlesql.TypeKindTypeTimestamp,
	})
	register("makedate", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDate),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64), mkArg(googlesql.TypeKindTypeInt64)})
	register("from_unixtime", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeTimestamp),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64)})
	register("sysdate", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeTimestamp), nil)
	register("utc_date", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDate), nil)
	register("utc_timestamp", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeTimestamp), nil)
	register("period_add", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64), mkArg(googlesql.TypeKindTypeInt64)})
	register("period_diff", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64), mkArg(googlesql.TypeKindTypeInt64)})
	// date_format / str_to_date accept DATE / DATETIME / TIMESTAMP
	// for the first argument; second is STRING. Build sig set
	// manually so each variant carries the right pair.
	for _, fn := range []string{"date_format"} {
		var sigs []*googlesql.FunctionSignature
		for _, k := range []googlesql.TypeKind{googlesql.TypeKindTypeDate, googlesql.TypeKindTypeDatetime, googlesql.TypeKindTypeTimestamp} {
			ret := mkArg(googlesql.TypeKindTypeString)
			a := mkArg(k)
			b := mkArg(googlesql.TypeKindTypeString)
			if ret == nil || a == nil || b == nil {
				continue
			}
			s, err := googlesql.NewFunctionSignature3(ret, []*googlesql.FunctionArgumentType{a, b}, 0)
			if err == nil && s != nil {
				sigs = append(sigs, s)
			}
		}
		if len(sigs) > 0 {
			fnObj, err := googlesql.NewFunction([]string{"mysql_" + fn}, "", googlesql.FunctionEnums_ModeScalar, sigs, nil)
			if err == nil && fnObj != nil {
				_ = mysqlCat.AddFunction2(fn, fnObj)
			}
		}
	}
	register("str_to_date", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeTimestamp),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})

	// mysql_timestamp aliases. datediff accepts DATE/DATETIME/TIMESTAMP.
	for _, fn := range []string{"datediff"} {
		var sigs []*googlesql.FunctionSignature
		for _, k := range []googlesql.TypeKind{googlesql.TypeKindTypeDate, googlesql.TypeKindTypeDatetime, googlesql.TypeKindTypeTimestamp} {
			ret := mkArg(googlesql.TypeKindTypeInt64)
			a := mkArg(k)
			b := mkArg(k)
			if ret == nil || a == nil || b == nil {
				continue
			}
			s, err := googlesql.NewFunctionSignature3(ret, []*googlesql.FunctionArgumentType{a, b}, 0)
			if err == nil && s != nil {
				sigs = append(sigs, s)
			}
		}
		if len(sigs) > 0 {
			fnObj, err := googlesql.NewFunction([]string{"mysql_" + fn}, "", googlesql.FunctionEnums_ModeScalar, sigs, nil)
			if err == nil && fnObj != nil {
				_ = mysqlCat.AddFunction2(fn, fnObj)
			}
		}
	}
	register("localtime", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeTimestamp), nil)
	register("localtimestamp", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeTimestamp), nil)

	// mysql_utility (network + uuid + ip predicates).
	register("inet_aton", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("inet_ntoa", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64)})
	register("inet6_aton", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("inet6_ntoa", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes)})
	register("is_ipv4", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("is_ipv6", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("is_ipv4_compat", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes)})
	register("is_ipv4_mapped", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes)})
	register("is_uuid", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("bin_to_uuid", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes)})
	register("uuid_to_bin", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})

	// mysql_encryption.
	register("sha2", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeInt64)})

	// mysql_json.
	register("json_quote", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("json_unquote", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})

	// mysql_numeric extras.
	register("degrees", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})
	register("radians", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})
	register("log2", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})
	register("truncate", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble), mkArg(googlesql.TypeKindTypeInt64)})
}

// registerSpannerSearchFuncs registers the Spanner search /
// tokenization functions into the root catalog. These live in the
// root catalog (not mysql.*), with TOKENLIST exposed as BYTES at the
// SQL surface. Each runtime delegate lives in
// internal/functions/spanner/search.go.
func registerSpannerSearchFuncs(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	mkArg := func(kind googlesql.TypeKind) *googlesql.FunctionArgumentType {
		return newSimpleArgType(factory, kind)
	}
	registerRoot := func(name string, mode googlesql.FunctionEnums_Mode, ret *googlesql.FunctionArgumentType, argTypes []*googlesql.FunctionArgumentType) {
		if ret == nil {
			return
		}
		for _, a := range argTypes {
			if a == nil {
				return
			}
		}
		sig, err := googlesql.NewFunctionSignature3(ret, argTypes, 0)
		if err != nil || sig == nil {
			return
		}
		fn, err := googlesql.NewFunction([]string{name}, "", mode, []*googlesql.FunctionSignature{sig}, nil)
		if err != nil || fn == nil {
			return
		}
		_ = catalog.AddFunction2(name, fn)
	}
	// COLLATE(value, collate_specification) — the function-call form of
	// the COLLATE clause. Upstream googlesql exposes it as docs-only
	// syntactic sugar (no FN_COLLATE function ID). We surface it as a
	// scalar so the analyzer can resolve the call sites used by the
	// upstream string_functions Examples.
	registerRoot("collate", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})
	registerRoot("tokenize_fulltext", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	registerRoot("tokenize_substring", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	registerRoot("tokenize_ngrams", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeInt64)})
	registerRoot("tokenize_number", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDouble)})
	registerRoot("tokenize_bool", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBool)})
	registerRoot("tokenize_json", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	registerRoot("token", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	registerRoot("tokenlist_concat", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes), mkArg(googlesql.TypeKindTypeBytes)})
	// `search` already exists as a BigQuery function with signature
	// (data, query [, analyzer, ...]); the Spanner overload takes
	// (TOKENLIST, STRING) but TOKENLIST coerces from BYTES so the
	// existing signature accepts both call patterns.
	registerRoot("search_substring", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes), mkArg(googlesql.TypeKindTypeString)})
	registerRoot("search_ngrams", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBool),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes), mkArg(googlesql.TypeKindTypeString)})
	registerRoot("score", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes), mkArg(googlesql.TypeKindTypeString)})
	registerRoot("score_ngrams", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDouble),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes), mkArg(googlesql.TypeKindTypeString)})
	registerRoot("snippet", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)})
	registerRoot("debug_tokenlist", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeBytes)})
}

// registerSpannerStandaloneScalarFuncs registers the Spanner-only
// standalone root-catalog scalars (sequence / fingerprint helpers)
// and the remaining `mysql.*` aliases (lcase / ucase / adddate /
// subdate) that the upstream analyzer does not ship as builtin
// signature IDs.
func registerSpannerStandaloneScalarFuncs(catalog *googlesql.SimpleCatalog, mysqlCat *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	mkArg := func(kind googlesql.TypeKind) *googlesql.FunctionArgumentType {
		return newSimpleArgType(factory, kind)
	}
	registerRoot := func(name string, mode googlesql.FunctionEnums_Mode, ret *googlesql.FunctionArgumentType, argTypes []*googlesql.FunctionArgumentType) {
		if ret == nil {
			return
		}
		for _, a := range argTypes {
			if a == nil {
				return
			}
		}
		sig, err := googlesql.NewFunctionSignature3(ret, argTypes, 0)
		if err != nil || sig == nil {
			return
		}
		fn, err := googlesql.NewFunction([]string{name}, "", mode, []*googlesql.FunctionSignature{sig}, nil)
		if err != nil || fn == nil {
			return
		}
		_ = catalog.AddFunction2(name, fn)
	}
	register := func(name string, mode googlesql.FunctionEnums_Mode, retArg *googlesql.FunctionArgumentType, argTypes []*googlesql.FunctionArgumentType) {
		if retArg == nil {
			return
		}
		for _, a := range argTypes {
			if a == nil {
				return
			}
		}
		sig, err := googlesql.NewFunctionSignature3(retArg, argTypes, 0)
		if err != nil || sig == nil {
			return
		}
		fn, err := googlesql.NewFunction([]string{"mysql_" + name}, "", mode, []*googlesql.FunctionSignature{sig}, nil)
		if err != nil || fn == nil {
			return
		}
		_ = mysqlCat.AddFunction2(name, fn)
	}
	// Spanner-only standalone scalars that the upstream analyzer
	// does not ship as builtin signature IDs. We register them
	// explicitly so the analyzer can resolve the call sites;
	// runtime UDFs are bound in function_register_normal.go under
	// the Spanner section.
	//
	// Functions that DO have upstream builtin IDs (dot_product,
	// approx_dot_product, approx_euclidean_distance,
	// approx_cosine_distance, zstd_compress / decompress_to_*,
	// pending_commit_timestamp) are opted in via
	// newDPBuiltinFunctionOptions's IncludeFunctionIds list and
	// resolved by the analyzer through the GoogleSQL builtin path.
	registerRoot("pending_commit_timestamp", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeTimestamp), nil)
	registerRoot("bit_reverse", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeInt64), mkArg(googlesql.TypeKindTypeBool)})
	registerRoot("highway_fingerprint128", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	registerRoot("get_next_sequence_value", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	registerRoot("get_internal_sequence_state", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeInt64),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	// NOTE: Spanner JSON FLOAT32 / FLOAT32_ARRAY / FLOAT64_ARRAY
	// helpers are not registered here. FLOAT32 collides with the
	// reserved type-name keyword and triggers a wasm-side crash; the
	// runtime UDFs remain available for callers that route through
	// an alternative dispatch path.

	// Spanner mysql.* aliases: lcase / ucase / adddate / subdate.
	register("lcase", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("ucase", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	register("adddate", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDate),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDate), mkArg(googlesql.TypeKindTypeInt64)})
	register("subdate", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeDate),
		[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeDate), mkArg(googlesql.TypeKindTypeInt64)})
}

// registerSpannerDLPFuncs registers the BigQuery DLP family
// (deterministic encrypt/decrypt with DLP_KEY_CHAIN wrapping) into
// the root catalog. Runtime bodies are in internal/functions/extras.
func registerSpannerDLPFuncs(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	mkArg := func(kind googlesql.TypeKind) *googlesql.FunctionArgumentType {
		return newSimpleArgType(factory, kind)
	}
	registerRoot := func(name string, mode googlesql.FunctionEnums_Mode, ret *googlesql.FunctionArgumentType, argTypes []*googlesql.FunctionArgumentType) {
		if ret == nil {
			return
		}
		for _, a := range argTypes {
			if a == nil {
				return
			}
		}
		sig, err := googlesql.NewFunctionSignature3(ret, argTypes, 0)
		if err != nil || sig == nil {
			return
		}
		fn, err := googlesql.NewFunction([]string{name}, "", mode, []*googlesql.FunctionSignature{sig}, nil)
		if err != nil || fn == nil {
			return
		}
		_ = catalog.AddFunction2(name, fn)
	}
	// Remaining longtail extensions that do not have upstream
	// FunctionSignatureIds: generate_range_array and range_sessionize.
	// The latter is normally a TVF; we accept its scalar form.
	rangeArg := func(elem googlesql.TypeKind) *googlesql.FunctionArgumentType {
		t, err := factory.MakeSimpleType(elem)
		if err != nil || t == nil {
			return nil
		}
		rt, err := factory.MakeRangeType2(t)
		if err != nil || rt == nil {
			return nil
		}
		opt, err := googlesql.NewFunctionArgumentTypeOptions()
		if err != nil || opt == nil {
			return nil
		}
		a, err := googlesql.NewFunctionArgumentType(rt, opt, -1)
		if err != nil {
			return nil
		}
		return a
	}
	_ = rangeArg

	// BigQuery DLP family (deterministic encrypt/decrypt with
	// DLP_KEY_CHAIN wrapping). Runtime bodies are in
	// internal/functions/extras.
	registerRoot("dlp_key_chain", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeBytes),
		[]*googlesql.FunctionArgumentType{
			mkArg(googlesql.TypeKindTypeString),
			mkArg(googlesql.TypeKindTypeBytes),
		})
	registerRoot("dlp_deterministic_encrypt", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{
			mkArg(googlesql.TypeKindTypeBytes),
			mkArg(googlesql.TypeKindTypeString),
			mkArg(googlesql.TypeKindTypeString),
		})
	registerRoot("dlp_deterministic_decrypt", googlesql.FunctionEnums_ModeScalar,
		mkArg(googlesql.TypeKindTypeString),
		[]*googlesql.FunctionArgumentType{
			mkArg(googlesql.TypeKindTypeBytes),
			mkArg(googlesql.TypeKindTypeString),
			mkArg(googlesql.TypeKindTypeString),
		})
}

// registerSpannerNamespaceFuncs registers the namespaced Spanner /
// BigQuery extension functions: the BigQuery ObjectRef family under
// the "obj" sub-catalog and the Spanner AI / ML stubs under the
// "ai" / "ml" sub-catalogs. The flattened name (formatter
// "<ns>.<name>" → "<ns>_<name>") dispatches to the runtime UDF.
func registerSpannerNamespaceFuncs(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	mkArg := func(kind googlesql.TypeKind) *googlesql.FunctionArgumentType {
		return newSimpleArgType(factory, kind)
	}
	mkAny := newAnyArgType
	// BigQuery ObjectRef family. Register under the dotted "obj"
	// sub-catalog so OBJ.MAKE_REF, OBJ.FETCH_METADATA,
	// OBJ.GET_ACCESS_URL, OBJ.GET_READ_URL resolve at analyze time.
	addSubFn := func(nsCat *googlesql.SimpleCatalog, ns, name string, mode googlesql.FunctionEnums_Mode, ret *googlesql.FunctionArgumentType, argTypes []*googlesql.FunctionArgumentType) {
		if nsCat == nil || ret == nil {
			return
		}
		for _, a := range argTypes {
			if a == nil {
				return
			}
		}
		sig, err := googlesql.NewFunctionSignature3(ret, argTypes, 0)
		if err != nil || sig == nil {
			return
		}
		fn, err := googlesql.NewFunction([]string{ns + "_" + name}, "", mode, []*googlesql.FunctionSignature{sig}, nil)
		if err != nil || fn == nil {
			return
		}
		_ = nsCat.AddFunction2(name, fn)
	}
	objCat, _ := googlesql.NewSimpleCatalog("obj", factory)
	if objCat != nil {
		_ = catalog.AddCatalog(objCat)
		addSubFn(objCat, "obj", "make_ref", googlesql.FunctionEnums_ModeScalar,
			mkArg(googlesql.TypeKindTypeString),
			[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
		addSubFn(objCat, "obj", "fetch_metadata", googlesql.FunctionEnums_ModeScalar,
			mkArg(googlesql.TypeKindTypeJson),
			[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
		addSubFn(objCat, "obj", "get_access_url", googlesql.FunctionEnums_ModeScalar,
			mkArg(googlesql.TypeKindTypeString),
			[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
		addSubFn(objCat, "obj", "get_read_url", googlesql.FunctionEnums_ModeScalar,
			mkArg(googlesql.TypeKindTypeString),
			[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	}

	// Spanner AI / ML stubs under the "ai" / "ml" namespaces.
	any1 := mkAny(googlesql.SignatureArgumentKindArgTypeAny1)
	aiCat, _ := googlesql.NewSimpleCatalog("ai", factory)
	if aiCat != nil {
		_ = catalog.AddCatalog(aiCat)
		addSubFn(aiCat, "ai", "if", googlesql.FunctionEnums_ModeScalar,
			any1,
			[]*googlesql.FunctionArgumentType{
				mkArg(googlesql.TypeKindTypeString),
				any1,
				any1,
			})
		addSubFn(aiCat, "ai", "classify", googlesql.FunctionEnums_ModeScalar,
			mkArg(googlesql.TypeKindTypeString),
			[]*googlesql.FunctionArgumentType{
				mkArg(googlesql.TypeKindTypeString),
				mkAny(googlesql.SignatureArgumentKindArgArrayTypeAny1),
			})
		addSubFn(aiCat, "ai", "score", googlesql.FunctionEnums_ModeScalar,
			mkArg(googlesql.TypeKindTypeDouble),
			[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)})
	}
	mlCat, _ := googlesql.NewSimpleCatalog("ml", factory)
	if mlCat != nil {
		_ = catalog.AddCatalog(mlCat)
		addSubFn(mlCat, "ml", "predict", googlesql.FunctionEnums_ModeScalar,
			any1,
			[]*googlesql.FunctionArgumentType{
				mkArg(googlesql.TypeKindTypeString),
				any1,
			})
	}
}

// registerBigQueryExtensionFunctions adds BigQuery-only functions
// that GoogleSQL upstream does not ship as builtins. Consumer code
// needs them, and the analyzer must
// recognise the call sites before we can route them at runtime.
//
// These are registered with mode=Scalar/Aggregate and explicit
// signatures, mirroring the way user CREATE FUNCTION statements
// register through addFunctionSpec.
func registerBigQueryExtensionFunctions(catalog *googlesql.SimpleCatalog) {
	factory := tf()
	// The BigQuery extension surface is split into focused per-family
	// sub-functions. They are invoked in the SAME ORDER as the
	// original flat registration sequence; catalog registration is
	// order-sensitive, so the order must not change.
	registerBigQueryStringAggFuncs(catalog, factory)
	registerBigQueryTextAnalysisFuncs(catalog, factory)
	registerBigQuerySearchFuncs(catalog, factory)
	registerBigQueryTfIdfFuncs(catalog, factory)
}

// bqArgBuilders bundles the FunctionArgumentType / signature
// construction helpers shared by the BigQuery extension family
// sub-functions. It is a mechanical extraction of the closures the
// original flat registerBigQueryExtensionFunctions defined inline.
type bqArgBuilders struct {
	catalog *googlesql.SimpleCatalog
	factory *googlesql.TypeFactory
}

func newBQArgBuilders(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) *bqArgBuilders {
	return &bqArgBuilders{catalog: catalog, factory: factory}
}

func (b *bqArgBuilders) mkArg(kind googlesql.TypeKind) *googlesql.FunctionArgumentType {
	return newSimpleArgType(b.factory, kind)
}

func (b *bqArgBuilders) mkArrayArg(elemKind googlesql.TypeKind) *googlesql.FunctionArgumentType {
	elem, err := b.factory.MakeSimpleType(elemKind)
	if err != nil || elem == nil {
		return nil
	}
	arr, err := b.factory.MakeArrayType2(elem)
	if err != nil || arr == nil {
		return nil
	}
	opt, err := googlesql.NewFunctionArgumentTypeOptions()
	if err != nil || opt == nil {
		return nil
	}
	arg, err := googlesql.NewFunctionArgumentType(arr, opt, -1)
	if err != nil {
		return nil
	}
	return arg
}

func (b *bqArgBuilders) mkArrayStructArg(fields []*googlesql.StructField) *googlesql.FunctionArgumentType {
	st, err := b.factory.MakeStructType2(fields)
	if err != nil || st == nil {
		return nil
	}
	arr, err := b.factory.MakeArrayType2(st)
	if err != nil || arr == nil {
		return nil
	}
	opt, err := googlesql.NewFunctionArgumentTypeOptions()
	if err != nil || opt == nil {
		return nil
	}
	arg, err := googlesql.NewFunctionArgumentType(arr, opt, -1)
	if err != nil {
		return nil
	}
	return arg
}

func (b *bqArgBuilders) mkSimpleType(kind googlesql.TypeKind) googlesql.Googlesql_TypeNode {
	t, err := b.factory.MakeSimpleType(kind)
	if err != nil {
		return nil
	}
	return t
}

func (b *bqArgBuilders) mkSig(retArg *googlesql.FunctionArgumentType, args []*googlesql.FunctionArgumentType) *googlesql.FunctionSignature {
	if retArg == nil {
		return nil
	}
	for _, a := range args {
		if a == nil {
			return nil
		}
	}
	sig, err := googlesql.NewFunctionSignature3(retArg, args, 0)
	if err != nil {
		return nil
	}
	return sig
}

func (b *bqArgBuilders) register(name string, mode googlesql.FunctionEnums_Mode, sigs ...*googlesql.FunctionSignature) {
	valid := len(sigs) > 0
	for _, s := range sigs {
		if s == nil {
			valid = false
			break
		}
	}
	if !valid {
		return
	}
	fn, err := googlesql.NewFunction([]string{name}, "", mode, sigs, nil)
	if err != nil || fn == nil {
		return
	}
	_ = b.catalog.AddFunction2(name, fn)
}

// registerWithOpts mirrors register() but threads an explicit
// FunctionOptions value — used by aggregate-shaped functions
// that need to also support OVER() (TF_IDF), etc.
func (b *bqArgBuilders) registerWithOpts(name string, mode googlesql.FunctionEnums_Mode, opts *googlesql.FunctionOptions, sigs ...*googlesql.FunctionSignature) {
	valid := len(sigs) > 0
	for _, s := range sigs {
		if s == nil {
			valid = false
			break
		}
	}
	if !valid {
		return
	}
	fn, err := googlesql.NewFunction([]string{name}, "", mode, sigs, opts)
	if err != nil || fn == nil {
		return
	}
	_ = b.catalog.AddFunction2(name, fn)
}

// registerBigQueryStringAggFuncs registers CONTAINS_SUBSTR and the
// MAX_BY / MIN_BY aggregates.
func registerBigQueryStringAggFuncs(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	b := newBQArgBuilders(catalog, factory)
	mkArg := b.mkArg
	mkAny := newAnyArgType
	mkSig := b.mkSig
	register := b.register
	// CONTAINS_SUBSTR(STRING, STRING) -> BOOL
	register(
		"contains_substr",
		googlesql.FunctionEnums_ModeScalar,
		mkSig(
			mkArg(googlesql.TypeKindTypeBool),
			[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString), mkArg(googlesql.TypeKindTypeString)},
		),
	)
	// MAX_BY(value: T1, key: T2) -> T1
	register(
		"max_by",
		googlesql.FunctionEnums_ModeAggregate,
		mkSig(
			mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
			[]*googlesql.FunctionArgumentType{
				mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
				mkAny(googlesql.SignatureArgumentKindArgTypeAny2),
			},
		),
	)
	// MIN_BY(value: T1, key: T2) -> T1
	register(
		"min_by",
		googlesql.FunctionEnums_ModeAggregate,
		mkSig(
			mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
			[]*googlesql.FunctionArgumentType{
				mkAny(googlesql.SignatureArgumentKindArgTypeAny1),
				mkAny(googlesql.SignatureArgumentKindArgTypeAny2),
			},
		),
	)
}

// registerBigQueryTextAnalysisFuncs registers TEXT_ANALYZE and
// BAG_OF_WORDS.
func registerBigQueryTextAnalysisFuncs(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	b := newBQArgBuilders(catalog, factory)
	mkArg := b.mkArg
	mkArrayArg := b.mkArrayArg
	mkArrayStructArg := b.mkArrayStructArg
	mkSimpleType := b.mkSimpleType
	mkSig := b.mkSig
	register := b.register
	// TEXT_ANALYZE(text: STRING) -> ARRAY<STRING>
	// Optional analyzer / analyzer_options arguments are accepted
	// positionally; the runtime UDF parses them from the trailing
	// STRING args.
	register(
		"text_analyze",
		googlesql.FunctionEnums_ModeScalar,
		mkSig(
			mkArrayArg(googlesql.TypeKindTypeString),
			[]*googlesql.FunctionArgumentType{mkArg(googlesql.TypeKindTypeString)},
		),
		mkSig(
			mkArrayArg(googlesql.TypeKindTypeString),
			[]*googlesql.FunctionArgumentType{
				mkArg(googlesql.TypeKindTypeString),
				mkArg(googlesql.TypeKindTypeString),
			},
		),
		mkSig(
			mkArrayArg(googlesql.TypeKindTypeString),
			[]*googlesql.FunctionArgumentType{
				mkArg(googlesql.TypeKindTypeString),
				mkArg(googlesql.TypeKindTypeString),
				mkArg(googlesql.TypeKindTypeString),
			},
		),
	)
	// BAG_OF_WORDS(tokens: ARRAY<STRING>) -> ARRAY<STRUCT<term STRING, count INT64>>
	bowFields := []*googlesql.StructField{
		{Name: "term", Type_: mkSimpleType(googlesql.TypeKindTypeString)},
		{Name: "count", Type_: mkSimpleType(googlesql.TypeKindTypeInt64)},
	}
	register(
		"bag_of_words",
		googlesql.FunctionEnums_ModeScalar,
		mkSig(
			mkArrayStructArg(bowFields),
			[]*googlesql.FunctionArgumentType{mkArrayArg(googlesql.TypeKindTypeString)},
		),
	)
}

// registerBigQuerySearchFuncs registers the polymorphic SEARCH
// function.
func registerBigQuerySearchFuncs(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	b := newBQArgBuilders(catalog, factory)
	mkArg := b.mkArg
	mkAny := newAnyArgType
	mkSig := b.mkSig
	register := b.register
	// SEARCH(data, query: STRING [, analyzer => STRING]
	//        [, analyzer_options => STRING]
	//        [, json_scope => STRING]) -> BOOL
	//
	// `data` accepts STRING / JSON / STRUCT / ARRAY<STRING> /
	// ARRAY<STRUCT>. We register polymorphic signatures via
	// ARG_TYPE_ARBITRARY so the analyzer accepts the call site
	// regardless of shape; the runtime then walks the actual value
	// in BindSearch / collectDataTokens.
	mkArbArg := mkAny(googlesql.SignatureArgumentKindArgTypeArbitrary)
	mkSearchSig := func(extraOpts int) *googlesql.FunctionSignature {
		args := []*googlesql.FunctionArgumentType{
			mkArbArg,
			mkArg(googlesql.TypeKindTypeString),
		}
		for range extraOpts {
			args = append(args, mkArg(googlesql.TypeKindTypeString))
		}
		return mkSig(mkArg(googlesql.TypeKindTypeBool), args)
	}
	register(
		"search",
		googlesql.FunctionEnums_ModeScalar,
		mkSearchSig(0),
		mkSearchSig(1),
		mkSearchSig(2),
		mkSearchSig(3),
	)
}

// registerBigQueryTfIdfFuncs registers the TF_IDF analytic
// aggregate.
func registerBigQueryTfIdfFuncs(catalog *googlesql.SimpleCatalog, factory *googlesql.TypeFactory) {
	b := newBQArgBuilders(catalog, factory)
	mkArg := b.mkArg
	mkArrayArg := b.mkArrayArg
	mkArrayStructArg := b.mkArrayStructArg
	mkSimpleType := b.mkSimpleType
	mkSig := b.mkSig
	registerWithOpts := b.registerWithOpts
	// TF_IDF(tokens: ARRAY<STRING>) -> ARRAY<STRUCT<term STRING, score FLOAT64>>
	// Registered ModeAggregate so it composes with OVER() per
	// BigQuery's TF_IDF analytic surface. Extra positional args
	// (max_distinct_tokens, frequency_threshold) accepted as INT64.
	tfidfFields := []*googlesql.StructField{
		{Name: "term", Type_: mkSimpleType(googlesql.TypeKindTypeString)},
		{Name: "score", Type_: mkSimpleType(googlesql.TypeKindTypeDouble)},
	}
	registerWithOpts(
		"tf_idf",
		googlesql.FunctionEnums_ModeAggregate,
		&googlesql.FunctionOptions{SupportsOverClause: true},
		mkSig(
			mkArrayStructArg(tfidfFields),
			[]*googlesql.FunctionArgumentType{mkArrayArg(googlesql.TypeKindTypeString)},
		),
		mkSig(
			mkArrayStructArg(tfidfFields),
			[]*googlesql.FunctionArgumentType{
				mkArrayArg(googlesql.TypeKindTypeString),
				mkArg(googlesql.TypeKindTypeInt64),
			},
		),
		mkSig(
			mkArrayStructArg(tfidfFields),
			[]*googlesql.FunctionArgumentType{
				mkArrayArg(googlesql.TypeKindTypeString),
				mkArg(googlesql.TypeKindTypeInt64),
				mkArg(googlesql.TypeKindTypeInt64),
			},
		),
	)
}

func NewCatalog(db *sql.DB) *Catalog {
	sc := newSimpleCatalog(catalogName)
	// Attach a DescriptorPool to the catalog at construction time so
	// the proto-related builtins (FROM_PROTO / TO_PROTO / EXTRACT /
	// FILTER_FIELDS / REPLACE_FIELDS / PROTO_DEFAULT_IF_NULL /
	// PROTO_MAP_* / ENUM_VALUE_DESCRIPTOR_PROTO) can resolve their
	// argument types against well-known proto descriptors.
	// SetDescriptorPool can only be called ONCE per catalog (per
	// upstream comment).
	var pool *googlesql.DescriptorPool
	if p, err := googlesql.NewDescriptorPool(); err == nil && p != nil && sc != nil {
		_ = sc.SetDescriptorPool(p)
		pool = p
	}
	c := &Catalog{
		db:                   db,
		catalog:              sc,
		descriptorPool:       pool,
		registeredProtos:     map[string]*googlesql.ProtoType{},
		registeredEnums:      map[string]*googlesql.EnumType{},
		tableMap:             map[string]*TableSpec{},
		funcMap:              map[string]*FunctionSpec{},
		tvfMap:               map[string]*TVFSpec{},
		subCatalogs:          map[*googlesql.SimpleCatalog]map[string]*googlesql.SimpleCatalog{},
		infoSchemaRegistered: map[string]struct{}{},
		infoSchemaTables:     map[string]*infoSchemaTableMeta{},
		propertyGraphs:       map[string]*propertyGraphSpec{},
	}
	// Auto-register protobuf well-known types (google.protobuf.* and
	// google.type.*). BigQuery and Spanner treat these as built-in,
	// so user SQL referencing them resolves without manual descriptor
	// wiring. Acquire wasmAnalyzeMu around the proto FFI block so
	// other goroutines that are mid-Analyze do not see a partially-
	// populated DescriptorPool / SimpleCatalog while AddType is in
	// flight.
	wasmAnalyzeMu.Lock()
	_ = registerWellKnownProtos(c)
	wasmAnalyzeMu.Unlock()
	return c
}

// DescriptorPool returns the catalog's protobuf DescriptorPool. The
// pool is created at NewCatalog time and shared across the catalog's
// lifetime; consumers register descriptors through RegisterProto.
func (c *Catalog) DescriptorPool() *googlesql.DescriptorPool {
	return c.descriptorPool
}

// RegisterProto adds a serialized FileDescriptorProto to the catalog's
// DescriptorPool. The bytes must be the binary protobuf encoding of a
// google.protobuf.FileDescriptorProto (matching the output of
// google.protobuf.FileDescriptorProto.SerializeToString in C++ or
// proto.Marshal on a *descriptorpb.FileDescriptorProto in Go).
//
// After successful registration the file's messages and enums are
// resolvable through Catalog.LookupProtoType / Catalog.LookupEnumType,
// which build *ProtoType / *EnumType via TypeFactory.MakeProtoType /
// MakeEnumType and pin them on the catalog under their fully-qualified
// names so subsequent SQL can reference them.
func (c *Catalog) RegisterProto(fileDescriptorProtoBytes []byte) error {
	if c.descriptorPool == nil {
		return fmt.Errorf("catalog has no DescriptorPool — proto registration unavailable")
	}
	if len(fileDescriptorProtoBytes) == 0 {
		return fmt.Errorf("RegisterProto: empty bytes")
	}
	fdProto, err := googlesql.NewFileDescriptorProto()
	if err != nil || fdProto == nil {
		return fmt.Errorf("RegisterProto: NewFileDescriptorProto: %w", err)
	}
	ok, err := fdProto.ParseFromString(string(fileDescriptorProtoBytes))
	if err != nil {
		return fmt.Errorf("RegisterProto: ParseFromString: %w", err)
	}
	if !ok {
		return fmt.Errorf("RegisterProto: invalid FileDescriptorProto wire format")
	}
	c.mu.Lock()
	fd, err := c.descriptorPool.BuildFile(fdProto)
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("RegisterProto: BuildFile: %w", err)
	}
	_ = fd
	c.mu.Unlock()
	// Mirror the registration on the Go-side proto registry so the
	// runtime (CAST(STRING AS Proto) text-format parsing, value
	// formatters, etc.) can resolve the descriptor without crossing
	// the wasm bridge. Failure here is non-fatal — the wasm pool
	// already has the FileDescriptor for analyzer use.
	_ = registerGoProtoFileFromBytes(fileDescriptorProtoBytes)
	// Auto-register every message declared in the file as a
	// SQL-visible proto type so callers don't have to follow each
	// RegisterProto with explicit RegisterProtoMessage calls. We
	// derive the names from the Go-side FileDescriptor (already built
	// above) which is cheaper than re-walking the wasm bridge.
	for _, fullName := range messageFullNamesFromGoFile(fileDescriptorProtoBytes) {
		if _, err := c.RegisterProtoMessage(fullName); err != nil {
			return fmt.Errorf("RegisterProto: auto-promote %q: %w", fullName, err)
		}
	}
	return nil
}

// RegisterProtoMessage promotes a message already present in the
// DescriptorPool (typically populated by an earlier RegisterProto
// call) into a SQL-visible PROTO type pinned at `fullName` on the
// catalog. After this call, `CREATE TABLE t (m proto<fullName>)` and
// `SELECT proto_field` accesses against columns of that type become
// resolvable; PROTO_DEFAULT_IF_NULL / EXTRACT / FROM_PROTO / TO_PROTO /
// FILTER_FIELDS / REPLACE_FIELDS / PROTO_MAP_* function calls bind
// against the resulting *ProtoType.
//
// `fullName` is the dotted google.protobuf-style name
// ("foo.bar.MyMessage"). The same `fullName` is used as the catalog
// key and is what SQL must reference.
func (c *Catalog) RegisterProtoMessage(fullName string) (*googlesql.ProtoType, error) {
	if c.descriptorPool == nil {
		return nil, fmt.Errorf("catalog has no DescriptorPool — proto registration unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if pt, ok := c.registeredProtos[fullName]; ok {
		return pt, nil
	}
	desc, err := c.descriptorPool.FindMessageTypeByName(fullName)
	if err != nil || desc == nil {
		return nil, fmt.Errorf("RegisterProtoMessage: FindMessageTypeByName(%q): %w", fullName, err)
	}
	pt, err := tf().MakeProtoType(desc, nil)
	if err != nil || pt == nil {
		return nil, fmt.Errorf("RegisterProtoMessage: MakeProtoType(%q): %w", fullName, err)
	}
	if err := c.catalog.AddType(fullName, pt); err != nil {
		return nil, fmt.Errorf("RegisterProtoMessage: AddType(%q): %w", fullName, err)
	}
	c.registeredProtos[fullName] = pt
	registerProtoTypeGlobal(fullName, pt)
	return pt, nil
}

// RegisterEnum is the analogue of RegisterProtoMessage for an ENUM
// declared inside the registered FileDescriptorProto. `fullName` is
// the dotted name of the enum type. The returned *EnumType is pinned
// on the catalog at that name.
func (c *Catalog) RegisterEnum(fullName string) (*googlesql.EnumType, error) {
	if c.descriptorPool == nil {
		return nil, fmt.Errorf("catalog has no DescriptorPool — enum registration unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if et, ok := c.registeredEnums[fullName]; ok {
		return et, nil
	}
	desc, err := c.descriptorPool.FindEnumTypeByName(fullName)
	if err != nil || desc == nil {
		return nil, fmt.Errorf("RegisterEnum: FindEnumTypeByName(%q): %w", fullName, err)
	}
	et, err := tf().MakeEnumType(desc, nil)
	if err != nil || et == nil {
		return nil, fmt.Errorf("RegisterEnum: MakeEnumType(%q): %w", fullName, err)
	}
	if err := c.catalog.AddType(fullName, et); err != nil {
		return nil, fmt.Errorf("RegisterEnum: AddType(%q): %w", fullName, err)
	}
	c.registeredEnums[fullName] = et
	registerEnumTypeGlobal(fullName, et)
	return et, nil
}

// LookupProtoType returns the previously-registered *ProtoType or nil
// if `fullName` has not been promoted yet.
func (c *Catalog) LookupProtoType(fullName string) *googlesql.ProtoType {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registeredProtos[fullName]
}

// AddPropertyGraph records a CREATE PROPERTY GRAPH statement.
// Idempotent on the canonical name path; CREATE_OR_REPLACE simply
// overwrites the previous entry.
func (c *Catalog) AddPropertyGraph(spec *propertyGraphSpec) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.propertyGraphs[spec.Name] = spec
}

// PropertyGraph returns the registered PropertyGraphSpec under
// `name`, or nil if no such graph exists.
func (c *Catalog) PropertyGraph(name string) *propertyGraphSpec {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.propertyGraphs[name]
}

// StorageNameForFunction returns the underlying SQLite UDF name for a
// Function handle that the analyzer resolved against this catalog.
// Each Function handle is constructed once per spec with Name() set
// to the storage path, then aliased into sub-catalogs via
// SimpleCatalog.AddFunction2, so handle.Name() is the storage name
// regardless of which catalog level the analyzer descended through.
func (c *Catalog) StorageNameForFunction(fn *googlesql.Function) string {
	if fn == nil {
		return ""
	}
	name, err := fn.Name()
	if err != nil {
		return ""
	}
	return name
}

// StorageNameForTable returns the underlying SQLite table name for a
// SimpleTable handle the analyzer resolved against this catalog.
// Same shape as StorageNameForFunction: the SimpleTable's Name() is
// pinned to the storage path at construction and the same handle is
// aliased into every sub-catalog level via AddTable2, so handle.Name()
// is the storage name everywhere.
func (c *Catalog) StorageNameForTable(table googlesql.TableNode) string {
	if table == nil {
		return ""
	}
	name, err := table.Name()
	if err != nil {
		return ""
	}
	return name
}

// getOrCreateSubCatalog returns the sub-catalog registered under
// parent+name, creating and attaching it on first use. AddCatalog is
// called exactly once per (parent, sub-name) so wasm-side
// deduplication traps are impossible.
func (c *Catalog) getOrCreateSubCatalog(parent *googlesql.SimpleCatalog, name string) *googlesql.SimpleCatalog {
	if parent == nil {
		return nil
	}
	subs := c.subCatalogs[parent]
	if subs == nil {
		subs = map[string]*googlesql.SimpleCatalog{}
		c.subCatalogs[parent] = subs
	}
	if existing, ok := subs[name]; ok {
		return existing
	}
	sub := newSimpleCatalog(name)
	if sub == nil {
		return nil
	}
	_ = parent.AddCatalog(sub)
	subs[name] = sub
	return sub
}

func (c *Catalog) FullName() string {
	s, _ := c.catalog.FullName()
	return s
}

// Find* methods are currently unsupported through the wasm bridge: the
// underlying googlesql::Catalog templated Find<T>(...) variants use
// output-pointer parameters the bridge cannot marshal today. Returning
// the stored SimpleCatalog handle via FindOptions would require a
// Go-side callback catalog wrapper, which is tracked separately.
func (c *Catalog) FindTable(path []string) (googlesql.TableNode, error) {
	if c.isWildcardTable(path) {
		return c.createWildcardTable(path)
	}
	return nil, fmt.Errorf("catalog: FindTable not yet supported via wasm bridge")
}

// registerWildcardTableByPath pre-creates a wildcard table for `path`
// and installs it into the googlesql catalog so the analyzer can
// resolve the reference. Called from the analyzer before a statement
// is analyzed (see preRegisterWildcardTables). Silent if the path
// does not match any registered tables.
func (c *Catalog) registerWildcardTableByPath(path []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isWildcardTable(path) {
		return
	}
	wt, err := c.createWildcardTableImpl(path)
	if err != nil || wt == nil {
		return
	}
	simpleTable, err := c.createSimpleTable(strings.Join(path, "."), wt.spec)
	if err != nil {
		return
	}
	registerWildcardTable(simpleTable, wt)
	// Install into the wasm-side catalog tree. Walk the leading path
	// segments as sub-catalogs so the analyzer resolves the reference
	// in either form (fully-qualified or dataset-qualified).
	c.installWildcardIntoCatalog(c.catalog, path, simpleTable)
}

// installWildcardIntoCatalog adds `table` under each leading namespace
// prefix of `path`, so `project.dataset.table_*` resolves regardless
// of whether the query omits the project identifier.
func (c *Catalog) installWildcardIntoCatalog(cat *googlesql.SimpleCatalog, path []string, table *googlesql.SimpleTable) {
	if len(path) == 0 {
		return
	}
	if len(path) == 1 {
		if c.existsTable(cat, path[0]) {
			return
		}
		_ = cat.AddTable2(path[0], table)
		return
	}
	subName := path[0]
	sub := c.getOrCreateSubCatalog(cat, subName)
	c.installWildcardIntoCatalog(cat, path[1:], table)
	c.installWildcardIntoCatalog(sub, path[1:], table)
}

func (c *Catalog) normalizeTablePath(path []string) []string {
	result := []string{}
	for _, p := range path {
		parts := strings.Split(p, ".")
		result = append(result, parts...)
	}
	return result
}

func (c *Catalog) FindModel(path []string) (googlesql.ModelNode, error) {
	return nil, fmt.Errorf("catalog: FindModel not yet supported via wasm bridge")
}

func (c *Catalog) FindConnection(path []string) (googlesql.ConnectionNode, error) {
	return nil, fmt.Errorf("catalog: FindConnection not yet supported via wasm bridge")
}

func (c *Catalog) FindFunction(path []string) (*googlesql.Function, error) {
	return nil, fmt.Errorf("catalog: FindFunction not yet supported via wasm bridge")
}

func (c *Catalog) FindTableValuedFunction(path []string) (*googlesql.TableValuedFunction, error) {
	return nil, fmt.Errorf("catalog: FindTableValuedFunction not yet supported via wasm bridge")
}

func (c *Catalog) FindProcedure(path []string) (*googlesql.Procedure, error) {
	return nil, fmt.Errorf("catalog: FindProcedure not yet supported via wasm bridge")
}

func (c *Catalog) FindType(path []string) (googlesql.Googlesql_TypeNode, error) {
	return nil, fmt.Errorf("catalog: FindType not yet supported via wasm bridge")
}

func (c *Catalog) FindConstant(path []string) (googlesql.ConstantNode, int, error) {
	return nil, 0, fmt.Errorf("catalog: FindConstant not yet supported via wasm bridge")
}

func (c *Catalog) FindConversion(from, to googlesql.Googlesql_TypeNode) (*googlesql.Conversion, error) {
	return nil, fmt.Errorf("catalog: FindConversion not yet supported via wasm bridge")
}

// ExtendedTypeSuperTypes used to return *TypeListView but that alias was
// dropped when the clang-AST parser started expanding the FileScope
// typedef (TypeListView = absl::Span<const Type *const>) at parse time.
// Callers of this method only ever checked the error, so returning an
// empty slice keeps the contract without depending on the removed alias.
func (c *Catalog) ExtendedTypeSuperTypes(typ googlesql.Googlesql_TypeNode) ([]googlesql.Googlesql_TypeNode, error) {
	return nil, fmt.Errorf("catalog: ExtendedTypeSuperTypes not yet supported via wasm bridge")
}

func (c *Catalog) SuggestTable(mistypedPath []string) string {
	s, _ := c.catalog.SuggestTable(strings.Join(mistypedPath, "."))
	return s
}

func (c *Catalog) SuggestModel(mistypedPath []string) string {
	// SuggestModel isn't exposed on SimpleCatalog via the wasm bridge.
	return ""
}

func (c *Catalog) SuggestFunction(mistypedPath []string) string {
	s, _ := c.catalog.SuggestFunction(strings.Join(mistypedPath, "."))
	return s
}

func (c *Catalog) SuggestTableValuedFunction(mistypedPath []string) string {
	s, _ := c.catalog.SuggestTableValuedFunction(strings.Join(mistypedPath, "."))
	return s
}

func (c *Catalog) SuggestConstant(mistypedPath []string) string {
	s, _ := c.catalog.SuggestConstant(strings.Join(mistypedPath, "."))
	return s
}

func (c *Catalog) formatNamePath(path []string) string {
	return strings.Join(path, "_")
}

func (c *Catalog) getTVFs(namePath *NamePath) []*TVFSpec {
	if namePath.empty() {
		return c.tvfs
	}
	key := c.formatNamePath(namePath.path)
	specs := make([]*TVFSpec, 0, len(c.tvfs))
	for _, tvf := range c.tvfs {
		if len(tvf.NamePath) == 1 {
			specs = append(specs, tvf)
			continue
		}
		pathPrefixKey := c.formatNamePath(c.trimmedLastPath(tvf.NamePath))
		if strings.Contains(pathPrefixKey, key) {
			specs = append(specs, tvf)
		}
	}
	return specs
}

func (c *Catalog) getFunctions(namePath *NamePath) []*FunctionSpec {
	if namePath.empty() {
		return c.functions
	}
	key := c.formatNamePath(namePath.path)
	specs := make([]*FunctionSpec, 0, len(c.functions))
	for _, fn := range c.functions {
		if len(fn.NamePath) == 1 {
			// function name only
			specs = append(specs, fn)
			continue
		}
		pathPrefixKey := c.formatNamePath(c.trimmedLastPath(fn.NamePath))
		if strings.Contains(pathPrefixKey, key) {
			specs = append(specs, fn)
		}
	}
	return specs
}

func (c *Catalog) Sync(ctx context.Context, conn *Conn) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.createCatalogTablesIfNotExists(ctx, conn); err != nil {
		return fmt.Errorf("failed to create catalog tables: %w", err)
	}
	now := time.Now()
	rows, err := conn.QueryContext(
		ctx,
		`SELECT name, kind, spec FROM googlesqlite_catalog WHERE updatedAt >= @lastUpdatedAt`,
		c.lastSyncedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to query load catalog: %w", err)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			name string
			kind catalogSpecKind
			spec string
		)
		if err := rows.Scan(&name, &kind, &spec); err != nil {
			return fmt.Errorf("failed to scan catalog values: %w", err)
		}
		switch kind {
		case TableSpecKind, ViewSpecKind:
			if err := c.loadTableSpec(spec); err != nil {
				return fmt.Errorf("failed to load table spec: %w", err)
			}
		case FunctionSpecKind:
			if err := c.loadFunctionSpec(spec); err != nil {
				return fmt.Errorf("failed to load function spec: %w", err)
			}
		case TVFSpecKind:
			if err := c.loadTVFSpec(spec); err != nil {
				return fmt.Errorf("failed to load TVF spec: %w", err)
			}
		default:
			return fmt.Errorf("unknown catalog spec kind %s", kind)
		}
	}
	c.lastSyncedAt = now
	return nil
}

func (c *Catalog) AddNewTableSpec(ctx context.Context, conn *Conn, spec *TableSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.addTableSpec(spec); err != nil {
		return err
	}
	if !spec.IsTemp {
		if err := c.saveTableSpec(ctx, conn, spec); err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) AddNewTVFSpec(ctx context.Context, conn *Conn, spec *TVFSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.addTVFSpec(spec); err != nil {
		return err
	}
	if !spec.IsTemp {
		if err := c.saveTVFSpec(ctx, conn, spec); err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) DeleteTVFSpec(ctx context.Context, conn *Conn, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.deleteTVFSpecByName(name); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, deleteCatalogQuery, sql.Named("name", name)); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) AddNewFunctionSpec(ctx context.Context, conn *Conn, spec *FunctionSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.addFunctionSpec(spec); err != nil {
		return err
	}
	if !spec.IsTemp {
		if err := c.saveFunctionSpec(ctx, conn, spec); err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) DeleteTableSpec(ctx context.Context, conn *Conn, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.deleteTableSpecByName(name); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, deleteCatalogQuery, sql.Named("name", name)); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) DeleteFunctionSpec(ctx context.Context, conn *Conn, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.deleteFunctionSpecByName(name); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, deleteCatalogQuery, sql.Named("name", name)); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) FunctionSpec(name string) (*FunctionSpec, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	spec, exists := c.funcMap[name]
	return spec, exists
}

func (c *Catalog) deleteTableSpecByName(name string) error {
	spec, exists := c.tableMap[name]
	if !exists {
		return fmt.Errorf("failed to find table spec from map by %s", name)
	}
	tables := make([]*TableSpec, 0, len(c.tables))
	specName := c.formatNamePath(spec.NamePath)
	for _, table := range c.tables {
		if specName == c.formatNamePath(table.NamePath) {
			continue
		}
		tables = append(tables, table)
	}
	if err := c.resetCatalog(tables, c.functions, c.tvfs); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) deleteFunctionSpecByName(name string) error {
	spec, exists := c.funcMap[name]
	if !exists {
		return fmt.Errorf("failed to find function spec from map by %s", name)
	}
	functions := make([]*FunctionSpec, 0, len(c.functions))
	specName := c.formatNamePath(spec.NamePath)
	for _, function := range c.functions {
		if specName == c.formatNamePath(function.NamePath) {
			continue
		}
		functions = append(functions, function)
	}
	if err := c.resetCatalog(c.tables, functions, c.tvfs); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) deleteTVFSpecByName(name string) error {
	spec, exists := c.tvfMap[name]
	if !exists {
		return fmt.Errorf("failed to find TVF spec from map by %s", name)
	}
	tvfs := make([]*TVFSpec, 0, len(c.tvfs))
	specName := c.formatNamePath(spec.NamePath)
	for _, tvf := range c.tvfs {
		if specName == c.formatNamePath(tvf.NamePath) {
			continue
		}
		tvfs = append(tvfs, tvf)
	}
	if err := c.resetCatalog(c.tables, c.functions, tvfs); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) resetCatalog(tables []*TableSpec, functions []*FunctionSpec, tvfs []*TVFSpec) error {
	c.catalog = newSimpleCatalog(catalogName)
	if c.catalog == nil {
		return fmt.Errorf("failed to create catalog")
	}
	// Rebuilding the SimpleCatalog throws away every type previously
	// registered through SimpleCatalog.AddType, so the well-known
	// proto / enum descriptors that NewCatalog put there at startup
	// vanish along with the dropped table/function/TVF. Re-attach
	// the existing DescriptorPool and re-pin every cached
	// *ProtoType / *EnumType under their full names — otherwise
	// a query that runs after a DROP fails with
	// "Type not found: google.type.Date" even though the descriptor
	// is still present in the Go-side DescriptorPool.
	if c.descriptorPool != nil {
		if err := c.catalog.SetDescriptorPool(c.descriptorPool); err != nil {
			return fmt.Errorf("resetCatalog: SetDescriptorPool: %w", err)
		}
	}
	for name, pt := range c.registeredProtos {
		if pt == nil {
			continue
		}
		if err := c.catalog.AddType(name, pt); err != nil {
			return fmt.Errorf("resetCatalog: AddType(%q): %w", name, err)
		}
	}
	for name, et := range c.registeredEnums {
		if et == nil {
			continue
		}
		if err := c.catalog.AddType(name, et); err != nil {
			return fmt.Errorf("resetCatalog: AddType(enum %q): %w", name, err)
		}
	}
	c.tables = []*TableSpec{}
	c.functions = []*FunctionSpec{}
	c.tvfs = []*TVFSpec{}
	c.tableMap = map[string]*TableSpec{}
	c.funcMap = map[string]*FunctionSpec{}
	c.tvfMap = map[string]*TVFSpec{}
	c.infoSchemaRegistered = map[string]struct{}{}
	c.infoSchemaTables = map[string]*infoSchemaTableMeta{}
	// Clear the sub-catalog cache. Its keys are the parent
	// *SimpleCatalog pointers — the OLD c.catalog is one of them.
	// Map keys keep their referents alive, so without this reset the
	// retired SimpleCatalog (plus every sub-catalog it spawned) would
	// stay reachable until the process ends. Under sustained DDL
	// pressure that pins thousands of catalogs in the wasm heap until
	// it exhausts and `wasm_alloc` starts returning OOB.
	c.subCatalogs = map[*googlesql.SimpleCatalog]map[string]*googlesql.SimpleCatalog{}
	for _, spec := range tables {
		if err := c.addTableSpec(spec); err != nil {
			return err
		}
	}
	for _, spec := range functions {
		if err := c.addFunctionSpec(spec); err != nil {
			return err
		}
	}
	for _, spec := range tvfs {
		if err := c.addTVFSpec(spec); err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) saveTableSpec(ctx context.Context, conn *Conn, spec *TableSpec) error {
	encoded, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to encode table spec: %w", err)
	}
	now := time.Now()
	kind := string(TableSpecKind)
	if spec.IsView {
		kind = string(ViewSpecKind)
	}
	if _, err := conn.ExecContext(
		ctx,
		upsertCatalogQuery,
		sql.Named("name", spec.TableName()),
		sql.Named("kind", kind),
		sql.Named("spec", string(encoded)),
		sql.Named("updatedAt", now),
		sql.Named("createdAt", now),
	); err != nil {
		return fmt.Errorf("failed to save a new table spec: %w", err)
	}
	return nil
}

func (c *Catalog) saveTVFSpec(ctx context.Context, conn *Conn, spec *TVFSpec) error {
	encoded, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to encode TVF spec: %w", err)
	}
	now := time.Now()
	if _, err := conn.ExecContext(
		ctx,
		upsertCatalogQuery,
		sql.Named("name", spec.TVFName()),
		sql.Named("kind", string(TVFSpecKind)),
		sql.Named("spec", string(encoded)),
		sql.Named("updatedAt", now),
		sql.Named("createdAt", now),
	); err != nil {
		return fmt.Errorf("failed to save a new TVF spec: %w", err)
	}
	return nil
}

func (c *Catalog) saveFunctionSpec(ctx context.Context, conn *Conn, spec *FunctionSpec) error {
	encoded, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to encode function spec: %w", err)
	}
	now := time.Now()
	if _, err := conn.ExecContext(
		ctx,
		upsertCatalogQuery,
		sql.Named("name", spec.FuncName()),
		sql.Named("kind", string(FunctionSpecKind)),
		sql.Named("spec", string(encoded)),
		sql.Named("updatedAt", now),
		sql.Named("createdAt", now),
	); err != nil {
		return fmt.Errorf("failed to save a new function spec: %w", err)
	}
	return nil
}

func (c *Catalog) createCatalogTablesIfNotExists(ctx context.Context, conn *Conn) error {
	if _, err := conn.ExecContext(ctx, createCatalogTableQuery); err != nil {
		return fmt.Errorf("failed to create catalog table: %w", err)
	}
	return nil
}

func (c *Catalog) loadTableSpec(spec string) error {
	var v TableSpec
	if err := json.Unmarshal([]byte(spec), &v); err != nil {
		return fmt.Errorf("failed to decode table spec: %w", err)
	}
	if err := c.addTableSpec(&v); err != nil {
		return fmt.Errorf("failed to add table spec to catalog: %w", err)
	}
	return nil
}

func (c *Catalog) loadFunctionSpec(spec string) error {
	var v FunctionSpec
	if err := json.Unmarshal([]byte(spec), &v); err != nil {
		return fmt.Errorf("failed to decode function spec: %w", err)
	}
	if err := c.addFunctionSpec(&v); err != nil {
		return fmt.Errorf("failed to add function spec to catalog: %w", err)
	}
	return nil
}

func (c *Catalog) loadTVFSpec(spec string) error {
	var v TVFSpec
	if err := json.Unmarshal([]byte(spec), &v); err != nil {
		return fmt.Errorf("failed to decode TVF spec: %w", err)
	}
	if err := c.addTVFSpec(&v); err != nil {
		return fmt.Errorf("failed to add TVF spec to catalog: %w", err)
	}
	return nil
}

func (c *Catalog) trimmedLastPath(path []string) []string {
	if len(path) == 0 {
		return path
	}
	return path[:len(path)-1]
}

func (c *Catalog) addTVFSpec(spec *TVFSpec) error {
	tvfName := spec.TVFName()
	if _, exists := c.tvfMap[tvfName]; exists {
		c.tvfMap[tvfName] = spec
		return nil
	}
	c.tvfs = append(c.tvfs, spec)
	c.tvfMap[tvfName] = spec
	if err := c.addTVFSpecRecursive(c.catalog, spec); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) addFunctionSpec(spec *FunctionSpec) error {
	funcName := spec.FuncName()
	if _, exists := c.funcMap[funcName]; exists {
		c.funcMap[funcName] = spec // update current spec
		return nil
	}
	c.functions = append(c.functions, spec)
	c.funcMap[funcName] = spec
	if err := c.addFunctionSpecRecursive(c.catalog, spec); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) addTableSpec(spec *TableSpec) error {
	tableName := spec.TableName()
	if _, exists := c.tableMap[tableName]; exists {
		// Refresh the entry in-place: CREATE OR REPLACE TABLE has to
		// update INFORMATION_SCHEMA's view of columns and options as
		// well as the resolver-side handle. Replace the existing
		// pointer in c.tables and re-register through the analyzer
		// catalog. Mapping the new pointer keeps any cached
		// re-analysis pinned to the latest spec.
		c.tableMap[tableName] = spec
		for i := range c.tables {
			if c.tables[i].TableName() == tableName {
				c.tables[i] = spec
				break
			}
		}
		return nil
	}
	c.tables = append(c.tables, spec)
	c.tableMap[tableName] = spec
	if err := c.addTableSpecRecursive(c.catalog, spec); err != nil {
		return err
	}
	// Make INFORMATION_SCHEMA views resolvable for every dataset that
	// hosts at least one table. Registration is idempotent; the vtab
	// handles row materialisation on query.
	//
	// BigQuery's table reference is `project.dataset.table` (or
	// `dataset.table` when the project is implicit). The schema /
	// dataset name in INFORMATION_SCHEMA is therefore the
	// SECOND-from-last segment, not always NamePath[1].
	if path := spec.NamePath; len(path) >= 2 {
		project, dataset := "", path[len(path)-2]
		if len(path) >= 3 {
			project = path[len(path)-3]
		}
		if err := c.ensureInfoSchemaForDataset(project, dataset); err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) addTableSpecRecursive(cat *googlesql.SimpleCatalog, spec *TableSpec) error {
	table, err := c.tableHandleForSpec(spec)
	if err != nil {
		return err
	}
	if table == nil {
		return fmt.Errorf("failed to build table %q", spec.TableName())
	}
	return c.addTableSpecRecursiveImpl(cat, spec, table)
}

// addTableSpecRecursiveImpl aliases the same SimpleTable handle into
// every (catalog, lookup-name) the analyzer might descend through:
// the root catalog under each prefix of the spec's NamePath
// ("project.dataset.table", "dataset.table", "table"), each
// sub-catalog under the rest of the path, and so on. Aliasing uses
// SimpleCatalog.AddTable2(name, table) which accepts a caller-provided
// lookup key instead of using table.Name(); the table's own Name()
// stays pinned to the storage path for the formatter.
func (c *Catalog) addTableSpecRecursiveImpl(cat *googlesql.SimpleCatalog, spec *TableSpec, table *googlesql.SimpleTable) error {
	if len(spec.NamePath) > 1 {
		subCatalogName := spec.NamePath[0]
		// Reuse the previously-added sub-catalog when the same name
		// reappears (e.g. "project" on the path of every table). Adding
		// a new SimpleCatalog handle under an already-registered name
		// traps the wasm runtime because SimpleCatalog::AddCatalog
		// rejects duplicates via a CHECK.
		subCatalog := c.getOrCreateSubCatalog(cat, subCatalogName)
		if subCatalog == nil {
			return fmt.Errorf("failed to register sub-catalog %q", subCatalogName)
		}
		fullTableName := strings.Join(spec.NamePath, ".")
		if !c.existsTable(cat, fullTableName) {
			if err := cat.AddTable2(fullTableName, table); err != nil {
				return fmt.Errorf("failed to register table %q: %w", fullTableName, err)
			}
		}
		newNamePath := spec.NamePath[1:]
		// add sub catalog to root catalog
		if err := c.addTableSpecRecursiveImpl(cat, c.copyTableSpec(spec, newNamePath), table); err != nil {
			return fmt.Errorf("failed to add table spec to root catalog: %w", err)
		}
		// add sub catalog to parent catalog
		if err := c.addTableSpecRecursiveImpl(subCatalog, c.copyTableSpec(spec, newNamePath), table); err != nil {
			return fmt.Errorf("failed to add table spec to parent catalog: %w", err)
		}
		return nil
	}
	if len(spec.NamePath) == 0 {
		return fmt.Errorf("table name is not found")
	}

	tableName := spec.NamePath[0]
	if c.existsTable(cat, tableName) {
		return nil
	}
	if err := cat.AddTable2(tableName, table); err != nil {
		return fmt.Errorf("failed to register table %q: %w", tableName, err)
	}
	return nil
}

// tableHandleForSpec builds a SimpleTable for spec with Name() pinned
// to the storage-format path (`spec.TableName()`). One handle is
// constructed per call to addTableSpecRecursive and aliased into
// every (catalog, lookup-name) combination the analyzer can descend
// through, so handle.Name() is the storage name everywhere — no
// side map / pointer-identity machinery needed at format time.
//
// We deliberately do NOT cache the handle across calls: a recreated
// table may reuse the same storage name with a different column set,
// and an old cached handle would then expose stale columns.
func (c *Catalog) tableHandleForSpec(spec *TableSpec) (*googlesql.SimpleTable, error) {
	storageName := spec.TableName()
	columns := []*googlesql.SimpleColumn{}
	colIndex := map[string]int32{}
	for i, column := range spec.Columns {
		typ, err := column.Type.ToGoogleSQLType()
		if err != nil {
			return nil, err
		}
		columns = append(columns, m1(googlesql.NewSimpleColumn(storageName, column.Name, typ, false, true)))
		colIndex[column.Name] = int32(i)
	}
	tbl := newSimpleTableWithColumns(storageName, columns)
	if tbl == nil {
		return nil, fmt.Errorf("failed to build table %q", storageName)
	}
	if len(spec.PrimaryKey) > 0 {
		pkIdx := make([]int32, 0, len(spec.PrimaryKey))
		for _, name := range spec.PrimaryKey {
			if idx, ok := colIndex[name]; ok {
				pkIdx = append(pkIdx, idx)
			}
		}
		if len(pkIdx) == len(spec.PrimaryKey) {
			_ = tbl.SetPrimaryKey(pkIdx)
		}
	}
	applyTableOptions(tbl, spec.Options)
	return tbl, nil
}

// applyTableOptions consumes the CREATE TABLE / CREATE VIEW OPTIONS
// list and applies any option that has a corresponding analyzer-
// visible setting on SimpleTable. Currently the only option honoured
// is `anonymization_userid_column`, which tells the resolver which
// column carries the privacy unit and unblocks `SELECT WITH
// ANONYMIZATION` queries. Other options are kept around for INFORMATION_SCHEMA
// but require no further analyzer wiring.
func applyTableOptions(tbl *googlesql.SimpleTable, opts []*tableOptionSpec) {
	for _, o := range opts {
		if o == nil {
			continue
		}
		switch strings.ToLower(o.Name) {
		case "anonymization_userid_column":
			s := stripQuotes(strings.TrimSpace(o.Value))
			if s != "" {
				_ = tbl.SetAnonymizationInfo2(s)
			}
		}
	}
}

func (c *Catalog) createSimpleTable(tableName string, spec *TableSpec) (*googlesql.SimpleTable, error) {
	columns := []*googlesql.SimpleColumn{}
	colIndex := map[string]int32{}
	for i, column := range spec.Columns {
		typ, err := column.Type.ToGoogleSQLType()
		if err != nil {
			return nil, err
		}
		columns = append(columns, m1(googlesql.NewSimpleColumn(
			tableName, column.Name, typ, false, true,
		)))
		colIndex[column.Name] = int32(i)
	}
	tbl := newSimpleTableWithColumns(tableName, columns)
	if tbl != nil && len(spec.PrimaryKey) > 0 {
		// Propagate PRIMARY KEY column ordinals so the analyzer can
		// resolve PropertyGraph node-table primary key requirements.
		// SQLite-side enforcement is unaffected; this is metadata
		// for analysis only.
		pkIdx := make([]int32, 0, len(spec.PrimaryKey))
		for _, name := range spec.PrimaryKey {
			if idx, ok := colIndex[name]; ok {
				pkIdx = append(pkIdx, idx)
			}
		}
		if len(pkIdx) == len(spec.PrimaryKey) {
			_ = tbl.SetPrimaryKey(pkIdx)
		}
	}
	return tbl, nil
}

// newSimpleTableWithColumns constructs a googlesql.SimpleTable with
// `name` and the supplied owned columns. Returns nil if the wasm
// bridge fails so a single trap can't tear down the whole binary.
//
// Each column is added with isOwned=true; v0.2.1 of go-googlesql
// neutralises the column's Go-side finalizer inside AddColumn2 itself,
// so we don't need to follow up with any finalizer suppression here.
func newSimpleTableWithColumns(name string, columns []*googlesql.SimpleColumn) *googlesql.SimpleTable {
	tbl, err := googlesql.NewSimpleTable(name, 0)
	if err != nil {
		return nil
	}
	for _, c := range columns {
		if c == nil {
			continue
		}
		_ = tbl.AddColumn2(c, true)
	}
	return tbl
}

func (c *Catalog) addFunctionSpecRecursive(cat *googlesql.SimpleCatalog, spec *FunctionSpec) error {
	fn, err := c.functionHandleForSpec(spec)
	if err != nil {
		return err
	}
	if err := c.addFunctionSpecRecursiveImpl(cat, spec, fn); err != nil {
		return err
	}
	// Alias the same Function handle under the `bqutil.fn.<name>`
	// namepath so the BigQuery community-UDF dataset prefix resolves
	// even when callers register the bare function (bqe#318). Apply
	// only to functions stored as a single-segment name; multi-segment
	// specs already encode their own dataset prefix.
	if len(spec.NamePath) == 1 {
		bqutilSpec := c.copyFunctionSpec(spec, []string{"bqutil", "fn", spec.NamePath[0]})
		if err := c.addFunctionSpecRecursiveImpl(cat, bqutilSpec, fn); err != nil {
			return err
		}
	}
	return nil
}

// addFunctionSpecRecursiveImpl aliases the same Function handle into
// every (catalog, lookup-name) combination the analyzer might descend
// through. The Function's own Name() stays the storage path; the
// SimpleCatalog.AddFunction2(lookupName, fn) call provides the
// per-level lookup key without needing a new Function handle.
func (c *Catalog) addFunctionSpecRecursiveImpl(cat *googlesql.SimpleCatalog, spec *FunctionSpec, fn *googlesql.Function) error {
	if len(spec.NamePath) > 1 {
		subCatalogName := spec.NamePath[0]
		subCatalog := c.getOrCreateSubCatalog(cat, subCatalogName)
		newNamePath := spec.NamePath[1:]
		// add sub catalog to root catalog
		if err := c.addFunctionSpecRecursiveImpl(cat, c.copyFunctionSpec(spec, newNamePath), fn); err != nil {
			return fmt.Errorf("failed to add function spec to root catalog: %w", err)
		}
		// add sub catalog to parent catalog
		if err := c.addFunctionSpecRecursiveImpl(subCatalog, c.copyFunctionSpec(spec, newNamePath), fn); err != nil {
			return fmt.Errorf("failed to add function spec to parent catalog: %w", err)
		}
		return nil
	}
	if len(spec.NamePath) == 0 {
		return fmt.Errorf("function name is not found")
	}

	funcName := spec.NamePath[0]
	if c.existsFunction(cat, funcName) {
		return nil
	}
	_ = cat.AddFunction2(funcName, fn)
	return nil
}

// functionHandleForSpec builds a Function handle for spec with Name()
// pinned to the storage-format path. One handle per call to
// addFunctionSpecRecursive; aliased into every sub-catalog level via
// AddFunction2. Not cached across calls — same reasoning as
// tableHandleForSpec.
func (c *Catalog) functionHandleForSpec(spec *FunctionSpec) (*googlesql.Function, error) {
	storageName := spec.FuncName()
	argTypes := []*googlesql.FunctionArgumentType{}
	for _, arg := range spec.Args {
		argType, err := arg.FunctionArgumentType()
		if err != nil {
			return nil, err
		}
		argTypes = append(argTypes, argType)
	}
	retType, err := spec.Return.FunctionArgumentType()
	if err != nil {
		return nil, err
	}
	sig := m1(googlesql.NewFunctionSignature3(retType, argTypes, 0))
	fn, err := googlesql.NewFunction([]string{storageName}, "", googlesql.FunctionEnums_ModeScalar, []*googlesql.FunctionSignature{sig}, nil)
	if err != nil {
		return nil, err
	}
	return fn, nil
}

// addTVFSpecRecursive aliases the same TableValuedFunction handle into
// every (catalog, lookup-name) the analyzer might descend through —
// mirroring addFunctionSpecRecursive but for TVFs.
func (c *Catalog) addTVFSpecRecursive(cat *googlesql.SimpleCatalog, spec *TVFSpec) error {
	tvf, err := c.tvfHandleForSpec(spec)
	if err != nil {
		return err
	}
	if tvf == nil {
		return fmt.Errorf("failed to build TVF %q", spec.TVFName())
	}
	return c.addTVFSpecRecursiveImpl(cat, spec, tvf)
}

func (c *Catalog) addTVFSpecRecursiveImpl(cat *googlesql.SimpleCatalog, spec *TVFSpec, tvf *googlesql.TableValuedFunction) error {
	if len(spec.NamePath) > 1 {
		subCatalogName := spec.NamePath[0]
		subCatalog := c.getOrCreateSubCatalog(cat, subCatalogName)
		newNamePath := spec.NamePath[1:]
		if err := c.addTVFSpecRecursiveImpl(cat, c.copyTVFSpec(spec, newNamePath), tvf); err != nil {
			return fmt.Errorf("failed to add TVF spec to root catalog: %w", err)
		}
		if err := c.addTVFSpecRecursiveImpl(subCatalog, c.copyTVFSpec(spec, newNamePath), tvf); err != nil {
			return fmt.Errorf("failed to add TVF spec to parent catalog: %w", err)
		}
		return nil
	}
	if len(spec.NamePath) == 0 {
		return fmt.Errorf("TVF name is not found")
	}
	tvfName := spec.NamePath[0]
	if c.existsTVF(cat, tvfName) {
		return nil
	}
	_ = cat.AddTableValuedFunction2(tvfName, tvf)
	return nil
}

func (c *Catalog) tvfHandleForSpec(spec *TVFSpec) (*googlesql.TableValuedFunction, error) {
	storageName := spec.TVFName()
	argTypes := []*googlesql.FunctionArgumentType{}
	for _, arg := range spec.Args {
		argType, err := arg.FunctionArgumentType()
		if err != nil {
			return nil, err
		}
		argTypes = append(argTypes, argType)
	}
	columns := make([]*googlesql.TVFSchemaColumn, 0, len(spec.OutputColumns))
	for _, col := range spec.OutputColumns {
		typ, err := col.Type.ToGoogleSQLType()
		if err != nil {
			return nil, err
		}
		columns = append(columns, &googlesql.TVFSchemaColumn{
			Name:  col.Name,
			Type_: typ,
		})
	}
	resultSchema, err := googlesql.NewTVFRelation(columns)
	if err != nil {
		return nil, fmt.Errorf("failed to build TVF result schema: %w", err)
	}
	relationArg, err := googlesql.NewFunctionArgumentTypeRelationWithSchema(resultSchema, false)
	if err != nil {
		return nil, fmt.Errorf("failed to build TVF result arg type: %w", err)
	}
	sig, err := googlesql.NewFunctionSignature3(relationArg, argTypes, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to build TVF signature: %w", err)
	}
	tvf, err := googlesql.NewFixedOutputSchemaTVF([]string{storageName}, sig, resultSchema, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build TVF handle: %w", err)
	}
	return tvf.TableValuedFunction, nil
}

func (c *Catalog) copyTVFSpec(spec *TVFSpec, newNamePath []string) *TVFSpec {
	return &TVFSpec{
		IsTemp:        spec.IsTemp,
		NamePath:      newNamePath,
		Args:          spec.Args,
		OutputColumns: spec.OutputColumns,
		Body:          spec.Body,
	}
}

func (c *Catalog) existsTVF(cat *googlesql.SimpleCatalog, name string) bool {
	names, err := cat.TableValuedFunctionNames()
	if err != nil {
		return false
	}
	target := strings.ToLower(name)
	for _, n := range names {
		if strings.ToLower(n) == target {
			return true
		}
	}
	return false
}

func (c *Catalog) existsTable(cat *googlesql.SimpleCatalog, name string) bool {
	// Checks presence on the *wasm-side* SimpleCatalog via TableNames().
	// The earlier implementation short-circuited against c.tableMap,
	// which caused AddTable to be skipped for tables that were already
	// recorded on the Go side but never propagated into the wasm
	// catalog — the analyzer then couldn't find them.
	names, err := cat.TableNames()
	if err != nil {
		return false
	}
	target := strings.ToLower(name)
	for _, n := range names {
		if strings.ToLower(n) == target {
			return true
		}
	}
	return false
}

func (c *Catalog) existsFunction(cat *googlesql.SimpleCatalog, name string) bool {
	names, err := cat.FunctionNames()
	if err != nil {
		return false
	}
	target := strings.ToLower(name)
	for _, n := range names {
		if strings.ToLower(n) == target {
			return true
		}
	}
	return false
}

func (c *Catalog) copyTableSpec(spec *TableSpec, newNamePath []string) *TableSpec {
	return &TableSpec{
		NamePath:   newNamePath,
		Columns:    spec.Columns,
		CreateMode: spec.CreateMode,
		Options:    spec.Options,
	}
}

func (c *Catalog) copyFunctionSpec(spec *FunctionSpec, newNamePath []string) *FunctionSpec {
	return &FunctionSpec{
		NamePath: newNamePath,
		Language: spec.Language,
		Args:     spec.Args,
		Return:   spec.Return,
		Code:     spec.Code,
		Body:     spec.Body,
	}
}
