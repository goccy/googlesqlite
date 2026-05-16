package text_analysis

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/functions/window"
	"github.com/goccy/googlesqlite/internal/value"
)

// --- BindTextAnalyze (LOG_ANALYZER default) ---

// TestBindTextAnalyzeLogAnalyzerDefault: BigQuery LOG_ANALYZER
// reference splits on non-alphanumerics and lowercases.
func TestBindTextAnalyzeLogAnalyzerDefault(t *testing.T) {
	got, err := BindTextAnalyze(value.StringValue("Hello, World!"))
	if err != nil {
		t.Fatalf("BindTextAnalyze: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 tokens, got %d (%v)", len(arr.Values), arr.Values)
	}
	t0, _ := arr.Values[0].ToString()
	t1, _ := arr.Values[1].ToString()
	if t0 != "hello" || t1 != "world" {
		t.Fatalf("want ['hello','world'], got [%q,%q]", t0, t1)
	}
}

func TestBindTextAnalyzeLogAnalyzerExplicit(t *testing.T) {
	got, _ := BindTextAnalyze(
		value.StringValue("foo bar"),
		value.StringValue("LOG_ANALYZER"),
	)
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 tokens, got %d", len(arr.Values))
	}
}

// TestBindTextAnalyzeNoOpAnalyzer keeps the input as a single token.
func TestBindTextAnalyzeNoOpAnalyzer(t *testing.T) {
	got, err := BindTextAnalyze(
		value.StringValue("Foo Bar"),
		value.StringValue("NO_OP_ANALYZER"),
	)
	if err != nil {
		t.Fatalf("BindTextAnalyze: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 1 {
		t.Fatalf("want 1 token, got %d", len(arr.Values))
	}
	s, _ := arr.Values[0].ToString()
	if s != "Foo Bar" {
		t.Fatalf("want 'Foo Bar', got %q", s)
	}
}

// TestBindTextAnalyzePatternAnalyzer extracts substrings matching
// each regex in the patterns list.
func TestBindTextAnalyzePatternAnalyzer(t *testing.T) {
	got, err := BindTextAnalyze(
		value.StringValue("foo123 bar456"),
		value.StringValue("PATTERN_ANALYZER"),
		value.StringValue(`{"patterns":["[0-9]+"]}`),
	)
	if err != nil {
		t.Fatalf("BindTextAnalyze: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 numeric tokens, got %d", len(arr.Values))
	}
}

// TestBindTextAnalyzePatternAnalyzerSinglePattern: the
// {"pattern": "..."} form is accepted as a single-element list.
func TestBindTextAnalyzePatternAnalyzerSinglePattern(t *testing.T) {
	got, err := BindTextAnalyze(
		value.StringValue("aaa bbb"),
		value.StringValue("PATTERN_ANALYZER"),
		value.StringValue(`{"pattern":"a+"}`),
	)
	if err != nil {
		t.Fatalf("BindTextAnalyze: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 1 {
		t.Fatalf("want 1 token, got %d", len(arr.Values))
	}
}

func TestBindTextAnalyzePatternAnalyzerMissingOptions(t *testing.T) {
	if _, err := BindTextAnalyze(
		value.StringValue("x"),
		value.StringValue("PATTERN_ANALYZER"),
	); err == nil {
		t.Fatalf("missing options should error")
	}
}

func TestBindTextAnalyzePatternAnalyzerInvalidJson(t *testing.T) {
	if _, err := BindTextAnalyze(
		value.StringValue("x"),
		value.StringValue("PATTERN_ANALYZER"),
		value.StringValue("not-json"),
	); err == nil {
		t.Fatalf("invalid JSON should error")
	}
}

func TestBindTextAnalyzePatternAnalyzerEmptyPatterns(t *testing.T) {
	if _, err := BindTextAnalyze(
		value.StringValue("x"),
		value.StringValue("PATTERN_ANALYZER"),
		value.StringValue(`{"patterns":[]}`),
	); err == nil {
		t.Fatalf("empty pattern list should error")
	}
}

func TestBindTextAnalyzePatternAnalyzerBadRegex(t *testing.T) {
	if _, err := BindTextAnalyze(
		value.StringValue("x"),
		value.StringValue("PATTERN_ANALYZER"),
		value.StringValue(`{"patterns":["[unclosed"]}`),
	); err == nil {
		t.Fatalf("bad regex should error")
	}
}

func TestBindTextAnalyzeUnknownAnalyzer(t *testing.T) {
	if _, err := BindTextAnalyze(
		value.StringValue("x"),
		value.StringValue("MAGIC"),
	); err == nil {
		t.Fatalf("unknown analyzer should error")
	}
}

func TestBindTextAnalyzeNullInput(t *testing.T) {
	got, err := BindTextAnalyze(nil)
	if err != nil {
		t.Fatalf("BindTextAnalyze NULL: %v", err)
	}
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
}

func TestBindTextAnalyzeArity(t *testing.T) {
	if _, err := BindTextAnalyze(); err == nil {
		t.Fatalf("arity error expected (0)")
	}
	if _, err := BindTextAnalyze(value.StringValue("a"), value.StringValue("b"), value.StringValue("c"), value.StringValue("d")); err == nil {
		t.Fatalf("arity error expected (4)")
	}
}

// TestLogAnalyzerCJK: CJK characters are per-character tokens.
func TestLogAnalyzerCJK(t *testing.T) {
	got, _ := BindTextAnalyze(value.StringValue("日本語"))
	arr, _ := got.ToArray()
	if len(arr.Values) != 3 {
		t.Fatalf("want 3 CJK tokens, got %d", len(arr.Values))
	}
}

// --- BindBagOfWords ---

func TestBindBagOfWordsBasic(t *testing.T) {
	arr := &value.ArrayValue{Values: []value.Value{
		value.StringValue("a"),
		value.StringValue("b"),
		value.StringValue("a"),
	}}
	got, err := BindBagOfWords(arr)
	if err != nil {
		t.Fatalf("BindBagOfWords: %v", err)
	}
	out, _ := got.ToArray()
	if len(out.Values) != 2 {
		t.Fatalf("want 2 rows, got %d", len(out.Values))
	}
	// Sorted lexicographically — first row term=a count=2.
	st0, _ := out.Values[0].ToStruct()
	term0, _ := st0.Values[0].ToString()
	n0, _ := st0.Values[1].ToInt64()
	if term0 != "a" || n0 != 2 {
		t.Fatalf("want first row a=2, got %s=%d", term0, n0)
	}
}

func TestBindBagOfWordsNullElement(t *testing.T) {
	arr := &value.ArrayValue{Values: []value.Value{
		nil,
		value.StringValue("x"),
		nil,
	}}
	got, _ := BindBagOfWords(arr)
	out, _ := got.ToArray()
	if len(out.Values) != 2 {
		t.Fatalf("want 2 rows (null + x), got %d", len(out.Values))
	}
	// First row has NULL term and count=2.
	st0, _ := out.Values[0].ToStruct()
	if st0.Values[0] != nil {
		t.Fatalf("first row term must be NULL, got %v", st0.Values[0])
	}
	n, _ := st0.Values[1].ToInt64()
	if n != 2 {
		t.Fatalf("null count = %d, want 2", n)
	}
}

func TestBindBagOfWordsEmpty(t *testing.T) {
	arr := &value.ArrayValue{Values: nil}
	got, err := BindBagOfWords(arr)
	if err != nil {
		t.Fatalf("BindBagOfWords: %v", err)
	}
	out, _ := got.ToArray()
	if len(out.Values) != 0 {
		t.Fatalf("empty array should yield 0 rows, got %d", len(out.Values))
	}
}

func TestBindBagOfWordsNullAndArity(t *testing.T) {
	got, _ := BindBagOfWords(nil)
	if got != nil {
		t.Fatalf("NULL input must produce NULL output")
	}
	if _, err := BindBagOfWords(); err == nil {
		t.Fatalf("arity error expected (0)")
	}
	if _, err := BindBagOfWords(value.IntValue(1), value.IntValue(2)); err == nil {
		t.Fatalf("arity error expected (2)")
	}
	// non-array argument → ToArray fails.
	if _, err := BindBagOfWords(value.IntValue(1)); err == nil {
		t.Fatalf("non-array arg should error")
	}
}

// --- BindSearch ---

func TestBindSearchBasic(t *testing.T) {
	got, err := BindSearch(
		value.StringValue("hello world"),
		value.StringValue("hello"),
	)
	if err != nil {
		t.Fatalf("BindSearch: %v", err)
	}
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want TRUE")
	}
}

func TestBindSearchMiss(t *testing.T) {
	got, _ := BindSearch(
		value.StringValue("hello world"),
		value.StringValue("missing"),
	)
	b, _ := got.ToBool()
	if b {
		t.Fatalf("want FALSE")
	}
}

func TestBindSearchAllTermsMustMatch(t *testing.T) {
	got, _ := BindSearch(
		value.StringValue("foo bar baz"),
		value.StringValue("foo baz"),
	)
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("all terms present, want TRUE")
	}
	got, _ = BindSearch(
		value.StringValue("foo bar"),
		value.StringValue("foo missing"),
	)
	b, _ = got.ToBool()
	if b {
		t.Fatalf("one term missing, want FALSE")
	}
}

