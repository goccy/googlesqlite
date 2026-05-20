package exportdata

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// init registers the built-in `gs://` writer. Drivers that want a different
// GCS stack can override it by calling RegisterURIWriter("gs", custom) after
// import.
func init() {
	RegisterURIWriter("gs", openGCSWriter)
}

// openGCSWriter opens a write stream against a `gs://bucket/object` URI.
//
// When STORAGE_EMULATOR_HOST is set (the convention every Google Cloud Go
// client follows for the standard fake-gcs-server / cloud emulators), the
// client is rewired at that endpoint with authentication disabled. Without
// it, the standard Application Default Credentials apply — i.e. the same
// behaviour as a plain `storage.NewClient(ctx)`.
//
// The `*` placeholder real BigQuery uses for shard numbers (a single export
// can produce many objects, one per shard) is collapsed to the 12-digit
// shard identifier BigQuery uses for the first shard, so a one-shard write
// against `gs://b/out/*.csv` lands at `gs://b/out/000000000000.csv`.
//
// WriterOpts.Overwrite=false (the BigQuery default) is enforced by a
// DoesNotExist precondition on the object; the write fails with HTTP 412
// PreconditionFailed if the object already exists. Overwrite=true skips the
// precondition so the object is replaced unconditionally.
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

	var clientOpts []option.ClientOption
	if host := os.Getenv("STORAGE_EMULATOR_HOST"); host != "" {
		clientOpts = append(clientOpts, option.WithEndpoint(host), option.WithoutAuthentication())
	}
	client, err := storage.NewClient(ctx, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("exportdata: open GCS client: %w", err)
	}
	obj := client.Bucket(bucket).Object(object)
	if !opts.Overwrite {
		// DoesNotExist makes the write atomic against concurrent
		// creators and matches BigQuery's `overwrite = false` default
		// (the statement must fail when the destination already exists
		// instead of silently clobbering it).
		obj = obj.If(storage.Conditions{DoesNotExist: true})
	}
	w := obj.NewWriter(ctx)
	return &gcsObjectWriter{Writer: w, client: client}, nil
}

// gcsObjectWriter pairs the per-object storage.Writer with the underlying
// storage.Client so closing the writer also releases the client.
type gcsObjectWriter struct {
	*storage.Writer
	client *storage.Client
}

func (w *gcsObjectWriter) Close() error {
	werr := w.Writer.Close()
	cerr := w.client.Close()
	if werr != nil {
		return werr
	}
	return cerr
}
