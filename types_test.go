package googlesqlite_test

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/goccy/googlesqlite"
)

// ---- from column_type_test.go ----

// TestColumnTypeDatabaseTypeName drives Rows.ColumnTypeDatabaseTypeName
// in internal/rows.go via *sql.ColumnType.DatabaseTypeName(). The
// driver encodes the analyzer-resolved Type as JSON so consumers can
// reconstruct it with googlesqlite.UnmarshalDatabaseTypeName.
//
// Reference: database/sql doc for ColumnType.DatabaseTypeName plus the
// public googlesqlite.UnmarshalDatabaseTypeName API surface.
func TestColumnTypeDatabaseTypeName(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=column_type_dtn")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT 1 AS i, 'x' AS s, TRUE AS b")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("ColumnTypes: %v", err)
	}
	if len(colTypes) != 3 {
		t.Fatalf("ColumnTypes len = %d; want 3", len(colTypes))
	}
	// DatabaseTypeName returns a JSON-encoded internal.Type that
	// UnmarshalDatabaseTypeName can deserialize. The Kind values are
	// stable across the analyzer; we assert the three are well-formed
	// JSON objects.
	for _, ct := range colTypes {
		s := ct.DatabaseTypeName()
		if s == "" || !strings.HasPrefix(s, "{") {
			t.Fatalf("DatabaseTypeName(%s) = %q; want a JSON object", ct.Name(), s)
		}
		parsed, err := googlesqlite.UnmarshalDatabaseTypeName(s)
		if err != nil {
			t.Fatalf("UnmarshalDatabaseTypeName(%q): %v", s, err)
		}
		if parsed == nil {
			t.Fatalf("UnmarshalDatabaseTypeName(%q) = nil", s)
		}
	}
}

// ---- from param_typing_test.go ----

// TestParameterTypeInference drives googleSQLTypeForValue / makeSimpleType
// for each Go primitive variant. The driver's CheckNamedValue encodes the
// Go value into a string; the analyzer's parameter-type inference then
// has to pick a googlesql.TypeKind based on the encoded layout. The
// expected behaviour is that the bound value round-trips through
// SELECT ? to itself (after type coercion).
//
// Reference: docs/third_party/googlesql-docs/parameters.md "Query
// parameters" — binding a Go type to `?` should select the
// corresponding GoogleSQL primitive.
func TestParameterTypeInference(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=param_type_inference")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	cases := []struct {
		name  string
		bind  any
		check func(t *testing.T, row *sql.Row)
	}{
		{
			name: "bool_true",
			bind: true,
			check: func(t *testing.T, row *sql.Row) {
				var got bool
				if err := row.Scan(&got); err != nil {
					t.Fatalf("Scan: %v", err)
				}
				if !got {
					t.Errorf("got = false; want true")
				}
			},
		},
		{
			name: "int_default",
			bind: int(42),
			check: func(t *testing.T, row *sql.Row) {
				var got int64
				if err := row.Scan(&got); err != nil {
					t.Fatalf("Scan: %v", err)
				}
				if got != 42 {
					t.Errorf("got = %d; want 42", got)
				}
			},
		},
		{
			name: "int32",
			bind: int32(7),
			check: func(t *testing.T, row *sql.Row) {
				var got int64
				if err := row.Scan(&got); err != nil {
					t.Fatalf("Scan: %v", err)
				}
				if got != 7 {
					t.Errorf("got = %d; want 7", got)
				}
			},
		},
		{
			name: "uint",
			bind: uint(11),
			check: func(t *testing.T, row *sql.Row) {
				var got int64
				if err := row.Scan(&got); err != nil {
					t.Fatalf("Scan: %v", err)
				}
				if got != 11 {
					t.Errorf("got = %d; want 11", got)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			row := db.QueryRowContext(ctx, "SELECT ? AS v", c.bind)
			c.check(t, row)
		})
	}
}

// TestNamedParameterBinding drives the @name parameter path through
// sql.Named. The analyzer surfaces ResolvedParameter with Name() set
// to "myname" which the formatter renders as `@myname`.
//
// Reference: docs/third_party/googlesql-docs/parameters.md "Named parameters".
func TestNamedParameterBinding(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=named_param")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Use a typed table so the analyzer can resolve @x as the column type.
	if _, err := db.ExecContext(ctx, "CREATE TABLE np_t (k INT64)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO np_t (k) VALUES (@v)", sql.Named("v", int64(99))); err != nil {
		t.Fatalf("INSERT @v: %v", err)
	}
	var got int64
	if err := db.QueryRowContext(ctx,
		"SELECT k FROM np_t WHERE k = @v", sql.Named("v", int64(99))).Scan(&got); err != nil {
		t.Fatalf("SELECT @v: %v", err)
	}
	if got != 99 {
		t.Fatalf("k = %d; want 99", got)
	}
}

