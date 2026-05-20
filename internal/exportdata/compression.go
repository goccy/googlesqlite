package exportdata

import (
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

// Compression identifies the EXPORT DATA `compression` option value applied
// on top of the format encoder.
type Compression string

const (
	CompressionNone Compression = "NONE"
	CompressionGZIP Compression = "GZIP"
)

// ParseCompression normalizes the `compression` option value against the
// chosen format. An empty string maps to NONE. Real BigQuery only documents
// NONE and GZIP for CSV / JSON exports; the format-incompatible codecs
// (SNAPPY / DEFLATE / ZSTD / LZ4 are AVRO- or PARQUET-only) and unknown
// values surface as a descriptive error rather than silently writing
// uncompressed bytes.
func ParseCompression(s string, format Format) (Compression, error) {
	c := Compression(strings.ToUpper(strings.TrimSpace(s)))
	if c == "" {
		c = CompressionNone
	}
	switch c {
	case CompressionNone:
		return CompressionNone, nil
	case CompressionGZIP:
		switch format {
		case FormatCSV, FormatNDJSON:
			return CompressionGZIP, nil
		}
		return "", fmt.Errorf("EXPORT DATA: compression GZIP is not valid for format %s", format)
	case "SNAPPY", "DEFLATE", "ZSTD", "LZ4":
		return "", fmt.Errorf("EXPORT DATA: compression %s is not supported by googlesqlite", c)
	default:
		return "", fmt.Errorf("EXPORT DATA: unknown compression %q", s)
	}
}

// WrapCompressor wraps the destination writer in the requested codec. The
// returned WriteCloser must be Closed before the underlying writer is
// closed so the compressed stream is flushed in full.
func WrapCompressor(w io.WriteCloser, c Compression) (io.WriteCloser, error) {
	switch c {
	case CompressionNone, "":
		return w, nil
	case CompressionGZIP:
		return &gzipWriteCloser{Writer: gzip.NewWriter(w), underlying: w}, nil
	}
	return nil, fmt.Errorf("EXPORT DATA: unsupported compression %q", c)
}

// gzipWriteCloser closes both the gzip writer (flushing the trailer) and
// the underlying destination so callers only need to defer one Close.
type gzipWriteCloser struct {
	*gzip.Writer
	underlying io.WriteCloser
}

func (w *gzipWriteCloser) Close() error {
	gerr := w.Writer.Close()
	uerr := w.underlying.Close()
	if gerr != nil {
		return gerr
	}
	return uerr
}
