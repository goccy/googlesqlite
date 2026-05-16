package text_analysis

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/goccy/go-json"
	"golang.org/x/text/unicode/norm"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

// TEXT_ANALYZE tokenises a STRING using the selected analyzer.
// LOG_ANALYZER (default): split on non-alphanumeric, lowercase,
//
//	drop empty tokens — matches BigQuery's default delimiter set
//	for the search/text-analysis surface.
//
// NO_OP_ANALYZER: return the input verbatim as a single token.
// PATTERN_ANALYZER: extract substrings matching a regex pattern
//
//	given by analyzer_options JSON {"patterns": ["regex", ...]}.
//
// Token order in the output follows discovery order; the spec
// states ordering is unspecified.
func TEXT_ANALYZE(text, analyzer, options string) (value.Value, error) {
	analyzer = strings.ToUpper(strings.TrimSpace(analyzer))
	if analyzer == "" {
		analyzer = "LOG_ANALYZER"
	}
	var tokens []string
	switch analyzer {
	case "LOG_ANALYZER":
		tokens = logAnalyzerTokens(text)
	case "NO_OP_ANALYZER":
		tokens = []string{text}
	case "PATTERN_ANALYZER":
		patterns, err := parsePatternAnalyzerOptions(options)
		if err != nil {
			return nil, err
		}
		tokens = patternAnalyzerTokens(text, patterns)
	default:
		return nil, fmt.Errorf("TEXT_ANALYZE: unsupported analyzer %q", analyzer)
	}
	out := make([]value.Value, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, value.StringValue(t))
	}
	return &value.ArrayValue{Values: out}, nil
}

// logAnalyzerTokens implements BigQuery's LOG_ANALYZER:
//
//  1. NFKC-normalise (compatibility decomposition + composition).
//  2. Case-fold to lowercase.
//  3. Split on every non-alphanumeric Unicode codepoint —
//     treating whitespace, punctuation, and symbols as delimiters.
//  4. Each CJK ideograph (Han/Hiragana/Katakana/Hangul Syllables)
//     becomes a single-character token: BigQuery's LOG_ANALYZER
//     handles CJK by per-character tokenisation because they
//     carry no whitespace separators.
//  5. Empty tokens are dropped; token order is the discovery
//     order in the input.
func logAnalyzerTokens(text string) []string {
	normalised := norm.NFKC.String(text)
	normalised = strings.ToLower(normalised)
	out := make([]string, 0, 8)
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			out = append(out, buf.String())
			buf.Reset()
		}
	}
	for _, r := range normalised {
		if isCJKUnit(r) {
			flush()
			out = append(out, string(r))
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

// isCJKUnit reports whether r is a CJK character that the
// LOG_ANALYZER treats as its own token (per-character
// tokenisation, no whitespace-grouping). Covers Han ideographs
// (CJK Unified, Extension A/B/C/D, Compatibility), Hiragana,
// Katakana, and Hangul Syllables.
func isCJKUnit(r rune) bool {
	switch {
	case r >= 0x3040 && r <= 0x309F: // Hiragana
		return true
	case r >= 0x30A0 && r <= 0x30FF: // Katakana
		return true
	case r >= 0x31F0 && r <= 0x31FF: // Katakana Phonetic Extensions
		return true
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified Ideographs
		return true
	case r >= 0x3400 && r <= 0x4DBF: // CJK Unified Ideographs Extension A
		return true
	case r >= 0x20000 && r <= 0x2A6DF: // CJK Unified Ideographs Extension B
		return true
	case r >= 0x2A700 && r <= 0x2B73F: // CJK Unified Ideographs Extension C
		return true
	case r >= 0x2B740 && r <= 0x2B81F: // CJK Unified Ideographs Extension D
		return true
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility Ideographs
		return true
	case r >= 0xAC00 && r <= 0xD7AF: // Hangul Syllables
		return true
	}
	return false
}

func parsePatternAnalyzerOptions(options string) ([]*regexp.Regexp, error) {
	if strings.TrimSpace(options) == "" {
		return nil, fmt.Errorf("TEXT_ANALYZE: PATTERN_ANALYZER requires analyzer_options")
	}
	var parsed struct {
		Patterns []string `json:"patterns"`
		Pattern  string   `json:"pattern"`
	}
	if err := json.Unmarshal([]byte(options), &parsed); err != nil {
		return nil, fmt.Errorf("TEXT_ANALYZE: invalid analyzer_options JSON: %w", err)
	}
	patterns := parsed.Patterns
	if len(patterns) == 0 && parsed.Pattern != "" {
		patterns = []string{parsed.Pattern}
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("TEXT_ANALYZE: PATTERN_ANALYZER requires at least one pattern")
	}
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("TEXT_ANALYZE: invalid pattern %q: %w", p, err)
		}
		compiled = append(compiled, r)
	}
	return compiled, nil
}

func patternAnalyzerTokens(text string, patterns []*regexp.Regexp) []string {
	out := make([]string, 0, 8)
	for _, r := range patterns {
		out = append(out, r.FindAllString(text, -1)...)
	}
	return out
}

// BindTextAnalyze unpacks the up-to-three positional STRING args
// (text, analyzer, analyzer_options) per the analyzer's
// resolution of the named-arg surface.
func BindTextAnalyze(args ...value.Value) (value.Value, error) {
	if len(args) == 0 || len(args) > 3 {
		return nil, fmt.Errorf("TEXT_ANALYZE: invalid number of arguments: got %d, want between 1 and 3", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	text, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	analyzer := ""
	options := ""
	if len(args) >= 2 && args[1] != nil {
		s, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		analyzer = s
	}
	if len(args) >= 3 && args[2] != nil {
		s, err := args[2].ToString()
		if err != nil {
			return nil, err
		}
		options = s
	}
	if helper.ExistsNull(args[:1]) {
		return nil, nil
	}
	return TEXT_ANALYZE(text, analyzer, options)
}