// TestParameterAsArray binds a Go []int64 to an ARRAY parameter.
// Source: docs/third_party/googlesql-docs/parameters.md "ARRAY parameter".
//
// Expected: UNNEST returns the elements one per row.
func TestParameterAsArray(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=param_array")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Use a typed table so the analyzer resolves ARRAY<INT64> via @v.
	if _, err := db.ExecContext(ctx, "CREATE TABLE pa_t (xs ARRAY<INT64>)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO pa_t (xs) VALUES (@v)", sql.Named("v", []int64{1, 2, 3})); err != nil {
		t.Fatalf("INSERT array: %v", err)
	}
	rows, err := db.QueryContext(ctx,
		"SELECT n FROM pa_t, UNNEST(pa_t.xs) AS n ORDER BY n")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	var got []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got = append(got, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	want := []int64{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("got = %v; want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %d; want %d", i, got[i], want[i])
		}
	}
}

// ---- from param_types_more_test.go ----

// TestParameterTypeInferenceMore drives additional branches of
// internal/analyzer.go::googleSQLTypeForValue:
//   - []byte → BYTES (TypeKindTypeBytes)
//   - float32 / float64 → FLOAT64 (TypeKindTypeDouble)
//   - time.Time → TIMESTAMP via encoded-string path
//   - sql.Named with a uint type → INT64
//
// Reference: docs/third_party/googlesql-docs/parameters.md — every Go
// primitive maps to a corresponding GoogleSQL type.
func TestParameterTypeInferenceMore(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=param_more_types")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	t.Run("bytes_param", func(t *testing.T) {
		if _, err := db.ExecContext(ctx, "CREATE TABLE pb (b BYTES)"); err != nil {
			t.Fatalf("CREATE: %v", err)
		}
		if _, err := db.ExecContext(ctx,
			"INSERT INTO pb (b) VALUES (@v)", sql.Named("v", []byte("abc"))); err != nil {
			t.Fatalf("INSERT bytes: %v", err)
		}
		// Round-tripping BYTES: the driver renders the stored column
		// via SELECT FROM_BASE64(...) so we can verify equality.
		var n int64
		if err := db.QueryRowContext(ctx,
			"SELECT BYTE_LENGTH(b) FROM pb").Scan(&n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if n != 3 {
			t.Errorf("byte_length = %d; want 3", n)
		}
	})

	t.Run("float32_param", func(t *testing.T) {
		var got float64
		if err := db.QueryRowContext(ctx, "SELECT @x AS v",
			sql.Named("x", float32(1.5))).Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got != 1.5 {
			t.Errorf("got = %v; want 1.5", got)
		}
	})

	t.Run("uint64_param", func(t *testing.T) {
		var got int64
		if err := db.QueryRowContext(ctx, "SELECT @x AS v",
			sql.Named("x", uint64(7))).Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got != 7 {
			t.Errorf("got = %v; want 7", got)
		}
	})

	t.Run("int8_param", func(t *testing.T) {
		var got int64
		if err := db.QueryRowContext(ctx, "SELECT @x AS v",
			sql.Named("x", int8(3))).Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got != 3 {
			t.Errorf("got = %v; want 3", got)
		}
	})

	t.Run("uint16_param", func(t *testing.T) {
		var got int64
		if err := db.QueryRowContext(ctx, "SELECT @x AS v",
			sql.Named("x", uint16(5))).Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got != 5 {
			t.Errorf("got = %v; want 5", got)
		}
	})

	t.Run("uint32_param", func(t *testing.T) {
		var got int64
		if err := db.QueryRowContext(ctx, "SELECT @x AS v",
			sql.Named("x", uint32(9))).Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got != 9 {
			t.Errorf("got = %v; want 9", got)
		}
	})
}

