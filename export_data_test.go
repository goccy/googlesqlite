package googlesqlite_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/goccy/googlesqlite"
)

// memCapture is the test-local store the test-only mem writer commits to
// on Close. Tests get one instance back from registerMemScheme so they can
// assert against the bytes that flowed through EXPORT DATA without any
// process-wide state.
type memCapture struct {
	mu   sync.Mutex
	data map[string][]byte
}

func (m *memCapture) get(key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.data[key]
	return b, ok
}

func (m *memCapture) keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0, len(m.data))
	for k := range m.data {
		out = append(out, k)
	}
	return out
}

// registerMemScheme installs a test-local URI writer keyed on a scheme
// derived from t.Name() so each test owns its own scheme and its own
// capture map. The scheme is unregistered when the test finishes, so
// nothing leaks into sibling tests. Returns (scheme, capture) so the test
// can compose its URIs and assert on the captured bytes.
func registerMemScheme(t *testing.T) (string, *memCapture) {
	t.Helper()
	capture := &memCapture{data: map[string][]byte{}}
	// URL schemes are restricted to ALPHA / DIGIT / "+" / "-" / "." by
	// RFC 3986. Strip everything else from the test name so net/url can
	// parse the resulting `mem-<name>://...` URI.
	var sanitized strings.Builder
	for _, r := range strings.ToLower(t.Name()) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '+', r == '.':
			sanitized.WriteRune(r)
		default:
			sanitized.WriteRune('-')
		}
	}
	scheme := "mem-" + sanitized.String()
	googlesqlite.RegisterExportURIWriter(scheme, func(_ context.Context, uri string, _ googlesqlite.ExportWriterOpts) (io.WriteCloser, error) {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("%s://: invalid uri %q: %w", scheme, uri, err)
		}
		return &memWriteCloser{
			key:     path.Join(u.Host, u.Path),
			buf:     &bytes.Buffer{},
			capture: capture,
		}, nil
	})
	t.Cleanup(func() {
		googlesqlite.RegisterExportURIWriter(scheme, nil)
	})
	return scheme, capture
}

type memWriteCloser struct {
	key     string
	buf     *bytes.Buffer
	capture *memCapture
}

func (w *memWriteCloser) Write(p []byte) (int, error) { return w.buf.Write(p) }

func (w *memWriteCloser) Close() error {
	w.capture.mu.Lock()
	defer w.capture.mu.Unlock()
	w.capture.data[w.key] = append([]byte(nil), w.buf.Bytes()...)
	return nil
}

