package internal

import (
	"context"
	"strconv"
	"strings"
	"unicode"
)

// formatScriptInt / formatScriptFloat keep the literal forms small
// and analyzer-friendly when DECLARE/SET RHS evaluates to a
// numeric.
func formatScriptInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func formatScriptFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// applyScriptVariables intercepts BigQuery script DECLARE / SET
// statements before they reach the upstream analyzer, which would
// otherwise reject them with "Statement not supported:
// VariableDeclaration / SingleAssignment". Resolved-AST level
// support for ResolvedScript is not exposed by go-googlesql yet.
//
// Behaviour:
//
//   - `DECLARE name [type] [DEFAULT expr] ;`
//     records (name, expr_text) on Conn.scriptVars (lowercased name)
//     and removes the statement from the rewritten SQL.
//
//   - `SET name = expr ;` (NOT @@name — those are system variables)
//     overwrites the recorded expr; removes the statement.
//
//   - Every other statement keeps its place in the rewritten SQL,
//     with bare references to known variable names replaced by the
//     current expr (case-insensitive identifier lookup, skipping
//     string / backtick / single-quote literals).
//
// Limitations:
//
//   - DECLARE list form (`DECLARE a, b INT64`) is parsed but only
//     the first identifier participates in the substitution map;
//     the others are silently ignored. Call out in test coverage if
//     the multi-ident form needs full support.
//   - Variable references are substituted positionally as text,
//     so an ARRAY-typed variable can be inlined as `[1, 2, 3]` and
//     pass through ARRAY-shaped contexts but a STRUCT variable
//     used in a complicated expression may need parens.
//
// Substitutions appear in the rewritten SQL parenthesised so binary
// expressions like `id = x` survive when x is e.g. `1 + 2`.
func applyScriptVariables(ctx context.Context, query string, conn *Conn) string {
	if conn == nil {
		return query
	}
	stmts := splitTopLevelStatements(query)
	if len(stmts) == 0 {
		return query
	}
	var keep []string
	for _, stmt := range stmts {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if name, expr, ok := parseDeclare(trimmed); ok {
			// Substitute any earlier-declared variables on the RHS,
			// evaluate the result through SQLite to a concrete
			// literal, and store the literal. This keeps repeated
			// SET m = m + 1 from growing the text quadratically.
			lit := evaluateScriptVariableExpr(ctx, conn, substituteScriptVariables(expr, conn))
			conn.SetScriptVariable(name, lit)
			continue
		}
		if name, expr, ok := parseAssignment(trimmed); ok {
			if _, declared := conn.ScriptVariable(name); declared {
				lit := evaluateScriptVariableExpr(ctx, conn, substituteScriptVariables(expr, conn))
				conn.SetScriptVariable(name, lit)
				continue
			}
			// Not a declared variable — fall through (handled by
			// AssignmentStmtAction for @@vars; user error otherwise).
		}
		// Substitute any known variable into the surviving statement.
		keep = append(keep, substituteScriptVariables(stmt, conn))
	}
	if len(keep) == 0 {
		// Pure DECLARE / SET script: feed the analyzer a no-op
		// SELECT so it has something to do.
		return "SELECT 1"
	}
	return strings.Join(keep, "; ")
}

// evaluateScriptVariableExpr asks the underlying SQLite to compute
// a single-row, single-column result for the given expression and
// returns the result as a SQL literal. Falls back to a parenthesised
// version of the original text when evaluation fails — that keeps
// callers correct even for expressions whose only use is in another
// statement (where SQLite's evaluator might reject them in isolation).
func evaluateScriptVariableExpr(ctx context.Context, conn *Conn, expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "NULL"
	}
	row := conn.QueryRowContext(ctx, "SELECT "+expr)
	var raw any
	if err := row.Scan(&raw); err != nil {
		return "(" + expr + ")"
	}
	switch v := raw.(type) {
	case nil:
		return "NULL"
	case int64:
		return formatScriptInt(v)
	case float64:
		return formatScriptFloat(v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case string:
		// May be an envelope-encoded value; if so the caller-site
		// substitution will store the envelope text and SQLite
		// will read it back via the standard collation. Quote with
		// double quotes (DQS_DML is enabled), matching the form
		// LiteralFromValue emits for non-primitive values.
		return "\"" + strings.ReplaceAll(v, "\"", "\\\"") + "\""
	case []byte:
		return "\"" + strings.ReplaceAll(string(v), "\"", "\\\"") + "\""
	}
	return "(" + expr + ")"
}

// splitTopLevelStatements splits on `;` while respecting single
// quotes, double quotes and backticks. Triple-quoted strings are
// not common in scripts; the simple form here is enough for the
// statement boundaries the rewriter needs.
func splitTopLevelStatements(query string) []string {
	var stmts []string
	start := 0
	for i := 0; i < len(query); i++ {
		c := query[i]
		switch c {
		case '\'', '"', '`':
			end := scanQuoted(query, i, c)
			i = end - 1
		case ';':
			stmts = append(stmts, query[start:i])
			start = i + 1
		}
	}
	if start < len(query) {
		stmts = append(stmts, query[start:])
	}
	return stmts
}

// scanQuoted returns the byte index past the closing quote at
// query[start]. Bare backslash escapes are honoured for ' and ";
// backticks have no escape mechanism but tolerate doubled “ to
// embed a backtick.
func scanQuoted(query string, start int, quote byte) int {
	i := start + 1
	for i < len(query) {
		c := query[i]
		if c == '\\' && quote != '`' && i+1 < len(query) {
			i += 2
			continue
		}
		if c == quote {
			return i + 1
		}
		i++
	}
	return len(query)
}

