package googlesqlite_test

import (
	"bytes"
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
	googlesqlite.RegisterExportURIWriter(scheme, func(_ context.Context, uri string) (io.WriteCloser, error) {
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

	t.Run("missing uri is rejected", func(t *testing.T) {
		_, err := conn.QueryContext(ctx,
			`EXPORT DATA OPTIONS(format = 'CSV') AS SELECT 1 AS id`)
		if err == nil {
			t.Fatal("EXPORT DATA without `uri` should fail")
		}
		if !containsFold(err.Error(), "uri") {
			t.Fatalf("error %q does not mention `uri`", err)
		}
	})

	t.Run("unsupported format is rejected", func(t *testing.T) {
		_, err := conn.QueryContext(ctx, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'AVRO')
             AS SELECT 1 AS id`, scheme))
		if err == nil {
			t.Fatal("EXPORT DATA with unsupported format should fail")
		}
		if !containsFold(err.Error(), "avro") {
			t.Fatalf("error %q does not name the offending format", err)
		}
	})

	t.Run("unregistered scheme is rejected", func(t *testing.T) {
		_, err := conn.QueryContext(ctx,
			`EXPORT DATA OPTIONS(uri = 's3://x/y', format = 'CSV')
             AS SELECT 1 AS id`)
		if err == nil {
			t.Fatal("EXPORT DATA against unregistered scheme should fail")
		}
		if !containsFold(err.Error(), "s3") {
			t.Fatalf("error %q does not name the missing scheme", err)
		}
	})

	t.Run("unknown option is rejected", func(t *testing.T) {
		// `overwrite` is a real BigQuery EXPORT DATA option, but the
		// engine does not honor it today. Silently dropping it would
		// let `overwrite = false` SQL pass through and overwrite the
		// destination anyway, with no signal that the option had no
		// effect — so unknown / unhonored options must fail explicitly.
		_, err := conn.QueryContext(ctx, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = '%s://x/y', format = 'CSV', overwrite = false)
             AS SELECT 1 AS id`, scheme))
		if err == nil {
			t.Fatal("EXPORT DATA with an unsupported option should fail")
		}
		if !containsFold(err.Error(), "overwrite") {
			t.Fatalf("error %q does not name the unsupported option", err)
		}
	})

	t.Run("known option as non-literal expression is rejected", func(t *testing.T) {
		// Concatenating the URI is a perfectly valid SQL expression
		// but the option-reader only consumes string LITERALS today
		// (the analyzer does not constant-fold OPTIONS expressions for
		// us). The failure must call out the offending option by name,
		// not drop the value and surface as "required option `uri` is
		// missing" — that would mislead the caller about what is
		// wrong.
		_, err := conn.QueryContext(ctx, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = CONCAT('%s://', 'x/y'), format = 'CSV')
             AS SELECT 1 AS id`, scheme))
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
}

// TestExportDataMemRoundTrip exercises the full EXPORT DATA path —
// OPTIONS parsing, the inner query, format encoding, scheme dispatch,
// writer Close — without any cloud dependency, by registering a
// test-local in-memory URIWriter for the duration of the test and
// asserting the bytes it captured.
func TestExportDataMemRoundTrip(t *testing.T) {
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
		`EXPORT DATA OPTIONS(uri = '%s://target/out.csv', format = 'CSV')
         AS SELECT 1 AS id, 'alice' AS name UNION ALL SELECT 2, 'bob'`,
		scheme,
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

	got, ok := capture.get("target/out.csv")
	if !ok {
		t.Fatalf("captured no bytes at target/out.csv (have %v)", capture.keys())
	}
	if want := "id,name\n1,alice\n2,bob\n"; string(got) != want {
		t.Errorf("captured = %q; want %q", got, want)
	}
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