// TestExportDataErrorPaths covers the per-statement validation that does
// not require any URI scheme to actually open: missing `uri`, unsupported
// `format`, an unregistered scheme, an unknown option, and a known option
// supplied as a non-literal expression. All of these error out before the
// EXPORT DATA action tries to open a writer, so they run anywhere
// (including under the race detector).
func TestExportDataErrorPaths(t *testing.T) {
	scheme, _ := registerMemScheme(t)

	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:?_test=exportdataerr")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// expectErr is the per-subtest invariant: the statement must fail
	// before the URI writer is opened, and the error string must name
	// the offending option / clause. The defensive Close on any
	// non-nil rows guards against a future regression where a
	// previously-rejected statement starts succeeding silently — without
	// it the leaked *sql.Rows would pin the parent connection and turn
	// the failure into a multi-minute conn.Close deadlock rather than a
	// crisp assertion failure.
	expectErr := func(t *testing.T, sql, mustContain string) {
		t.Helper()
		rows, err := conn.QueryContext(ctx, sql)
		if rows != nil {
			rows.Close()
		}
		if err == nil {
			t.Fatalf("statement should have failed but succeeded: %s", sql)
		}
		if !containsFold(err.Error(), mustContain) {
			t.Fatalf("error %q does not mention %q", err, mustContain)
		}
	}

	t.Run("missing uri is rejected", func(t *testing.T) {
		expectErr(t, `EXPORT DATA OPTIONS(format = 'CSV') AS SELECT 1 AS id`, "uri")
	})

	t.Run("unsupported format is rejected", func(t *testing.T) {
		expectErr(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'AVRO')
             AS SELECT 1 AS id`, scheme), "avro")
	})

	t.Run("unregistered scheme is rejected", func(t *testing.T) {
		expectErr(t,
			`EXPORT DATA OPTIONS(uri = 's3://x/y', format = 'CSV')
             AS SELECT 1 AS id`, "s3")
	})

	t.Run("unknown option is rejected", func(t *testing.T) {
		// Use a name that is definitively not in BigQuery's EXPORT DATA
		// vocabulary so the test does not regress the day an option we
		// happen to pick becomes supported (this exact bug bit us once:
		// the test used `overwrite = false` as the "unknown" stand-in,
		// then started silently succeeding the moment overwrite landed
		// in the option-reader, leaking rows and deadlocking the
		// connection on Close).
		expectErr(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'CSV', not_a_real_option = false)
             AS SELECT 1 AS id`, scheme), "not_a_real_option")
	})

	t.Run("known option as non-literal expression is rejected", func(t *testing.T) {
		// Concatenating the URI is a perfectly valid SQL expression
		// but the option-reader only consumes string LITERALS today
		// (the analyzer does not constant-fold OPTIONS expressions for
		// us). The failure must call out the offending option by name,
		// not drop the value and surface as "required option `uri` is
		// missing" — that would mislead the caller about what is
		// wrong.
		rows, err := conn.QueryContext(ctx, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = CONCAT('%s://', 'x/y'), format = 'CSV')
             AS SELECT 1 AS id`, scheme))
		if rows != nil {
			rows.Close()
		}
		if err == nil {
			t.Fatal("EXPORT DATA with non-literal `uri` should fail")
		}
		if !containsFold(err.Error(), "uri") {
			t.Fatalf("error %q does not pin the offending `uri` option", err)
		}
		if containsFold(err.Error(), "missing") {
			t.Fatalf("error %q wrongly reports `uri` as missing when it was supplied non-literal", err)
		}
	})

	t.Run("header on non-CSV format is rejected", func(t *testing.T) {
		// `header` is documented as CSV-only; silently honouring it for
		// JSON output would either drop the value (losing intent) or
		// emit a column-name row that NDJSON readers cannot parse.
		expectErr(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'JSON', header = true)
             AS SELECT 1 AS id`, scheme), "header")
	})

	t.Run("field_delimiter on non-CSV format is rejected", func(t *testing.T) {
		expectErr(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'JSON', field_delimiter = '|')
             AS SELECT 1 AS id`, scheme), "field_delimiter")
	})

	t.Run("incompatible compression for format is rejected", func(t *testing.T) {
		// SNAPPY is documented for AVRO/PARQUET, not CSV. The encoder
		// would happily write uncompressed bytes if we dropped the value
		// — that produces a file that does not match what the caller
		// asked for.
		expectErr(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'CSV', compression = 'SNAPPY')
             AS SELECT 1 AS id`, scheme), "snappy")
	})

	t.Run("use_avro_logical_types names AVRO format gap", func(t *testing.T) {
		// `use_avro_logical_types` is a real BigQuery option but only
		// meaningful with AVRO output, which the encoder does not
		// implement. The diagnostic should name the option (so the
		// caller knows their input was understood) AND the format
		// (so they know which gap to file).
		expectErr(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'CSV', use_avro_logical_types = true)
             AS SELECT 1 AS id`, scheme), "use_avro_logical_types")
	})

	t.Run("WITH CONNECTION is rejected", func(t *testing.T) {
		// EXPORT DATA WITH CONNECTION `proj.region.conn` routes Omni
		// exports through a caller-supplied IAM/credential binding.
		// Silently ignoring the connection would route the export to
		// whatever the URIWriter for the scheme picks by default,
		// which is almost never what the caller meant.
		expectErr(t, fmt.Sprintf(
			`EXPORT DATA WITH CONNECTION ` + "`proj.region.conn`" + ` OPTIONS(uri = '%s://x/y', format = 'CSV')
             AS SELECT 1 AS id`, scheme), "connection")
	})
}

// TestExportDataMemRoundTrip exercises the full EXPORT DATA path —
// OPTIONS parsing, the inner query, format encoding, scheme dispatch,
// writer Close — without any cloud dependency, by registering a
// test-local in-memory URIWriter for the duration of the test and
// asserting the bytes it captured. The subtests cover every OPTIONS
// field that influences the emitted bytes (format, header,
// field_delimiter, compression) so a regression in option plumbing
// surfaces as a byte-level mismatch, not as silently-correct-looking
// output that ignores the caller's intent.
func TestExportDataMemRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		options string  // body of OPTIONS(...) less the uri
		key     string  // expected capture key (host/path of the uri)
		decode  func(t *testing.T, raw []byte) []byte
		want    string
	}{
		{
			name:    "csv default",
			options: `format = 'CSV'`,
			key:     "target/out.csv",
			want:    "id,name\n1,alice\n2,bob\n",
		},
		{
			name:    "csv header suppressed",
			options: `format = 'CSV', header = false`,
			key:     "target/out.csv",
			want:    "1,alice\n2,bob\n",
		},
		{
			name:    "csv field_delimiter pipe",
			options: `format = 'CSV', field_delimiter = '|'`,
			key:     "target/out.csv",
			want:    "id|name\n1|alice\n2|bob\n",
		},
		{
			name:    "csv gzip compression",
			options: `format = 'CSV', compression = 'GZIP'`,
			key:     "target/out.csv.gz",
			decode:  gunzip,
			want:    "id,name\n1,alice\n2,bob\n",
		},
		{
			name:    "newline-delimited json",
			options: `format = 'JSON'`,
			key:     "target/out.json",
			want:    "{\"id\":1,\"name\":\"alice\"}\n{\"id\":2,\"name\":\"bob\"}\n",
		},
		{
			name:    "json gzip compression",
			options: `format = 'JSON', compression = 'GZIP'`,
			key:     "target/out.json.gz",
			decode:  gunzip,
			want:    "{\"id\":1,\"name\":\"alice\"}\n{\"id\":2,\"name\":\"bob\"}\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheme, capture := registerMemScheme(t)
			ctx := context.Background()
			db, err := sql.Open("googlesqlite", ":memory:?_test=exportdatamem")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			conn, err := db.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			rows, err := conn.QueryContext(ctx, fmt.Sprintf(
				`EXPORT DATA OPTIONS(uri = '%s://%s', %s)
                 AS SELECT 1 AS id, 'alice' AS name UNION ALL SELECT 2, 'bob'`,
				scheme, tc.key, tc.options,
			))
			if err != nil {
				t.Fatalf("exec: %v", err)
			}
			defer rows.Close()
			if rows.Next() {
				t.Fatal("EXPORT DATA returned a row; want none")
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("rows.Err: %v", err)
			}

			raw, ok := capture.get(tc.key)
			if !ok {
				t.Fatalf("captured no bytes at %s (have %v)", tc.key, capture.keys())
			}
			got := raw
			if tc.decode != nil {
				got = tc.decode(t, raw)
			}
			if string(got) != tc.want {
				t.Errorf("captured = %q; want %q", got, tc.want)
			}
		})
	}
}

// TestExportDataOverwriteOption checks that the `overwrite` option
// reaches the URIWriter via ExportWriterOpts. The mem writer here
// refuses a second open at the same key unless opts.Overwrite is
// true — modelling the GCS Conditions{DoesNotExist:true} the
// built-in gs:// writer attaches when overwrite=false. The test
// fails if EXPORT DATA strips the option somewhere between the
// analyzer and the writer (which would silently overwrite real
// objects in production).
func TestExportDataOverwriteOption(t *testing.T) {
	scheme := "mem-overwrite"
	var (
		mu     sync.Mutex
		exists = map[string]bool{}
	)
	googlesqlite.RegisterExportURIWriter(scheme, func(_ context.Context, uri string, opts googlesqlite.ExportWriterOpts) (io.WriteCloser, error) {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, err
		}
		key := path.Join(u.Host, u.Path)
		mu.Lock()
		defer mu.Unlock()
		if exists[key] && !opts.Overwrite {
			return nil, fmt.Errorf("destination %s exists and overwrite=false", key)
		}
		exists[key] = true
		return nopWriteCloser{}, nil
	})
	t.Cleanup(func() { googlesqlite.RegisterExportURIWriter(scheme, nil) })

	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:?_test=exportoverwrite")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	stmt := fmt.Sprintf(
		`EXPORT DATA OPTIONS(uri = '%s://b/out.csv', format = 'CSV', overwrite = %%s)
         AS SELECT 1 AS id`, scheme)

	// First write — fresh destination; should succeed regardless of
	// the overwrite flag.
	r1, err := conn.QueryContext(ctx, fmt.Sprintf(stmt, "false"))
	if err != nil {
		t.Fatalf("first export: %v", err)
	}
	r1.Close()

	// Second write at the same URI with overwrite=false — the writer
	// is supposed to refuse. If the option is silently dropped the
	// writer would see opts.Overwrite=false but already-exists=true
	// and the inner refusal still fires, so this also fails. The bug
	// we are guarding against is the opposite: an opts.Overwrite=true
	// value silently being downgraded to false (or vice versa). Cover
	// both directions below.
	r2, err := conn.QueryContext(ctx, fmt.Sprintf(stmt, "false"))
	if r2 != nil {
		r2.Close()
	}
	if err == nil {
		t.Fatal("second EXPORT DATA with overwrite=false should fail; URIWriter returned no error so the option must have been silently flipped to true")
	}

	// Same destination, this time with overwrite=true — must succeed.
	r3, err := conn.QueryContext(ctx, fmt.Sprintf(stmt, "true"))
	if err != nil {
		t.Fatalf("EXPORT DATA with overwrite=true should succeed: %v", err)
	}
	r3.Close()
}

func gunzip(t *testing.T, raw []byte) []byte {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("gzip header: %v", err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	return out
}

func containsFold(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			a, b := haystack[i+j], needle[j]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