// TestBindSearchArray: array of strings.
func TestBindSearchArray(t *testing.T) {
	arr := &value.ArrayValue{Values: []value.Value{
		value.StringValue("alpha"),
		value.StringValue("beta gamma"),
	}}
	got, _ := BindSearch(arr, value.StringValue("gamma"))
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want TRUE for gamma in array element")
	}
}

// TestBindSearchStruct: struct flattening.
func TestBindSearchStruct(t *testing.T) {
	st := &value.StructValue{
		Keys:   []string{"a"},
		Values: []value.Value{value.StringValue("inner text")},
		M:      map[string]value.Value{"a": value.StringValue("inner text")},
	}
	got, _ := BindSearch(st, value.StringValue("inner"))
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want TRUE for inner in struct")
	}
}

// TestBindSearchJson: JSON values participate via the default
// JSON_VALUES scope.
func TestBindSearchJson(t *testing.T) {
	got, _ := BindSearch(
		value.JsonValue(`{"name":"alice","city":"paris"}`),
		value.StringValue("alice"),
	)
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want TRUE for alice in JSON values")
	}
	// Default scope is JSON_VALUES so a key-only query misses.
	got, _ = BindSearch(
		value.JsonValue(`{"name":"alice"}`),
		value.StringValue("name"),
	)
	b, _ = got.ToBool()
	if b {
		t.Fatalf("want FALSE for key 'name' under JSON_VALUES")
	}
	// JSON_KEYS scope matches the key but not the value.
	got, _ = BindSearch(
		value.JsonValue(`{"name":"alice"}`),
		value.StringValue("name"),
		value.StringValue(""),
		value.StringValue(""),
		value.StringValue("JSON_KEYS"),
	)
	b, _ = got.ToBool()
	if !b {
		t.Fatalf("want TRUE for key 'name' under JSON_KEYS")
	}
}

