package internal

import (
	"fmt"
	"strings"
)

// applyRangeSessionizeRewrite rewrites every `RANGE_SESSIONIZE(...)`
// invocation in `query` to its equivalent SQL using window
// functions over the source table's `RANGE` column. The upstream
// analyzer does not surface RANGE_SESSIONIZE as a TVF for us, so
// we lower it before AnalyzeStatement runs — the same mechanism
// applyGapFillRewrite uses.
//
// Supported form (per the BigQuery / GoogleSQL spec):
//
//	RANGE_SESSIONIZE(
//	  TABLE <name>,           | (<subquery>),
//	  '<range_column>',
//	  ['<partition_col>', ...],
//	  ['MEETS' | 'OVERLAPS']
//	)
//
// MEETS (default): ranges that meet or overlap form a session — a
// new session starts when current.range_start > prev.max(range_end).
//
// OVERLAPS: only ranges that strictly overlap form a session — a
// new session starts when current.range_start >= prev.max(range_end).
//
// String literals and backtick-quoted identifiers are honoured by the
// scanner, so embedded `RANGE_SESSIONIZE` substrings stay untouched.
func applyRangeSessionizeRewrite(query string) string {
	if !containsCaseInsensitive(query, "RANGE_SESSIONIZE") {
		return query
	}
	for {
		start, end, args, ok := findRangeSessionizeCall(query)
		if !ok {
			return query
		}
		replacement, err := rewriteRangeSessionize(args)
		if err != nil {
			// On parse failure, leave the call untouched — the analyzer
			// will surface a clearer "Table-valued function not found"
			// error than we could here.
			return query
		}
		query = query[:start] + replacement + query[end:]
	}
}

// findRangeSessionizeCall locates the next `RANGE_SESSIONIZE(...)`
// invocation, returning the byte range covering it plus the raw
// inner argument string.
func findRangeSessionizeCall(s string) (start, end int, args string, ok bool) {
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '\'' || c == '"':
			i = scanStringLiteral(s, i)
			continue
		case c == '`':
			j := i + 1
			for j < len(s) && s[j] != '`' {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j
			continue
		case c == '-' && i+1 < len(s) && s[i+1] == '-':
			j := i
			for j < len(s) && s[j] != '\n' {
				j++
			}
			i = j
			continue
		}
		if isIdentStart(c) {
			j := i + 1
			for j < len(s) && isIdentCont(s[j]) {
				j++
			}
			word := strings.ToUpper(s[i:j])
			if word == "RANGE_SESSIONIZE" {
				// Skip whitespace before the opening paren.
				k := j
				for k < len(s) && (s[k] == ' ' || s[k] == '\t' || s[k] == '\n' || s[k] == '\r') {
					k++
				}
				if k < len(s) && s[k] == '(' {
					close := matchingCloseParen(s, k)
					if close > k {
						return i, close + 1, s[k+1 : close], true
					}
				}
			}
			i = j
			continue
		}
		i++
	}
	return 0, 0, "", false
}

type rangeSessionizeArgs struct {
	inputTable       string // bare table name, when used in `TABLE name` form
	inputSubquery    string // `( ... )` SQL, when used in subquery form
	rangeColumn      string
	partitioningCols []string
	mode             string // "MEETS" | "OVERLAPS"
}

func parseRangeSessionizeArgs(raw string) (*rangeSessionizeArgs, error) {
	parts := splitTopLevelCommas(raw)
	if len(parts) < 3 {
		return nil, fmt.Errorf("RANGE_SESSIONIZE: need at least 3 arguments, got %d", len(parts))
	}
	args := &rangeSessionizeArgs{mode: "MEETS"}
	first := strings.TrimSpace(parts[0])
	if up := strings.ToUpper(first); strings.HasPrefix(up, "TABLE ") || strings.HasPrefix(up, "TABLE\t") || strings.HasPrefix(up, "TABLE\n") {
		args.inputTable = strings.TrimSpace(first[len("TABLE"):])
	} else if strings.HasPrefix(first, "(") {
		args.inputSubquery = strings.TrimSpace(first[1 : len(first)-1])
	} else {
		args.inputTable = first
	}
	args.rangeColumn = stripQuotes(strings.TrimSpace(parts[1]))
	args.partitioningCols = parseStringArray(strings.TrimSpace(parts[2]))
	if len(parts) >= 4 {
		mode := strings.ToUpper(stripQuotes(strings.TrimSpace(parts[3])))
		switch mode {
		case "MEETS", "OVERLAPS":
			args.mode = mode
		default:
			return nil, fmt.Errorf("RANGE_SESSIONIZE: invalid mode %q", mode)
		}
	}
	if args.rangeColumn == "" {
		return nil, fmt.Errorf("RANGE_SESSIONIZE: range_column is required")
	}
	if args.inputTable == "" && args.inputSubquery == "" {
		return nil, fmt.Errorf("RANGE_SESSIONIZE: input table required")
	}
	return args, nil
}

