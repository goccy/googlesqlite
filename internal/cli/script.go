package cli

import "strings"

// SplitStatements splits SQL text into individual statements on
// top-level semicolons. It is GoogleSQL-lexer-aware: a `;` inside a
// string literal, a quoted identifier or a comment does not terminate
// a statement. Comments and quoted text are preserved verbatim in the
// returned statements; only the separating `;` is dropped. Empty
// statements (whitespace / comments only) are skipped.
//
// A trailing `\G` group-mode marker is left attached to the statement
// text — the runner strips it — because `\` and `G` are ordinary
// characters to the splitter.
func SplitStatements(text string) []string {
	runes := []rune(text)
	n := len(runes)
	var (
		stmts []string
		cur   strings.Builder
	)
	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" && hasSQLContent(s) {
			stmts = append(stmts, s)
		}
		cur.Reset()
	}
	writeRange := func(from, to int) {
		for k := from; k < to && k < n; k++ {
			cur.WriteRune(runes[k])
		}
	}
	for i := 0; i < n; {
		c := runes[i]
		switch {
		case c == '-' && i+1 < n && runes[i+1] == '-',
			c == '#':
			// Line comment: consume to end of line (newline kept).
			j := i
			for j < n && runes[j] != '\n' {
				j++
			}
			writeRange(i, j)
			i = j
		case c == '/' && i+1 < n && runes[i+1] == '*':
			// Block comment: consume through the closing */.
			j := i + 2
			for j < n && (runes[j] != '*' || j+1 >= n || runes[j+1] != '/') {
				j++
			}
			if j < n {
				j += 2
			}
			writeRange(i, j)
			i = j
		case c == '`':
			// Backtick-quoted identifier.
			j := i + 1
			for j < n && runes[j] != '`' {
				j++
			}
			if j < n {
				j++
			}
			writeRange(i, j)
			i = j
		case c == ';':
			i++
			flush()
		default:
			if raw, quote, ok := stringLiteralAt(runes, i); ok {
				end := consumeString(runes, quote, raw)
				writeRange(i, end)
				i = end
				continue
			}
			cur.WriteRune(c)
			i++
		}
	}
	flush()
	return stmts
}

// SplitComplete splits text like SplitStatements but reports any
// unterminated trailing statement separately. It is the REPL's
// multi-line accumulator: complete holds statements that ended with
// `;` (or a trailing \G), and remainder holds text the user is still
// typing. remainder is returned verbatim so the caller can append the
// next input line to it.
func SplitComplete(text string) (complete []string, remainder string) {
	runes := []rune(text)
	last := lastTopLevelSemicolon(runes)
	head := ""
	tailRaw := text
	if last >= 0 {
		head = string(runes[:last+1])
		tailRaw = string(runes[last+1:])
	}
	complete = SplitStatements(head)
	tail := strings.TrimSpace(tailRaw)
	if !hasSQLContent(tail) {
		return complete, ""
	}
	// A trailing \G terminates a statement even without a semicolon.
	if strings.HasSuffix(tail, `\G`) {
		complete = append(complete, tail)
		return complete, ""
	}
	return complete, tailRaw
}

