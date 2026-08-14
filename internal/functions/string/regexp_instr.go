package string

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// reportedSpanIndex picks the match index REGEXP_INSTR reports on:
// occurrencePos selects the start or the end of the span, and the span
// is the capturing group when the pattern has one.
func reportedSpanIndex(re *regexp.Regexp, occurrencePos int) int {
	if re.NumSubexp() == 1 {
		return occurrencePos + 2
	}
	return occurrencePos
}

func REGEXP_INSTR(sourceValue, exprValue value.Value, position, occurrence, occurrencePos int64) (value.Value, error) {
	if position <= 0 {
		return nil, fmt.Errorf("REGEXP_INSTR: unexpected position number. position must be positive number")
	}
	if occurrence <= 0 {
		return nil, fmt.Errorf("REGEXP_INSTR: unexpected occurrence number. occurrence must be positive number")
	}
	if occurrencePos != 0 && occurrencePos != 1 {
		return nil, fmt.Errorf("REGEXP_INSTR: invalid return_position_after_match: must be 0 or 1")
	}
	posInt, err := helper.SafeInt(position)
	if err != nil {
		return nil, err
	}
	pos := posInt - 1
	occ, err := helper.SafeInt(occurrence)
	if err != nil {
		return nil, err
	}
	occPos, err := helper.SafeInt(occurrencePos)
	if err != nil {
		return nil, err
	}
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
		off, ok := charOffset(source, pos)
		if !ok || expr == "" {
			return value.IntValue(0), nil
		}
		if err := checkExtractionGroups("REGEXP_INSTR", re); err != nil {
			return nil, err
		}
		rest := source[off:]
		matches := re.FindAllStringSubmatchIndex(rest, occ)
		if len(matches) < occ {
			return value.IntValue(0), nil
		}
		idx := reportedSpanIndex(re, occPos)
		match := matches[occ-1]
		if len(match) <= idx || match[idx] < 0 {
			return value.IntValue(0), nil
		}
		return value.IntValue(pos + utf8.RuneCountInString(rest[:match[idx]]) + 1), nil
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
		if pos >= len(source) || len(expr) == 0 {
			return value.IntValue(0), nil
		}
		if err := checkExtractionGroups("REGEXP_INSTR", re); err != nil {
			return nil, err
		}
		matches := re.FindAllSubmatchIndex(source[pos:], occ)
		if len(matches) < occ {
			return value.IntValue(0), nil
		}
		idx := reportedSpanIndex(re, occPos)
		match := matches[occ-1]
		if len(match) <= idx || match[idx] < 0 {
			return value.IntValue(0), nil
		}
		return value.IntValue(pos + match[idx] + 1), nil
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