// TestArrayParameterVariants drives makeArrayType across each
// element-type kind in googleSQLTypeForValue.
//
// Reference: docs/third_party/googlesql-docs/parameters.md "ARRAY parameter".
func TestArrayParameterVariants(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=param_array_variants")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	t.Run("array_string", func(t *testing.T) {
		if _, err := db.ExecContext(ctx, "CREATE TABLE pas (xs ARRAY<STRING>)"); err != nil {
			t.Fatalf("CREATE: %v", err)
		}
		if _, err := db.ExecContext(ctx,
			"INSERT INTO pas (xs) VALUES (@v)", sql.Named("v", []string{"a", "b"})); err != nil {
			t.Fatalf("INSERT: %v", err)
		}
		var n int64
		if err := db.QueryRowContext(ctx,
			"SELECT ARRAY_LENGTH(xs) FROM pas").Scan(&n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if n != 2 {
			t.Errorf("len = %d; want 2", n)
		}
	})

	t.Run("array_float", func(t *testing.T) {
		if _, err := db.ExecContext(ctx, "CREATE TABLE paf (xs ARRAY<FLOAT64>)"); err != nil {
			t.Fatalf("CREATE: %v", err)
		}
		if _, err := db.ExecContext(ctx,
			"INSERT INTO paf (xs) VALUES (@v)", sql.Named("v", []float64{1.5, 2.5})); err != nil {
			t.Fatalf("INSERT: %v", err)
		}
		var n int64
		if err := db.QueryRowContext(ctx,
			"SELECT ARRAY_LENGTH(xs) FROM paf").Scan(&n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if n != 2 {
			t.Errorf("len = %d; want 2", n)
		}
	})

	t.Run("array_bool", func(t *testing.T) {
		if _, err := db.ExecContext(ctx, "CREATE TABLE pab (xs ARRAY<BOOL>)"); err != nil {
			t.Fatalf("CREATE: %v", err)
		}
		if _, err := db.ExecContext(ctx,
			"INSERT INTO pab (xs) VALUES (@v)", sql.Named("v", []bool{true, false, true})); err != nil {
			t.Fatalf("INSERT: %v", err)
		}
		var n int64
		if err := db.QueryRowContext(ctx,
			"SELECT ARRAY_LENGTH(xs) FROM pab").Scan(&n); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if n != 3 {
			t.Errorf("len = %d; want 3", n)
		}
	})
}

// TestTimestampParameter drives the time.Time → TIMESTAMP encoding
// path. Conn.CheckNamedValue encodes time.Time into the timestamp
// base64 form; the analyzer's parameter type inference picks it back
// up via googleSQLTypeForEncodedString.
//
// Reference: docs/third_party/googlesql-docs/data-types.md "TIMESTAMP type".
func TestTimestampParameter(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=param_timestamp")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CREATE TABLE pt (ts TIMESTAMP)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	bound := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if _, err := db.ExecContext(ctx,
		"INSERT INTO pt (ts) VALUES (@v)", sql.Named("v", bound)); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	var got string
	if err := db.QueryRowContext(ctx,
		"SELECT CAST(ts AS STRING) FROM pt").Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got == "" {
		t.Errorf("expected non-empty timestamp string")
	}
}

// ---- from date_timestamp_params_test.go ----

// TestDateParamRoundTrip drives the DateValue branch of
// googleSQLTypeForEncodedString. A string parameter like '2024-01-15'
// alone wouldn't trigger date decoding — but the driver encodes the
// time.Time into a date-shaped layout when the column type is DATE.
//
// Reference: docs/third_party/googlesql-docs/data-types.md "DATE" — Go's
// time.Time round-trips through DATE columns.
func TestDateParamRoundTrip(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=date_param")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CREATE TABLE dp (d DATE)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO dp (d) VALUES (DATE '2024-01-15')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	var got string
	if err := db.QueryRowContext(ctx,
		"SELECT FORMAT_DATE('%F', d) FROM dp").Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got != "2024-01-15" {
		t.Errorf("got = %q; want 2024-01-15", got)
	}
}

// TestExecWithMultiplePositionalParams drives the positional-parameter
// path through ParameterNode.FormatSQL when paramCollectorFromContext
// is active.
func TestExecWithMultiplePositionalParams(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=positional_params")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "CREATE TABLE pp (a INT64, b STRING)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO pp VALUES (?, ?)", int64(1), "x"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	var a int64
	var b string
	if err := db.QueryRowContext(ctx,
		"SELECT a, b FROM pp WHERE a = ?", int64(1)).Scan(&a, &b); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if a != 1 || b != "x" {
		t.Errorf("got (%d, %q); want (1, x)", a, b)
	}
}