func TestBindSearchAnalyzers(t *testing.T) {
	// NO_OP_ANALYZER: query must match data verbatim.
	got, _ := BindSearch(
		value.StringValue("hello world"),
		value.StringValue("hello world"),
		value.StringValue("NO_OP_ANALYZER"),
	)
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("NO_OP_ANALYZER full-match want TRUE")
	}

	// PATTERN_ANALYZER: tokens defined by regex.
	got, _ = BindSearch(
		value.StringValue("aaa bbb"),
		value.StringValue("aaa"),
		value.StringValue("PATTERN_ANALYZER"),
		value.StringValue(`{"patterns":["a+"]}`),
	)
	b, _ = got.ToBool()
	if !b {
		t.Fatalf("PATTERN_ANALYZER want TRUE")
	}
}

func TestBindSearchNullQuery(t *testing.T) {
	if _, err := BindSearch(value.StringValue("data"), nil); err == nil {
		t.Fatalf("NULL query should error")
	}
}

func TestBindSearchNullData(t *testing.T) {
	got, _ := BindSearch(nil, value.StringValue("any"))
	b, _ := got.ToBool()
	if b {
		t.Fatalf("NULL data must produce FALSE")
	}
}

func TestBindSearchEmptyQueryTokens(t *testing.T) {
	if _, err := BindSearch(value.StringValue("data"), value.StringValue("   ")); err == nil {
		t.Fatalf("whitespace-only query (no tokens) should error")
	}
}

