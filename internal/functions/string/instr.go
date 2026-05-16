package string

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func INSTR(source, search value.Value, position, occurrence int64) (value.Value, error) {
	if position == 0 {
		return nil, fmt.Errorf("INSTR: invalid position number. position is must be large than zero value")
	}
	if occurrence <= 0 {
		return nil, fmt.Errorf("INSTR: invalid occurrence number. occurrence is must be large than zero value. but specified %d", occurrence)
	}
	pos := int(math.Abs(float64(position)))
	if _, ok := source.(value.StringValue); ok {
		if _, ok := search.(value.StringValue); !ok {
			return nil, fmt.Errorf("INSTR: source and search are must be same type")
		}
		src, err := source.ToString()
		if err != nil {
			return nil, err
		}
		search, err := search.ToString()
		if err != nil {
			return nil, err
		}
		if pos >= len(src) {
			return nil, fmt.Errorf("INSTR: invalid position number. position %d is larger than source value length %d", pos, len(src))
		}
		length := len(src)
		if position < 0 {
			src = src[:len(src)-pos+1]
		} else {
			src = src[pos-1:]
		}
		var found int64
		for i := 0; i < len(src); i++ {
			idx := strings.Index(src[i:], search)
			if idx >= 0 {
				found++
				i += idx
			}
			if found == occurrence {
				if position < 0 {
					return value.IntValue(length - i - 1), nil
				}
				return value.IntValue(pos + i), nil
			}
		}
		return value.IntValue(0), nil
	}
	if _, ok := source.(value.BytesValue); ok {
		if _, ok := search.(value.BytesValue); !ok {
			return nil, fmt.Errorf("INSTR: source and search are must be same type")
		}
		src, err := source.ToBytes()
		if err != nil {
			return nil, err
		}
		search, err := search.ToBytes()
		if err != nil {
			return nil, err
		}
		if pos >= len(src) {
			return nil, fmt.Errorf("INSTR: invalid position number. position %d is larger than source value length %d", pos, len(src))
		}
		length := len(src)
		if position < 0 {
			src = src[:len(src)-pos+1]
		} else {
			src = src[pos-1:]
		}
		var found int64
		for i := 0; i < len(src); i++ {
			idx := bytes.Index(src[i:], search)
			if idx >= 0 {
				found++
				i += idx
			}
			if found == occurrence {
				if position < 0 {
					return value.IntValue(length - i - 1), nil
				}
				return value.IntValue(pos + i), nil
			}
		}
		return value.IntValue(0), nil
	}
	return nil, fmt.Errorf("INSTR: source and search type are must be STRING or BYTES type")
}

var BindInstr = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 2 && len(args) != 3 && len(args) != 4 {
		return nil, fmt.Errorf("INSTR: invalid number of arguments: got %d, want one of 2, 3, 4", len(args))
	}
	var (
		// Per BigQuery docs, position defaults to 1 (start from the
		// first character). The previous default of 0 tripped the
		// `position > 0` guard in INSTR for 2-arg calls.
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
