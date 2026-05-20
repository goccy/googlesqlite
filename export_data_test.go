package googlesqlite_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/goccy/googlesqlite"
)

// memExportStore captures bytes written to the test-only `mem://` scheme so
// the spec suite can exercise EXPORT DATA end to end without a real cloud
// dependency. Each closed writer publishes its full payload keyed by
// `<host>/<path>` so tests can inspect it.
var (
	memExportStoreMu sync.Mutex
	memExportStore   = map[string][]byte{}
)

// init registers the test-only `mem://` writer. _test.go init runs only in
// the test binary, so the registration cannot escape into production
// callers of the googlesqlite package.
func init() {
	googlesqlite.RegisterExportURIWriter("mem", func(_ context.Context, uri string) (io.WriteCloser, error) {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("mem://: invalid uri %q: %w", uri, err)
		}
		key := path.Join(u.Host, u.Path)
		return &memWriteCloser{key: key, buf: &bytes.Buffer{}}, nil
	})
}

type memWriteCloser struct {
	key string
	buf *bytes.Buffer
}

func (w *memWriteCloser) Write(p []byte) (int, error) { return w.buf.Write(p) }

func (w *memWriteCloser) Close() error {
	memExportStoreMu.Lock()
	defer memExportStoreMu.Unlock()
	memExportStore[w.key] = append([]byte(nil), w.buf.Bytes()...)
	return nil
}

// TestExportDataErrorPaths covers the per-statement validation that does
// not require any URI scheme to actually open: missing `uri`, unsupported
// `format`, and an unregistered scheme. These all error out before the
// EXPORT DATA action tries to open a writer, so they run anywhere
// (including under the race detector).
func TestExportDataErrorPaths(t *testing.T) {
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
		_, err := conn.QueryContext(ctx,
			`EXPORT DATA OPTIONS(uri = 'mem://x/y', format = 'AVRO')
             AS SELECT 1 AS id`)
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

// TestExportDataMemScheme guards the spec-suite path: the `mem://` writer
// registered in init() must capture the bytes for the spec EXPORT DATA case.
func TestExportDataMemScheme(t *testing.T) {
	// Sanity check that init() landed and the spec-only sink is usable.
	if _, err := os.Stat("testdata/specs/googlesql/syntax/io/export_data.yaml"); err != nil {
		t.Skipf("export_data spec not present: %v", err)
	}
	memExportStoreMu.Lock()
	had := len(memExportStore) > 0
	memExportStoreMu.Unlock()
	if had {
		// TestSpec may have run before this and already populated the
		// store, in which case the registration is obviously working.
		return
	}
	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:?_test=memexport")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn, _ := db.Conn(ctx)
	defer conn.Close()

	rows, err := conn.QueryContext(ctx,
		`EXPORT DATA OPTIONS(uri = 'mem://probe/out.csv', format = 'CSV')
         AS SELECT 1 AS x`)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("EXPORT DATA returned a row")
	}

	memExportStoreMu.Lock()
	defer memExportStoreMu.Unlock()
	if _, ok := memExportStore["probe/out.csv"]; !ok {
		t.Fatalf("mem:// writer did not capture the export (keys: %v)", keysOf(memExportStore))
	}
}

func keysOf(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
