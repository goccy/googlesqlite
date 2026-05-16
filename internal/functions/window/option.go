package window

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	googlesql "github.com/goccy/go-googlesql"
	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"
)

type WindowFuncOptionType string

const (
	WindowFuncOptionUnknown   WindowFuncOptionType = "window_unknown"
	WindowFuncOptionFrameUnit WindowFuncOptionType = "window_frame_unit"
	WindowFuncOptionStart     WindowFuncOptionType = "window_boundary_start"
	WindowFuncOptionEnd       WindowFuncOptionType = "window_boundary_end"
	WindowFuncOptionPartition WindowFuncOptionType = "window_partition"
	WindowFuncOptionRowID     WindowFuncOptionType = "window_rowid"
	WindowFuncOptionOrderBy   WindowFuncOptionType = "window_order_by"
)

type WindowFuncOption struct {
	Type  WindowFuncOptionType `json:"type"`
	Value any                  `json:"value"`
}

func (o *WindowFuncOption) UnmarshalJSON(b []byte) error {
	type windowFuncOption WindowFuncOption

	var v windowFuncOption
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	o.Type = v.Type
	switch v.Type {
	case WindowFuncOptionFrameUnit:
		var val struct {
			Value WindowFrameUnitType `json:"value"`
		}
		if err := json.Unmarshal(b, &val); err != nil {
			return err
		}
		o.Value = val.Value
	case WindowFuncOptionStart, WindowFuncOptionEnd:
		var val struct {
			Value *WindowBoundary `json:"value"`
		}
		if err := json.Unmarshal(b, &val); err != nil {
			return err
		}
		o.Value = val.Value
	case WindowFuncOptionRowID:
		var val struct {
			Value int64 `json:"value"`
		}
		if err := json.Unmarshal(b, &val); err != nil {
			return err
		}
		o.Value = val.Value
	case WindowFuncOptionPartition:
		val, err := value.DecodeValue(v.Value)
		if err != nil {
			return fmt.Errorf("failed to convert %v to Value: %w", v.Value, err)
		}
		o.Value = val
	case WindowFuncOptionOrderBy:
		var val struct {
			Value *WindowOrderBy `json:"value"`
		}
		if err := json.Unmarshal(b, &val); err != nil {
			return err
		}
		o.Value = val.Value
	}
	return nil
}

type WindowFrameUnitType int

const (
	WindowFrameUnitUnknown WindowFrameUnitType = 0
	WindowFrameUnitRows    WindowFrameUnitType = 1
	WindowFrameUnitRange   WindowFrameUnitType = 2
)

type WindowBoundaryType int

const (
	WindowBoundaryTypeUnknown    WindowBoundaryType = 0
	WindowUnboundedPrecedingType WindowBoundaryType = 1
	WindowOffsetPrecedingType    WindowBoundaryType = 2
	WindowCurrentRowType         WindowBoundaryType = 3
	WindowOffsetFollowingType    WindowBoundaryType = 4
	WindowUnboundedFollowingType WindowBoundaryType = 5
)

type WindowBoundary struct {
	Type   WindowBoundaryType `json:"type"`
	Offset int64              `json:"offset"`
}

func GetWindowFrameUnitOptionFuncSQL(frameUnit googlesql.ResolvedWindowFrameEnums_FrameUnit) string {
	var typ WindowFrameUnitType
	switch frameUnit {
	case googlesql.ResolvedWindowFrameEnums_FrameUnitRows:
		typ = WindowFrameUnitRows
	case googlesql.ResolvedWindowFrameEnums_FrameUnitRange:
		typ = WindowFrameUnitRange
	}
	return fmt.Sprintf("googlesqlite_window_frame_unit(%d)", typ)
}

func toWindowBoundaryType(boundaryType googlesql.ResolvedWindowFrameExprEnums_BoundaryType) WindowBoundaryType {
	switch boundaryType {
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedPreceding:
		return WindowUnboundedPrecedingType
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetPreceding:
		return WindowOffsetPrecedingType
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeCurrentRow:
		return WindowCurrentRowType
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeOffsetFollowing:
		return WindowOffsetFollowingType
	case googlesql.ResolvedWindowFrameExprEnums_BoundaryTypeUnboundedFollowing:
		return WindowUnboundedFollowingType
	}
	return WindowBoundaryTypeUnknown
}

