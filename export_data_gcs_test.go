//go:build !race

// fake-gcs-server spins up an in-process HTTP server whose handler
// goroutines combine with wazero's signal-driven stack growth to trip
// Go's race detector with "fatal: bad g in signal handler" before any
// test logic runs. The crash is upstream (wazero + the race runtime),
// not in the EXPORT DATA path under test, so this end-to-end gs:// case
// is gated behind !race. The behaviour under -race is still covered:
// TestExportDataErrorPaths exercises the OPTIONS validation paths, and
// TestW_ExportData / TestSpec route through the `mem://` writer
// registered in export_data_test.go.

package googlesqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"google.golang.org/api/option"
)

// TestExportDataStatementGCS asserts that an `EXPORT DATA OPTIONS(uri =
// 'gs://...', format = '...') AS <query>` statement writes the inner
// query's rows to the destination URI in the requested format and returns
// no rows to the caller. The destination is a `gs://` URI backed by
// fake-gcs-server, exercising the full path through the built-in GCS
// writer (the standard cloud.google.com/go/storage client wired at
// STORAGE_EMULATOR_HOST).
func TestExportDataStatementGCS(t *testing.T) {
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
	storageServer.CreateBucketWithOpts(fakestorage.CreateBucketOpts{Name: bucket})
	u, err := url.Parse(storageServer.URL())
	if err != nil {
		t.Fatal(err)
	}
	emulatorHost := fmt.Sprintf("http://%s:%s", publicHost, u.Port())
	t.Setenv("STORAGE_EMULATOR_HOST", emulatorHost)

	ctx := context.Background()
	db, err := sql.Open("googlesqlite", ":memory:?_test=exportdatagcs")
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
		if got, want := readObject(t, "csv/000000000000.csv"), "id,name\n1,a\n2,b\n"; got != want {
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
		// json.Encoder serializes map keys in alphabetical order, so
		// `id` comes before `name` in the encoded body.
		want := `{"id":1,"name":"a"}` + "\n" + `{"id":2,"name":"b"}` + "\n"
		if got != want {
			t.Errorf("json body = %q; want %q", got, want)
		}
	})
}
