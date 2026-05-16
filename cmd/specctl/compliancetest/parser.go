// Package compliancetest parses the GoogleSQL compliance "*.test"
// fixture format.
//
// Each file holds a sequence of cases separated by "==" lines. A
// single case is roughly:
//
//	# Optional comment block.
//	[name=<id>]
//	[required_features=FEAT_A,FEAT_B]   (optional, repeatable)
//	<SQL statement on one or more lines>
//	--
//	<expected value in GoogleSQL value notation, multi-line>
//
// The expected block typically takes the shape
//
//	ARRAY<STRUCT<TYPE, TYPE, ...>>[{value, value, ...}, ...]
//
// or
//
//	ERROR: <kind>: <message>
//
// This package extracts the structural pieces. Translating the value
// notation into testdata YAML rows is a separate concern; see
// cmd/specctl/extract_compliance.go for the policy of which cases are
// safe to emit as rows-equality assertions.
package compliancetest

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// Case is a single test case parsed from a .test file.
type Case struct {
	// Name from `[name=...]`. Empty when the file omitted it.
	Name string
	// Features required to run this case, gathered from
	// `[required_features=...]` lines. Multiple lines unionised.
	Features []string
	// Other `[key=value]` attributes (e.g. parameters, default_time_zone).
	Attrs map[string]string
	// Comment lines that preceded the case (without their leading "# ").
	Comment string
	// SQL block between the headers and the `--` separator, trimmed.
	SQL string
	// Expected block after `--`, trimmed. Multi-line preserved.
	Expected string
}

// ParseFile reads path and returns every case.
func ParseFile(path string) ([]Case, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// Parse reads from r and returns every case. The file is treated as
// UTF-8 with `\n` line endings; CRLF is tolerated by trimming `\r`
// suffixes.
func Parse(r io.Reader) ([]Case, error) {
	const sep = "=="

	br := bufio.NewScanner(r)
	br.Buffer(make([]byte, 1<<16), 1<<22)
	var (
		out     []Case
		raw     []string
		caseBuf []string
	)
	for br.Scan() {
		line := strings.TrimRight(br.Text(), "\r")
		raw = append(raw, line)
		if strings.TrimSpace(line) == sep {
			out = appendCaseIfAny(out, caseBuf)
			caseBuf = nil
			continue
		}
		caseBuf = append(caseBuf, line)
	}
	if err := br.Err(); err != nil {
		return nil, err
	}
	out = appendCaseIfAny(out, caseBuf)
	if len(out) == 0 && len(raw) > 0 {
		return nil, errors.New("compliancetest: no cases parsed; file may be malformed")
	}
	return out, nil
}

func appendCaseIfAny(out []Case, lines []string) []Case {
	c, ok := parseCase(lines)
	if !ok {
		return out
	}
	return append(out, c)
}

// parseCase turns one chunk (lines between `==` boundaries) into a
// Case. Returns ok=false when the chunk has no SQL+expected pair.
func parseCase(lines []string) (Case, bool) {
	var c Case
	c.Attrs = map[string]string{}

	// Collect leading comment + header lines until we hit the first
	// non-empty line that is not a comment or attribute marker.
	idx := 0
	var commentLines []string
	for idx < len(lines) {
		l := strings.TrimSpace(lines[idx])
		if l == "" {
			idx++
			continue
		}
		if strings.HasPrefix(l, "#") {
			commentLines = append(commentLines, strings.TrimSpace(strings.TrimPrefix(l, "#")))
			idx++
			continue
		}
		if strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]") {
			body := l[1 : len(l)-1]
			eq := strings.Index(body, "=")
			if eq < 0 {
				idx++
				continue
			}
			k := strings.TrimSpace(body[:eq])
			v := strings.TrimSpace(body[eq+1:])
			switch k {
			case "name":
				c.Name = v
			case "required_features":
				c.Features = append(c.Features, splitCSV(v)...)
			default:
				c.Attrs[k] = v
			}
			idx++
			continue
		}
		break
	}
	c.Comment = strings.Join(commentLines, "\n")

	// Find the `--` separator within the rest.
	sepIdx := -1
	for j := idx; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == "--" {
			sepIdx = j
			break
		}
	}
	if sepIdx < 0 {
		return Case{}, false
	}
	c.SQL = strings.TrimSpace(strings.Join(lines[idx:sepIdx], "\n"))
	c.Expected = strings.TrimSpace(strings.Join(lines[sepIdx+1:], "\n"))
	if c.SQL == "" {
		return Case{}, false
	}
	return c, true
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
