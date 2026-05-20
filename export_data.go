package googlesqlite

import (
	"github.com/goccy/googlesqlite/internal/exportdata"
)

// ExportURIWriter opens an io.WriteCloser for an EXPORT DATA destination
// URI. The engine uses it to materialize the rows produced by an
// `EXPORT DATA OPTIONS(uri = '...', format = '...', ...) AS <query>`
// statement. See RegisterExportURIWriter.
type ExportURIWriter = exportdata.URIWriter

// ExportWriterOpts carries the EXPORT DATA OPTIONS that influence how the
// destination is opened. Implementations of ExportURIWriter should honor
// each field where it is meaningful to their scheme and ignore the rest.
type ExportWriterOpts = exportdata.WriterOpts

// RegisterExportURIWriter installs w as the writer for EXPORT DATA URIs
// whose scheme matches scheme (case-insensitive). The "gs" scheme is
// registered by default and writes to Google Cloud Storage (honoring
// STORAGE_EMULATOR_HOST so a fake-gcs-server instance is reachable
// without any extra wiring). Register additional schemes — typically
// "s3" or "azure" for BigQuery Omni destinations — when the application
// targets them; an EXPORT DATA against an unregistered scheme fails with
// a descriptive error rather than silently dropping the export.
//
// Passing a nil writer unregisters the scheme. Registering an
// already-registered scheme replaces the previous writer; this is by
// design so an application can swap out the built-in.
func RegisterExportURIWriter(scheme string, w ExportURIWriter) {
	exportdata.RegisterURIWriter(scheme, w)
}
