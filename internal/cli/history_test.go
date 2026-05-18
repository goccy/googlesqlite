package cli

import (
	"context"
	"strings"
	"testing"
)

func TestHistoryExport(t *testing.T) {
	r := newTestRunner(t)
	ctx := context.Background()

	var h History
	h.Add(r.Exec(ctx, "CREATE TABLE t (id INT64)"))
	h.Add(r.Exec(ctx, "INSERT INTO t VALUES (1), (2)"))
	h.Add(r.Exec(ctx, "SELECT id FROM t ORDER BY id"))

	if h.Len() != 3 {
		t.Fatalf("history length = %d, want 3", h.Len())
	}

	jsonOut, err := h.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON: %v", err)
	}
	if !strings.Contains(jsonOut, `"statement"`) {
		t.Errorf("JSON export missing statement field:\n%s", jsonOut)
	}

	sqlOut := h.ExportSQL()
	if !strings.Contains(sqlOut, "CREATE TABLE t") || !strings.Contains(sqlOut, ";") {
		t.Errorf("SQL export looks wrong:\n%s", sqlOut)
	}

	mdOut := h.ExportMarkdown()
	if !strings.Contains(mdOut, "```sql") || !strings.Contains(mdOut, "| id |") {
		t.Errorf("Markdown export looks wrong:\n%s", mdOut)
	}

	csvOut, err := h.ExportCSV()
	if err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}
	if !strings.Contains(csvOut, "id") || !strings.Contains(csvOut, "1") {
		t.Errorf("CSV export looks wrong:\n%s", csvOut)
	}

	if _, err := h.Export("bogus"); err == nil {
		t.Errorf("Export(bogus) should fail")
	}
}
