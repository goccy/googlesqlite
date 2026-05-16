// Tests for the option-marker SQL emitters and the resolved-tree
// boundary / frame enum translators in option.go.

package window

import (
	"strings"
	"testing"

	googlesql "github.com/goccy/go-googlesql"
)

// TestGetWindowFrameUnitOptionFuncSQLAllUnits drives every branch of
// GetWindowFrameUnitOptionFuncSQL.
func TestGetWindowFrameUnitOptionFuncSQLAllUnits(t *testing.T) {
	t.Parallel()
	for _, fu := range []googlesql.ResolvedWindowFrameEnums_FrameUnit{
		googlesql.ResolvedWindowFrameEnums_FrameUnitRows,
		googlesql.ResolvedWindowFrameEnums_FrameUnitRange,
	} {
		s := GetWindowFrameUnitOptionFuncSQL(fu)
		if !strings.HasPrefix(s, "googlesqlite_window_frame_unit(") {
			t.Errorf("FrameUnit(%v): got %q; want googlesqlite_window_frame_unit(...)", fu, s)
		}
	}
}

// TestGetWindowBoundaryStartEndOptionFuncSQLAllBoundaries drives
// every branch of toWindowBoundaryType / GetWindowBoundaryStart-
// OptionFuncSQL / GetWindowBoundaryEndOptionFuncSQL, including the
// offset-bearing OffsetPreceding / OffsetFollowing cases.
func TestGetWindowBoundaryStartEndOptionFuncSQLAllBoundaries(t *testing.T) {
	t.Parallel()
	for _, bt := range []googlesql.ResolvedWindowFrameExprEnums_BoundaryType{
		googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedPreceding,
		googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetPreceding,
		googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeCurrentRow,
		googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetFollowing,
		googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedFollowing,
	} {
		startSQL := GetWindowBoundaryStartOptionFuncSQL(bt, "1")
		if !strings.HasPrefix(startSQL, "googlesqlite_window_boundary_start(") {
			t.Errorf("start SQL for %v: got %q", bt, startSQL)
		}
		endSQL := GetWindowBoundaryEndOptionFuncSQL(bt, "1")
		if !strings.HasPrefix(endSQL, "googlesqlite_window_boundary_end(") {
			t.Errorf("end SQL for %v: got %q", bt, endSQL)
		}
		// Empty offset -> default "0" on both start and end.
		emptyStart := GetWindowBoundaryStartOptionFuncSQL(bt, "")
		if !strings.Contains(emptyStart, "0") {
			t.Errorf("start SQL for %v empty offset: got %q; want default 0", bt, emptyStart)
		}
		emptyEnd := GetWindowBoundaryEndOptionFuncSQL(bt, "")
		if !strings.Contains(emptyEnd, "0") {
			t.Errorf("end SQL for %v empty offset: got %q; want default 0", bt, emptyEnd)
		}
	}
}

// TestGetWindowPartitionOrderByRowIDOptionFuncSQL drives the simpler
// option-marker emitters.
func TestGetWindowPartitionOrderByRowIDOptionFuncSQL(t *testing.T) {
	t.Parallel()
	if s := GetWindowPartitionOptionFuncSQL("col"); !strings.HasPrefix(s, "googlesqlite_window_partition(") {
		t.Errorf("partition: got %q", s)
	}
	if s := GetWindowOrderByOptionFuncSQL("col", true); !strings.HasPrefix(s, "googlesqlite_window_order_by(") {
		t.Errorf("order_by: got %q", s)
	}
	if s := GetWindowRowIDOptionFuncSQL(); !strings.HasPrefix(s, "googlesqlite_window_rowid(") {
		t.Errorf("rowid: got %q", s)
	}
}