// parseDeclare matches `DECLARE name [type] [DEFAULT expr]` and
// returns the (lower-cased) identifier and the DEFAULT expr text.
// Returns ok=false when the statement is not a DECLARE we can
// handle.
func parseDeclare(stmt string) (name, expr string, ok bool) {
	rest, hasKW := matchKeyword(stmt, "DECLARE")
	if !hasKW {
		return "", "", false
	}
	rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
	identEnd := scanIdent(rest)
	if identEnd == 0 {
		return "", "", false
	}
	name = strings.ToLower(rest[:identEnd])
	rest = rest[identEnd:]
	rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
	// `DECLARE x DEFAULT 5` (no type)? `DECLARE x INT64 DEFAULT 5`?
	// Walk forward: skip until DEFAULT or end of statement.
	defaultIdx := indexOfKeyword(rest, "DEFAULT")
	if defaultIdx < 0 {
		// No DEFAULT — start the variable as NULL.
		return name, "NULL", true
	}
	expr = strings.TrimSpace(rest[defaultIdx+len("DEFAULT"):])
	expr = strings.TrimSuffix(expr, ";")
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return name, "NULL", true
	}
	return name, expr, true
}

// parseAssignment matches `SET name = expr` for non-@@ targets.
func parseAssignment(stmt string) (name, expr string, ok bool) {
	rest, hasKW := matchKeyword(stmt, "SET")
	if !hasKW {
		return "", "", false
	}
	rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
	if strings.HasPrefix(rest, "@@") {
		// System-variable assignment is handled elsewhere.
		return "", "", false
	}
	identEnd := scanIdent(rest)
	if identEnd == 0 {
		return "", "", false
	}
	name = strings.ToLower(rest[:identEnd])
	rest = strings.TrimLeftFunc(rest[identEnd:], unicode.IsSpace)
	if !strings.HasPrefix(rest, "=") {
		return "", "", false
	}
	expr = strings.TrimSpace(rest[1:])
	expr = strings.TrimSuffix(expr, ";")
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", "", false
	}
	return name, expr, true
}

// substituteScriptVariables walks `stmt` and replaces every bare
// identifier that matches a declared script variable name (case-
// insensitive) with the variable's stored expression. Quoted
// strings, backticked identifiers, and tokens prefixed with `.`
// (struct field access) or `@` (parameter) are left alone.
func substituteScriptVariables(stmt string, conn *Conn) string {
	if conn == nil || len(conn.scriptVars) == 0 {
		return stmt
	}
	var b strings.Builder
	b.Grow(len(stmt))
	i := 0
	for i < len(stmt) {
		c := stmt[i]
		switch c {
		case '\'', '"', '`':
			end := scanQuoted(stmt, i, c)
			b.WriteString(stmt[i:end])
			i = end
			continue
		}
		// Skip identifiers that are field accessors (`a.x`) — only
		// substitute when the name stands alone.
		if c == '.' {
			b.WriteByte(c)
			i++
			// Skip the following identifier verbatim.
			j := i
			for j < len(stmt) && (isIdentCont(stmt[j])) {
				j++
			}
			b.WriteString(stmt[i:j])
			i = j
			continue
		}
		if isIdentStart(c) && (i == 0 || !isIdentCont(stmt[i-1])) {
			j := i + 1
			for j < len(stmt) && isIdentCont(stmt[j]) {
				j++
			}
			word := stmt[i:j]
			if expr, ok := conn.ScriptVariable(strings.ToLower(word)); ok {
				// Don't substitute when followed by `(` — that would
				// rewrite a function call whose name happens to
				// collide with a declared variable.
				k := j
				for k < len(stmt) && (stmt[k] == ' ' || stmt[k] == '\t') {
					k++
				}
				if k < len(stmt) && stmt[k] == '(' {
					b.WriteString(word)
				} else {
					b.WriteString(expr)
				}
			} else {
				b.WriteString(word)
			}
			i = j
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
}

// matchKeyword returns the suffix after a leading <kw> keyword if
// the statement starts with it (case-insensitive, word-boundary
// aware). Returns the remainder and true on match.
func matchKeyword(stmt, kw string) (string, bool) {
	stmt = strings.TrimLeftFunc(stmt, unicode.IsSpace)
	if len(stmt) < len(kw) {
		return "", false
	}
	if !strings.EqualFold(stmt[:len(kw)], kw) {
		return "", false
	}
	if len(stmt) > len(kw) && isIdentCont(stmt[len(kw)]) {
		return "", false
	}
	return stmt[len(kw):], true
}

// indexOfKeyword finds the byte offset of a top-level <kw> keyword
// in s (case-insensitive, word-boundary aware), respecting string
// literals. Returns -1 when absent.
func indexOfKeyword(s, kw string) int {
	for i := 0; i < len(s); {
		c := s[i]
		switch c {
		case '\'', '"', '`':
			i = scanQuoted(s, i, c)
			continue
		}
		if isIdentStart(c) {
			j := i + 1
			for j < len(s) && isIdentCont(s[j]) {
				j++
			}
			if strings.EqualFold(s[i:j], kw) {
				return i
			}
			i = j
			continue
		}
		i++
	}
	return -1
}

// scanIdent returns the byte length of the leading identifier in s,
// or 0 when s starts with something that is not an identifier.
func scanIdent(s string) int {
	if len(s) == 0 || !isIdentStart(s[0]) {
		return 0
	}
	i := 1
	for i < len(s) && isIdentCont(s[i]) {
		i++
	}
	return i
}