func GetWindowBoundaryStartOptionFuncSQL(boundaryType googlesql.ResolvedWindowFrameExprEnums_BoundaryType, offset string) string {
	typ := toWindowBoundaryType(boundaryType)
	if offset == "" {
		offset = "0"
	}
	return fmt.Sprintf("googlesqlite_window_boundary_start(%d, %s)", typ, offset)
}

func GetWindowBoundaryEndOptionFuncSQL(boundaryType googlesql.ResolvedWindowFrameExprEnums_BoundaryType, offset string) string {
	typ := toWindowBoundaryType(boundaryType)
	if offset == "" {
		offset = "0"
	}
	return fmt.Sprintf("googlesqlite_window_boundary_end(%d, %s)", typ, offset)
}

func GetWindowPartitionOptionFuncSQL(column string) string {
	return fmt.Sprintf("googlesqlite_window_partition(%s)", column)
}

func GetWindowRowIDOptionFuncSQL() string {
	return "googlesqlite_window_rowid(`row_id`)"
}

func GetWindowOrderByOptionFuncSQL(column string, isAsc bool) string {
	return fmt.Sprintf("googlesqlite_window_order_by(%s, %t)", column, isAsc)
}

func WINDOW_FRAME_UNIT(frameUnit int64) (value.Value, error) {
	b, err := json.Marshal(&WindowFuncOption{
		Type:  WindowFuncOptionFrameUnit,
		Value: frameUnit,
	})
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

func WINDOW_BOUNDARY_START(boundaryType, offset int64) (value.Value, error) {
	b, err := json.Marshal(&WindowFuncOption{
		Type: WindowFuncOptionStart,
		Value: &WindowBoundary{
			Type:   WindowBoundaryType(boundaryType),
			Offset: offset,
		},
	})
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

func WINDOW_BOUNDARY_END(boundaryType, offset int64) (value.Value, error) {
	b, err := json.Marshal(&WindowFuncOption{
		Type: WindowFuncOptionEnd,
		Value: &WindowBoundary{
			Type:   WindowBoundaryType(boundaryType),
			Offset: offset,
		},
	})
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

func WINDOW_PARTITION(partition value.Value) (value.Value, error) {
	v, err := value.EncodeValue(partition)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(&WindowFuncOption{
		Type:  WindowFuncOptionPartition,
		Value: v,
	})
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

func WINDOW_ROWID(id int64) (value.Value, error) {
	b, err := json.Marshal(&WindowFuncOption{
		Type:  WindowFuncOptionRowID,
		Value: id,
	})
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

type WindowOrderBy struct {
	Value value.Value `json:"value"`
	IsAsc bool        `json:"isAsc"`
}

func (w *WindowOrderBy) UnmarshalJSON(b []byte) error {
	var v struct {
		Value any  `json:"value"`
		IsAsc bool `json:"isAsc"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	val, err := value.DecodeValue(v.Value)
	if err != nil {
		return err
	}
	w.Value = val
	w.IsAsc = v.IsAsc
	return nil
}

func WINDOW_ORDER_BY(val value.Value, isAsc bool) (value.Value, error) {
	v, err := value.EncodeValue(val)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(&WindowFuncOption{
		Type: WindowFuncOptionOrderBy,
		Value: struct {
			Value any  `json:"value"`
			IsAsc bool `json:"isAsc"`
		}{
			Value: v,
			IsAsc: isAsc,
		},
	})
	if err != nil {
		return nil, err
	}
	return value.StringValue(string(b)), nil
}

type WindowFuncStatus struct {
	FrameUnit  WindowFrameUnitType
	Start      *WindowBoundary
	End        *WindowBoundary
	Partitions []value.Value
	RowID      int64
	OrderBy    []*WindowOrderBy
}

func (s *WindowFuncStatus) Partition() (string, error) {
	partitions := make([]string, 0, len(s.Partitions))
	for _, p := range s.Partitions {
		text, err := p.ToString()
		if err != nil {
			return "", err
		}
		partitions = append(partitions, text)
	}
	return strings.Join(partitions, "_"), nil
}

func parseWindowOptions(args ...value.Value) ([]value.Value, *WindowFuncStatus) {
	var (
		filteredArgs []value.Value
		opt          = &WindowFuncStatus{}
	)
	for _, arg := range args {
		if arg == nil {
			filteredArgs = append(filteredArgs, nil)
			continue
		}
		text, err := arg.ToString()
		if err != nil {
			filteredArgs = append(filteredArgs, arg)
			continue
		}
		var v WindowFuncOption
		if err := json.Unmarshal([]byte(text), &v); err != nil {
			filteredArgs = append(filteredArgs, arg)
			continue
		}
		switch v.Type {
		case WindowFuncOptionFrameUnit:
			opt.FrameUnit = v.Value.(WindowFrameUnitType)
		case WindowFuncOptionStart:
			opt.Start = v.Value.(*WindowBoundary)
		case WindowFuncOptionEnd:
			opt.End = v.Value.(*WindowBoundary)
		case WindowFuncOptionPartition:
			opt.Partitions = append(opt.Partitions, v.Value.(value.Value))
		case WindowFuncOptionRowID:
			opt.RowID = v.Value.(int64)
		case WindowFuncOptionOrderBy:
			opt.OrderBy = append(opt.OrderBy, v.Value.(*WindowOrderBy))
		default:
			filteredArgs = append(filteredArgs, arg)
			continue
		}
	}
	return filteredArgs, opt
}

type WindowOrderedValue struct {
	OrderBy []*WindowOrderBy
	Value   value.Value
}

type PartitionedValue struct {
	Partition string
	Value     *WindowOrderedValue
}

type WindowFuncAggregatedStatus struct {
	FrameUnit            WindowFrameUnitType
	Start                *WindowBoundary
	End                  *WindowBoundary
	RowID                int64
	once                 sync.Once
	PartitionToValuesMap map[string][]*WindowOrderedValue
	PartitionedValues    []*PartitionedValue
	Values               []*WindowOrderedValue
	SortedValues         []*WindowOrderedValue
	// IgnoreNullsOpt and DistinctOpt mirror the corresponding fields
	// of the surrounding helper.Option. Storing them as plain bools
	// (instead of holding a *helper.Option pointer) keeps this
	// type free of an aggregator-runtime import — the only reason
	// helper.Option was reachable from here in the first place.
	IgnoreNullsOpt bool
	DistinctOpt    bool
}

func newWindowFuncAggregatedStatus() *WindowFuncAggregatedStatus {
	return &WindowFuncAggregatedStatus{
		PartitionToValuesMap: map[string][]*WindowOrderedValue{},
	}
}

func (s *WindowFuncAggregatedStatus) Step(val value.Value, status *WindowFuncStatus) error {
	s.once.Do(func() {
		s.FrameUnit = status.FrameUnit
		s.Start = status.Start
		s.End = status.End
		s.RowID = status.RowID
	})
	if s.FrameUnit != status.FrameUnit {
		return fmt.Errorf("mismatch frame unit type %d != %d", s.FrameUnit, status.FrameUnit)
	}
	if s.Start != nil {
		if s.Start.Type != status.Start.Type {
			return fmt.Errorf("mismatch boundary type %d != %d", s.Start.Type, status.Start.Type)
		}
	}
	if s.End != nil {
		if s.End.Type != status.End.Type {
			return fmt.Errorf("mismatch boundary type %d != %d", s.End.Type, status.End.Type)
		}
	}
	if s.RowID != status.RowID {
		return fmt.Errorf("mismatch rowid %d != %d", s.RowID, status.RowID)
	}
	v := &WindowOrderedValue{
		OrderBy: status.OrderBy,
		Value:   val,
	}
	if len(status.Partitions) != 0 {
		partition, err := status.Partition()
		if err != nil {
			return fmt.Errorf("failed to get partition: %w", err)
		}
		s.PartitionToValuesMap[partition] = append(s.PartitionToValuesMap[partition], v)
		s.PartitionedValues = append(s.PartitionedValues, &PartitionedValue{
			Partition: partition,
			Value:     v,
		})
	}
	s.Values = append(s.Values, v)
	return nil
}

func (s *WindowFuncAggregatedStatus) Done(cb func([]value.Value, int, int) error) error {
	if s.RowID <= 0 {
		return fmt.Errorf("invalid rowid. rowid must be greater than zero")
	}
	values := s.FilteredValues()
	sortedValues := make([]*WindowOrderedValue, len(values))
	copy(sortedValues, values)
	if len(sortedValues) != 0 {
		sort.Slice(sortedValues, func(i, j int) bool {
			for orderBy := 0; orderBy < len(sortedValues[0].OrderBy); orderBy++ {
				iV := sortedValues[i].OrderBy[orderBy].Value
				jV := sortedValues[j].OrderBy[orderBy].Value
				isAsc := sortedValues[0].OrderBy[orderBy].IsAsc
				if iV == nil {
					return true
				}
				if jV == nil {
					return false
				}
				isEqual, _ := iV.EQ(jV)
				if isEqual {
					// break tie with subsequent fields
					continue
				}
				if isAsc {
					cond, _ := iV.LT(jV)
					return cond
				} else {
					cond, _ := iV.GT(jV)
					return cond
				}
			}
			return false
		})
	}
	s.SortedValues = sortedValues
	start, err := s.getIndexFromBoundary(s.Start)
	if err != nil {
		return fmt.Errorf("failed to get start index: %w", err)
	}
	end, err := s.getIndexFromBoundary(s.End)
	if err != nil {
		return fmt.Errorf("failed to get end index: %w", err)
	}
	resultValues := make([]value.Value, 0, len(sortedValues))
	for _, val := range sortedValues {
		resultValues = append(resultValues, val.Value)
	}
	if start >= len(resultValues) || end < 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}
	if end >= len(resultValues) {
		end = len(resultValues) - 1
	}
	return cb(resultValues, start, end)
}

func (s *WindowFuncAggregatedStatus) IgnoreNulls() bool {
	return s.IgnoreNullsOpt
}

func (s *WindowFuncAggregatedStatus) Distinct() bool {
	return s.DistinctOpt
}

func (s *WindowFuncAggregatedStatus) FilteredValues() []*WindowOrderedValue {
	if len(s.PartitionedValues) != 0 {
		return s.PartitionToValuesMap[s.Partition()]
	}
	return s.Values
}

func (s *WindowFuncAggregatedStatus) Partition() string {
	return s.PartitionedValues[s.RowID-1].Partition
}

func (s *WindowFuncAggregatedStatus) getIndexFromBoundary(boundary *WindowBoundary) (int, error) {
	switch s.FrameUnit {
	case WindowFrameUnitRows:
		return s.getIndexFromBoundaryByRows(boundary)
	case WindowFrameUnitRange:
		return s.getIndexFromBoundaryByRange(boundary)
	default:
		return s.currentIndexByRows()
	}
}

func (s *WindowFuncAggregatedStatus) getIndexFromBoundaryByRows(boundary *WindowBoundary) (int, error) {
	switch boundary.Type {
	case WindowUnboundedPrecedingType:
		return 0, nil
	case WindowCurrentRowType:
		return s.currentIndexByRows()
	case WindowUnboundedFollowingType:
		return len(s.FilteredValues()) - 1, nil
	case WindowOffsetPrecedingType:
		cur, err := s.currentIndexByRows()
		if err != nil {
			return 0, err
		}
		return cur - int(boundary.Offset), nil
	case WindowOffsetFollowingType:
		cur, err := s.currentIndexByRows()
		if err != nil {
			return 0, err
		}
		return cur + int(boundary.Offset), nil
	}
	return 0, fmt.Errorf("unsupported boundary type %d", boundary.Type)
}

func (s *WindowFuncAggregatedStatus) currentIndexByRows() (int, error) {
	if len(s.PartitionedValues) != 0 {
		return s.partitionedCurrentIndexByRows()
	}
	curRowID := int(s.RowID - 1)
	curValue := s.Values[curRowID]
	for idx, val := range s.SortedValues {
		if val == curValue {
			return idx, nil
		}
	}
	return 0, fmt.Errorf("failed to find current index")
}

func (s *WindowFuncAggregatedStatus) partitionedCurrentIndexByRows() (int, error) {
	curRowID := int(s.RowID - 1)
	curValue := s.PartitionedValues[curRowID]
	for idx, val := range s.SortedValues {
		if val == curValue.Value {
			return idx, nil
		}
	}
	return 0, fmt.Errorf("failed to find current index")
}

func (s *WindowFuncAggregatedStatus) getIndexFromBoundaryByRange(boundary *WindowBoundary) (int, error) {
	switch boundary.Type {
	case WindowUnboundedPrecedingType:
		return 0, nil
	case WindowUnboundedFollowingType:
		return len(s.FilteredValues()) - 1, nil
	case WindowCurrentRowType:
		val, err := s.currentRangeValue()
		if err != nil {
			return 0, err
		}
		return s.lookupMaxIndexFromRangeValue(val)
	case WindowOffsetPrecedingType:
		val, err := s.currentRangeValue()
		if err != nil {
			return 0, err
		}
		sub, err := val.Sub(value.IntValue(boundary.Offset))
		if err != nil {
			return 0, err
		}
		return s.lookupMinIndexFromRangeValue(sub)
	case WindowOffsetFollowingType:
		val, err := s.currentRangeValue()
		if err != nil {
			return 0, err
		}
		add, err := val.Add(value.IntValue(boundary.Offset))
		if err != nil {
			return 0, err
		}
		return s.lookupMaxIndexFromRangeValue(add)
	}
	return 0, fmt.Errorf("unsupported boundary type %d", boundary.Type)
}

func (s *WindowFuncAggregatedStatus) currentRangeValue() (value.Value, error) {
	if len(s.PartitionedValues) != 0 {
		return s.partitionedCurrentRangeValue()
	}
	curRowID := int(s.RowID - 1)
	curValue := s.Values[curRowID]
	if len(curValue.OrderBy) == 0 {
		return nil, fmt.Errorf("required order by column for analytic range scanning")
	}
	return curValue.OrderBy[len(curValue.OrderBy)-1].Value, nil
}

func (s *WindowFuncAggregatedStatus) partitionedCurrentRangeValue() (value.Value, error) {
	curRowID := int(s.RowID - 1)
	curValue := s.PartitionedValues[curRowID]
	if len(curValue.Value.OrderBy) == 0 {
		return nil, fmt.Errorf("required order by column for analytic range scanning")
	}
	return curValue.Value.OrderBy[len(curValue.Value.OrderBy)-1].Value, nil
}

func (s *WindowFuncAggregatedStatus) lookupMinIndexFromRangeValue(rangeValue value.Value) (int, error) {
	minIndex := -1
	for idx, val := range slices.Backward(s.SortedValues) {

		if len(val.OrderBy) == 0 {
			continue
		}
		target := val.OrderBy[len(val.OrderBy)-1].Value
		cond, err := rangeValue.LTE(target)
		if err != nil {
			return 0, err
		}
		if cond {
			minIndex = idx
		}
	}
	return minIndex, nil
}

func (s *WindowFuncAggregatedStatus) lookupMaxIndexFromRangeValue(rangeValue value.Value) (int, error) {
	maxIndex := -1
	for idx := 0; idx < len(s.SortedValues); idx++ {
		val := s.SortedValues[idx]
		if len(val.OrderBy) == 0 {
			continue
		}
		target := val.OrderBy[len(val.OrderBy)-1].Value
		cond, err := rangeValue.GTE(target)
		if err != nil {
			return 0, err
		}
		if cond {
			maxIndex = idx
		}
	}
	return maxIndex, nil
}
