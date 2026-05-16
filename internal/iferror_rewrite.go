package internal

import (
	"strings"
)

// applyIferrorTypePropagation wraps the inner argument of an
// `IFERROR(IFERROR(ERROR(...), ERROR(...)), <typed-expr>)` call with
// `SAFE_CAST(... AS <T>)` so the outer IFERROR's templated argument
// resolution succeeds.
//
// Why this is here:
//
//	IFERROR's builtin signature is `IFERROR(T1, T1) -> T1`. When both
//	arguments to a call are templated calls themselves (`ERROR(...)`)
//	the resolver has no peer constraint for `T1`, so it falls back
//	to INT64. The OUTER call then sees `IFERROR(INT64, STRING)` (or
//	any other concrete type for the second arg) and rejects with
//	"Unable to find common supertype for templated argument <T1>".
//
//	Wrapping the inner call with `SAFE_CAST(<inner> AS <T>)` where
//	`<T>` is the static type of the outer's second argument lets the
//	resolver propagate the outer's expected type into the inner's
//	placeholder, which is what the upstream Examples document and
//	expect. The CAST is safe — its only observable effect is to
//	re-type the resolved-INT64 inner expression as the outer's type;
//	the runtime evaluator always reaches the CAST with a NULL
//	(because every reachable `ERROR()` folds to NULL inside the
//	safe-eval sub-context, see formatter.go), so the SAFE_CAST is
//	also NULL and the outer IFERROR uses its catch arm.
//
// Scope:
//
//	The rewrite intentionally fires only when both inner arguments
//	are bare `ERROR(...)` calls and the outer second argument is a
//	string or numeric literal — that's the precise pattern that
//	defaults T1 to INT64. Any other shape is left untouched so the
//	rewrite cannot accidentally re-type a well-typed inner result.
func applyIferrorTypePropagation(query string) string {
	if !containsCaseInsensitive(query, "IFERROR") {
		return query
	}
	// Loop because rewrites near the start of the query can expose
	// new candidates further along (multiple statements / nested
	// CTEs can each contain the pattern).
	for {
		next, ok := rewriteOneIferrorPattern(query)
		if !ok {
			return query
		}
		query = next
	}
}

// rewriteOneIferrorPattern finds the first outer IFERROR call whose
// shape matches the pattern and returns the rewritten query. When no
// candidate is left, returns (_, false).
func rewriteOneIferrorPattern(s string) (string, bool) {
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '\'' || c == '"':
			i = scanStringLiteral(s, i)
			continue
		case c == '`':
			j := i + 1
			for j < len(s) && s[j] != '`' {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j
			continue
		case c == '-' && i+1 < len(s) && s[i+1] == '-':
			j := i
			for j < len(s) && s[j] != '\n' {
				j++
			}
			i = j
			continue
		}
		if isIdentStart(c) {
			j := i + 1
			for j < len(s) && isIdentCont(s[j]) {
				j++
			}
			word := s[i:j]
			if strings.EqualFold(word, "IFERROR") {
				if rewritten, ok := tryRewriteOuterIferror(s, i, j); ok {
					return rewritten, true
				}
			}
			i = j
			continue
		}
		i++
	}
	return s, false
}

// tryRewriteOuterIferror runs the pattern check on a single IFERROR
// invocation that starts at byte `nameStart` with the identifier
// ending at byte `afterName`. Returns the rewritten query on a match.
func tryRewriteOuterIferror(s string, nameStart, afterName int) (string, bool) {
	k := afterName
	for k < len(s) && (s[k] == ' ' || s[k] == '\t' || s[k] == '\n' || s[k] == '\r') {
		k++
	}
	if k >= len(s) || s[k] != '(' {
		return "", false
	}
	closeIdx := matchingCloseParen(s, k)
	if closeIdx <= 0 {
		return "", false
	}
	args := splitTopLevelArgs(s[k+1 : closeIdx])
	if len(args) != 2 {
		return "", false
	}
	innerExpr := strings.TrimSpace(args[0])
	outerCatch := strings.TrimSpace(args[1])

	// Outer catch arm must be a literal whose type we can name.
	castType, ok := literalTypeName(outerCatch)
	if !ok {
		return "", false
	}

	// Inner must be IFERROR(<a>, <b>) where <a> and <b> are both
	// bare ERROR(...) calls.
	if !isIferrorOfTwoErrors(innerExpr) {
		return "", false
	}

	rewrittenInner := "SAFE_CAST(" + innerExpr + " AS " + castType + ")"
	rewritten := s[:k+1] + rewrittenInner + ", " + outerCatch + s[closeIdx:]
	return rewritten, true
}