// rewriteRangeSessionize lowers RANGE_SESSIONIZE(...) into a
// parenthesised SELECT that produces the same shape — every input
// row plus a `session_range` column.
//
// The lowered SQL chains a few subqueries:
//
//	__rs_extracted: every input row tagged with RANGE_START / RANGE_END
//	                of the chosen range column.
//	__rs_marked:    each row carries the max of the prev rows'
//	                range_end within its partition (the LAG-style
//	                bound), so we can detect session boundaries.
//	__rs_grouped:   a running SUM of new-session indicators yields
//	                a stable per-partition session_id.
//	final SELECT:   project all original columns plus a session_range
//	                computed as RANGE(MIN(start), MAX(end)) within
//	                each (partition, session_id) group.
func rewriteRangeSessionize(rawArgs string) (string, error) {
	args, err := parseRangeSessionizeArgs(rawArgs)
	if err != nil {
		return "", err
	}
	inputRef := args.inputTable
	if inputRef == "" {
		inputRef = "(" + args.inputSubquery + ")"
	}
	rc := args.rangeColumn
	cmp := ">"
	if args.mode == "OVERLAPS" {
		cmp = ">="
	}

	var partitionExprs []string
	for _, c := range args.partitioningCols {
		partitionExprs = append(partitionExprs, "`"+c+"`")
	}
	partitionList := strings.Join(partitionExprs, ", ")
	partitionClause := ""
	if partitionList != "" {
		partitionClause = "PARTITION BY " + partitionList
	}
	partitionForSession := partitionList
	if partitionForSession != "" {
		partitionForSession += ", "
	}

	// Level 0: extract range bounds onto every input row.
	extracted := fmt.Sprintf(
		"SELECT *, RANGE_START(`%s`) AS __rs_start, RANGE_END(`%s`) AS __rs_end FROM %s",
		rc, rc, inputRef,
	)

	// Level 1: per-row "previous max end" within the partition.
	marked := fmt.Sprintf(
		"SELECT *, MAX(__rs_end) OVER (%s ORDER BY __rs_start ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING) AS __rs_prev_max_end FROM (%s)",
		partitionClause, extracted,
	)

	// Level 2: tag each row with a new-session flag, then accumulate.
	// The new-session indicator is 1 when there is no preceding row
	// in the partition OR when the current row's start is past the
	// previous max end (strictly past for MEETS, equal-or-past for
	// OVERLAPS).
	indicator := fmt.Sprintf(
		"SELECT *, CASE WHEN __rs_prev_max_end IS NULL THEN 1 WHEN __rs_start %s __rs_prev_max_end THEN 1 ELSE 0 END AS __rs_new_session FROM (%s)",
		cmp, marked,
	)
	grouped := fmt.Sprintf(
		"SELECT *, SUM(__rs_new_session) OVER (%s ORDER BY __rs_start ROWS UNBOUNDED PRECEDING) AS __rs_session_id FROM (%s)",
		partitionClause, indicator,
	)

	// Level 3: derive session_range. Cannot wrap the RANGE() result
	// directly around two window aggregates because RANGE expects
	// scalar args; let SQLite evaluate the window aggregates first.
	final := fmt.Sprintf(
		"SELECT * EXCEPT(__rs_start, __rs_end, __rs_prev_max_end, __rs_new_session, __rs_session_id), RANGE(MIN(__rs_start) OVER (PARTITION BY %s__rs_session_id), MAX(__rs_end) OVER (PARTITION BY %s__rs_session_id)) AS session_range FROM (%s)",
		partitionForSession, partitionForSession, grouped,
	)

	return "(" + final + ")", nil
}