func TestBindSearchInvalidJsonScope(t *testing.T) {
	if _, err := BindSearch(
		value.StringValue("data"),
		value.StringValue("data"),
		value.StringValue(""),
		value.StringValue(""),
		value.StringValue("BAD_SCOPE"),
	); err == nil {
		t.Fatalf("invalid json_scope should error")
	}
}

func TestBindSearchUnknownAnalyzer(t *testing.T) {
	if _, err := BindSearch(
		value.StringValue("data"),
		value.StringValue("data"),
		value.StringValue("UNKNOWN"),
	); err == nil {
		t.Fatalf("unknown analyzer should error")
	}
}

func TestBindSearchArity(t *testing.T) {
	if _, err := BindSearch(value.StringValue("a")); err == nil {
		t.Fatalf("arity error expected (1)")
	}
}

// --- TF_IDF (driven on the struct directly) ---

// tfidfNewAgg constructs the minimal WindowFuncAggregatedStatus
// state the TF_IDF.Done callback expects, mirroring the runtime
// shape the SQLite-side aggregator produces. We bypass the
// WindowAggregator wrapper because its Step expects the SQLite
// ConvertArgs-decoded option markers; the struct surface is the
// stable contract that survives refactors.
func tfidfNewAgg() *window.WindowFuncAggregatedStatus {
	return &window.WindowFuncAggregatedStatus{
		PartitionToValuesMap: map[string][]*window.WindowOrderedValue{},
	}
}

func tfidfRowStatus(rowID int64) *window.WindowFuncStatus {
	return &window.WindowFuncStatus{
		FrameUnit: window.WindowFrameUnitRows,
		Start:     &window.WindowBoundary{Type: window.WindowUnboundedPrecedingType},
		End:       &window.WindowBoundary{Type: window.WindowUnboundedFollowingType},
		RowID:     rowID,
	}
}

func TestTfIdfBasic(t *testing.T) {
	fn := &TF_IDF{maxDistinctTokens: 32000, freqThreshold: 0}
	doc1 := &value.ArrayValue{Values: []value.Value{
		value.StringValue("apple"),
		value.StringValue("banana"),
	}}
	doc2 := &value.ArrayValue{Values: []value.Value{
		value.StringValue("banana"),
		value.StringValue("cherry"),
	}}
	agg := tfidfNewAgg()
	for _, d := range []value.Value{doc1, doc2} {
		if err := fn.Step(d, 32000, 0, tfidfRowStatus(1), agg); err != nil {
			t.Fatalf("TF_IDF.Step: %v", err)
		}
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatalf("TF_IDF.Done: %v", err)
	}
	arr, err := got.ToArray()
	if err != nil {
		t.Fatalf("ToArray: %v", err)
	}
	if len(arr.Values) != 2 {
		t.Fatalf("want 2 scored entries (apple,banana), got %d", len(arr.Values))
	}
	st, _ := arr.Values[0].ToStruct()
	if len(st.Keys) != 2 || st.Keys[0] != "term" || st.Keys[1] != "score" {
		t.Fatalf("unexpected struct: %v", st.Keys)
	}
}

// TestTfIdfMaxDistinctMerge: a tight cap merges the displaced terms
// into the __unknown__ bucket.
func TestTfIdfMaxDistinctMerge(t *testing.T) {
	fn := &TF_IDF{maxDistinctTokens: 1, freqThreshold: 0}
	doc1 := &value.ArrayValue{Values: []value.Value{
		value.StringValue("a"),
		value.StringValue("b"),
		value.StringValue("c"),
	}}
	agg := tfidfNewAgg()
	if err := fn.Step(doc1, 1, 0, tfidfRowStatus(1), agg); err != nil {
		t.Fatalf("TF_IDF.Step: %v", err)
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatalf("TF_IDF.Done: %v", err)
	}
	arr, _ := got.ToArray()
	// Top term + __unknown__ bucket = at most 2 entries.
	if len(arr.Values) == 0 {
		t.Fatalf("expected at least one scored entry")
	}
	// Walk entries; expect __unknown__ to appear.
	foundUnknown := false
	for _, e := range arr.Values {
		st, _ := e.ToStruct()
		s, _ := st.Values[0].ToString()
		if s == "__unknown__" {
			foundUnknown = true
			break
		}
	}
	if !foundUnknown {
		t.Fatalf("expected __unknown__ bucket after max_distinct merge")
	}
}