// splitTopLevelArgs splits an argument string at top-level commas,
// respecting string literals, backtick identifiers, and nested
// parens.
func splitTopLevelArgs(s string) []string {
	var out []string
	depth := 0
	start := 0
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '\'' || c == '"':
			i = scanStringLiteral(s, i)
			continue
		case c == '`':
			j := i + 1
			for j < len(s) && s[j] != '`' {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j
			continue
		}
		switch c {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
		i++
	}
	out = append(out, s[start:])
	return out
}

// isIferrorOfTwoErrors reports whether `expr` is `IFERROR(ERROR(...),
// ERROR(...))` (case-insensitive, allowing whitespace).
func isIferrorOfTwoErrors(expr string) bool {
	expr = strings.TrimSpace(expr)
	if !strings.HasSuffix(expr, ")") {
		return false
	}
	// Find the leading IFERROR identifier.
	if len(expr) < len("IFERROR(") {
		return false
	}
	if !strings.EqualFold(expr[:len("IFERROR")], "IFERROR") {
		return false
	}
	k := len("IFERROR")
	for k < len(expr) && (expr[k] == ' ' || expr[k] == '\t' || expr[k] == '\n' || expr[k] == '\r') {
		k++
	}
	if k >= len(expr) || expr[k] != '(' {
		return false
	}
	closeIdx := matchingCloseParen(expr, k)
	if closeIdx != len(expr)-1 {
		// Trailing junk after the IFERROR call disqualifies it.
		return false
	}
	args := splitTopLevelArgs(expr[k+1 : closeIdx])
	if len(args) != 2 {
		return false
	}
	for _, a := range args {
		a = strings.TrimSpace(a)
		if !startsWithCallIdent(a, "ERROR") {
			return false
		}
	}
	return true
}

// startsWithCallIdent reports whether `expr` is a call to the named
// identifier (i.e. `<name>(...)` with the closing paren as the last
// character).
func startsWithCallIdent(expr, name string) bool {
	if len(expr) < len(name)+2 {
		return false
	}
	if !strings.EqualFold(expr[:len(name)], name) {
		return false
	}
	k := len(name)
	for k < len(expr) && (expr[k] == ' ' || expr[k] == '\t' || expr[k] == '\n' || expr[k] == '\r') {
		k++
	}
	if k >= len(expr) || expr[k] != '(' {
		return false
	}
	closeIdx := matchingCloseParen(expr, k)
	return closeIdx == len(expr)-1
}

// literalTypeName returns the canonical GoogleSQL type name for a
// bare literal expression, or `("", false)` if the expression is not a
// recognised literal. Only literals where we can statically name the
// type are accepted — column references, function calls, parameters,
// etc. all fall through.
func literalTypeName(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	if len(expr) == 0 {
		return "", false
	}
	// String literal: 'abc', "abc", or the r/b/rb-prefixed variants.
	stripped := stripStringPrefix(expr)
	if len(stripped) > 0 && (stripped[0] == '\'' || stripped[0] == '"') {
		// Must end with a matching quote and be a single literal
		// (no concatenation), so the scanner-consumed length equals
		// the trimmed expression's length.
		end := scanStringLiteral(stripped, 0)
		if end == len(stripped) {
			return "STRING", true
		}
		return "", false
	}
	// Integer / floating literals.
	if looksLikeNumberLiteral(expr) {
		if strings.ContainsAny(expr, ".eE") {
			return "FLOAT64", true
		}
		return "INT64", true
	}
	// Boolean literals.
	if strings.EqualFold(expr, "TRUE") || strings.EqualFold(expr, "FALSE") {
		return "BOOL", true
	}
	return "", false
}

// stripStringPrefix removes a leading STRING / BYTES literal prefix
// (e.g. `R`, `B`, `RB`, case-insensitive) before the opening quote so
// that the caller can identify the literal kind.
func stripStringPrefix(expr string) string {
	i := 0
	for i < len(expr) && (expr[i] == 'r' || expr[i] == 'R' || expr[i] == 'b' || expr[i] == 'B') {
		i++
		if i > 2 {
			return expr
		}
	}
	return expr[i:]
}

func looksLikeNumberLiteral(expr string) bool {
	if len(expr) == 0 {
		return false
	}
	i := 0
	if expr[0] == '-' || expr[0] == '+' {
		i = 1
	}
	if i == len(expr) {
		return false
	}
	hasDigit := false
	for ; i < len(expr); i++ {
		c := expr[i]
		switch {
		case c >= '0' && c <= '9':
			hasDigit = true
		case c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-':
			// Permissive — caller already verified it isn't an
			// identifier/call.
		default:
			return false
		}
	}
	return hasDigit
}
