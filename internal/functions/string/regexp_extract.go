package string

import (
	"fmt"
	"regexp"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func REGEXP_EXTRACT(val value.Value, expr string, position, occurrence int64) (value.Value, error) {
	if position <= 0 {
		return nil, fmt.Errorf("REGEXP_EXTRACT: unexpected position number. position must be positive number")
	}
	if occurrence <= 0 {
		return nil, fmt.Errorf("REGEXP_EXTRACT: unexpected occurrence number. occurrence must be positive number")
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
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
	switch val.(type) {
	case value.StringValue:
		v, err := val.ToString()
		if err != nil {
			return nil, err
		}
		// Position 1 is a valid start for an empty value even though it
		// has no first character, so an empty pattern still matches it.
		atStartOfEmpty := v == "" && pos == 0
		off, ok := charOffset(v, pos)
		if !ok && !atStartOfEmpty {
			return nil, nil
		}
		if err := checkExtractionGroups("REGEXP_EXTRACT", re); err != nil {
			return nil, err
		}
		rest := v[off:]
		matches := re.FindAllStringSubmatchIndex(rest, occ)
		if len(matches) < occ {
			return nil, nil
		}
		start, end := extractedSpan(re, matches[occ-1])
		if start < 0 {
			return nil, nil
		}
		return value.StringValue(rest[start:end]), nil
	case value.BytesValue:
		v, err := val.ToBytes()
		if err != nil {
			return nil, err
		}
		atStartOfEmpty := len(v) == 0 && pos == 0
		if pos >= len(v) && !atStartOfEmpty {
			return nil, nil
		}
		if err := checkExtractionGroups("REGEXP_EXTRACT", re); err != nil {
			return nil, err
		}
		rest := v[pos:]
		matches := re.FindAllSubmatchIndex(rest, occ)
		if len(matches) < occ {
			return nil, nil
		}
		start, end := extractedSpan(re, matches[occ-1])
		if start < 0 {
			return nil, nil
		}
		return value.BytesValue(rest[start:end]), nil
	}
	return nil, fmt.Errorf("REGEXP_EXTRACT: val argument must be STRING or BYTES")
}

// extractedSpan picks the span REGEXP_EXTRACT returns for one match:
// the capturing group when the pattern has one, the whole match
// otherwise. A group that did not participate in the match has a
// negative start, which the callers turn into NULL.
func extractedSpan(re *regexp.Regexp, match []int) (int, int) {
	if re.NumSubexp() == 1 {
		return match[2], match[3]
	}
	return match[0], match[1]
}

var BindRegexpExtract = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	regexp, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	var pos int64 = 1
	if len(args) > 2 {
		p, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		pos = p
	}
	var occurrence int64 = 1
	if len(args) > 3 {
		o, err := args[3].ToInt64()
		if err != nil {
			return nil, err
		}
		occurrence = o
	}
	return REGEXP_EXTRACT(args[0], regexp, pos, occurrence)
})