// TestTfIdfRowOutOfRange yields an empty array.
func TestTfIdfRowOutOfRange(t *testing.T) {
	fn := &TF_IDF{maxDistinctTokens: 32000, freqThreshold: 0}
	doc := &value.ArrayValue{Values: []value.Value{value.StringValue("a")}}
	agg := tfidfNewAgg()
	if err := fn.Step(doc, 32000, 0, tfidfRowStatus(99), agg); err != nil {
		t.Fatalf("TF_IDF.Step: %v", err)
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatalf("TF_IDF.Done: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 0 {
		t.Fatalf("out-of-range row should yield 0 entries, got %d", len(arr.Values))
	}
}

// TestTfIdfNullToken: nil tokens are skipped silently.
func TestTfIdfNullToken(t *testing.T) {
	fn := &TF_IDF{maxDistinctTokens: 32000, freqThreshold: 0}
	doc := &value.ArrayValue{Values: []value.Value{
		nil,
		value.StringValue("apple"),
	}}
	agg := tfidfNewAgg()
	if err := fn.Step(doc, 32000, 0, tfidfRowStatus(1), agg); err != nil {
		t.Fatalf("TF_IDF.Step: %v", err)
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatalf("TF_IDF.Done: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 1 {
		t.Fatalf("want 1 entry (apple), got %d", len(arr.Values))
	}
}

// TestTfIdfEmptyPartition: zero rows in the partition → empty result.
func TestTfIdfEmptyPartition(t *testing.T) {
	fn := &TF_IDF{maxDistinctTokens: 32000, freqThreshold: 0}
	agg := tfidfNewAgg()
	// Drive a single nil document so RowID picks slot 0 of an empty
	// doc list.
	if err := fn.Step(nil, 32000, 0, tfidfRowStatus(1), agg); err != nil {
		t.Fatalf("TF_IDF.Step: %v", err)
	}
	got, err := fn.Done(agg)
	if err != nil {
		t.Fatalf("TF_IDF.Done: %v", err)
	}
	arr, _ := got.ToArray()
	if len(arr.Values) != 0 {
		t.Fatalf("nil document should yield 0 scored entries, got %d", len(arr.Values))
	}
}

// TestBindTfIdfArity: zero arguments must error.
func TestBindTfIdfArity(t *testing.T) {
	agg := BindTfIdf()()
	if err := agg.Step(); err == nil {
		t.Fatalf("missing tokens argument should error")
	}
}

// TestBindSearchTokenList covers the Spanner-style TOKENLIST branch.
func TestBindSearchTokenList(t *testing.T) {
	// JSON-encoded tokens payload.
	tl := value.BytesValue([]byte(`{"tokens":["alpha","beta"]}`))
	got, err := BindSearch(tl, value.StringValue("alpha"))
	if err != nil {
		t.Fatalf("BindSearch tokenlist: %v", err)
	}
	b, _ := got.ToBool()
	if !b {
		t.Fatalf("want TRUE for present token")
	}

	got, _ = BindSearch(tl, value.StringValue("gamma"))
	b, _ = got.ToBool()
	if b {
		t.Fatalf("want FALSE for missing token")
	}

	// Empty tokenlist returns FALSE.
	got, _ = BindSearch(value.BytesValue([]byte{}), value.StringValue("alpha"))
	b, _ = got.ToBool()
	if b {
		t.Fatalf("empty tokenlist must be FALSE")
	}

	// Invalid JSON tokenlist returns FALSE.
	got, _ = BindSearch(value.BytesValue([]byte("not-json")), value.StringValue("alpha"))
	b, _ = got.ToBool()
	if b {
		t.Fatalf("invalid tokenlist must be FALSE")
	}
}
