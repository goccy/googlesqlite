package internal

import (
	"strings"
)

// applyBigQueryTvfStubs rewrites built-in BigQuery TVFs that have no
// natural local execution path into empty / minimal subqueries that
// the analyzer can resolve.
//
// The following call sites are handled by lowering the TVF to plain
// SQL BEFORE the analyzer runs (analogous to applyGapFillRewrite):
//
//	APPENDS(TABLE t, ...)              → (SELECT *, NULL AS _CHANGE_TYPE,
//	                                               NULL AS _CHANGE_TIMESTAMP
//	                                      FROM t WHERE FALSE)
//	CHANGES(TABLE t, ...)              → (SELECT *, NULL AS _CHANGE_TYPE,
//	                                               NULL AS _CHANGE_TIMESTAMP,
//	                                               NULL AS _CHANGE_SEQUENCE_NUMBER
//	                                      FROM t WHERE FALSE)
//	VECTOR_INDEX.STATISTICS(TABLE t)   → (SELECT CAST(NULL AS FLOAT64) AS drift
//	                                      WHERE FALSE)
//
// These functions report on time-travel snapshots and vector-index
// drift that googlesqlite does not maintain; the rewrite preserves
// the call-site shape so consumers can parse / analyze without
// crashing, returning an empty result set. EXTERNAL_QUERY,
// VECTOR_SEARCH, and EXTERNAL_OBJECT_TRANSFORM are not handled here
// because their output schema depends on caller-supplied information
// that no rewrite can derive.
func applyBigQueryTvfStubs(query string) string {
	upper := strings.ToUpper(query)
	if !strings.Contains(upper, "APPENDS") &&
		!strings.Contains(upper, "CHANGES") &&
		!strings.Contains(upper, "VECTOR_INDEX") {
		return query
	}
	for {
		changed := false
		query, changed = rewriteTvfCall(query, "APPENDS", buildAppendsBody)
		if changed {
			continue
		}
		query, changed = rewriteTvfCall(query, "CHANGES", buildChangesBody)
		if changed {
			continue
		}
		query, changed = rewriteTvfCall(query, "VECTOR_INDEX.STATISTICS", buildVectorIndexStatisticsBody)
		if changed {
			continue
		}
		break
	}
	return query
}

// rewriteTvfCall locates the next case-insensitive `name (` call and,
// if it matches a TVF stub, replaces the entire `name(...)` span with
// `bodyFn(args)`. The args slice contains the comma-separated raw
// token sequences from the call (whitespace-trimmed), with TABLE
// keyword preserved on the leading position.
func rewriteTvfCall(s, name string, bodyFn func(args []string) string) (string, bool) {
	for i := 0; i+len(name) <= len(s); i++ {
		// Skip string literals / quoted identifiers / line comments to
		// avoid matching the keyword inside them.
		switch s[i] {
		case '\'', '"':
			i = scanStringLiteral(s, i) - 1
			continue
		case '`':
			j := i + 1
			for j < len(s) && s[j] != '`' {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j - 1
			continue
		case '-':
			if i+1 < len(s) && s[i+1] == '-' {
				j := i
				for j < len(s) && s[j] != '\n' {
					j++
				}
				i = j - 1
				continue
			}
		}
		if !strings.EqualFold(s[i:i+len(name)], name) {
			continue
		}
		// Boundary check: previous char must not be alnum/_ (avoid
		// matching CHANGES inside MYCHANGES etc).
		if i > 0 {
			c := s[i-1]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_' || c == '.' {
				continue
			}
		}
		// Skip whitespace.
		j := i + len(name)
		for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
			j++
		}
		if j >= len(s) || s[j] != '(' {
			continue
		}
		// Locate matching close paren.
		end, args, ok := scanBalancedArgs(s, j)
		if !ok {
			continue
		}
		return s[:i] + bodyFn(args) + s[end:], true
	}
	return s, false
}

// scanBalancedArgs reads from `s[start]` (which must be '(') through
// the matching ')' and returns the index one past ')', the
// comma-split argument list (trimmed, paren-balanced), and ok=true on
// success. String literals and back-ticked identifiers are honoured.
func scanBalancedArgs(s string, start int) (int, []string, bool) {
	if start >= len(s) || s[start] != '(' {
		return 0, nil, false
	}
	depth := 0
	args := []string{}
	cur := strings.Builder{}
	for i := start; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(':
			depth++
			if depth > 1 {
				cur.WriteByte(c)
			}
		case ')':
			depth--
			if depth == 0 {
				if cur.Len() > 0 || len(args) > 0 {
					args = append(args, strings.TrimSpace(cur.String()))
				}
				return i + 1, args, true
			}
			cur.WriteByte(c)
		case ',':
			if depth == 1 {
				args = append(args, strings.TrimSpace(cur.String()))
				cur.Reset()
			} else {
				cur.WriteByte(c)
			}
		case '\'', '"':
			endLit := scanStringLiteral(s, i)
			cur.WriteString(s[i:endLit])
			i = endLit - 1
		case '`':
			j := i + 1
			for j < len(s) && s[j] != '`' {
				j++
			}
			if j < len(s) {
				j++
			}
			cur.WriteString(s[i:j])
			i = j - 1
		default:
			cur.WriteByte(c)
		}
	}
	return 0, nil, false
}

func buildAppendsBody(args []string) string {
	tableRef := tvfTableRef(args)
	return "(SELECT *, CAST(NULL AS STRING) AS _CHANGE_TYPE, " +
		"CAST(NULL AS TIMESTAMP) AS _CHANGE_TIMESTAMP " +
		"FROM " + tableRef + " WHERE FALSE)"
}

func buildChangesBody(args []string) string {
	tableRef := tvfTableRef(args)
	return "(SELECT *, CAST(NULL AS STRING) AS _CHANGE_TYPE, " +
		"CAST(NULL AS TIMESTAMP) AS _CHANGE_TIMESTAMP, " +
		"CAST(NULL AS STRING) AS _CHANGE_SEQUENCE_NUMBER " +
		"FROM " + tableRef + " WHERE FALSE)"
}

func buildVectorIndexStatisticsBody(args []string) string {
	return "(SELECT CAST(NULL AS FLOAT64) AS drift FROM (SELECT 1) WHERE FALSE)"
}

// tvfTableRef pulls the table reference out of the first positional
// argument of an APPENDS / CHANGES call. The arg is expected to be
// "TABLE <name>"; we drop the leading TABLE keyword if present.
func tvfTableRef(args []string) string {
	if len(args) == 0 {
		return "(SELECT 1 WHERE FALSE)"
	}
	first := strings.TrimSpace(args[0])
	if strings.EqualFold(first[:min(len(first), 5)], "TABLE") {
		first = strings.TrimSpace(first[5:])
	}
	if first == "" {
		return "(SELECT 1 WHERE FALSE)"
	}
	return first
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
