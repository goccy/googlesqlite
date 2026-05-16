package internal

import (
	"github.com/goccy/googlesqlite/internal/functions/aead"
	arrfn "github.com/goccy/googlesqlite/internal/functions/array"
	"github.com/goccy/googlesqlite/internal/functions/bit"
	"github.com/goccy/googlesqlite/internal/functions/conditional"
	"github.com/goccy/googlesqlite/internal/functions/date"
	"github.com/goccy/googlesqlite/internal/functions/datetime"
	"github.com/goccy/googlesqlite/internal/functions/debugging"
	"github.com/goccy/googlesqlite/internal/functions/extras"
	geofn "github.com/goccy/googlesqlite/internal/functions/geography"
	"github.com/goccy/googlesqlite/internal/functions/hash"
	"github.com/goccy/googlesqlite/internal/functions/hll"
	"github.com/goccy/googlesqlite/internal/functions/interval"
	jsonfn "github.com/goccy/googlesqlite/internal/functions/json"
	"github.com/goccy/googlesqlite/internal/functions/kll"
	"github.com/goccy/googlesqlite/internal/functions/longtail"
	mathfn "github.com/goccy/googlesqlite/internal/functions/math"
	netfn "github.com/goccy/googlesqlite/internal/functions/net"
	numfn "github.com/goccy/googlesqlite/internal/functions/numeric"
	"github.com/goccy/googlesqlite/internal/functions/operator"
	protofn "github.com/goccy/googlesqlite/internal/functions/proto"
	rangefn "github.com/goccy/googlesqlite/internal/functions/rangefn"
	"github.com/goccy/googlesqlite/internal/functions/security"
	spannerfn "github.com/goccy/googlesqlite/internal/functions/spanner"
	str "github.com/goccy/googlesqlite/internal/functions/string"
	"github.com/goccy/googlesqlite/internal/functions/structs"
	textfn "github.com/goccy/googlesqlite/internal/functions/text_analysis"
	"github.com/goccy/googlesqlite/internal/functions/time"
	"github.com/goccy/googlesqlite/internal/functions/timestamp"
	"github.com/goccy/googlesqlite/internal/functions/uuid"

	"github.com/goccy/googlesqlite/internal/functions/helper"

	"github.com/goccy/googlesqlite/internal/value"
)

// normalFuncs lists every scalar / non-aggregate function the
// driver registers as a SQLite UDF. Organized roughly by category
// (arithmetic / comparison / bit; struct + array accessors; type
// system; string; numeric; date / datetime / time / timestamp;
// json; hash; net; geography; helpers).

