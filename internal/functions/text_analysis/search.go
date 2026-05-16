package text_analysis

import (
	"fmt"
	"strings"

	gjson "github.com/goccy/go-json"

	stdjson "encoding/json"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// SEARCH returns TRUE when every term of the LOG_ANALYZER
// tokenisation of `query` appears in the tokenisation of `data`.
// NULL `query` raises an error per the upstream spec.
//
// `data` may be STRING, STRUCT (recursively flattened), JSON
// (controlled by `jsonScope`), or ARRAY<STRING> / ARRAY<STRUCT>.
// `analyzer` selects the tokeniser: LOG_ANALYZER (default),
// NO_OP_ANALYZER, or PATTERN_ANALYZER (with regex pattern from
// `analyzerOptions`).
//
// `jsonScope` (default JSON_VALUES) chooses whether JSON keys,
// values, or both are tokenised. STRUCT inputs always flatten
// every STRING/JSON leaf, ignoring numeric / boolean leaves.
func SEARCH(data value.Value, query, analyzer, analyzerOptions, jsonScope string) (value.Value, error) {
	if strings.TrimSpace(jsonScope) == "" {
		jsonScope = "JSON_VALUES"
	}
	jsonScope = strings.ToUpper(jsonScope)
	switch jsonScope {
	case "JSON_VALUES", "JSON_KEYS", "JSON_KEYS_AND_VALUES":
	default:
		return nil, fmt.Errorf("SEARCH: invalid json_scope %q", jsonScope)
	}
	queryTokens, err := tokenise(query, analyzer, analyzerOptions)
	if err != nil {
		return nil, err
	}
	if len(queryTokens) == 0 {
		return nil, fmt.Errorf("SEARCH: search_query produced no tokens")
	}
	have := map[string]struct{}{}
	if err := collectDataTokens(data, analyzer, analyzerOptions, jsonScope, have); err != nil {
		return nil, err
	}
	for _, q := range queryTokens {
		if _, ok := have[q]; !ok {
			return value.BoolValue(false), nil
		}
	}
	return value.BoolValue(true), nil
}

// tokenise dispatches by analyzer name (case-insensitive). Empty
// analyzer ⇒ LOG_ANALYZER.
func tokenise(text, analyzer, analyzerOptions string) ([]string, error) {
	analyzer = strings.ToUpper(strings.TrimSpace(analyzer))
	switch analyzer {
	case "", "LOG_ANALYZER":
		return logAnalyzerTokens(text), nil
	case "NO_OP_ANALYZER":
		return []string{text}, nil
	case "PATTERN_ANALYZER":
		patterns, err := parsePatternAnalyzerOptions(analyzerOptions)
		if err != nil {
			return nil, err
		}
		return patternAnalyzerTokens(text, patterns), nil
	default:
		return nil, fmt.Errorf("SEARCH: unsupported analyzer %q", analyzer)
	}
}

// collectDataTokens flattens any data shape that BigQuery accepts
// for the first SEARCH argument into a token set. STRINGs go
// through the chosen analyzer; STRUCTs recurse over their fields;
// ARRAYs recurse over their elements; JSON values dispatch on
// json_scope; numeric / boolean leaves contribute no tokens.
func collectDataTokens(v value.Value, analyzer, analyzerOptions, jsonScope string, out map[string]struct{}) error {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case value.StringValue:
		tokens, err := tokenise(string(x), analyzer, analyzerOptions)
		if err != nil {
			return err
		}
		for _, t := range tokens {
			out[t] = struct{}{}
		}
		return nil
	case *value.ArrayValue:
		for _, e := range x.Values {
			if err := collectDataTokens(e, analyzer, analyzerOptions, jsonScope, out); err != nil {
				return err
			}
		}
		return nil
	case *value.StructValue:
		for _, fv := range x.Values {
			if err := collectDataTokens(fv, analyzer, analyzerOptions, jsonScope, out); err != nil {
				return err
			}
		}
		return nil
	case value.JsonValue:
		return collectJsonTokens([]byte(x), analyzer, analyzerOptions, jsonScope, out)
	}
	// Other scalar types (INT64, FLOAT64, BOOL, DATE, ...) do not
	// participate in BigQuery's SEARCH semantics: the function
	// returns FALSE for them per spec.
	return nil
}