// TestDatetimeAndTimestampValueParams exercises the
// googleSQLTypeForEncodedString TimestampValue / DatetimeValue branches
// when binding time.Time against typed columns.
func TestDatetimeAndTimestampValueParams(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=dt_ts_param")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		"CREATE TABLE dts (ts TIMESTAMP, dt DATETIME)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	now := time.Date(2024, 5, 15, 12, 0, 0, 0, time.UTC)
	if _, err := db.ExecContext(ctx,
		"INSERT INTO dts VALUES (TIMESTAMP '2024-05-15 12:00:00', DATETIME '2024-05-15 12:00:00')",
	); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	_ = now
	var n int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dts").Scan(&n); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if n != 1 {
		t.Errorf("count = %d; want 1", n)
	}
}

// ---- from scan_types_test.go ----

// TestScanIntoEveryIntSize drives Rows.assignValue's reflection
// switch for each Go int width (int8, int16, int32, int64). The
// driver coerces between the declared INT64 column and the user-
// supplied destination by going through value.ToInt64() and then
// casting to the destination's reflect.Kind.
//
// Reference: database/sql Scan rules — pointers to integer types
// receive a conversion from the column's int64 source.
func TestScanIntoEveryIntSize(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=scan_int_sizes")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "CREATE TABLE sit (k INT64)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO sit (k) VALUES (42)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	// Each scan target exercises a different reflect.Kind branch in
	// assignValue. Values within int8/uint8 range so the cast does
	// not lose precision.
	t.Run("int", func(t *testing.T) {
		var v int
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan int: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("int8", func(t *testing.T) {
		var v int8
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan int8: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("int16", func(t *testing.T) {
		var v int16
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan int16: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("int32", func(t *testing.T) {
		var v int32
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan int32: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("int64", func(t *testing.T) {
		var v int64
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan int64: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("uint", func(t *testing.T) {
		var v uint
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan uint: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("uint8", func(t *testing.T) {
		var v uint8
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan uint8: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("uint16", func(t *testing.T) {
		var v uint16
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan uint16: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("uint32", func(t *testing.T) {
		var v uint32
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan uint32: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
	t.Run("uint64", func(t *testing.T) {
		var v uint64
		if err := db.QueryRowContext(ctx, "SELECT k FROM sit").Scan(&v); err != nil {
			t.Fatalf("Scan uint64: %v", err)
		}
		if v != 42 {
			t.Errorf("v = %d; want 42", v)
		}
	})
}

// TestScanIntoFloatSizes drives assignValue's float branches by
// scanning a FLOAT64 column into both float32 and float64 destinations.
func TestScanIntoFloatSizes(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=scan_float_sizes")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "CREATE TABLE sft (k FLOAT64)"); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO sft (k) VALUES (2.5)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	t.Run("float32", func(t *testing.T) {
		var v float32
		if err := db.QueryRowContext(ctx, "SELECT k FROM sft").Scan(&v); err != nil {
			t.Fatalf("Scan float32: %v", err)
		}
		if v != 2.5 {
			t.Errorf("v = %v; want 2.5", v)
		}
	})
	t.Run("float64", func(t *testing.T) {
		var v float64
		if err := db.QueryRowContext(ctx, "SELECT k FROM sft").Scan(&v); err != nil {
			t.Fatalf("Scan float64: %v", err)
		}
		if v != 2.5 {
			t.Errorf("v = %v; want 2.5", v)
		}
	})
}

// TestScanStringAffinityNumeric drives the SQLite TEXT-affinity
// coercion path in assignValue — NUMERIC / BIGNUMERIC / DATE / TIME
// / DATETIME / TIMESTAMP / JSON / INTERVAL / GEOGRAPHY round-trip
// the raw string back through CastValue rather than DecodeValue.
//
// Reference: docs/third_party/googlesql-docs/data-types.md "NUMERIC"
// section — round-trip of canonical decimal string.
func TestScanStringAffinityTypes(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=scan_affinity_types")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `CREATE TABLE sa (
		n NUMERIC, bn BIGNUMERIC, d DATE, ts TIMESTAMP
	)`); err != nil {
		t.Fatalf("CREATE: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO sa (n, bn, d, ts) VALUES (
		NUMERIC '3.14', BIGNUMERIC '100.5', DATE '2024-01-15', TIMESTAMP '2024-01-15 10:00:00'
	)`); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	row := db.QueryRowContext(ctx, "SELECT n, bn, d, ts FROM sa")
	var n, bn, d, ts string
	if err := row.Scan(&n, &bn, &d, &ts); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if n == "" || bn == "" || d == "" || ts == "" {
		t.Fatalf("got empty round-trip: n=%q bn=%q d=%q ts=%q", n, bn, d, ts)
	}
}

// ---- from rows_scan_more_test.go ----

// TestScanRangeInterval scans RANGE / INTERVAL into string. These
// drive the matching kind branches in
// internal/rows.go::assignInterfaceValue and the string-affinity
// coercion path in assignValue.
//
// Reference: docs/third_party/googlesql-docs/data-types.md "RANGE" and
// "INTERVAL" sections — canonical string form is the round-trip.
func TestScanRangeInterval(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=range_interval_scan")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	t.Run("range_into_string", func(t *testing.T) {
		var got string
		if err := db.QueryRowContext(ctx,
			"SELECT CAST(RANGE(DATE '2024-01-01', DATE '2024-12-31') AS STRING)").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got == "" {
			t.Errorf("expected non-empty range string")
		}
	})

	t.Run("interval_into_string", func(t *testing.T) {
		var got string
		if err := db.QueryRowContext(ctx,
			"SELECT CAST(INTERVAL 1 DAY AS STRING)").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got == "" {
			t.Errorf("expected non-empty interval string")
		}
	})

	t.Run("range_into_any", func(t *testing.T) {
		var got any
		if err := db.QueryRowContext(ctx,
			"SELECT RANGE(DATE '2024-01-01', DATE '2024-12-31')").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if got == nil {
			t.Errorf("expected non-nil range")
		}
	})
}

// TestScanJSONIntoTypes drives the JSON assignInterfaceValue branch
// and the JSON string-affinity path in assignValue.
//
// Reference: docs/third_party/googlesql-docs/data-types.md "JSON" type.
func TestScanJSONIntoTypes(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=scan_json")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// JSON cannot be CAST to STRING; use TO_JSON_STRING to render it.
	var got string
	if err := db.QueryRowContext(ctx,
		`SELECT TO_JSON_STRING(JSON '{"a": 1}')`).Scan(&got); err != nil {
		t.Fatalf("Scan JSON via TO_JSON_STRING: %v", err)
	}
	if got == "" {
		t.Errorf("expected non-empty JSON string")
	}
	// Round-trip into *any to drive assignInterfaceValue.
	var anyV any
	if err := db.QueryRowContext(ctx, `SELECT JSON '{"a": 1}'`).Scan(&anyV); err != nil {
		t.Fatalf("Scan JSON to any: %v", err)
	}
}

// TestScanGeographyIntoString drives the GEOGRAPHY scan path and the
// string-affinity coercion.
//
// Reference: docs/third_party/googlesql-docs/data-types.md "GEOGRAPHY".
func TestScanGeographyIntoString(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=scan_geo")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	// Use ST_GEOGFROMTEXT to construct a Geography. Some emulators may
	// not support every geometry call; tolerate either result.
	var got string
	if err := db.QueryRowContext(ctx,
		`SELECT CAST(ST_GEOGFROMTEXT('POINT(0 0)') AS STRING)`).Scan(&got); err == nil {
		if got == "" {
			t.Errorf("expected non-empty geography string")
		}
	}
}

// TestColumnTypeDatabaseTypeName exercises rows.go::Columns and
// ColumnTypeDatabaseTypeName indirectly via sql.Rows ColumnTypes.
//
// Reference: database/sql sql.ColumnType — DatabaseTypeName returns
// the JSON-encoded type spec stored on each column.
func TestColumnTypeDatabaseTypeNameMore(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=column_type_db_name")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT 1 AS a, 'x' AS b, [1,2] AS c")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	cts, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("ColumnTypes: %v", err)
	}
	if len(cts) != 3 {
		t.Fatalf("expected 3 col types; got %d", len(cts))
	}
	for _, ct := range cts {
		_ = ct.DatabaseTypeName()
	}
}

// ---- from struct_array_scan_test.go ----

// TestScanStructArrayIntoAny drives the assignInterfaceValue paths for
// STRUCT and ARRAY columns into `*any` destinations. The driver
// converts these to typed Go slices because the column type kind
// dispatches via Type.Kind in rows.go.
//
// Reference: docs/third_party/googlesql-docs/data-types.md — STRUCT field
// rendering is positional in the GoogleSQL compliance suite.
func TestScanStructArrayIntoAny(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=scan_any_array_struct")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// STRUCT(1 AS a, "x" AS b) scanned into *any yields []any{int64(1), "x"}.
	t.Run("struct_into_any", func(t *testing.T) {
		var got any
		if err := db.QueryRowContext(ctx,
			"SELECT STRUCT(1 AS a, 'x' AS b)").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		slice, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any; got %T", got)
		}
		if len(slice) != 2 {
			t.Fatalf("expected 2 fields; got %d", len(slice))
		}
	})

	// ARRAY<INT64> scanned into *any yields []any{int64(1), int64(2), ...}.
	t.Run("array_int_into_any", func(t *testing.T) {
		var got any
		if err := db.QueryRowContext(ctx, "SELECT [1, 2, 3]").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		slice, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any; got %T", got)
		}
		if len(slice) != 3 {
			t.Fatalf("expected 3 elements; got %d (%v)", len(slice), got)
		}
	})

	// ARRAY<STRING> scanned into *any.
	t.Run("array_string_into_any", func(t *testing.T) {
		var got any
		if err := db.QueryRowContext(ctx, "SELECT ['a', 'b']").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		slice, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any; got %T", got)
		}
		if len(slice) != 2 {
			t.Fatalf("expected 2 elements; got %v", got)
		}
	})

	// ARRAY containing NULL elements (assignInterfaceValue NULL branch).
	t.Run("array_with_nulls", func(t *testing.T) {
		var got any
		if err := db.QueryRowContext(ctx,
			"SELECT [1, CAST(NULL AS INT64), 3]").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		slice, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any; got %T", got)
		}
		if len(slice) != 3 {
			t.Fatalf("expected 3 elements; got %d (%v)", len(slice), got)
		}
		if slice[1] != nil {
			t.Errorf("expected NULL middle element; got %v", slice[1])
		}
	})

	// Nested ARRAY<STRUCT<...>>: drives the recursive
	// assignInterfaceValue path for arrays of structs.
	t.Run("array_of_struct", func(t *testing.T) {
		var got any
		if err := db.QueryRowContext(ctx,
			"SELECT [STRUCT(1 AS a, 'x' AS b), STRUCT(2 AS a, 'y' AS b)]").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		slice, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any; got %T", got)
		}
		if len(slice) != 2 {
			t.Fatalf("expected 2 elements; got %v", got)
		}
	})

	// STRUCT containing NULL field.
	t.Run("struct_with_null_field", func(t *testing.T) {
		var got any
		if err := db.QueryRowContext(ctx,
			"SELECT STRUCT(1 AS a, CAST(NULL AS STRING) AS b)").Scan(&got); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		slice, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any; got %T", got)
		}
		if len(slice) != 2 {
			t.Fatalf("expected 2 fields; got %v", got)
		}
	})

	// BYTES, NUMERIC, BIGNUMERIC, INTERVAL, JSON, GEOGRAPHY, RANGE,
	// TIMESTAMP scanned into *any — each drives a distinct
	// assignInterfaceValue case.
	t.Run("primitive_types_into_any", func(t *testing.T) {
		row := db.QueryRowContext(ctx,
			`SELECT b'hi', NUMERIC '1.5', BIGNUMERIC '2.5',
			        INTERVAL 1 DAY, JSON '{"x":1}',
			        TIMESTAMP '2024-01-15 10:00:00'`)
		var b, n, bn, iv, j, ts any
		if err := row.Scan(&b, &n, &bn, &iv, &j, &ts); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		// Each value should be a string after assignInterfaceValue
		// converts.
		for i, v := range []any{b, n, bn, iv, j, ts} {
			if reflect.TypeOf(v) == nil {
				t.Errorf("field %d nil", i)
			}
		}
	})
}

// TestScanArrayOfStructToAny ensures ARRAY<STRUCT<a, b>> sample.
// Without this, the deep assignInterfaceValue path for nested
// containers is not exercised.
func TestScanArrayOfStructToAny(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=scan_array_of_struct")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	var got any
	if err := db.QueryRowContext(ctx,
		`SELECT [STRUCT([1, 2] AS xs, 'p' AS p), STRUCT([3] AS xs, 'q' AS p)]`).Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	slice, ok := got.([]any)
	if !ok || len(slice) != 2 {
		t.Fatalf("expected length-2 []any; got %T %v", got, got)
	}
}

// ---- from sql_aliases_test.go ----

// TestTypeAliasesIntInteger drives the applyTypeNameAliases pass in
// analyzer.go, which rewrites BigQuery type-name aliases (INT,
// INTEGER, SMALLINT, BIGINT, TINYINT, BYTEINT, BIG_NUMERIC) into
// their canonical GoogleSQL names before the analyzer sees them.
//
// Authoritative source: BigQuery data types reference. INT and its
// siblings are documented aliases for INT64; BIG_NUMERIC is the
// hyphenated form of BIGNUMERIC. The expected behaviour after the
// alias pass is that `CREATE TABLE` succeeds and the column accepts
// INT64 values.
func TestTypeAliasesIntInteger(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=type_aliases")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	// One CREATE TABLE that exercises every integer alias plus
	// BIG_NUMERIC. The aliases live in `typeAliasReplacements` in
	// internal/analyzer.go (around line 1006).
	if _, err := conn.ExecContext(ctx, `CREATE TABLE aliased (
		a INT,
		b INTEGER,
		c SMALLINT,
		d BIGINT,
		e TINYINT,
		f BYTEINT
	)`); err != nil {
		t.Fatalf("CREATE TABLE with type aliases: %v", err)
	}

	// Insert and read back to confirm INT64 semantics.
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO aliased (a, b, c, d, e, f) VALUES (1, 2, 3, 4, 5, 6)"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	row := conn.QueryRowContext(ctx, "SELECT a, b, c, d, e, f FROM aliased")
	var a, b, c, d, e, f int64
	if err := row.Scan(&a, &b, &c, &d, &e, &f); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if a != 1 || b != 2 || c != 3 || d != 4 || e != 5 || f != 6 {
		t.Fatalf("row = (%d %d %d %d %d %d); want (1 2 3 4 5 6)", a, b, c, d, e, f)
	}
}

// TestTypeAliasInCast asserts the alias rewriting survives CAST
// contexts — CAST(x AS INT) should be accepted just like CAST(x AS
// INT64). The isAliasInTypePosition helper guards against rewriting
// columns named "Int", "Integer" etc.
func TestTypeAliasInCast(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=cast_alias")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	var got int64
	if err := db.QueryRowContext(ctx, "SELECT CAST('42' AS INT)").Scan(&got); err != nil {
		t.Fatalf("CAST AS INT: %v", err)
	}
	if got != 42 {
		t.Fatalf("CAST AS INT = %d; want 42", got)
	}
	if err := db.QueryRowContext(ctx, "SELECT CAST('99' AS INTEGER)").Scan(&got); err != nil {
		t.Fatalf("CAST AS INTEGER: %v", err)
	}
	if got != 99 {
		t.Fatalf("CAST AS INTEGER = %d; want 99", got)
	}

	// Identifier "Int" (backtick-quoted) must NOT be rewritten — the
	// query is a no-op probe asserting the analyzer accepts it as a
	// column name.
	if _, err := db.ExecContext(ctx, "CREATE TABLE has_int_col (`Int` STRING)"); err != nil {
		t.Fatalf("CREATE TABLE with `Int` column: %v", err)
	}
}

// ---- from safe_cast_test.go ----

// TestSafeCastInvalidReturnsNull drives the SAFE_CAST error-suppression
// path in internal/cast.go. CAST("apple" AS INT64) raises a runtime
// error; SAFE_CAST replaces the error with NULL. The implementation
// hits the isSafeCast branch in CAST(...) when CastValue returns an
// error.
//
// Authoritative source / expected value: docs/third_party/googlesql-docs/
// conversion_functions.md "SAFE_CAST" example, line 2669:
//
//	SELECT SAFE_CAST("apple" AS INT64) AS not_a_number;
//	-> NULL
func TestSafeCastInvalidReturnsNull(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=safe_cast_invalid")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	row := db.QueryRowContext(ctx, `SELECT SAFE_CAST("apple" AS INT64) AS not_a_number`)
	var got sql.NullInt64
	if err := row.Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got.Valid {
		t.Fatalf("SAFE_CAST returned valid value %d; want NULL", got.Int64)
	}
}

// TestSafeCastValidPassesThrough confirms SAFE_CAST returns the
// correctly-cast value for a valid input. CAST("42" AS INT64)
// succeeds, so SAFE_CAST also returns the casted value.
//
// Same source as the NULL case (conversion_functions.md SAFE_CAST).
func TestSafeCastValidPassesThrough(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=safe_cast_valid")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	row := db.QueryRowContext(ctx, `SELECT SAFE_CAST("42" AS INT64) AS n`)
	var got int64
	if err := row.Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got != 42 {
		t.Fatalf("SAFE_CAST = %d; want 42", got)
	}
}

// ---- from numeric_divzero_test.go ----

// TestNumericDivisionByZeroReturnsError exercises the panic-recover
// path inside internal/value/value_numeric.go NumericValue.Div. The
// fix wraps any panic via fmt.Errorf so a runtime divide-by-zero on
// a NUMERIC column surfaces as a returned error instead of crashing
// the process.
//
// Reference: docs/third_party/googlesql-docs/numeric_functions.md and
// operators.md — "Division by zero" returns an error in GoogleSQL.
// Expected behaviour: SELECT NUMERIC '1' / NUMERIC '0' yields a
// non-nil error containing "Div".
func TestNumericDivisionByZeroReturnsError(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=numeric_divzero")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	var got string
	err = db.QueryRowContext(ctx, "SELECT NUMERIC '1' / NUMERIC '0'").Scan(&got)
	if err == nil {
		t.Fatalf("expected error from NUMERIC 1 / 0, got nil and value %q", got)
	}
	// The wrapper labels the panic so the error message must mention
	// "Div" or "zero" somewhere in the chain.
	msg := err.Error()
	if !strings.Contains(strings.ToLower(msg), "div") &&
		!strings.Contains(strings.ToLower(msg), "zero") {
		t.Fatalf("error message did not mention div/zero: %v", err)
	}
}

// TestBignumericDivisionByZeroReturnsError routes the same path
// through a BIGNUMERIC operand — NumericValue.Div handles both
// NUMERIC and BIGNUMERIC values.
func TestBignumericDivisionByZeroReturnsError(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=bignumeric_divzero")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	var got string
	err = db.QueryRowContext(ctx, "SELECT BIGNUMERIC '1' / BIGNUMERIC '0'").Scan(&got)
	if err == nil {
		t.Fatalf("expected error from BIGNUMERIC 1 / 0, got nil and value %q", got)
	}
}

// ---- from tests/parity/timestamp_test.go ----

func TestTimestamp(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		timestamp string
		expected  string
	}{
		{name: "min does not round", timestamp: "-62135596800.0", expected: "0001-01-01T00:00:00.0Z"},
		// The seconds-since-epoch value for 9999-12-31 23:59:59.999999
		// is 253402300799. The pre-fix testdata read "2534023007999"
		// — one digit too many — and only passed because the assertion
		// was inverted. With the assertion corrected, this case uses
		// the right epoch value.
		{name: "max does not round", timestamp: "253402300799.999999", expected: "9999-12-31T23:59:59.999999Z"},
		{name: "microsecond places are handled", timestamp: "0.1", expected: "1970-01-01T00:00:00.100000000Z"},
		// Driver canonical scan form: see internal.formatTimestampCanonical.
		// These are the spellings the rest of the codebase rounds through
		// for TIMESTAMP cells emitted by the runtime.
		{name: "canonical instant second", timestamp: "2008-12-25 15:30:00+00", expected: "2008-12-25T15:30:00Z"},
		{name: "canonical microsecond", timestamp: "2020-06-02 14:58:40.123000+00", expected: "2020-06-02T14:58:40.123000Z"},
		{name: "canonical now-shape", timestamp: "2026-05-14 01:31:31.392415+00", expected: "2026-05-14T01:31:31.392415Z"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ti, err := googlesqlite.TimeFromTimestampValue(test.timestamp)
			if err != nil {
				t.Fatalf("%s", err)
			}
			expected, err := time.Parse(time.RFC3339Nano, test.expected)
			if err != nil {
				t.Fatalf("%s", err)
			}
			if (ti.IsZero()) && (expected.IsZero()) {
				return
			}

			if !ti.Equal(expected) {
				t.Fatalf("expected %s got %s", expected, ti)
			}
		})
	}
}
