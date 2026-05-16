package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// frontmatter is the YAML-ish header on every spec markdown file.
//
// We use a hand-rolled writer so the output is stable and obvious; the
// fields are simple scalars and we never need YAML's full grammar.
type frontmatter struct {
	Name        string
	Dialect     string // "googlesql", "bigquery", "spanner"
	Category    string // e.g. "functions/string", "types"
	Status      string // "drafted", "reviewed", "tested", "implemented", "partial", "unsupported"
	SourceURL   string // upstream snapshot URL (docs/third_party-anchored)
	UpstreamURL string // public-facing cloud.google.com URL
	LastSynced  string // YYYY-MM-DD
	Testdata    string // path to the testdata YAML
	Notes       string // optional, required when Status is partial or unsupported
}

func (f frontmatter) writeTo(w io.Writer) error {
	bw := bufio.NewWriter(w)
	if _, err := bw.WriteString("---\n"); err != nil {
		return err
	}
	pairs := []struct{ k, v string }{
		{"name", f.Name},
		{"dialect", f.Dialect},
		{"category", f.Category},
		{"status", f.Status},
		{"source_url", f.SourceURL},
		{"upstream_url", f.UpstreamURL},
		{"last_synced", f.LastSynced},
		{"testdata", f.Testdata},
	}
	for _, p := range pairs {
		if _, err := fmt.Fprintf(bw, "%s: %s\n", p.k, p.v); err != nil {
			return err
		}
	}
	if f.Notes != "" {
		if _, err := fmt.Fprintf(bw, "notes: %s\n", f.Notes); err != nil {
			return err
		}
	}
	if _, err := bw.WriteString("---\n"); err != nil {
		return err
	}
	return bw.Flush()
}

// section represents one H2 chunk extracted from an upstream markdown file.
type section struct {
	heading string // the raw H2 line, e.g. "## `CONCAT`"
	body    string // everything between this H2 (exclusive) and the next H2
	startLn int    // 1-based line number of the heading in the source file
}

// splitH2 splits the contents of a markdown file into H2 sections. The
// preamble (everything before the first H2) is returned separately.
func splitH2(src string) (preamble string, sections []section) {
	lines := strings.Split(src, "\n")
	var (
		preLines []string
		curHead  string
		curStart int
		curBody  []string
		flush    = func() {
			if curHead == "" {
				return
			}
			sections = append(sections, section{
				heading: curHead,
				body:    strings.Join(curBody, "\n"),
				startLn: curStart,
			})
			curBody = nil
		}
	)
	for i, line := range lines {
		ln := i + 1
		if isH2(line) {
			if curHead != "" {
				flush()
			}
			curHead = line
			curStart = ln
			continue
		}
		if curHead == "" {
			preLines = append(preLines, line)
		} else {
			curBody = append(curBody, line)
		}
	}
	flush()
	preamble = strings.Join(preLines, "\n")
	return preamble, sections
}

func isH2(line string) bool {
	if !strings.HasPrefix(line, "## ") {
		return false
	}
	if strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "#### ") {
		return false
	}
	return true
}

// extractFunctionName returns the function name from an H2 like
// "## `CONCAT`" or "## `REGEXP_MATCH` (Deprecated)" — the first
// backtick-quoted token after the ##.
//
// Returns "" when the heading is not a backtick-quoted function name.
func extractFunctionName(heading string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(heading, "##"))
	if !strings.HasPrefix(rest, "`") {
		return ""
	}
	rest = rest[1:]
	before, _, ok := strings.Cut(rest, "`")
	if !ok {
		return ""
	}
	return strings.TrimSpace(before)
}

// extractTypeName returns the canonical name of a data-type heading like
// "## Array type" or "## Numeric types ". It returns the leading words
// up to the trailing "type"/"types" token, lowercased and joined by
// spaces.
//
// Returns "" when the heading is not a "<words> type[s]" entry.
func extractTypeName(heading string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(heading, "##"))
	rest = strings.TrimSpace(rest)
	lower := strings.ToLower(rest)
	switch {
	case strings.HasSuffix(lower, " type"):
		return strings.TrimSpace(rest[:len(rest)-len(" type")])
	case strings.HasSuffix(lower, " types"):
		return strings.TrimSpace(rest[:len(rest)-len(" types")])
	}
	return ""
}

// slugify produces a filesystem-safe filename for a function or type
// name. It lowercases the input and replaces every run of non-alnum
// characters with a single underscore.
func slugify(name string) string {
	var (
		b        strings.Builder
		lastDash bool
	)
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('_')
				lastDash = true
			}
		}
	}
	out := b.String()
	return strings.Trim(out, "_")
}

// upstreamAnchor builds a GitHub anchor URL that points at the H2 we
// originated from in the upstream snapshot.
//
// GitHub's anchor algorithm: lowercase, replace spaces with hyphens,
// strip non-alphanumeric except hyphens. We approximate that here.
func upstreamAnchor(repoBlob, file, heading string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(heading, "##"))
	rest = strings.TrimSpace(rest)
	// Strip backticks.
	rest = strings.ReplaceAll(rest, "`", "")
	// GitHub anchor: lowercase, hyphenated.
	var b strings.Builder
	for _, r := range strings.ToLower(rest) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_':
			b.WriteRune(r)
		}
	}
	return fmt.Sprintf("%s/%s#%s", repoBlob, file, b.String())
}
