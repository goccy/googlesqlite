package json

import (
	"fmt"

	gjson "github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"
)

// JSON_CONTAINS(target, candidate) returns TRUE when every
// element of `candidate` is present in `target`. Object
// containment recurses; array containment requires every
// candidate element to appear in the target array.
func JSON_CONTAINS(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("JSON_CONTAINS: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	targetBody, err := args[0].ToJSON()
	if err != nil {
		return nil, err
	}
	candidateBody, err := args[1].ToJSON()
	if err != nil {
		return nil, err
	}
	var target, candidate any
	if err := gjson.Unmarshal([]byte(targetBody), &target); err != nil {
		return nil, err
	}
	if err := gjson.Unmarshal([]byte(candidateBody), &candidate); err != nil {
		return nil, err
	}
	return value.BoolValue(jsonContains(target, candidate)), nil
}

func jsonContains(target, candidate any) bool {
	switch c := candidate.(type) {
	case map[string]any:
		t, ok := target.(map[string]any)
		if !ok {
			return false
		}
		for k, v := range c {
			tv, exists := t[k]
			if !exists {
				return false
			}
			if !jsonContains(tv, v) {
				return false
			}
		}
		return true
	case []any:
		t, ok := target.([]any)
		if !ok {
			return false
		}
		for _, c := range c {
			matched := false
			for _, tv := range t {
				if jsonContains(tv, c) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
		return true
	default:
		return jsonEquals(target, candidate)
	}
}

func jsonEquals(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !jsonEquals(v, bv[k]) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !jsonEquals(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