func collectJsonTokens(body []byte, analyzer, analyzerOptions, jsonScope string, out map[string]struct{}) error {
	var node any
	if err := gjson.Unmarshal(body, &node); err != nil {
		return nil
	}
	wantKeys := jsonScope == "JSON_KEYS" || jsonScope == "JSON_KEYS_AND_VALUES"
	wantValues := jsonScope == "JSON_VALUES" || jsonScope == "JSON_KEYS_AND_VALUES"
	var walk func(n any) error
	walk = func(n any) error {
		switch v := n.(type) {
		case map[string]any:
			for k, child := range v {
				if wantKeys {
					tokens, err := tokenise(k, analyzer, analyzerOptions)
					if err != nil {
						return err
					}
					for _, t := range tokens {
						out[t] = struct{}{}
					}
				}
				if err := walk(child); err != nil {
					return err
				}
			}
		case []any:
			for _, e := range v {
				if err := walk(e); err != nil {
					return err
				}
			}
		case string:
			if wantValues {
				tokens, err := tokenise(v, analyzer, analyzerOptions)
				if err != nil {
					return err
				}
				for _, t := range tokens {
					out[t] = struct{}{}
				}
			}
		}
		return nil
	}
	return walk(node)
}

// BindSearch handles the runtime argument list. BigQuery's SEARCH
// uses named arguments (analyzer => ..., analyzer_options => ...,
// json_scope => ...) which the analyzer normalises to trailing
// positional STRING arguments after `data, query`.
func BindSearch(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 5 {
		return nil, fmt.Errorf("SEARCH: invalid number of arguments: got %d, want between 2 and 5", len(args))
	}
	if args[1] == nil {
		return nil, fmt.Errorf("SEARCH: search_query must not be NULL")
	}
	if helper.ExistsNull(args[:1]) {
		return value.BoolValue(false), nil
	}
	// Spanner-style call: first argument is a TOKENLIST (BYTES). The
	// runtime carries the kind discriminator inside the serialised
	// payload, so route to the Spanner runtime when we see a
	// BytesValue here.
	if _, isBytes := args[0].(value.BytesValue); isBytes {
		return spannerSearchOnTokenList(args[0], args[1])
	}
	query, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	analyzer := ""
	analyzerOptions := ""
	jsonScope := ""
	if len(args) >= 3 && args[2] != nil {
		s, err := args[2].ToString()
		if err != nil {
			return nil, err
		}
		analyzer = s
	}
	if len(args) >= 4 && args[3] != nil {
		s, err := args[3].ToString()
		if err != nil {
			return nil, err
		}
		analyzerOptions = s
	}
	if len(args) >= 5 && args[4] != nil {
		s, err := args[4].ToString()
		if err != nil {
			return nil, err
		}
		jsonScope = s
	}
	return SEARCH(args[0], query, analyzer, analyzerOptions, jsonScope)
}

// spannerSearchOnTokenList runs the Spanner-style SEARCH against
// a TOKENLIST-serialised BYTES value. It mirrors the behaviour of
// internal/functions/spanner.BindSpannerSearch but avoids an import
// cycle by re-decoding inline.
func spannerSearchOnTokenList(tlVal, qVal value.Value) (value.Value, error) {
	b, err := tlVal.ToBytes()
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return value.BoolValue(false), nil
	}
	var tl struct {
		Tokens []string `json:"tokens"`
	}
	if err := stdjson.Unmarshal(b, &tl); err != nil {
		return value.BoolValue(false), nil
	}
	query, err := qVal.ToString()
	if err != nil {
		return nil, err
	}
	idx := map[string]struct{}{}
	for _, t := range tl.Tokens {
		idx[t] = struct{}{}
	}
	for _, q := range searchTokenizeWords(query) {
		if _, ok := idx[q]; !ok {
			return value.BoolValue(false), nil
		}
	}
	return value.BoolValue(true), nil
}

// searchTokenizeWords is the same simple word-splitter used by
// the Spanner TOKENIZE_FULLTEXT runtime.
func searchTokenizeWords(s string) []string {
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = cur[:0]
		}
	}
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r >= 0x80:
			cur = append(cur, r)
		case r >= 'A' && r <= 'Z':
			cur = append(cur, r+32)
		default:
			flush()
		}
	}
	flush()
	return out
}
