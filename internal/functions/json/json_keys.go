package json

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// JSON_KEYS returns the unique JSON object keys reachable from a
// JSON value as an ARRAY<STRING>. Optional max_depth caps recursion
// depth; optional mode ("strict" / "lax" / "lax recursive") controls
// how keys inside arrays are handled per the spec
// (docs/specs/bigquery/functions/json/json_keys.md).
//
// Strict (default): keys nested inside any array are skipped.
// Lax: keys inside non-consecutively nested arrays are included.
// Lax recursive: keys inside consecutively nested arrays too.
func JSON_KEYS(jsonText string, maxDepth int, mode string) (value.Value, error) {
	var node any
	if err := json.Unmarshal([]byte(jsonText), &node); err != nil {
		return nil, err
	}
	collector := &jsonKeysCollector{
		seen:     map[string]struct{}{},
		maxDepth: maxDepth,
		mode:     strings.ToLower(strings.TrimSpace(mode)),
	}
	collector.walk(node, "", 0, false)
	keys := make([]string, 0, len(collector.seen))
	for k := range collector.seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]value.Value, 0, len(keys))
	for _, k := range keys {
		out = append(out, value.StringValue(k))
	}
	return &value.ArrayValue{Values: out}, nil
}

type jsonKeysCollector struct {
	seen     map[string]struct{}
	maxDepth int    // 0 = unbounded
	mode     string // "" or "strict" / "lax" / "lax recursive"
}

func (c *jsonKeysCollector) walk(node any, path string, depth int, insideArray bool) {
	if c.maxDepth > 0 && depth >= c.maxDepth {
		return
	}
	switch v := node.(type) {
	case map[string]any:
		for k, child := range v {
			// In "strict" mode, keys nested inside any array are
			// skipped. "lax" allows a single array level; "lax
			// recursive" allows arbitrary array nesting.
			if insideArray {
				switch c.mode {
				case "", "strict":
					// Skip — keys inside arrays are excluded.
					continue
				}
			}
			full := joinJSONKey(path, k)
			c.seen[full] = struct{}{}
			c.walk(child, full, depth+1, insideArray)
		}
	case []any:
		// Array enters; mode controls whether further nested keys
		// are visible.
		nextInside := true
		if c.mode == "lax recursive" {
			nextInside = insideArray // unchanged: stays at the same nesting level
		}
		// "lax" allows only a single array level: subsequent array
		// nesting hides keys again, which we model by treating
		// already-insideArray as a hard stop in non-recursive modes.
		if c.mode == "lax" && insideArray {
			return
		}
		for _, child := range v {
			c.walk(child, path, depth, nextInside)
		}
	}
}

// joinJSONKey appends `key` onto `parent`, quoting keys that contain
// non-identifier characters per the spec ("Keys containing special
// characters are escaped using double quotes").
func joinJSONKey(parent, key string) string {
	out := key
	if needsJSONKeyQuote(key) {
		out = strconv.Quote(key)
	}
	if parent == "" {
		return out
	}
	return parent + "." + out
}

func needsJSONKeyQuote(s string) bool {
	if s == "" {
		return true
	}
	for i, r := range s {
		if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return true
	}
	return false
}

// BindJsonKeys parses positional args and any kwargs the analyzer
// passes through. The signature variants are:
//
//	JSON_KEYS(json_expr)
//	JSON_KEYS(json_expr, max_depth)
//	JSON_KEYS(json_expr, mode => '...')
//	JSON_KEYS(json_expr, max_depth, mode => '...')
//
// The analyzer normalises named arguments to positional, with the
// mode landing at the trailing slot.
func BindJsonKeys(args ...value.Value) (value.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("JSON_KEYS: missing JSON argument")
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	jsonText, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	maxDepth := 0
	mode := ""
	for _, a := range args[1:] {
		if a == nil {
			continue
		}
		switch v := a.(type) {
		case value.IntValue:
			maxDepth = int(v)
		case value.StringValue:
			mode = string(v)
		default:
			// Unknown trailing arg; fall through.
		}
	}
	return JSON_KEYS(jsonText, maxDepth, mode)
}
