package spanner

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// Spanner's TOKENLIST is exposed as BYTES at the SQL surface. We
// encode it as a JSON document carrying the kind (fulltext /
// substring / ngrams / number / bool / json) and the normalised
// token list. The JSON shape is stable across all SEARCH /
// TOKENIZE / SCORE / SNIPPET runtime calls in this package.
type tokenList struct {
	Kind   string    `json:"kind"`
	Tokens []string  `json:"tokens"`
	NGSize int       `json:"ng,omitempty"`
	Raw    string    `json:"raw,omitempty"`
	Pos    []int     `json:"pos,omitempty"`
	Num    []float64 `json:"num,omitempty"`
}

func encodeTokenList(tl *tokenList) (value.Value, error) {
	b, err := json.Marshal(tl)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

func decodeTokenList(v value.Value) (*tokenList, error) {
	if v == nil {
		return nil, nil
	}
	b, err := v.ToBytes()
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return &tokenList{}, nil
	}
	tl := &tokenList{}
	if err := json.Unmarshal(b, tl); err != nil {
		return nil, fmt.Errorf("invalid TOKENLIST payload: %w", err)
	}
	return tl, nil
}

// tokenizeWords lowercases and splits on non-letter / non-digit
// boundaries, returning the surviving tokens.
func tokenizeWords(s string) []string {
	var out []string
	var cur strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r >= 0x80 {
			cur.WriteRune(r)
		} else if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// BindSpannerTokenizeFullText produces a TOKENLIST from STRING.
func BindSpannerTokenizeFullText(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 4 {
		return nil, fmt.Errorf("TOKENIZE_FULLTEXT: invalid number of arguments: got %d, want between 1 and 4", len(args))
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return encodeTokenList(&tokenList{Kind: "fulltext", Tokens: tokenizeWords(s), Raw: s})
}

// BindSpannerTokenizeSubstring produces a substring-indexed
// TOKENLIST. We store the lower-cased haystack so SEARCH_SUBSTRING
// can fall back to direct string matching.
func BindSpannerTokenizeSubstring(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 4 {
		return nil, fmt.Errorf("TOKENIZE_SUBSTRING: invalid number of arguments: got %d, want between 1 and 4", len(args))
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return encodeTokenList(&tokenList{Kind: "substring", Raw: strings.ToLower(s)})
}

// BindSpannerTokenizeNGrams produces an n-gram TOKENLIST.
func BindSpannerTokenizeNGrams(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 4 {
		return nil, fmt.Errorf("TOKENIZE_NGRAMS: invalid number of arguments: got %d, want between 1 and 4", len(args))
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	n := 3
	if len(args) >= 2 && args[1] != nil {
		x, err := args[1].ToInt64()
		if err != nil {
			return nil, err
		}
		n, err = helper.SafeInt(x)
		if err != nil {
			return nil, err
		}
		if n < 1 {
			n = 1
		}
	}
	low := strings.ToLower(s)
	runes := []rune(low)
	var grams []string
	for i := 0; i+n <= len(runes); i++ {
		grams = append(grams, string(runes[i:i+n]))
	}
	return encodeTokenList(&tokenList{Kind: "ngrams", NGSize: n, Tokens: grams, Raw: low})
}

// BindSpannerTokenizeNumber stores the numeric value as a single
// numeric token. Useful for numeric range searches.
func BindSpannerTokenizeNumber(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TOKENIZE_NUMBER: missing argument")
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	f, err := args[0].ToFloat64()
	if err != nil {
		return nil, err
	}
	return encodeTokenList(&tokenList{Kind: "number", Num: []float64{f}})
}

// BindSpannerTokenizeBool stores the boolean value.
func BindSpannerTokenizeBool(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TOKENIZE_BOOL: missing argument")
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	b, err := args[0].ToBool()
	if err != nil {
		return nil, err
	}
	tok := "false"
	if b {
		tok = "true"
	}
	return encodeTokenList(&tokenList{Kind: "bool", Tokens: []string{tok}})
}

// BindSpannerTokenizeJSON tokenizes every leaf string value.
func BindSpannerTokenizeJSON(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TOKENIZE_JSON: missing argument")
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	var node any
	if err := json.Unmarshal([]byte(s), &node); err != nil {
		return nil, fmt.Errorf("TOKENIZE_JSON: invalid JSON: %w", err)
	}
	tokens := tokenizeJSONNode(node)
	return encodeTokenList(&tokenList{Kind: "json", Tokens: tokens, Raw: s})
}

func tokenizeJSONNode(n any) []string {
	var out []string
	switch x := n.(type) {
	case string:
		out = append(out, tokenizeWords(x)...)
	case map[string]any:
		for k, v := range x {
			out = append(out, tokenizeWords(k)...)
			out = append(out, tokenizeJSONNode(v)...)
		}
	case []any:
		for _, v := range x {
			out = append(out, tokenizeJSONNode(v)...)
		}
	case float64:
		out = append(out, strconv.FormatFloat(x, 'g', -1, 64))
	case bool:
		if x {
			out = append(out, "true")
		} else {
			out = append(out, "false")
		}
	}
	return out
}

// BindSpannerToken returns a one-element TOKENLIST containing
// a single literal token.
func BindSpannerToken(args ...value.Value) (value.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TOKEN: missing argument")
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return encodeTokenList(&tokenList{Kind: "fulltext", Tokens: []string{strings.ToLower(s)}, Raw: s})
}

// BindSpannerTokenListConcat merges multiple TOKENLISTs by
// concatenating their token slices.
func BindSpannerTokenListConcat(args ...value.Value) (value.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("TOKENLIST_CONCAT: needs at least 1 argument")
	}
	out := &tokenList{Kind: "fulltext"}
	for _, a := range args {
		if a == nil {
			continue
		}
		tl, err := decodeTokenList(a)
		if err != nil {
			return nil, err
		}
		if tl == nil {
			continue
		}
		out.Tokens = append(out.Tokens, tl.Tokens...)
		out.Raw += " " + tl.Raw
	}
	return encodeTokenList(out)
}

// BindSpannerSearch returns true when every query token appears
// in the TOKENLIST.
func BindSpannerSearch(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 6 {
		return nil, fmt.Errorf("SEARCH: invalid number of arguments: got %d, want between 2 and 6", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	tl, err := decodeTokenList(args[0])
	if err != nil {
		return nil, err
	}
	query, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return value.BoolValue(false), nil
	}
	idx := map[string]struct{}{}
	for _, t := range tl.Tokens {
		idx[t] = struct{}{}
	}
	for _, q := range tokenizeWords(query) {
		if _, ok := idx[q]; !ok {
			return value.BoolValue(false), nil
		}
	}
	return value.BoolValue(true), nil
}

// BindSpannerSearchSubstring reports whether `query` appears as a
// substring in the original payload of a TOKENIZE_SUBSTRING
// TOKENLIST.
func BindSpannerSearchSubstring(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("SEARCH_SUBSTRING: invalid number of arguments: got %d, want between 2 and 4", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	tl, err := decodeTokenList(args[0])
	if err != nil {
		return nil, err
	}
	query, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return value.BoolValue(false), nil
	}
	return value.BoolValue(strings.Contains(tl.Raw, strings.ToLower(query))), nil
}

// BindSpannerSearchNGrams reports whether every n-gram of `query`
// (at the TOKENLIST's stored size) appears in the TOKENLIST.
func BindSpannerSearchNGrams(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("SEARCH_NGRAMS: invalid number of arguments: got %d, want between 2 and 4", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	tl, err := decodeTokenList(args[0])
	if err != nil {
		return nil, err
	}
	query, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	if tl == nil || tl.NGSize == 0 {
		return value.BoolValue(false), nil
	}
	low := strings.ToLower(query)
	runes := []rune(low)
	if len(runes) < tl.NGSize {
		return value.BoolValue(false), nil
	}
	idx := map[string]struct{}{}
	for _, t := range tl.Tokens {
		idx[t] = struct{}{}
	}
	for i := 0; i+tl.NGSize <= len(runes); i++ {
		gram := string(runes[i : i+tl.NGSize])
		if _, ok := idx[gram]; !ok {
			return value.BoolValue(false), nil
		}
	}
	return value.BoolValue(true), nil
}

// BindSpannerScore returns a simple match-count score for the
// TOKENLIST against the query. Zero means no match.
func BindSpannerScore(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("SCORE: invalid number of arguments: got %d, want between 2 and 4", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	tl, err := decodeTokenList(args[0])
	if err != nil {
		return nil, err
	}
	query, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return value.FloatValue(0), nil
	}
	idx := map[string]int{}
	for _, t := range tl.Tokens {
		idx[t]++
	}
	score := 0.0
	for _, q := range tokenizeWords(query) {
		if c, ok := idx[q]; ok {
			score += float64(c)
		}
	}
	return value.FloatValue(score), nil
}

// BindSpannerScoreNGrams returns the n-gram overlap count.
func BindSpannerScoreNGrams(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("SCORE_NGRAMS: invalid number of arguments: got %d, want between 2 and 4", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	tl, err := decodeTokenList(args[0])
	if err != nil {
		return nil, err
	}
	query, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	if tl == nil || tl.NGSize == 0 {
		return value.FloatValue(0), nil
	}
	low := strings.ToLower(query)
	runes := []rune(low)
	idx := map[string]int{}
	for _, t := range tl.Tokens {
		idx[t]++
	}
	score := 0.0
	for i := 0; i+tl.NGSize <= len(runes); i++ {
		gram := string(runes[i : i+tl.NGSize])
		if c, ok := idx[gram]; ok {
			score += float64(c)
		}
	}
	return value.FloatValue(score), nil
}

// BindSpannerSnippet extracts a windowed slice of the raw text
// around the first match of `query`.
func BindSpannerSnippet(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("SNIPPET: invalid number of arguments: got %d, want between 2 and 4", len(args))
	}
	if helper.ExistsNull(args[:2]) {
		return nil, nil
	}
	text, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	query, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	hay := strings.ToLower(text)
	idx := strings.Index(hay, q)
	if idx < 0 {
		return value.StringValue(""), nil
	}
	pre := idx - 20
	if pre < 0 {
		pre = 0
	}
	post := idx + len(q) + 20
	if post > len(text) {
		post = len(text)
	}
	return value.StringValue(text[pre:post]), nil
}

// BindSpannerDebugTokenList renders the TOKENLIST as a JSON
// document for inspection.
func BindSpannerDebugTokenList(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("DEBUG_TOKENLIST: invalid number of arguments: got %d, want 1", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	b, err := args[0].ToBytes()
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}