// lastTopLevelSemicolon returns the index of the last `;` that is not
// inside a string, quoted identifier or comment, or -1 if there is
// none.
func lastTopLevelSemicolon(runes []rune) int {
	n := len(runes)
	last := -1
	for i := 0; i < n; {
		c := runes[i]
		switch {
		case c == '-' && i+1 < n && runes[i+1] == '-', c == '#':
			for i < n && runes[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < n && runes[i+1] == '*':
			i += 2
			for i < n && (runes[i] != '*' || i+1 >= n || runes[i+1] != '/') {
				i++
			}
			if i < n {
				i += 2
			}
		case c == '`':
			i++
			for i < n && runes[i] != '`' {
				i++
			}
			if i < n {
				i++
			}
		case c == ';':
			last = i
			i++
		default:
			if raw, q, ok := stringLiteralAt(runes, i); ok {
				i = consumeString(runes, q, raw)
			} else {
				i++
			}
		}
	}
	return last
}

// stringLiteralAt reports whether a GoogleSQL string literal begins at
// runes[i]. It accepts an optional case-insensitive r/b prefix (in
// either order, e.g. r”, b"", rb”, br""). On success it returns
// whether the literal is raw (no backslash escapes) and the index of
// the opening quote rune.
func stringLiteralAt(runes []rune, i int) (raw bool, quote int, ok bool) {
	n := len(runes)
	j := i
	hasR, hasB := false, false
	for j < n && j-i < 2 {
		c := runes[j]
		switch {
		case (c == 'r' || c == 'R') && !hasR:
			hasR = true
			j++
		case (c == 'b' || c == 'B') && !hasB:
			hasB = true
			j++
		default:
			j = n // stop the prefix scan
		}
		if j >= n {
			break
		}
	}
	// j-i may have advanced past the prefix; re-derive the quote index
	// as the first non-prefix rune.
	q := i
	if hasR {
		q++
	}
	if hasB {
		q++
	}
	if q < n && (runes[q] == '\'' || runes[q] == '"') {
		return hasR, q, true
	}
	return false, 0, false
}

// consumeString returns the index just past the end of the string
// literal whose opening quote is at runes[quote]. It handles both
// single- and triple-quoted forms; backslash escapes are honoured
// unless raw is true. An unterminated literal is consumed to the end
// of input (single-line forms stop at a newline).
func consumeString(runes []rune, quote int, raw bool) int {
	n := len(runes)
	q := runes[quote]
	triple := quote+2 < n && runes[quote+1] == q && runes[quote+2] == q
	if triple {
		for i := quote + 3; i < n; {
			if !raw && runes[i] == '\\' {
				i += 2
				continue
			}
			if runes[i] == q && i+2 < n && runes[i+1] == q && runes[i+2] == q {
				return i + 3
			}
			i++
		}
		return n
	}
	for i := quote + 1; i < n; {
		if !raw && runes[i] == '\\' {
			i += 2
			continue
		}
		if runes[i] == q {
			return i + 1
		}
		if runes[i] == '\n' {
			return i
		}
		i++
	}
	return n
}

// leadingKeyword returns the first SQL keyword of a statement in upper
// case, skipping leading whitespace and comments. A statement that
// starts with `(` reports "(" so the classifier can treat a
// parenthesised query as a query.
func leadingKeyword(stmt string) string {
	runes := []rune(stmt)
	n := len(runes)
	i := 0
	for i < n {
		c := runes[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '-' && i+1 < n && runes[i+1] == '-', c == '#':
			for i < n && runes[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < n && runes[i+1] == '*':
			i += 2
			for i < n && (runes[i] != '*' || i+1 >= n || runes[i+1] != '/') {
				i++
			}
			if i < n {
				i += 2
			}
		default:
			if c == '(' {
				return "("
			}
			j := i
			for j < n && isWordRune(runes[j]) {
				j++
			}
			return strings.ToUpper(string(runes[i:j]))
		}
	}
	return ""
}

// hasSQLContent reports whether s contains anything other than
// whitespace and comments — i.e. whether it is a real statement worth
// executing rather than a stray comment block.
func hasSQLContent(s string) bool {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n; {
		c := runes[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '-' && i+1 < n && runes[i+1] == '-', c == '#':
			for i < n && runes[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < n && runes[i+1] == '*':
			i += 2
			for i < n && (runes[i] != '*' || i+1 >= n || runes[i+1] != '/') {
				i++
			}
			if i < n {
				i += 2
			}
		default:
			return true
		}
	}
	return false
}

func isWordRune(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

// queryLeadKeywords are the statement-leading keywords that produce a
// result set and should be run through QueryContext rather than
// ExecContext.
var queryLeadKeywords = map[string]bool{
	"SELECT": true,
	"WITH":   true,
	"VALUES": true,
	"TABLE":  true,
	"FROM":   true, // GoogleSQL pipe syntax
	"GRAPH":  true,
	"(":      true,
}

// isQueryStatement reports whether stmt should be executed as a query.
func isQueryStatement(stmt string) bool {
	return queryLeadKeywords[leadingKeyword(stmt)]
}
