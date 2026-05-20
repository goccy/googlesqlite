// Package exportdata holds the EXPORT DATA execution machinery: the URI
// writer registry, built-in writers (gs://), the format encoders and the
// compression wrappers the driver uses to materialize an
// `EXPORT DATA OPTIONS(...) AS SELECT ...` statement to its destination.
//
// The public API surface — the URIWriter / WriterOpts types and
// RegisterURIWriter — is re-exported from the root googlesqlite package so
// callers do not need to import internal/ paths.
package exportdata

import (
	"context"
	"io"
	"strings"
	"sync"
)

// WriterOpts carries scheme-independent EXPORT DATA OPTIONS that influence
// how the destination is opened. URIWriter implementations should honor
// each field where it is meaningful for their scheme and ignore the rest.
type WriterOpts struct {
	// Overwrite controls what happens when the destination already
	// exists. The BigQuery default is false (the statement fails when
	// the destination is present); set true to replace any existing
	// object atomically.
	Overwrite bool
}

// URIWriter opens an io.WriteCloser for an EXPORT DATA destination URI.
// Implementations should honor the URI's scheme, resolve any
// environment-driven host overrides (e.g. STORAGE_EMULATOR_HOST for gs://)
// themselves, and apply the WriterOpts fields meaningful to that scheme.
// Closing the returned writer is the caller's responsibility and is what
// commits the bytes — the contract follows io.WriteCloser semantics.
type URIWriter func(ctx context.Context, uri string, opts WriterOpts) (io.WriteCloser, error)

var (
	writersMu sync.RWMutex
	writers   = map[string]URIWriter{}
)

// RegisterURIWriter installs w as the writer for URIs whose scheme matches
// scheme (case-insensitive — schemes are compared lower-cased). Passing a
// nil writer unregisters the scheme. Registering an already-registered
// scheme replaces the previous writer; this is intentional so a driver can
// override a built-in (e.g. point gs:// at a different storage stack).
func RegisterURIWriter(scheme string, w URIWriter) {
	writersMu.Lock()
	defer writersMu.Unlock()
	key := strings.ToLower(scheme)
	if w == nil {
		delete(writers, key)
		return
	}
	writers[key] = w
}

// LookupURIWriter returns the writer registered for the given scheme, or nil
// when none is registered. ExportDataStmtAction uses this to dispatch the
// open at execution time so the registry can be populated after Init.
func LookupURIWriter(scheme string) URIWriter {
	writersMu.RLock()
	defer writersMu.RUnlock()
	return writers[strings.ToLower(scheme)]
}
