package string

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func INSTR(source, search value.Value, position, occurrence int64) (value.Value, error) {
	if position == 0 {
		return nil, fmt.Errorf("INSTR: position must be non-zero")
	}
	if occurrence <= 0 {
		return nil, fmt.Errorf("INSTR: occurrence must be positive, but got %d", occurrence)
	}
	if _, ok := source.(value.StringValue); ok {
		if _, ok := search.(value.StringValue); !ok {
			return nil, fmt.Errorf("INSTR: value and subvalue must be the same type")
		}
		src, err := source.ToString()
		if err != nil {
			return nil, err
		}
		sub, err := search.ToString()
		if err != nil {
			return nil, err
		}
		return value.IntValue(instr([]rune(src), []rune(sub), position, occurrence)), nil
	}
	if _, ok := source.(value.BytesValue); ok {
		if _, ok := search.(value.BytesValue); !ok {
			return nil, fmt.Errorf("INSTR: value and subvalue must be the same type")
		}
		src, err := source.ToBytes()
		if err != nil {
			return nil, err
		}
		sub, err := search.ToBytes()
		if err != nil {
			return nil, err
		}
		return value.IntValue(instr(src, sub, position, occurrence)), nil
	}
	return nil, fmt.Errorf("INSTR: value and subvalue must be STRING or BYTES")
}

// An empty sub also matches just past the last element, hence len(src)+1
// addressable positions.
func instr[T comparable](src, sub []T, position, occurrence int64) int64 {
	n := int64(len(src))
	if len(sub) == 0 {
		n++
	}
	matchesAt := func(s int64) bool {
		if s+int64(len(sub)) > int64(len(src)) {
			return false
		}
		for i, c := range sub {
			if src[s+int64(i)] != c {
				return false
			}
		}
		return true
	}
	if position > 0 {
		if position > n {
			return 0
		}
		for s := position - 1; s < n; s++ {
			if matchesAt(s) {
				occurrence--
				if occurrence == 0 {
					return s + 1
				}
			}
		}
		return 0
	}
	if -position > n {
		return 0
	}
	for s := n + position; s >= 0; s-- {
		if matchesAt(s) {
			occurrence--
			if occurrence == 0 {
				return s + 1
			}
		}
	}
	return 0
}

var BindInstr = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 2 && len(args) != 3 && len(args) != 4 {
		return nil, fmt.Errorf("INSTR: invalid number of arguments: got %d, want one of 2, 3, 4", len(args))
	}
	var (
		position   int64 = 1
		occurrence int64 = 1
	)
	if len(args) >= 3 {
		pos, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		position = pos
	}
	if len(args) == 4 {
		occur, err := args[3].ToInt64()
		if err != nil {
			return nil, err
		}
		occurrence = occur
	}
	return INSTR(args[0], args[1], position, occurrence)
})
