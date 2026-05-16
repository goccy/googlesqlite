package string

import (
	"fmt"
	"regexp"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func REGEXP_INSTR(sourceValue, exprValue value.Value, position, occurrence, occurrencePos int64) (value.Value, error) {
	if position <= 0 {
		return nil, fmt.Errorf("REGEXP_INSTR: unexpected position number. position must be positive number")
	}
	if occurrence <= 0 {
		return nil, fmt.Errorf("REGEXP_INSTR: unexpected occurrence number. occurrence must be positive number")
	}
	pos := int(position) - 1
	switch sourceValue.(type) {
	case value.StringValue:
		source, err := sourceValue.ToString()
		if err != nil {
			return nil, err
		}
		expr, err := exprValue.ToString()
		if err != nil {
			return nil, err
		}
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		if pos >= len([]rune(source)) {
			return value.IntValue(0), nil
		}
		matches := re.FindAllStringSubmatchIndex(source[pos:], int(occurrence))
		if len(matches) < int(occurrence) {
			return value.IntValue(0), nil
		}
		match := matches[occurrence-1]
		if len(match) <= int(occurrencePos) {
			return value.IntValue(0), nil
		}
		return value.IntValue(pos + match[occurrencePos] + 1), nil
	case value.BytesValue:
		source, err := sourceValue.ToBytes()
		if err != nil {
			return nil, err
		}
		expr, err := exprValue.ToBytes()
		if err != nil {
			return nil, err
		}
		re, err := regexp.Compile(string(expr))
		if err != nil {
			return nil, err
		}
		if pos >= len(source) {
			return value.IntValue(0), nil
		}
		matches := re.FindAllSubmatchIndex(source[pos:], int(occurrence))
		if len(matches) < int(occurrence) {
			return value.IntValue(0), nil
		}
		match := matches[occurrence-1]
		if len(match) <= int(occurrencePos) {
			return value.IntValue(0), nil
		}
		return value.IntValue(pos + match[occurrencePos] + 1), nil
	}
	return nil, fmt.Errorf("REGEXP_INSTR: source value must be STRING or BYTES")
}

var BindRegexpInstr = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	var (
		pos           int64 = 1
		occurrence    int64 = 1
		occurrencePos int64 = 0
	)
	if len(args) > 2 {
		p, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		pos = p
	}
	if len(args) > 3 {
		o, err := args[3].ToInt64()
		if err != nil {
			return nil, err
		}
		occurrence = o
	}
	if len(args) > 4 {
		p, err := args[4].ToInt64()
		if err != nil {
			return nil, err
		}
		occurrencePos = p
	}
	return REGEXP_INSTR(args[0], args[1], pos, occurrence, occurrencePos)
})
