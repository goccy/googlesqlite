package googlesqlite_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/goccy/googlesqlite"
)

// TestToJSONStringQuotesDateTimeTypes covers the JSON encodings of DATE,
// DATETIME, TIME and TIMESTAMP, per
// https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#json_encodings
// (each is a JSON string, e.g. DATE '2017-03-06' -> "2017-03-06"). They were
// emitted bare, so TO_JSON_STRING(STRUCT(DATE '2017-03-06' AS d)) returned
// {"d":2017-03-06}, which is not valid JSON.
func TestToJSONStringQuotesDateTimeTypes(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("googlesqlite", ":memory:?_test=to_json_string_datetime")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	for _, tc := range []struct {
		query string
		want  string
	}{
		{"SELECT TO_JSON_STRING(DATE '2017-03-06')", `"2017-03-06"`},
		{"SELECT TO_JSON_STRING(DATETIME '2017-03-06 12:34:56.789012')", `"2017-03-06T12:34:56.789012"`},
		{"SELECT TO_JSON_STRING(TIME '12:34:56.789012')", `"12:34:56.789012"`},
		{"SELECT TO_JSON_STRING(TIMESTAMP '2017-03-06 12:34:56.789012')", `"2017-03-06T12:34:56.789012Z"`},
		{"SELECT TO_JSON_STRING(STRUCT(DATE '2017-03-06' AS d, 1 AS n))", `{"d":"2017-03-06","n":1}`},
		{"SELECT TO_JSON_STRING([DATE '2017-03-06', DATE '2017-03-07'])", `["2017-03-06","2017-03-07"]`},
		{"SELECT TO_JSON_STRING(TO_JSON(STRUCT(DATE '2017-03-06' AS d)))", `{"d":"2017-03-06"}`},
	} {
		var got string
		if err := db.QueryRowContext(ctx, tc.query).Scan(&got); err != nil {
			t.Errorf("%s: %v", tc.query, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s = %s; want %s", tc.query, got, tc.want)
		}
		if !json.Valid([]byte(got)) {
			t.Errorf("%s = %s: not valid JSON", tc.query, got)
		}
	}
}
