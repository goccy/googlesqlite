package string

import (
	"fmt"
	"slices"
)

func isDelim(v rune, delimiters []rune) bool {
	return slices.Contains(delimiters, v)
}

// normalizeReplacement rewrites a BigQuery / GoogleSQL REGEXP_REPLACE
// replacement string into the template syntax consumed by Go's
// regexp.Expand (ReplaceAllString / ReplaceAll).
//
// The BigQuery replacement grammar (string_functions.md#regexp_replace)
// is:
//
//   - \0 .. \9  insert the text captured by the corresponding group,
//     where \0 is the entire match. The index is a SINGLE digit; a
//     following digit is a literal (e.g. \10 is group 1 then "0").
//   - \\        a literal backslash.
//   - \ + other an error: "'\' must be followed by a digit or '\'".
//   - every other byte, INCLUDING '$', is a literal.
//
// Go's Expand template instead spells group references "${d}" and uses
// '$' as its sigil, so a literal '$' must be doubled to "$$".
func normalizeReplacement(repl string) (string, error) {
	var normalized []byte
	for i := 0; i < len(repl); i++ {
		switch c := repl[i]; c {
		case '\\':
			if i+1 >= len(repl) {
				return "", fmt.Errorf("REGEXP_REPLACE: '\\' must be followed by a digit or '\\'")
			}
			next := repl[i+1]
			switch {
			case next >= '0' && next <= '9':
				normalized = append(normalized, '$', '{', next, '}')
			case next == '\\':
				normalized = append(normalized, '\\')
			default:
				return "", fmt.Errorf("REGEXP_REPLACE: '\\' must be followed by a digit or '\\'")
			}
			i++
		case '$':
			// '$' is an ordinary literal in a BigQuery replacement, but
			// the group sigil in Go's Expand template; escape it.
			normalized = append(normalized, '$', '$')
		default:
			normalized = append(normalized, c)
		}
	}
	return string(normalized), nil
}

var soundexMap = map[byte]byte{
	'A': '0', 'B': '1', 'C': '2', 'D': '3',
	'E': '0', 'F': '1', 'G': '2', 'H': '0',
	'I': '0', 'J': '2', 'K': '2', 'L': '4',
	'M': '5', 'N': '5', 'O': '0', 'P': '1',
	'Q': '2', 'R': '6', 'S': '2', 'T': '3',
	'U': '0', 'V': '1', 'W': '0', 'X': '2',
	'Y': '0', 'Z': '2',
}

func substrPos(pos int64, strlen int64) int64 {
	if pos == 0 || pos < -strlen {
		return 0
	}
	if pos > strlen {
		return strlen
	}
	if pos > 0 {
		return pos - 1
	}
	// pos is negative number
	return strlen + pos
}

func substrLen(length *int64, strlen int64) (int64, error) {
	if length == nil {
		return strlen, nil
	}
	if *length < 0 {
		return 0, fmt.Errorf("SUBSTR: length must be positive number")
	}
	if *length > strlen {
		return strlen, nil
	}
	return *length, nil
}