var normalFuncs = []*funcInfo{
	{Name: "add", BindFunc: helper.Scalar2(mathfn.ADD)},
	{Name: "subtract", BindFunc: helper.Scalar2(mathfn.SUB)},
	{Name: "multiply", BindFunc: helper.Scalar2(mathfn.MUL)},
	{Name: "divide", BindFunc: helper.Scalar2(mathfn.OP_DIV)},
	{Name: "equal", BindFunc: helper.Scalar2(operator.EQ)},
	{Name: "not_equal", BindFunc: helper.Scalar2(operator.NOT_EQ)},
	{Name: "greater", BindFunc: helper.Scalar2(operator.GT)},
	{Name: "greater_or_equal", BindFunc: helper.Scalar2(operator.GTE)},
	{Name: "less", BindFunc: helper.Scalar2(operator.LT)},
	{Name: "less_or_equal", BindFunc: helper.Scalar2(operator.LTE)},
	{Name: "bitwise_not", BindFunc: helper.Scalar1(bit.BIT_NOT)},
	{Name: "bitwise_left_shift", BindFunc: helper.Scalar2(bit.BIT_LEFT_SHIFT)},
	{Name: "bitwise_right_shift", BindFunc: helper.Scalar2(bit.BIT_RIGHT_SHIFT)},
	{Name: "bitwise_and", BindFunc: helper.Scalar2(bit.BIT_AND)},
	{Name: "bitwise_or", BindFunc: helper.Scalar2(bit.BIT_OR)},
	{Name: "bitwise_xor", BindFunc: helper.Scalar2(bit.BIT_XOR)},
	{Name: "bit_cast_to_int64", BindFunc: helper.Scalar1(bit.BIT_CAST_TO_INT64)},
	{Name: "bit_cast_to_uint64", BindFunc: helper.Scalar1(bit.BIT_CAST_TO_UINT64)},
	{Name: "bit_cast_to_int32", BindFunc: helper.Scalar1(bit.BIT_CAST_TO_INT32)},
	{Name: "bit_cast_to_uint32", BindFunc: helper.Scalar1(bit.BIT_CAST_TO_UINT32)},
	{Name: "in_array", BindFunc: arrfn.BindInArray},
	{Name: "array_is_distinct", BindFunc: arrfn.ARRAY_IS_DISTINCT},
	{Name: "get_struct_field", BindFunc: structs.BindStructField},
	{Name: "get_proto_field", BindFunc: protofn.BindGetProtoField},
	{Name: "get_proto_field_repeated", BindFunc: protofn.BindGetProtoFieldRepeated},
	{Name: "from_proto", BindFunc: protofn.BindFromProto},
	{Name: "to_proto", BindFunc: protofn.BindToProto},
	{Name: "make_proto", BindFunc: protofn.BindMakeProto},
	{Name: "filter_fields", BindFunc: protofn.BindFilterFields},
	{Name: "replace_fields", BindFunc: protofn.BindReplaceFields},
	{Name: "proto_map_contains_key", BindFunc: protofn.BindProtoMapContainsKey},
	{Name: "proto_modify_map", BindFunc: protofn.BindProtoModifyMap},
	{Name: "enum_value_descriptor_proto", BindFunc: protofn.BindEnumValueDescriptorProto},
	{Name: "struct_with_field_set", BindFunc: structs.BindStructWithFieldSet},
	{Name: "get_json_field", BindFunc: jsonfn.BindJsonField},
	{Name: "subscript", BindFunc: jsonfn.BindSubscript},
	{Name: "array_at_offset", BindFunc: arrfn.BindArrayAtOffset},
	{Name: "array_at_ordinal", BindFunc: arrfn.BindArrayAtOrdinal},
	{Name: "safe_array_at_offset", BindFunc: arrfn.BindSafeArrayAtOffset},
	{Name: "safe_array_at_ordinal", BindFunc: arrfn.BindSafeArrayAtOrdinal},
	{Name: "is_distinct_from", BindFunc: operator.BindIsDistinctFrom},
	{Name: "is_not_distinct_from", BindFunc: operator.BindIsNotDistinctFrom},

	// security functions
	{Name: "session_user", BindFunc: func(args ...value.Value) (value.Value, error) { return security.SESSION_USER() }},

	// uuid functions
	{
		Name:             "generate_uuid",
		BindFunc:         func(args ...value.Value) (value.Value, error) { return uuid.GENERATE_UUID() },
		NonDeterministic: true,
	},
	{
		// NEW_UUID is GoogleSQL's V1.4 spec-canonical alias for
		// GENERATE_UUID; both return a random UUID v4 STRING.
		Name:             "new_uuid",
		BindFunc:         func(args ...value.Value) (value.Value, error) { return uuid.GENERATE_UUID() },
		NonDeterministic: true,
	},

	// debugging functions
	{Name: "error", BindFunc: debugging.BindError},

	// date functions
	{Name: "current_date", BindFunc: date.BindCurrentDate},
	{Name: "extract", BindFunc: date.BindExtract},
	{Name: "extract_date", BindFunc: date.BindExtractDate},
	{Name: "date", BindFunc: date.BindDate},
	{Name: "date_add", BindFunc: date.BindDateAdd},
	{Name: "date_sub", BindFunc: date.BindDateSub},
	{Name: "date_diff", BindFunc: date.BindDateDiff},
	{Name: "date_trunc", BindFunc: date.BindDateTrunc},
	{Name: "date_from_unix_date", BindFunc: date.BindDateFromUnixDate},
	{Name: "format_date", BindFunc: date.BindFormatDate},
	{Name: "last_day", BindFunc: date.BindLastDay},
	{Name: "parse_date", BindFunc: date.BindParseDate},
	{Name: "unix_date", BindFunc: date.BindUnixDate},
	{Name: "date_bucket", BindFunc: date.BindDateBucket},

	// datetime functions
	{Name: "current_datetime", BindFunc: datetime.BindCurrentDatetime},
	{Name: "datetime", BindFunc: datetime.BindDatetime},
	{Name: "datetime_add", BindFunc: datetime.BindDatetimeAdd},
	{Name: "datetime_sub", BindFunc: datetime.BindDatetimeSub},
	{Name: "datetime_diff", BindFunc: datetime.BindDatetimeDiff},
	{Name: "datetime_trunc", BindFunc: datetime.BindDatetimeTrunc},
	{Name: "format_datetime", BindFunc: datetime.BindFormatDatetime},
	{Name: "parse_datetime", BindFunc: datetime.BindParseDatetime},
	{Name: "datetime_bucket", BindFunc: datetime.BindDatetimeBucket},

	// time functions
	{Name: "current_time", BindFunc: time.BindCurrentTime},
	{Name: "time", BindFunc: time.BindTime},
	{Name: "time_add", BindFunc: time.BindTimeAdd},
	{Name: "time_sub", BindFunc: time.BindTimeSub},
	{Name: "time_diff", BindFunc: time.BindTimeDiff},
	{Name: "time_trunc", BindFunc: time.BindTimeTrunc},
	{Name: "format_time", BindFunc: time.BindFormatTime},
	{Name: "parse_time", BindFunc: time.BindParseTime},

	// timestamp functions
	{Name: "current_timestamp", BindFunc: timestamp.BindCurrentTimestamp},
	{Name: "string", BindFunc: timestamp.BindString},
	{Name: "timestamp", BindFunc: timestamp.BindTimestamp},
	{Name: "timestamp_add", BindFunc: timestamp.BindTimestampAdd},
	{Name: "timestamp_sub", BindFunc: timestamp.BindTimestampSub},
	{Name: "timestamp_diff", BindFunc: timestamp.BindTimestampDiff},
	{Name: "timestamp_trunc", BindFunc: timestamp.BindTimestampTrunc},
	{Name: "format_timestamp", BindFunc: timestamp.BindFormatTimestamp},
	{Name: "parse_timestamp", BindFunc: timestamp.BindParseTimestamp},
	{Name: "timestamp_seconds", BindFunc: timestamp.BindTimestampSeconds},
	{Name: "timestamp_millis", BindFunc: timestamp.BindTimestampMillis},
	{Name: "timestamp_micros", BindFunc: timestamp.BindTimestampMicros},
	{Name: "timestamp_from_unix_seconds", BindFunc: timestamp.BindTimestampSeconds},
	{Name: "timestamp_from_unix_millis", BindFunc: timestamp.BindTimestampMillis},
	{Name: "timestamp_from_unix_micros", BindFunc: timestamp.BindTimestampMicros},
	{Name: "unix_seconds", BindFunc: timestamp.BindUnixSeconds},
	{Name: "unix_millis", BindFunc: timestamp.BindUnixMillis},
	{Name: "unix_micros", BindFunc: timestamp.BindUnixMicros},
	{Name: "timestamp_bucket", BindFunc: timestamp.BindTimestampBucket},
	{Name: "like", BindFunc: str.BindLike},
	{Name: "between", BindFunc: operator.BindBetween},
	{Name: "in", BindFunc: operator.BindIn},
	{Name: "is_null", BindFunc: operator.BindIsNull},
	{Name: "is_true", BindFunc: helper.Scalar1(operator.IS_TRUE)},
	{Name: "is_false", BindFunc: helper.Scalar1(operator.IS_FALSE)},
	{Name: "not", BindFunc: helper.Scalar1(operator.NOT)},
	{Name: "and", BindFunc: operator.BindAnd},
	{Name: "or", BindFunc: operator.BindOr},
	{Name: "coalesce", BindFunc: conditional.BindCoalesce},
	{Name: "if", BindFunc: conditional.BindIf},
	{Name: "ifnull", BindFunc: conditional.BindIfNull},
	{Name: "nullif", BindFunc: conditional.BindNullIf},
	{Name: "length", BindFunc: str.BindLength},
	{Name: "cast", BindFunc: bindCast},

	// interval functions
	{Name: "interval", BindFunc: interval.BindInterval},
	{Name: "make_interval", BindFunc: interval.BindMakeInterval},
	{Name: "justify_days", BindFunc: interval.BindJustifyDays},
	{Name: "justify_hours", BindFunc: interval.BindJustifyHours},
	{Name: "justify_interval", BindFunc: interval.BindJustifyInterval},

	// geography functions
	{Name: "st_geogpoint", BindFunc: geofn.BindStGeogPoint},
	{Name: "st_geogfromtext", BindFunc: geofn.BindStGeogFromText},
	{Name: "st_distance", BindFunc: geofn.BindStDistance},
	{Name: "st_x", BindFunc: geofn.BindStX},
	{Name: "st_y", BindFunc: geofn.BindStY},
	{Name: "st_astext", BindFunc: geofn.BindStAsText},
	{Name: "st_equals", BindFunc: geofn.BindStEquals},
	{Name: "st_dwithin", BindFunc: geofn.BindStDWithin},
	{Name: "st_intersects", BindFunc: geofn.BindStIntersects},
	// Measurements + accessors.
	{Name: "st_length", BindFunc: geofn.BindStLength},
	{Name: "st_perimeter", BindFunc: geofn.BindStPerimeter},
	{Name: "st_area", BindFunc: geofn.BindStArea},
	{Name: "st_numpoints", BindFunc: geofn.BindStNumPoints},
	{Name: "st_numgeometries", BindFunc: geofn.BindStNumGeometries},
	{Name: "st_dimension", BindFunc: geofn.BindStDimension},
	{Name: "st_geometrytype", BindFunc: geofn.BindStGeometryType},
	{Name: "st_isempty", BindFunc: geofn.BindStIsEmpty},
	{Name: "st_iscollection", BindFunc: geofn.BindStIsCollection},
	{Name: "st_startpoint", BindFunc: geofn.BindStStartPoint},
	{Name: "st_endpoint", BindFunc: geofn.BindStEndPoint},
	{Name: "st_pointn", BindFunc: geofn.BindStPointN},
	{Name: "st_geometryn", BindFunc: geofn.BindStGeometryN},
	{Name: "st_boundary", BindFunc: geofn.BindStBoundary},
	// Topology.
	{Name: "st_disjoint", BindFunc: geofn.BindStDisjoint},
	{Name: "st_contains", BindFunc: geofn.BindStContains},
	{Name: "st_within", BindFunc: geofn.BindStWithin},
	{Name: "st_covers", BindFunc: geofn.BindStCovers},
	{Name: "st_coveredby", BindFunc: geofn.BindStCoveredBy},
	{Name: "st_touches", BindFunc: geofn.BindStTouches},
	{Name: "st_maxdistance", BindFunc: geofn.BindStMaxDistance},
	// Serialization.
	{Name: "st_asgeojson", BindFunc: geofn.BindStAsGeoJSON},
	{Name: "st_geogfromgeojson", BindFunc: geofn.BindStGeogFromGeoJSON},
	{Name: "st_asbinary", BindFunc: geofn.BindStAsBinary},
	{Name: "st_geogfromwkb", BindFunc: geofn.BindStGeogFromWKB},
	// Constructors.
	{Name: "st_makeline", BindFunc: geofn.BindStMakeLine},
	{Name: "st_makepolygon", BindFunc: geofn.BindStMakePolygon},
	{Name: "st_makepolygonoriented", BindFunc: geofn.BindStMakePolygonOriented},
	{Name: "st_centroid", BindFunc: geofn.BindStCentroid},
	{Name: "st_envelope", BindFunc: geofn.BindStEnvelope},
	{Name: "st_boundingbox", BindFunc: geofn.BindStBoundingBox},
	// Long-tail geography functions.
	{Name: "st_npoints", BindFunc: geofn.BindStNPoints},
	{Name: "st_exteriorring", BindFunc: geofn.BindStExteriorRing},
	{Name: "st_interiorrings", BindFunc: geofn.BindStInteriorRings},
	{Name: "st_dump", BindFunc: geofn.BindStDump},
	{Name: "st_dumppoints", BindFunc: geofn.BindStDumpPoints},
	{Name: "st_isclosed", BindFunc: geofn.BindStIsClosed},
	{Name: "st_isring", BindFunc: geofn.BindStIsRing},
	{Name: "st_azimuth", BindFunc: geofn.BindStAzimuth},
	{Name: "st_angle", BindFunc: geofn.BindStAngle},
	{Name: "st_hausdorffdistance", BindFunc: geofn.BindStHausdorffDistance},
	{Name: "st_hausdorffdwithin", BindFunc: geofn.BindStHausdorffDWithin},
	{Name: "s2_cellidfrompoint", BindFunc: geofn.BindS2CellIDFromPoint},
	{Name: "s2_coveringcellids", BindFunc: geofn.BindS2CoveringCellIDs},
	{Name: "st_closestpoint", BindFunc: geofn.BindStClosestPoint},
	{Name: "st_lineinterpolatepoint", BindFunc: geofn.BindStLineInterpolatePoint},
	{Name: "st_linelocatepoint", BindFunc: geofn.BindStLineLocatePoint},
	{Name: "st_linesubstring", BindFunc: geofn.BindStLineSubstring},
	{Name: "st_geohash", BindFunc: geofn.BindStGeoHash},
	{Name: "st_geogpointfromgeohash", BindFunc: geofn.BindStGeogPointFromGeoHash},
	{Name: "st_askml", BindFunc: geofn.BindStAsKML},
	{Name: "st_geogfromkml", BindFunc: geofn.BindStGeogFromKML},
	{Name: "st_geogfrom", BindFunc: geofn.BindStGeogFrom},
	// st_extent: aggregate form moved to function_register_aggregate.go.
	// Scalar variadic BindStExtent is still defined for the rare
	// callers that invoke it as a fixed-arity scalar.
	{Name: "st_snaptogrid", BindFunc: geofn.BindStSnapToGrid},
	{Name: "st_intersectsbox", BindFunc: geofn.BindStIntersectsBox},
	// Polygon set-ops (planar approximations).
	{Name: "st_union", BindFunc: geofn.BindStUnion},
	{Name: "st_intersection", BindFunc: geofn.BindStIntersection},
	{Name: "st_difference", BindFunc: geofn.BindStDifference},
	{Name: "st_buffer", BindFunc: geofn.BindStBuffer},
	{Name: "st_bufferwithtolerance", BindFunc: geofn.BindStBufferWithTolerance},
	{Name: "st_convexhull", BindFunc: geofn.BindStConvexHull},
	{Name: "st_simplify", BindFunc: geofn.BindStSimplify},
	// st_clusterdbscan is registered as a window aggregator instead
	// (see windowFuncs); the analytic emulation path reaches it as
	// `(SELECT st_clusterdbscan(...) FROM input)` per outer row, which
	// requires aggregate semantics to fold the inner scan into a
	// single value.
	// Raster.
	{Name: "st_regionstats", BindFunc: geofn.BindStRegionStats},

	// range functions (TYPE_RANGE = 29 in googlesql/public/type.proto,
	// gated by FEATURE_RANGE_TYPE in NewLanguageOptions).
	{Name: "range", BindFunc: rangefn.BindRange},
	{Name: "range_start", BindFunc: rangefn.BindRangeStart},
	{Name: "range_end", BindFunc: rangefn.BindRangeEnd},
	{Name: "range_is_start_unbounded", BindFunc: rangefn.BindRangeIsStartUnbounded},
	{Name: "range_is_end_unbounded", BindFunc: rangefn.BindRangeIsEndUnbounded},
	{Name: "range_contains", BindFunc: rangefn.BindRangeContains},
	{Name: "range_overlaps", BindFunc: rangefn.BindRangeOverlaps},
	{Name: "range_intersect", BindFunc: rangefn.BindRangeIntersect},
	{Name: "generate_range_array", BindFunc: rangefn.BindGenerateRangeArray},
	{Name: "range_sessionize", BindFunc: rangefn.BindRangeSessionize},

	// numeric/bignumeric functions
	{Name: "parse_numeric", BindFunc: numfn.BindParseNumeric},
	{Name: "parse_bignumeric", BindFunc: numfn.BindParseBigNumeric},

	// hash functions
	{Name: "farm_fingerprint", BindFunc: helper.Scalar1(hash.FarmFingerprintBind)},
	{Name: "md5", BindFunc: helper.Scalar1(hash.MD5Bind)},
	{Name: "sha1", BindFunc: helper.Scalar1(hash.SHA1Bind)},
	{Name: "sha256", BindFunc: helper.Scalar1(hash.SHA256Bind)},
	{Name: "sha512", BindFunc: helper.Scalar1(hash.SHA512Bind)},

	// string functions
	{Name: "ascii", BindFunc: str.BindAscii},
	{Name: "byte_length", BindFunc: str.BindByteLength},
	{Name: "char_length", BindFunc: str.BindCharLength},
	{Name: "chr", BindFunc: str.BindChr},
	{Name: "code_points_to_bytes", BindFunc: str.BindCodePointsToBytes},
	{Name: "code_points_to_string", BindFunc: str.BindCodePointsToString},
	{Name: "collate", BindFunc: str.BindCollate},
	{Name: "concat", BindFunc: str.BindConcat},
	{Name: "contains_substr", BindFunc: str.BindContainsSubstr},
	{Name: "edit_distance", BindFunc: str.BindEditDistance},
	{Name: "ends_with", BindFunc: str.BindEndsWith},
	{Name: "format", BindFunc: str.BindFormat},
	{Name: "from_base32", BindFunc: str.BindFromBase32},
	{Name: "from_base64", BindFunc: str.BindFromBase64},
	{Name: "from_hex", BindFunc: str.BindFromHex},
	{Name: "initcap", BindFunc: str.BindInitcap},
	{Name: "instr", BindFunc: str.BindInstr},
	{Name: "left", BindFunc: str.BindLeft},
	{Name: "length", BindFunc: str.BindLength},
	{Name: "lpad", BindFunc: str.BindLpad},
	{Name: "lower", BindFunc: str.BindLower},
	{Name: "ltrim", BindFunc: str.BindLtrim},
	{Name: "normalize", BindFunc: str.BindNormalize},
	{Name: "normalize_and_casefold", BindFunc: str.BindNormalizeAndCasefold},
	{Name: "regexp_contains", BindFunc: str.BindRegexpContains},
	{Name: "regexp_extract", BindFunc: str.BindRegexpExtract},
	{Name: "regexp_extract_all", BindFunc: str.BindRegexpExtractAll},
	{Name: "regexp_instr", BindFunc: str.BindRegexpInstr},
	{Name: "regexp_replace", BindFunc: helper.Scalar3(str.REGEXP_REPLACE)},
	{Name: "replace", BindFunc: helper.Scalar3(str.REPLACE)},
	{Name: "repeat", BindFunc: str.BindRepeat},
	{Name: "reverse", BindFunc: helper.Scalar1(str.REVERSE)},
	{Name: "right", BindFunc: str.BindRight},
	{Name: "rpad", BindFunc: str.BindRpad},
	{Name: "rtrim", BindFunc: str.BindRtrim},
	{Name: "safe_convert_bytes_to_string", BindFunc: str.BindSafeConvertBytesToString},
	{Name: "soundex", BindFunc: str.BindSoundex},
	{Name: "split", BindFunc: str.BindSplit},
	{Name: "starts_with", BindFunc: str.BindStartsWith},
	{Name: "strpos", BindFunc: str.BindStrpos},
	{Name: "substr", BindFunc: str.BindSubstr},
	{Name: "to_base32", BindFunc: str.BindToBase32},
	{Name: "to_base64", BindFunc: str.BindToBase64},
	{Name: "to_code_points", BindFunc: str.BindToCodePoints},
	{Name: "to_hex", BindFunc: str.BindToHex},
	{Name: "translate", BindFunc: str.BindTranslate},
	{Name: "trim", BindFunc: str.BindTrim},
	{Name: "unicode", BindFunc: str.BindUnicode},
	{Name: "upper", BindFunc: str.BindUpper},

	// json functions
	{Name: "json_extract", BindFunc: jsonfn.BindJsonExtract},
	{Name: "json_extract_scalar", BindFunc: jsonfn.BindJsonExtractScalar},
	{Name: "json_extract_array", BindFunc: jsonfn.BindJsonExtractArray},
	{Name: "json_extract_string_array", BindFunc: jsonfn.BindJsonExtractStringArray},
	{Name: "json_query", BindFunc: jsonfn.BindJsonQuery},
	{Name: "json_value", BindFunc: jsonfn.BindJsonValue},
	{Name: "lax_int64", BindFunc: jsonfn.BindLaxInt64},
	{Name: "lax_float64", BindFunc: jsonfn.BindLaxFloat64},
	{Name: "lax_bool", BindFunc: jsonfn.BindLaxBool},
	{Name: "lax_string", BindFunc: jsonfn.BindLaxString},
	{Name: "json_query_array", BindFunc: jsonfn.BindJsonQueryArray},
	{Name: "json_value_array", BindFunc: jsonfn.BindJsonValueArray},
	{Name: "parse_json", BindFunc: jsonfn.BindParseJson},
	{Name: "to_json", BindFunc: jsonfn.BindToJson},
	{Name: "to_json_string", BindFunc: jsonfn.BindToJsonString},
	{Name: "bool", BindFunc: bindBool},
	{Name: "int64", BindFunc: bindInt64},
	{Name: "double", BindFunc: bindDouble},
	{Name: "float64", BindFunc: bindDouble},
	{Name: "json_type", BindFunc: jsonfn.BindJsonType},
	{Name: "json_strip_nulls", BindFunc: jsonfn.BindJsonStripNulls},
	{Name: "json_remove", BindFunc: jsonfn.BindJsonRemove},
	{Name: "json_set", BindFunc: jsonfn.BindJsonSet},
	{Name: "json_keys", BindFunc: jsonfn.BindJsonKeys},
	{Name: "json_array", BindFunc: jsonfn.JSON_ARRAY},
	{Name: "json_object", BindFunc: jsonfn.JSON_OBJECT},
	{Name: "json_contains", BindFunc: jsonfn.JSON_CONTAINS},
	{Name: "json_flatten", BindFunc: jsonfn.JSON_FLATTEN},
	{Name: "json_path_exists", BindFunc: jsonfn.JSON_PATH_EXISTS},

	// JSON strict scalar accessors.
	{Name: "int32", BindFunc: jsonfn.BindInt32},
	{Name: "uint32", BindFunc: jsonfn.BindUint32},
	{Name: "uint64", BindFunc: jsonfn.BindUint64},
	{Name: "float32", BindFunc: jsonfn.BindFloat},

	// JSON LAX scalar accessors (extending the existing set).
	{Name: "lax_int32", BindFunc: jsonfn.BindLaxInt32},
	{Name: "lax_uint32", BindFunc: jsonfn.BindLaxUint32},
	{Name: "lax_uint64", BindFunc: jsonfn.BindLaxUint64},
	{Name: "lax_float32", BindFunc: jsonfn.BindLaxFloat},

	// JSON strict array accessors.
	{Name: "bool_array", BindFunc: jsonfn.BindBoolArray},
	{Name: "int32_array", BindFunc: jsonfn.BindInt32Array},
	{Name: "int64_array", BindFunc: jsonfn.BindInt64Array},
	{Name: "uint32_array", BindFunc: jsonfn.BindUint32Array},
	{Name: "uint64_array", BindFunc: jsonfn.BindUint64Array},
	{Name: "float32_array", BindFunc: jsonfn.BindFloatArray},
	{Name: "double_array", BindFunc: jsonfn.BindDoubleArray},
	{Name: "float64_array", BindFunc: jsonfn.BindDoubleArray},
	{Name: "string_array", BindFunc: jsonfn.BindStringArray},

	// JSON LAX array accessors.
	{Name: "lax_bool_array", BindFunc: jsonfn.BindLaxBoolArray},
	{Name: "lax_int32_array", BindFunc: jsonfn.BindLaxInt32Array},
	{Name: "lax_int64_array", BindFunc: jsonfn.BindLaxInt64Array},
	{Name: "lax_uint32_array", BindFunc: jsonfn.BindLaxUint32Array},
	{Name: "lax_uint64_array", BindFunc: jsonfn.BindLaxUint64Array},
	{Name: "lax_float32_array", BindFunc: jsonfn.BindLaxFloatArray},
	{Name: "lax_double_array", BindFunc: jsonfn.BindLaxDoubleArray},
	{Name: "lax_float64_array", BindFunc: jsonfn.BindLaxDoubleArray},
	{Name: "lax_string_array", BindFunc: jsonfn.BindLaxStringArray},

	// KLL quantile scalar extractors (the aggregate
	// variants live in function_register_aggregate.go).
	{Name: "kll_quantiles_extract_int64", BindFunc: kll.BindExtractInt64},
	{Name: "kll_quantiles_extract_float64", BindFunc: kll.BindExtractFloat64},
	{Name: "kll_quantiles_extract_point_int64", BindFunc: kll.BindExtractPointInt64},
	{Name: "kll_quantiles_extract_point_float64", BindFunc: kll.BindExtractPointFloat64},

	// Spanner mysql.* aliases. The analyzer recognises the
	// dotted path through the `mysql` sub-catalog (registered in
	// registerSpannerExtensionFunctions); the formatter
	// flattens "mysql.<name>" to "mysql_<name>" so these UDFs
	// are the runtime targets. Each one simply delegates to the
	// equivalent GoogleSQL bind function.
	{Name: "mysql_length", BindFunc: str.BindLength},
	{Name: "mysql_char_length", BindFunc: str.BindCharLength},
	{Name: "mysql_lower", BindFunc: str.BindLower},
	{Name: "mysql_upper", BindFunc: str.BindUpper},
	{Name: "mysql_trim", BindFunc: str.BindTrim},
	{Name: "mysql_ltrim", BindFunc: str.BindLtrim},
	{Name: "mysql_rtrim", BindFunc: str.BindRtrim},
	{Name: "mysql_concat", BindFunc: str.BindConcat},
	{Name: "mysql_reverse", BindFunc: helper.Scalar1(str.REVERSE)},
	{Name: "mysql_md5", BindFunc: helper.Scalar1(hash.MD5Bind)},
	{Name: "mysql_sha1", BindFunc: helper.Scalar1(hash.SHA1Bind)},
	{Name: "mysql_abs", BindFunc: helper.Scalar1(mathfn.ABS)},
	{Name: "mysql_floor", BindFunc: helper.Scalar1(mathfn.FLOOR)},
	{Name: "mysql_ceil", BindFunc: helper.Scalar1(mathfn.CEIL)},
	{Name: "mysql_ceiling", BindFunc: helper.Scalar1(mathfn.CEIL)},
	{Name: "mysql_round", BindFunc: mathfn.BindRound},
	{Name: "mysql_coalesce", BindFunc: conditional.BindCoalesce},
	{Name: "mysql_curdate", BindFunc: date.BindCurrentDate},
	{Name: "mysql_current_date", BindFunc: date.BindCurrentDate},
	{Name: "mysql_now", BindFunc: timestamp.BindCurrentTimestamp},
	{Name: "mysql_current_timestamp", BindFunc: timestamp.BindCurrentTimestamp},
	{Name: "mysql_unix_timestamp", BindFunc: timestamp.BindMysqlUnixTimestamp},
	// mysql_datetime extras (EXTRACT-like helpers).
	{Name: "mysql_day", BindFunc: spannerfn.BindDay},
	{Name: "mysql_month", BindFunc: spannerfn.BindMonth},
	{Name: "mysql_year", BindFunc: spannerfn.BindYear},
	{Name: "mysql_hour", BindFunc: spannerfn.BindHour},
	{Name: "mysql_minute", BindFunc: spannerfn.BindMinute},
	{Name: "mysql_second", BindFunc: spannerfn.BindSecond},
	{Name: "mysql_microsecond", BindFunc: spannerfn.BindMicrosecond},
	{Name: "mysql_quarter", BindFunc: spannerfn.BindQuarter},
	{Name: "mysql_dayofweek", BindFunc: spannerfn.BindDayOfWeek},
	{Name: "mysql_dayofyear", BindFunc: spannerfn.BindDayOfYear},
	{Name: "mysql_dayofmonth", BindFunc: spannerfn.BindDayOfMonth},
	{Name: "mysql_week", BindFunc: spannerfn.BindWeek},
	{Name: "mysql_weekofyear", BindFunc: spannerfn.BindWeekOfYear},
	{Name: "mysql_weekday", BindFunc: spannerfn.BindWeekday},
	// mysql_string extras.
	{Name: "mysql_bit_length", BindFunc: spannerfn.BindBitLength},
	{Name: "mysql_hex", BindFunc: spannerfn.BindHex},
	{Name: "mysql_space", BindFunc: spannerfn.BindSpace},
	{Name: "mysql_position", BindFunc: spannerfn.BindPosition},
	// mysql_numeric extras.
	{Name: "mysql_degrees", BindFunc: spannerfn.BindDegrees},
	{Name: "mysql_radians", BindFunc: spannerfn.BindRadians},
	{Name: "mysql_log2", BindFunc: spannerfn.BindLog2},
	{Name: "mysql_truncate", BindFunc: spannerfn.BindTruncate},
	// More mysql_string aliases.
	{Name: "mysql_char", BindFunc: spannerfn.BindChar},
	{Name: "mysql_concat_ws", BindFunc: spannerfn.BindConcatWs},
	{Name: "mysql_insert", BindFunc: spannerfn.BindInsert},
	{Name: "mysql_locate", BindFunc: spannerfn.BindLocate},
	{Name: "mysql_mid", BindFunc: spannerfn.BindMid},
	{Name: "mysql_oct", BindFunc: spannerfn.BindOct},
	{Name: "mysql_ord", BindFunc: spannerfn.BindOrd},
	{Name: "mysql_quote", BindFunc: spannerfn.BindQuote},
	{Name: "mysql_regexp_like", BindFunc: spannerfn.BindRegexpLike},
	{Name: "mysql_strcmp", BindFunc: spannerfn.BindStrcmp},
	{Name: "mysql_substring_index", BindFunc: spannerfn.BindSubstringIndex},
	{Name: "mysql_unhex", BindFunc: spannerfn.BindUnhex},
	// More mysql_datetime aliases.
	{Name: "mysql_dayname", BindFunc: spannerfn.BindDayName},
	{Name: "mysql_monthname", BindFunc: spannerfn.BindMonthName},
	{Name: "mysql_from_days", BindFunc: spannerfn.BindFromDays},
	{Name: "mysql_to_days", BindFunc: spannerfn.BindToDays},
	{Name: "mysql_to_seconds", BindFunc: spannerfn.BindToSeconds},
	{Name: "mysql_makedate", BindFunc: spannerfn.BindMakeDate},
	{Name: "mysql_from_unixtime", BindFunc: spannerfn.BindFromUnixtime},
	{Name: "mysql_sysdate", BindFunc: spannerfn.BindSysDate},
	{Name: "mysql_utc_date", BindFunc: spannerfn.BindUtcDate},
	{Name: "mysql_utc_timestamp", BindFunc: spannerfn.BindUtcTimestamp},
	{Name: "mysql_period_add", BindFunc: spannerfn.BindPeriodAdd},
	{Name: "mysql_period_diff", BindFunc: spannerfn.BindPeriodDiff},
	{Name: "mysql_date_format", BindFunc: spannerfn.BindDateFormat},
	{Name: "mysql_str_to_date", BindFunc: spannerfn.BindStrToDate},
	// mysql_timestamp.
	{Name: "mysql_datediff", BindFunc: spannerfn.BindDateDiff},
	{Name: "mysql_localtime", BindFunc: timestamp.BindCurrentTimestamp},
	{Name: "mysql_localtimestamp", BindFunc: timestamp.BindCurrentTimestamp},
	// mysql_utility.
	{Name: "mysql_inet_aton", BindFunc: spannerfn.BindInetAton},
	{Name: "mysql_inet_ntoa", BindFunc: spannerfn.BindInetNtoa},
	{Name: "mysql_inet6_aton", BindFunc: spannerfn.BindInet6Aton},
	{Name: "mysql_inet6_ntoa", BindFunc: spannerfn.BindInet6Ntoa},
	{Name: "mysql_is_ipv4", BindFunc: spannerfn.BindIsIPv4},
	{Name: "mysql_is_ipv6", BindFunc: spannerfn.BindIsIPv6},
	{Name: "mysql_is_ipv4_compat", BindFunc: spannerfn.BindIsIPv4Compat},
	{Name: "mysql_is_ipv4_mapped", BindFunc: spannerfn.BindIsIPv4Mapped},
	{Name: "mysql_is_uuid", BindFunc: spannerfn.BindIsUUID},
	{Name: "mysql_bin_to_uuid", BindFunc: spannerfn.BindBinToUUID},
	{Name: "mysql_uuid_to_bin", BindFunc: spannerfn.BindUUIDToBin},
	// mysql_encryption.
	{Name: "mysql_sha2", BindFunc: spannerfn.BindSHA2},
	// mysql_json.
	{Name: "mysql_json_quote", BindFunc: spannerfn.BindJsonQuote},
	{Name: "mysql_json_unquote", BindFunc: spannerfn.BindJsonUnquote},

	// Spanner search functions.
	{Name: "tokenize_fulltext", BindFunc: spannerfn.BindSpannerTokenizeFullText},
	{Name: "tokenize_substring", BindFunc: spannerfn.BindSpannerTokenizeSubstring},
	{Name: "tokenize_ngrams", BindFunc: spannerfn.BindSpannerTokenizeNGrams},
	{Name: "tokenize_number", BindFunc: spannerfn.BindSpannerTokenizeNumber},
	{Name: "tokenize_bool", BindFunc: spannerfn.BindSpannerTokenizeBool},
	{Name: "tokenize_json", BindFunc: spannerfn.BindSpannerTokenizeJSON},
	{Name: "token", BindFunc: spannerfn.BindSpannerToken},
	{Name: "tokenlist_concat", BindFunc: spannerfn.BindSpannerTokenListConcat},
	{Name: "search_substring", BindFunc: spannerfn.BindSpannerSearchSubstring},
	{Name: "search_ngrams", BindFunc: spannerfn.BindSpannerSearchNGrams},
	{Name: "score", BindFunc: spannerfn.BindSpannerScore},
	{Name: "score_ngrams", BindFunc: spannerfn.BindSpannerScoreNGrams},
	{Name: "snippet", BindFunc: spannerfn.BindSpannerSnippet},
	{Name: "debug_tokenlist", BindFunc: spannerfn.BindSpannerDebugTokenList},

	// text_analysis functions
	{Name: "text_analyze", BindFunc: textfn.BindTextAnalyze},
	{Name: "bag_of_words", BindFunc: textfn.BindBagOfWords},
	{Name: "search", BindFunc: textfn.BindSearch},

	// math functions

	{Name: "abs", BindFunc: helper.Scalar1(mathfn.ABS)},
	{Name: "sign", BindFunc: helper.Scalar1(mathfn.SIGN)},
	{Name: "is_inf", BindFunc: helper.Scalar1(mathfn.IS_INF)},
	{Name: "is_nan", BindFunc: helper.Scalar1(mathfn.IS_NAN)},
	{Name: "ieee_divide", BindFunc: helper.Scalar2(mathfn.IEEE_DIVIDE)},
	{Name: "rand", BindFunc: mathfn.BindRand},
	{Name: "sqrt", BindFunc: helper.Scalar1(mathfn.SQRT)},
	{Name: "pow", BindFunc: helper.Scalar2(mathfn.POW)},
	{Name: "power", BindFunc: helper.Scalar2(mathfn.POW)},
	{Name: "exp", BindFunc: helper.Scalar1(mathfn.EXP)},
	{Name: "ln", BindFunc: helper.Scalar1(mathfn.LN)},
	{Name: "log", BindFunc: mathfn.BindLog},
	{Name: "log10", BindFunc: helper.Scalar1(mathfn.LOG10)},
	{Name: "greatest", BindFunc: mathfn.BindGreatest},
	{Name: "least", BindFunc: mathfn.BindLeast},
	{Name: "div", BindFunc: helper.Scalar2(mathfn.DIV)},
	{Name: "safe_divide", BindFunc: helper.Scalar2(mathfn.SAFE_DIVIDE)},
	{Name: "safe_multiply", BindFunc: helper.Scalar2(mathfn.SAFE_MULTIPLY)},
	{Name: "safe_negate", BindFunc: helper.Scalar1(mathfn.SAFE_NEGATE)},
	{Name: "safe_add", BindFunc: helper.Scalar2(mathfn.SAFE_ADD)},
	{Name: "safe_subtract", BindFunc: helper.Scalar2(mathfn.SAFE_SUBTRACT)},
	{Name: "mod", BindFunc: helper.Scalar2(mathfn.MOD)},
	{Name: "round", BindFunc: mathfn.BindRound},
	{Name: "trunc", BindFunc: helper.Scalar1(mathfn.TRUNC)},
	{Name: "ceil", BindFunc: helper.Scalar1(mathfn.CEIL)},
	{Name: "ceiling", BindFunc: helper.Scalar1(mathfn.CEIL)},
	{Name: "floor", BindFunc: helper.Scalar1(mathfn.FLOOR)},
	{Name: "cos", BindFunc: helper.Scalar1(mathfn.COS)},
	{Name: "cosh", BindFunc: helper.Scalar1(mathfn.COSH)},
	{Name: "acos", BindFunc: helper.Scalar1(mathfn.ACOS)},
	{Name: "acosh", BindFunc: helper.Scalar1(mathfn.ACOSH)},
	{Name: "sin", BindFunc: helper.Scalar1(mathfn.SIN)},
	{Name: "sinh", BindFunc: helper.Scalar1(mathfn.SINH)},
	{Name: "asin", BindFunc: helper.Scalar1(mathfn.ASIN)},
	{Name: "asinh", BindFunc: helper.Scalar1(mathfn.ASINH)},
	{Name: "tan", BindFunc: helper.Scalar1(mathfn.TAN)},
	{Name: "tanh", BindFunc: helper.Scalar1(mathfn.TANH)},
	{Name: "atan", BindFunc: helper.Scalar1(mathfn.ATAN)},
	{Name: "cot", BindFunc: helper.Scalar1(mathfn.COT)},
	{Name: "sec", BindFunc: helper.Scalar1(mathfn.SEC)},
	{Name: "csc", BindFunc: helper.Scalar1(mathfn.CSC)},
	{Name: "coth", BindFunc: helper.Scalar1(mathfn.COTH)},
	{Name: "sech", BindFunc: helper.Scalar1(mathfn.SECH)},
	{Name: "csch", BindFunc: helper.Scalar1(mathfn.CSCH)},
	{Name: "cbrt", BindFunc: helper.Scalar1(mathfn.CBRT)},
	{Name: "euclidean_distance", BindFunc: mathfn.EUCLIDEAN_DISTANCE},
	{Name: "cosine_distance", BindFunc: mathfn.COSINE_DISTANCE},
	{Name: "atanh", BindFunc: helper.Scalar1(mathfn.ATANH)},
	{Name: "atan2", BindFunc: helper.Scalar2(mathfn.ATAN2)},
	{Name: "range_bucket", BindFunc: arrfn.BindRangeBucket},

	// array functions
	{Name: "array_concat", BindFunc: arrfn.BindArrayConcat},
	{Name: "array_length", BindFunc: arrfn.BindArrayLength},
	{Name: "array_to_string", BindFunc: arrfn.BindArrayToString},
	{Name: "generate_array", BindFunc: arrfn.BindGenerateArray},
	{Name: "generate_date_array", BindFunc: date.BindGenerateDateArray},
	{Name: "generate_timestamp_array", BindFunc: timestamp.BindGenerateTimestampArray},
	{Name: "array_reverse", BindFunc: arrfn.BindArrayReverse},
	{Name: "make_array", BindFunc: arrfn.BindMakeArray},
	{Name: "make_struct", BindFunc: structs.BindMakeStruct},

	// hyperloglog++ functions
	{Name: "hll_count_extract", BindFunc: hll.BindHllCountExtract},

	// bit functions
	{Name: "bit_count", BindFunc: helper.Scalar1(bit.BIT_COUNT)},

	// aggregate option funcs
	{Name: "distinct", BindFunc: bindDistinct},
	{Name: "ignore_nulls", BindFunc: bindIgnoreNulls},

	// window option funcs
	{Name: "window_rowid", BindFunc: bindWindowRowID},

	// javascript funcs
	{Name: "eval_javascript", BindFunc: bindEvalJavaScript},

	// net funcs
	{Name: "net_host", BindFunc: netfn.BindNetHost},
	{Name: "net_ip_from_string", BindFunc: netfn.BindNetIpFromString},
	{Name: "net_ip_net_mask", BindFunc: netfn.BindNetIpNetMask},
	{Name: "net_ip_to_string", BindFunc: netfn.BindNetIpToString},
	{Name: "net_ip_trunc", BindFunc: netfn.BindNetIpTrunc},
	{Name: "net_ipv4_from_int64", BindFunc: netfn.BindNetIpv4FromInt64},
	{Name: "net_ipv4_to_int64", BindFunc: netfn.BindNetIpv4ToInt64},
	{Name: "net_public_suffix", BindFunc: netfn.BindNetPublicSuffix},
	{Name: "net_reg_domain", BindFunc: netfn.BindNetRegDomain},
	{Name: "net_safe_ip_from_string", BindFunc: netfn.BindNetSafeIpFromString},
	{Name: "net_ip_in_net", BindFunc: longtail.BindNetIpInNet},
	{Name: "net_parse_ip", BindFunc: longtail.BindNetParseIp},
	{Name: "net_format_ip", BindFunc: longtail.BindNetFormatIp},
	{Name: "net_parse_packed_ip", BindFunc: longtail.BindNetParsePackedIp},
	{Name: "net_format_packed_ip", BindFunc: longtail.BindNetFormatPackedIp},
	{Name: "net_make_net", BindFunc: longtail.BindNetMakeNet},

	// longtail GoogleSQL scalars (debug, string, json, hash, array,
	// aggregate-grouping). The aggregate `grouping` lands in
	// normalFuncs because it folds during ANALYZE — by the time the
	// runtime sees it the column is no longer a grouping column.
	{Name: "iferror", BindFunc: longtail.BindIfError},
	{Name: "iserror", BindFunc: longtail.BindIsError},
	{Name: "nulliferror", BindFunc: longtail.BindNullIfError},
	{Name: "regexp_match", BindFunc: longtail.BindRegexpMatch},
	{Name: "regexp_extract_groups", BindFunc: longtail.BindRegexpExtractGroups},
	{Name: "split_substr", BindFunc: longtail.BindSplitSubstr},
	// `collate` is registered earlier from internal/functions/string
	// (alias of BindCollate). longtail.BindCollate would conflict.
	{Name: "safe_to_json", BindFunc: longtail.BindSafeToJson},
	{Name: "json_array_append", BindFunc: longtail.BindJsonArrayAppend},
	{Name: "json_array_insert", BindFunc: longtail.BindJsonArrayInsert},
	{Name: "highway_fingerprint128", BindFunc: longtail.BindHighwayFingerprint128},
	{Name: "flatten", BindFunc: longtail.BindFlatten},
	{Name: "array_zip", BindFunc: longtail.BindArrayZip},
	{Name: "grouping", BindFunc: longtail.BindGrouping},

	// Spanner-only singletons. These do not live behind mysql.* —
	// callers use the bare name.
	{Name: "pending_commit_timestamp", BindFunc: spannerfn.BindPendingCommitTimestamp},
	{Name: "bit_reverse", BindFunc: spannerfn.BindBitReverse},
	{Name: "dot_product", BindFunc: spannerfn.BindDotProduct},
	{Name: "approx_dot_product", BindFunc: spannerfn.BindApproxDotProduct},
	{Name: "approx_euclidean_distance", BindFunc: spannerfn.BindApproxEuclideanDistance},
	{Name: "approx_cosine_distance", BindFunc: spannerfn.BindApproxCosineDistance},
	{Name: "zstd_compress", BindFunc: spannerfn.BindZstdCompress},
	{Name: "zstd_decompress_to_bytes", BindFunc: spannerfn.BindZstdDecompressToBytes},
	{Name: "zstd_decompress_to_string", BindFunc: spannerfn.BindZstdDecompressToString},
	{Name: "get_next_sequence_value", BindFunc: spannerfn.BindGetNextSequenceValue},
	{Name: "get_internal_sequence_state", BindFunc: spannerfn.BindGetInternalSequenceState},

	// Spanner mysql.* aliases over already-implemented scalars.
	{Name: "mysql_lcase", BindFunc: str.BindLower},
	{Name: "mysql_ucase", BindFunc: str.BindUpper},
	{Name: "mysql_adddate", BindFunc: spannerfn.BindAddDate},
	{Name: "mysql_subdate", BindFunc: spannerfn.BindSubDate},

	// BigQuery AEAD encryption (KEYS.* + AEAD.* + DETERMINISTIC_*).
	// Upstream registers these as namespaced functions under "keys"
	// and "aead" sub-catalogs when the Encryption LanguageFeature is
	// enabled. We enable that feature both on the catalog construction
	// path (newSimpleCatalog) and on the analyzer-side LanguageOptions
	// (internal/analyzer.go) — the latter is the one the resolver
	// consults at analyze time. Runtime bodies live in
	// internal/functions/aead.
	{Name: "keys_new_keyset", NonDeterministic: true, BindFunc: aead.BindKeysNewKeyset},
	{Name: "keys_add_key_from_raw_bytes", NonDeterministic: true, BindFunc: aead.BindKeysAddKeyFromRawBytes},
	{Name: "keys_keyset_length", BindFunc: aead.BindKeysKeysetLength},
	{Name: "keys_keyset_to_json", BindFunc: aead.BindKeysKeysetToJson},
	{Name: "keys_keyset_from_json", BindFunc: aead.BindKeysKeysetFromJson},
	{Name: "keys_rotate_keyset", NonDeterministic: true, BindFunc: aead.BindKeysRotateKeyset},
	{Name: "keys_keyset_chain", BindFunc: aead.BindKeysKeysetChain},
	{Name: "keys_new_wrapped_keyset", NonDeterministic: true, BindFunc: aead.BindKeysNewWrappedKeyset},
	{Name: "keys_rewrap_keyset", BindFunc: aead.BindKeysRewrapKeyset},
	{Name: "keys_rotate_wrapped_keyset", BindFunc: aead.BindKeysRotateWrappedKeyset},
	{Name: "aead_encrypt", NonDeterministic: true, BindFunc: aead.BindAeadEncrypt},
	{Name: "aead_decrypt_bytes", BindFunc: aead.BindAeadDecryptBytes},
	{Name: "aead_decrypt_string", BindFunc: aead.BindAeadDecryptString},
	{Name: "deterministic_encrypt", BindFunc: aead.BindDeterministicEncrypt},
	{Name: "deterministic_decrypt_bytes", BindFunc: aead.BindDeterministicDecryptBytes},
	{Name: "deterministic_decrypt_string", BindFunc: aead.BindDeterministicDecryptString},

	// BigQuery DLP family (deterministic local AES-GCM stand-in for
	// Cloud DLP). DLP_KEY_CHAIN serialises (resource, wrapped) into a
	// JSON envelope BYTES; the encrypt/decrypt pair uses an HMAC-
	// derived nonce so the same (key, plaintext, surrogate, context)
	// always yields the same ciphertext, matching production
	// determinism guarantees.
	{Name: "dlp_key_chain", BindFunc: extras.BindDlpKeyChain},
	{Name: "dlp_deterministic_encrypt", BindFunc: extras.BindDlpDeterministicEncrypt},
	{Name: "dlp_deterministic_decrypt", BindFunc: extras.BindDlpDeterministicDecrypt},

	// BigQuery ObjectRef. The runtime ObjectRef is a STRING-typed
	// JSON envelope; metadata / signed-URL functions return values
	// deterministically derived from that envelope.
	{Name: "obj_make_ref", BindFunc: extras.BindObjMakeRef},
	{Name: "obj_fetch_metadata", BindFunc: extras.BindObjFetchMetadata},
	{Name: "obj_get_access_url", BindFunc: extras.BindObjGetAccessUrl},
	{Name: "obj_get_read_url", BindFunc: extras.BindObjGetReadUrl},

	// Spanner ML / AI. These stubs return deterministic placeholder
	// values; production semantics delegate to Vertex AI which is
	// out of scope for an in-process backend.
	{Name: "ai_if", BindFunc: extras.BindAiIf},
	{Name: "ai_classify", BindFunc: extras.BindAiClassify},
	{Name: "ai_score", BindFunc: extras.BindAiScore},
	{Name: "ml_predict", BindFunc: extras.BindMlPredict},
}
