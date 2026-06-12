package exportdata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// init registers the built-in `gs://` writer. Drivers that want a different
// GCS stack can override it by calling RegisterURIWriter("gs", custom) after
// import.
func init() {
	RegisterURIWriter("gs", openGCSWriter)
}

// gcsWriteScope is the OAuth 2.0 scope required to PUT objects into a GCS
// bucket. Hard-coded here so the GCS writer can call into ADC
// (`oauth2/google.DefaultTokenSource`) without dragging in the full
// `cloud.google.com/go/storage` stack (grpc → envoyproxy →
// opentelemetry, ~120 MB live in the linked binary) for the constant.
const gcsWriteScope = "https://www.googleapis.com/auth/devstorage.read_write"

// gcsProductionHost is the canonical public endpoint for the GCS JSON API.
// Overridable through STORAGE_EMULATOR_HOST — the convention every Google
// Cloud client follows for the fake-gcs-server / cloud emulators flow.
const gcsProductionHost = "https://storage.googleapis.com"

// openGCSWriter opens a write stream against a `gs://bucket/object` URI.
//
// When STORAGE_EMULATOR_HOST is set (the convention every Google Cloud Go
// client follows for the standard fake-gcs-server / cloud emulators), the
// upload is rewired at that endpoint with no authentication. Without it,
// the standard Application Default Credentials chain
// (GOOGLE_APPLICATION_CREDENTIALS → gcloud config → GCE metadata server)
// is consulted via `oauth2/google.DefaultTokenSource`.
//
// The `*` placeholder real BigQuery uses for shard numbers (a single export
// can produce many objects, one per shard) is collapsed to the 12-digit
// shard identifier BigQuery uses for the first shard, so a one-shard write
// against `gs://b/out/*.csv` lands at `gs://b/out/000000000000.csv`.
//
// WriterOpts.Overwrite=false (the BigQuery default) is enforced by an
// `ifGenerationMatch=0` query parameter on the upload, which translates to
// the JSON API's "create-only" precondition: the write fails with HTTP 412
// PreconditionFailed if the object already exists. Overwrite=true skips
// the precondition so the object is replaced unconditionally.
//
// Implementation note: we used to wrap `cloud.google.com/go/storage`, which
// is the canonical client but pulls in grpc / envoyproxy / opentelemetry —
// none of which the GCS JSON upload path actually needs. Talking JSON over
// `net/http` directly drops ~120 MB of transitive deps from the binary
// without losing any of the behaviour the emulator's EXPORT DATA flow
// relies on.
func openGCSWriter(ctx context.Context, uri string, opts WriterOpts) (io.WriteCloser, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("exportdata: parse gs:// URI %q: %w", uri, err)
	}
	if u.Scheme != "gs" {
		return nil, fmt.Errorf("exportdata: not a gs:// URI: %q", uri)
	}
	bucket := u.Host
	object := strings.TrimPrefix(u.Path, "/")
	if bucket == "" || object == "" {
		return nil, fmt.Errorf("exportdata: malformed gs:// URI %q (expected gs://bucket/path)", uri)
	}
	// BigQuery uses `*` as the per-shard placeholder in the URI. The
	// emulator always writes a single shard so `*` becomes the 12-digit
	// shard identifier BigQuery uses for the first shard.
	object = strings.ReplaceAll(object, "*", "000000000000")

	host := os.Getenv("STORAGE_EMULATOR_HOST")
	noAuth := host != ""
	if host == "" {
		host = gcsProductionHost
	}

	// Build the simple-media upload URL. The JSON API's "simple" upload
	// path streams a single PUT body (`uploadType=media`), which maps
	// cleanly onto an io.Pipe — no resumable-session handshake needed.
	q := url.Values{
		"uploadType": {"media"},
		"name":       {object},
	}
	if !opts.Overwrite {
		// ifGenerationMatch=0 enforces "object must not exist" — same
		// semantics as cloud.google.com/go/storage's
		// `If(Conditions{DoesNotExist: true})`. The server returns 412
		// PreconditionFailed on overwrite.
		q.Set("ifGenerationMatch", "0")
	}
	uploadURL := fmt.Sprintf("%s/upload/storage/v1/b/%s/o?%s",
		strings.TrimRight(host, "/"),
		url.PathEscape(bucket),
		q.Encode())

	// HTTP transport: against the production GCS endpoint we wrap the
	// default transport with an oauth2 token source so every request
	// carries a fresh bearer token. Against STORAGE_EMULATOR_HOST we use
	// the unwrapped default transport — fake-gcs-server's HTTPS-disabled
	// endpoint rejects an Authorization header.
	client := http.DefaultClient
	if !noAuth {
		ts, err := google.DefaultTokenSource(ctx, gcsWriteScope)
		if err != nil {
			return nil, fmt.Errorf("exportdata: GCS DefaultTokenSource: %w", err)
		}
		client = oauth2.NewClient(ctx, ts)
	}

	// Stream the body through an io.Pipe. The caller writes to pw; a
	// background goroutine consumes pr and feeds it to the HTTP PUT.
	// Close() on the writer signals EOF to the goroutine and waits for
	// the request's status, so an HTTP-side error (auth, precondition,
	// network) surfaces synchronously to the EXPORT DATA statement.
	pr, pw := io.Pipe()
	done := make(chan error, 1)
	go func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, pr)
		if err != nil {
			done <- fmt.Errorf("exportdata: build GCS upload request: %w", err)
			// Drain the pipe so the writing goroutine does not block.
			_, _ = io.Copy(io.Discard, pr)
			return
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		resp, err := client.Do(req)
		if err != nil {
			done <- fmt.Errorf("exportdata: GCS upload: %w", err)
			_, _ = io.Copy(io.Discard, pr)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			// Read up to 4 KiB of the error body so the failing
			// EXPORT DATA statement sees the GCS error message.
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			done <- fmt.Errorf("exportdata: GCS upload %s: %s",
				resp.Status, strings.TrimSpace(string(body)))
			return
		}
		// Drain any remaining body so the HTTP connection can be
		// returned to the pool.
		_, _ = io.Copy(io.Discard, resp.Body)
		done <- nil
	}()
	return &gcsObjectWriter{pw: pw, done: done}, nil
}

// gcsObjectWriter is an io.WriteCloser fronted by an io.Pipe; Close signals
// EOF to the background uploader goroutine and waits for the HTTP request
// to land so any non-2xx response surfaces synchronously.
type gcsObjectWriter struct {
	pw   *io.PipeWriter
	done <-chan error
}

func (w *gcsObjectWriter) Write(p []byte) (int, error) {
	return w.pw.Write(p)
}

func (w *gcsObjectWriter) Close() error {
	// Closing the writer end of the pipe signals EOF to the goroutine,
	// which finishes reading and returns the request status on `done`.
	if err := w.pw.Close(); err != nil {
		return err
	}
	return <-w.done
}
