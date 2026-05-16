package spanner

import (
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// Spanner SEARCH/TOKENIZE compliance tests. Tokenization is
// case-insensitive, splits on punctuation, and keeps Unicode
// code-points (so non-ASCII letters survive).

func TestBindSpannerTokenizeFullText(t *testing.T) {
	t.Parallel()

	tl, err := BindSpannerTokenizeFullText(value.StringValue("The Quick brown FOX"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := decodeTokenList(tl)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(parsed.Tokens, "|") != "the|quick|brown|fox" {
		t.Fatalf("got %v", parsed.Tokens)
	}

	if v, _ := BindSpannerTokenizeFullText(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerTokenizeFullText(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerTokenizeSubstring(t *testing.T) {
	t.Parallel()

	tl, err := BindSpannerTokenizeSubstring(value.StringValue("Foo Bar"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := decodeTokenList(tl)
	if parsed.Raw != "foo bar" {
		t.Fatalf("got raw %q", parsed.Raw)
	}

	if v, _ := BindSpannerTokenizeSubstring(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerTokenizeSubstring(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindSpannerTokenizeNGrams(t *testing.T) {
	t.Parallel()

	tl, err := BindSpannerTokenizeNGrams(value.StringValue("abcd"), value.IntValue(2))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := decodeTokenList(tl)
	if strings.Join(parsed.Tokens, "|") != "ab|bc|cd" {
		t.Fatalf("got %v", parsed.Tokens)
	}
	if parsed.NGSize != 2 {
		t.Fatalf("got ngsize %d", parsed.NGSize)
	}

	// Default size is 3.
	tl, err = BindSpannerTokenizeNGrams(value.StringValue("abcd"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ = decodeTokenList(tl)
	if parsed.NGSize != 3 {
		t.Fatalf("default ngsize: %d", parsed.NGSize)
	}

	// n < 1 clamped to 1.
	tl, err = BindSpannerTokenizeNGrams(value.StringValue("ab"), value.IntValue(0))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ = decodeTokenList(tl)
	if parsed.NGSize != 1 {
		t.Fatalf("got ngsize %d", parsed.NGSize)
	}

	if v, _ := BindSpannerTokenizeNGrams(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerTokenizeNGrams(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindSpannerTokenizeNumberBool(t *testing.T) {
	t.Parallel()

	tl, err := BindSpannerTokenizeNumber(value.FloatValue(1.5))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := decodeTokenList(tl)
	if len(parsed.Num) != 1 || parsed.Num[0] != 1.5 {
		t.Fatalf("got %v", parsed.Num)
	}

	tl, err = BindSpannerTokenizeBool(value.BoolValue(true))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ = decodeTokenList(tl)
	if len(parsed.Tokens) != 1 || parsed.Tokens[0] != "true" {
		t.Fatalf("got %v", parsed.Tokens)
	}

	tl, err = BindSpannerTokenizeBool(value.BoolValue(false))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ = decodeTokenList(tl)
	if parsed.Tokens[0] != "false" {
		t.Fatalf("got %v", parsed.Tokens)
	}

	if v, _ := BindSpannerTokenizeNumber(nil); v != nil {
		t.Fatal("expected null")
	}
	if v, _ := BindSpannerTokenizeBool(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerTokenizeNumber(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := BindSpannerTokenizeBool(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindSpannerTokenizeJSON(t *testing.T) {
	t.Parallel()

	tl, err := BindSpannerTokenizeJSON(value.StringValue(`{"name": "alice", "age": 30, "ok": true}`))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := decodeTokenList(tl)
	// Expect tokens drawn from keys + leaf strings + numbers + bools.
	joined := strings.Join(parsed.Tokens, " ")
	for _, want := range []string{"name", "alice", "age", "30", "ok", "true"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected %q in %q", want, joined)
		}
	}

	if _, err := BindSpannerTokenizeJSON(value.StringValue("not-json")); err == nil {
		t.Fatal("expected error on invalid JSON")
	}
	if v, _ := BindSpannerTokenizeJSON(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerTokenizeJSON(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerToken(t *testing.T) {
	t.Parallel()

	tl, err := BindSpannerToken(value.StringValue("Hello"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := decodeTokenList(tl)
	if len(parsed.Tokens) != 1 || parsed.Tokens[0] != "hello" {
		t.Fatalf("got %v", parsed.Tokens)
	}

	if v, _ := BindSpannerToken(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerToken(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBindSpannerTokenListConcat(t *testing.T) {
	t.Parallel()

	a, _ := BindSpannerTokenizeFullText(value.StringValue("foo"))
	b, _ := BindSpannerTokenizeFullText(value.StringValue("bar"))
	out, err := BindSpannerTokenListConcat(a, b)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := decodeTokenList(out)
	if strings.Join(parsed.Tokens, "|") != "foo|bar" {
		t.Fatalf("got %v", parsed.Tokens)
	}

	// Nil arg is skipped silently.
	out, err = BindSpannerTokenListConcat(a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ = decodeTokenList(out)
	if strings.Join(parsed.Tokens, "|") != "foo|bar" {
		t.Fatalf("got %v", parsed.Tokens)
	}

	if _, err := BindSpannerTokenListConcat(); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerSearch(t *testing.T) {
	t.Parallel()

	tl, _ := BindSpannerTokenizeFullText(value.StringValue("The quick brown fox"))

	got, err := BindSpannerSearch(tl, value.StringValue("quick fox"))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected match")
	}

	got, err = BindSpannerSearch(tl, value.StringValue("missing"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected no match")
	}

	// NULL TOKENLIST -> false.
	emptyTL, _ := encodeTokenList(nil)
	got, err = BindSpannerSearch(emptyTL, value.StringValue("foo"))
	if err != nil {
		t.Fatal(err)
	}
	// emptyTL encodes empty struct, so it will deserialize as kind=""; search returns true if no tokens match (only "foo" doesn't exist).
	if mustBool(t, got) {
		t.Fatal("expected false against empty token list")
	}

	if v, _ := BindSpannerSearch(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerSearch(tl); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerSearchSubstring(t *testing.T) {
	t.Parallel()

	tl, _ := BindSpannerTokenizeSubstring(value.StringValue("Hello World"))
	got, err := BindSpannerSearchSubstring(tl, value.StringValue("orld"))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected match")
	}

	got, err = BindSpannerSearchSubstring(tl, value.StringValue("xyz"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected no match")
	}

	if v, _ := BindSpannerSearchSubstring(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerSearchSubstring(tl); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerSearchNGrams(t *testing.T) {
	t.Parallel()

	tl, _ := BindSpannerTokenizeNGrams(value.StringValue("abcdef"), value.IntValue(2))
	got, err := BindSpannerSearchNGrams(tl, value.StringValue("bcde"))
	if err != nil {
		t.Fatal(err)
	}
	if !mustBool(t, got) {
		t.Fatal("expected match")
	}

	// Query shorter than ngram size -> false.
	got, err = BindSpannerSearchNGrams(tl, value.StringValue("a"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false for query < ngram size")
	}

	if v, _ := BindSpannerSearchNGrams(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerSearchNGrams(tl); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerScore(t *testing.T) {
	t.Parallel()

	tl, _ := BindSpannerTokenizeFullText(value.StringValue("foo bar foo"))
	// "foo" appears 2x in tokens, "bar" 1x, query "foo" -> 2.
	got, err := BindSpannerScore(tl, value.StringValue("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 2 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	got, err = BindSpannerScore(tl, value.StringValue("zzz"))
	if err != nil {
		t.Fatal(err)
	}
	if mustFloat64(t, got) != 0 {
		t.Fatal("expected 0")
	}

	if v, _ := BindSpannerScore(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerScore(tl); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerScoreNGrams(t *testing.T) {
	t.Parallel()

	tl, _ := BindSpannerTokenizeNGrams(value.StringValue("abcdef"), value.IntValue(2))
	got, err := BindSpannerScoreNGrams(tl, value.StringValue("bcde"))
	if err != nil {
		t.Fatal(err)
	}
	// "bcde" with n=2 yields 3 grams: bc, cd, de, all present -> score 3.
	if mustFloat64(t, got) != 3 {
		t.Fatalf("got %f", mustFloat64(t, got))
	}

	if _, err := BindSpannerScoreNGrams(tl); err == nil {
		t.Fatal("expected arg count error")
	}
	if v, _ := BindSpannerScoreNGrams(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
}

func TestBindSpannerSnippet(t *testing.T) {
	t.Parallel()

	got, err := BindSpannerSnippet(value.StringValue("This is a very long sample text that contains hello somewhere"), value.StringValue("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "hello") {
		t.Fatalf("got %q", mustString(t, got))
	}

	// Not found -> empty.
	got, err = BindSpannerSnippet(value.StringValue("abc"), value.StringValue("xyz"))
	if err != nil {
		t.Fatal(err)
	}
	if mustString(t, got) != "" {
		t.Fatal("expected empty")
	}

	if v, _ := BindSpannerSnippet(nil, value.StringValue("x")); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerSnippet(value.StringValue("a")); err == nil {
		t.Fatal("expected arg count error")
	}
}

func TestBindSpannerSearchInvalidPayload(t *testing.T) {
	t.Parallel()

	// Non-JSON BYTES payload yields decode error.
	if _, err := BindSpannerSearch(value.BytesValue("not-json"), value.StringValue("x")); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := BindSpannerSearchSubstring(value.BytesValue("not-json"), value.StringValue("x")); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := BindSpannerSearchNGrams(value.BytesValue("not-json"), value.StringValue("x")); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := BindSpannerScore(value.BytesValue("not-json"), value.StringValue("x")); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := BindSpannerScoreNGrams(value.BytesValue("not-json"), value.StringValue("x")); err == nil {
		t.Fatal("expected decode error")
	}

	// Empty BYTES yields an empty TOKENLIST that matches no query.
	empty := value.BytesValue([]byte{})
	got, err := BindSpannerSearch(empty, value.StringValue("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if mustBool(t, got) {
		t.Fatal("expected false from empty")
	}
}

func TestBindSpannerDebugTokenList(t *testing.T) {
	t.Parallel()

	tl, _ := BindSpannerTokenizeFullText(value.StringValue("foo bar"))
	got, err := BindSpannerDebugTokenList(tl)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mustString(t, got), "fulltext") {
		t.Fatalf("expected kind in output, got %q", mustString(t, got))
	}

	if v, _ := BindSpannerDebugTokenList(nil); v != nil {
		t.Fatal("expected null")
	}
	if _, err := BindSpannerDebugTokenList(); err == nil {
		t.Fatal("expected arg count error")
	}
}
