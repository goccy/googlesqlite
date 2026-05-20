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
	"strings"
	"sync"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/goccy/googlesqlite"
	"google.golang.org/api/option"
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

// TestExportDataStatement asserts that an `EXPORT DATA OPTIONS(...) AS
// <query>` statement (i) actually writes the inner query's rows to the
// destination URI in the requested format and (ii) returns no rows to the
// caller. The destination is a `gs://` URI backed by fake-gcs-server, so
// the test covers the full path through the built-in GCS writer.
func TestExportDataStatement(t *testing.T) {
	const (
		bucket     = "test-export-bucket"
		publicHost = "127.0.0.1"
	)
	storageServer, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		PublicHost: publicHost,
		Scheme:     "http",
	})
	if err != nil {
		t.Fatalf("fake-gcs-server: %v", err)
	}
	defer storageServer.Stop()
	storageServer.CreateBucket(bucket)
	u, err := url.Parse(storageServer.URL())
	if err != nil {
		t.Fatal(err)
	}
	emulatorHost := fmt.Sprintf("http://%s:%s", publicHost, u.Port())
	t.Setenv("STORAGE_EMULATOR_HOST", emulatorHost)

	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:?_test=exportdata")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close()

	gcs, err := storage.NewClient(ctx,
		option.WithEndpoint(emulatorHost),
		option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("storage.NewClient: %v", err)
	}
	defer gcs.Close()

	exec := func(t *testing.T, sqlText string) {
		t.Helper()
		rows, err := conn.QueryContext(ctx, sqlText)
		if err != nil {
			t.Fatalf("exec: %v", err)
		}
		defer rows.Close()
		// EXPORT DATA must yield zero result rows.
		if rows.Next() {
			t.Fatal("EXPORT DATA returned at least one row; want none")
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows.Err: %v", err)
		}
	}

	readObject := func(t *testing.T, name string) string {
		t.Helper()
		r, err := gcs.Bucket(bucket).Object(name).NewReader(ctx)
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer r.Close()
		b, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return string(b)
	}

	t.Run("CSV", func(t *testing.T) {
		exec(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = 'gs://%s/csv/*.csv', format = 'CSV')
             AS SELECT 1 AS id, 'a' AS name UNION ALL SELECT 2, 'b'`,
			bucket,
		))
		// `*` is the per-shard placeholder; the engine writes a single
		// shard at BigQuery's default 12-digit id.
		got := readObject(t, "csv/000000000000.csv")
		want := "id,name\n1,a\n2,b\n"
		if got != want {
			t.Errorf("csv body = %q; want %q", got, want)
		}
	})

	t.Run("JSON", func(t *testing.T) {
		exec(t, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = 'gs://%s/json/out.json', format = 'JSON')
             AS SELECT 1 AS id, 'a' AS name UNION ALL SELECT 2, 'b'`,
			bucket,
		))
		got := readObject(t, "json/out.json")
		// json.Encoder writes one object per line in deterministic field
		// order (map iteration in Go is randomized, but the encoder
		// sorts map keys — so the order is alphabetical: id, name).
		want := `{"id":1,"name":"a"}` + "\n" + `{"id":2,"name":"b"}` + "\n"
		if got != want {
			t.Errorf("json body = %q; want %q", got, want)
		}
	})

	t.Run("missing uri is rejected", func(t *testing.T) {
		_, err := conn.QueryContext(ctx,
			`EXPORT DATA OPTIONS(format = 'CSV') AS SELECT 1 AS id`)
		if err == nil {
			t.Fatal("EXPORT DATA without `uri` should fail")
		}
		if !strings.Contains(err.Error(), "uri") {
			t.Fatalf("error %q does not mention `uri`", err)
		}
	})

	t.Run("unsupported format is rejected", func(t *testing.T) {
		_, err := conn.QueryContext(ctx, fmt.Sprintf(
			`EXPORT DATA OPTIONS(uri = 'gs://%s/x/out', format = 'AVRO')
             AS SELECT 1 AS id`,
			bucket,
		))
		if err == nil {
			t.Fatal("EXPORT DATA with unsupported format should fail")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "avro") {
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
		if !strings.Contains(strings.ToLower(err.Error()), "s3") {
			t.Fatalf("error %q does not name the missing scheme", err)
		}
	})
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
