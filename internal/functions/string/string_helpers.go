package string

import (
	"fmt"
	"slices"
)

func isDelim(v rune, delimiters []rune) bool {
	return slices.Contains(delimiters, v)
}

func normalizeReplacement(repl string) string {
	var normalized []byte
	for i := 0; i < len(repl); i++ {
		switch repl[i] {
		case '\\':
			i++
			var tmp []byte
			switch repl[i] {
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				tmp = []byte{'$', '{', repl[i]}
				for j := i + 1; j < len(repl); j++ {
					switch repl[j] {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						tmp = append(tmp, repl[j])
						continue
					}
					tmp = append(tmp, '}')
					i = j - 1
					break
				}
			default:
				tmp = []byte{'\\', repl[i]}
			}
			normalized = append(normalized, tmp...)
		default:
			normalized = append(normalized, repl[i])
		}
	}
	return string(normalized)
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
