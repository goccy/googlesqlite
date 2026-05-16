package string

import (
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// EDIT_DISTANCE computes the Levenshtein distance between two strings.
// Counted in Unicode code points, matching BigQuery's default
// (byte_semantics => FALSE). The optional max_distance / byte_semantics
// named arguments are not yet wired through the analyzer; they are
// follow-ups.
func EDIT_DISTANCE(a, b string) (value.Value, error) {
	ra, rb := []rune(a), []rune(b)
	n, m := len(ra), len(rb)
	if n == 0 {
		return value.IntValue(int64(m)), nil
	}
	if m == 0 {
		return value.IntValue(int64(n)), nil
	}
	prev := make([]int, m+1)
	curr := make([]int, m+1)
	for j := 0; j <= m; j++ {
		prev[j] = j
	}
	for i := 1; i <= n; i++ {
		curr[0] = i
		for j := 1; j <= m; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			min := del
			if ins < min {
				min = ins
			}
			if sub < min {
				min = sub
			}
			curr[j] = min
		}
		prev, curr = curr, prev
	}
	return value.IntValue(int64(prev[m])), nil
}

func BindEditDistance(args ...value.Value) (value.Value, error) {
	// EDIT_DISTANCE accepts up to four positional arguments after the
	// analyzer normalises the optional named arguments
	// (max_distance, byte_semantics). The first two are the strings
	// to compare; if max_distance is set we honour it as an upper
	// bound on the result; byte_semantics is currently ignored
	// because the runtime always counts in Unicode code points.
	if len(args) < 2 {
		return nil, nil
	}
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	a, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	b, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	out, err := EDIT_DISTANCE(a, b)
	if err != nil || out == nil {
		return out, err
	}
	if len(args) >= 3 && args[2] != nil {
		maxDistance, err := args[2].ToInt64()
		if err != nil {
			return nil, err
		}
		got, err := out.ToInt64()
		if err != nil {
			return nil, err
		}
		if got > maxDistance {
			return value.IntValue(maxDistance), nil
		}
	}
	_ = helper.ExistsNull
	return out, nil
}
