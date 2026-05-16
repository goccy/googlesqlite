package internal

import (
	"fmt"
	"strings"
)

// applyGapFillRewrite rewrites every `GAP_FILL(...)` invocation in
// `query` to its equivalent SQL using GENERATE_TIMESTAMP_ARRAY +
// LEFT JOIN + window-function gap fill. The upstream analyzer does
// not know GAP_FILL natively, so we lower it before
// AnalyzeStatement runs.
//
// Supported forms (per the BigQuery spec):
//
//	GAP_FILL(
//	  TABLE <name>,          | (<subquery>),
//	  ts_column => '<col>',
//	  bucket_width => INTERVAL <n> <part>,
//	  [partitioning_columns => ['c1', 'c2', ...]],
//	  [value_columns => [('col1', 'null'|'locf'|'linear'), ...]],
//	  [origin => <DATE|DATETIME|TIMESTAMP literal>],
//	  [ignore_null_values => TRUE|FALSE]
//	)
//
// String literals and backtick-quoted identifiers are honoured by
// the scanner, so embedded `GAP_FILL` substrings stay untouched.
func applyGapFillRewrite(query string) string {
	if !containsCaseInsensitive(query, "GAP_FILL") {
		return query
	}
	for {
		start, end, args, ok := findGapFillCall(query)
		if !ok {
			return query
		}
		replacement, err := rewriteGapFill(args)
		if err != nil {
			// On parse failure, leave the call untouched — the
			// analyzer will surface a clearer "Table-valued function
			// not found: GAP_FILL" error than we could here.
			return query
		}
		query = query[:start] + replacement + query[end:]
	}
}

func containsCaseInsensitive(s, needle string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(needle))
}

// findGapFillCall locates the next `GAP_FILL(...)` invocation,
// returning the byte range covering `GAP_FILL ( ... )` plus the raw
// inner argument string. Skips string literals and quoted
// identifiers; tracks paren depth.
func findGapFillCall(s string) (start, end int, args string, ok bool) {
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
			word := s[i:j]
			if strings.EqualFold(word, "GAP_FILL") {
				// Skip whitespace before `(`.
				k := j
				for k < len(s) && (s[k] == ' ' || s[k] == '\t' || s[k] == '\n' || s[k] == '\r') {
					k++
				}
				if k < len(s) && s[k] == '(' {
					closeIdx := matchingCloseParen(s, k)
					if closeIdx > 0 {
						return i, closeIdx + 1, s[k+1 : closeIdx], true
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

// matchingCloseParen returns the offset of the `)` matching the
// `(` at `open`, honouring nested parens, string literals, and
// backtick-quoted identifiers. Returns -1 when unmatched.
func matchingCloseParen(s string, open int) int {
	depth := 0
	i := open
	for i < len(s) {
		c := s[i]
		switch {
		case c == '(':
			depth++
		case c == ')':
			depth--
			if depth == 0 {
				return i
			}
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
		}
		i++
	}
	return -1
}

// gapFillArgs is the parsed argument set extracted from a
// GAP_FILL invocation.
type gapFillArgs struct {
	inputTable         string // bare table name, when used in `TABLE name` form
	inputSubquery      string // `( ... )` SQL, when used in subquery form
	tsColumn           string
	bucketWidth        string // raw SQL of INTERVAL expression
	partitioningCols   []string
	valueCols          []gapFillValueCol
	origin             string // raw SQL of origin expression, may be empty
	ignoreNullValues   bool
	ignoreNullExplicit bool
}

type gapFillValueCol struct {
	name   string
	method string // "null" | "locf" | "linear"
}

func parseGapFillArgs(raw string) (*gapFillArgs, error) {
	parts := splitTopLevelCommas(raw)
	if len(parts) < 1 {
		return nil, fmt.Errorf("GAP_FILL: empty argument list")
	}
	args := &gapFillArgs{ignoreNullValues: true}
	// First positional: TABLE name or (subquery).
	first := strings.TrimSpace(parts[0])
	if up := strings.ToUpper(first); strings.HasPrefix(up, "TABLE ") || strings.HasPrefix(up, "TABLE\t") || strings.HasPrefix(up, "TABLE\n") {
		args.inputTable = strings.TrimSpace(first[len("TABLE"):])
	} else if strings.HasPrefix(first, "(") {
		// Drop the outermost parens.
		args.inputSubquery = strings.TrimSpace(first[1 : len(first)-1])
	} else {
		// Some callers omit the TABLE keyword.
		args.inputTable = first
	}
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		eq := strings.Index(p, "=>")
		if eq < 0 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(p[:eq]))
		valStr := strings.TrimSpace(p[eq+2:])
		switch name {
		case "ts_column":
			args.tsColumn = stripQuotes(valStr)
		case "bucket_width":
			args.bucketWidth = valStr
		case "partitioning_columns":
			args.partitioningCols = parseStringArray(valStr)
		case "value_columns":
			args.valueCols = parseValueColumnsArray(valStr)
		case "origin":
			args.origin = valStr
		case "ignore_null_values":
			args.ignoreNullExplicit = true
			args.ignoreNullValues = strings.EqualFold(valStr, "TRUE")
		}
	}
	if args.tsColumn == "" {
		return nil, fmt.Errorf("GAP_FILL: ts_column is required")
	}
	if args.bucketWidth == "" {
		return nil, fmt.Errorf("GAP_FILL: bucket_width is required")
	}
	if args.inputTable == "" && args.inputSubquery == "" {
		return nil, fmt.Errorf("GAP_FILL: input table required")
	}
	return args, nil
}

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

// splitTopLevelCommas splits a raw argument list at commas that are
// not inside parens / brackets / string literals.
func splitTopLevelCommas(s string) []string {
	var out []string
	depth := 0
	last := 0
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '(' || c == '[':
			depth++
		case c == ')' || c == ']':
			depth--
		case c == '\'' || c == '"':
			i = scanStringLiteral(s, i)
			continue
		case c == ',' && depth == 0:
			out = append(out, s[last:i])
			last = i + 1
		}
		i++
	}
	if last < len(s) {
		out = append(out, s[last:])
	}
	return out
}

// parseStringArray accepts `['a', 'b', "c"]` style ARRAY literals.
func parseStringArray(s string) []string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}
	parts := splitTopLevelCommas(s)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := stripQuotes(strings.TrimSpace(p))
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// parseValueColumnsArray accepts the
// `[('col1','locf'), ('col2','linear')]` shape.
func parseValueColumnsArray(s string) []gapFillValueCol {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}
	parts := splitTopLevelCommas(s)
	var out []gapFillValueCol
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if !strings.HasPrefix(p, "(") || !strings.HasSuffix(p, ")") {
			continue
		}
		inner := p[1 : len(p)-1]
		fields := splitTopLevelCommas(inner)
		if len(fields) < 2 {
			continue
		}
		col := stripQuotes(strings.TrimSpace(fields[0]))
		method := strings.ToLower(stripQuotes(strings.TrimSpace(fields[1])))
		if col == "" {
			continue
		}
		out = append(out, gapFillValueCol{name: col, method: method})
	}
	return out
}

// rewriteGapFill produces a SQL subquery that has the same shape as
// `GAP_FILL(...)` and can be referenced directly. The result is
// wrapped in parens so callers can embed it in FROM positions.
//
// The lowered SQL is structured as a chain of nested subqueries:
//
//	level 0: __gf_aligned   — input rows tagged with their bucket
//	level 1: __gf_ticks     — generated bucket boundaries
//	level 2: __gf_joined    — LEFT JOIN of ticks vs aligned
//	level 3: __gf_grouped   — COUNT(col) OVER as a per-column LOCF
//	                          group id (the standard SQL idiom for
//	                          IGNORE NULLS without engine support)
//	level 4: final SELECT   — null / locf / linear projection
func rewriteGapFill(rawArgs string) (string, error) {
	args, err := parseGapFillArgs(rawArgs)
	if err != nil {
		return "", err
	}
	inputRef := args.inputTable
	if inputRef == "" {
		inputRef = "(" + args.inputSubquery + ")"
	}
	ts := args.tsColumn
	bw := args.bucketWidth
	origin := args.origin
	if origin == "" {
		origin = "TIMESTAMP '1950-01-01 00:00:00+00'"
	}

	var partitionExprs []string
	for _, c := range args.partitioningCols {
		partitionExprs = append(partitionExprs, "`"+c+"`")
	}
	partitionList := strings.Join(partitionExprs, ", ")

	// Aligned source: every input row tagged with the bucket
	// boundary that contains its timestamp.
	alignedSelect := fmt.Sprintf(
		"SELECT TIMESTAMP_BUCKET(`%s`, %s, %s) AS `%s`, * EXCEPT(`%s`) FROM %s",
		ts, bw, origin, ts, ts, inputRef,
	)

	// Bucket-tick generation. Single-partition form when no
	// partitioning_columns are present, per-partition CROSS JOIN
	// against the partition bounds otherwise.
	var ticks string
	if partitionList == "" {
		ticks = fmt.Sprintf(
			"SELECT bucket_ts AS `%s` "+
				"FROM (SELECT MIN(`%s`) AS mn, MAX(`%s`) AS mx FROM (%s)), "+
				"UNNEST(GENERATE_TIMESTAMP_ARRAY(TIMESTAMP_BUCKET(mn, %s, %s), TIMESTAMP_BUCKET(mx, %s, %s), %s)) AS bucket_ts",
			ts, ts, ts, alignedSelect, bw, origin, bw, origin, bw,
		)
	} else {
		ticks = fmt.Sprintf(
			"SELECT bucket_ts AS `%s`, %s "+
				"FROM (SELECT %s, MIN(`%s`) AS mn, MAX(`%s`) AS mx FROM (%s) GROUP BY %s), "+
				"UNNEST(GENERATE_TIMESTAMP_ARRAY(TIMESTAMP_BUCKET(mn, %s, %s), TIMESTAMP_BUCKET(mx, %s, %s), %s)) AS bucket_ts",
			ts, partitionList, partitionList, ts, ts, alignedSelect, partitionList, bw, origin, bw, origin, bw,
		)
	}

	// LEFT JOIN ticks ↔ aligned. We hand-craft the join key list so
	// that USING() picks up the partitioning_columns plus ts.
	joinUsing := "`" + ts + "`"
	for _, c := range args.partitioningCols {
		joinUsing += ", `" + c + "`"
	}
	joined := fmt.Sprintf(
		"SELECT * FROM (%s) AS `__gf_ticks` LEFT JOIN (%s) AS `__gf_src` USING (%s)",
		ticks, alignedSelect, joinUsing,
	)

	// For LOCF / linear we need a per-column "group id" computed as
	// COUNT(col) OVER (ORDER BY ts). Rows that share the same group
	// id all sit downstream of the same last-non-null value.
	partClause := ""
	if partitionList != "" {
		partClause = "PARTITION BY " + partitionList + " "
	}
	var groupedCols []string
	groupedCols = append(groupedCols, "*")
	needGroups := false
	for _, vc := range args.valueCols {
		if vc.method == "locf" || vc.method == "linear" {
			needGroups = true
			groupedCols = append(groupedCols, fmt.Sprintf(
				"COUNT(`%s`) OVER (%sORDER BY `%s` ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS `__gf_grp_%s`",
				vc.name, partClause, ts, vc.name,
			))
		}
		if vc.method == "linear" {
			// Also need the forward-looking group id, for the
			// "next non-null" lookup.
			groupedCols = append(groupedCols, fmt.Sprintf(
				"COUNT(`%s`) OVER (%sORDER BY `%s` ROWS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING) AS `__gf_fgrp_%s`",
				vc.name, partClause, ts, vc.name,
			))
		}
	}

	grouped := joined
	if needGroups {
		grouped = fmt.Sprintf("SELECT %s FROM (%s) AS `__gf_j`", strings.Join(groupedCols, ", "), joined)
	}

	// Final projection.
	var projCols []string
	projCols = append(projCols, "`"+ts+"`")
	for _, p := range args.partitioningCols {
		projCols = append(projCols, "`"+p+"`")
	}
	for _, vc := range args.valueCols {
		switch vc.method {
		case "locf":
			// LOCF: take MAX of the non-null value within the
			// LOCF group. The window partition includes the
			// per-row group id so each group sees a single
			// non-null value (or NULL if no preceding value).
			locfPart := "PARTITION BY `__gf_grp_" + vc.name + "`"
			if partitionList != "" {
				locfPart = "PARTITION BY " + partitionList + ", `__gf_grp_" + vc.name + "`"
			}
			projCols = append(projCols, fmt.Sprintf(
				"MAX(`%s`) OVER (%s) AS `%s`", vc.name, locfPart, vc.name,
			))
		case "linear":
			// Linear interpolation: find previous non-null
			// (value + timestamp) via PARTITION BY __gf_grp_X,
			// next non-null via PARTITION BY __gf_fgrp_X, then
			// linearly interpolate.
			prevPart := "PARTITION BY `__gf_grp_" + vc.name + "`"
			nextPart := "PARTITION BY `__gf_fgrp_" + vc.name + "`"
			if partitionList != "" {
				prevPart = "PARTITION BY " + partitionList + ", `__gf_grp_" + vc.name + "`"
				nextPart = "PARTITION BY " + partitionList + ", `__gf_fgrp_" + vc.name + "`"
			}
			prevV := fmt.Sprintf("MAX(`%s`) OVER (%s)", vc.name, prevPart)
			nextV := fmt.Sprintf("MAX(`%s`) OVER (%s)", vc.name, nextPart)
			prevT := fmt.Sprintf("MAX(IF(`%s` IS NULL, NULL, `%s`)) OVER (%s)", vc.name, ts, prevPart)
			nextT := fmt.Sprintf("MAX(IF(`%s` IS NULL, NULL, `%s`)) OVER (%s)", vc.name, ts, nextPart)
			interp := fmt.Sprintf(
				"CASE WHEN `%s` IS NOT NULL THEN `%s` "+
					"WHEN (%s) IS NULL OR (%s) IS NULL THEN NULL "+
					"WHEN TIMESTAMP_DIFF((%s), (%s), SECOND) = 0 THEN (%s) "+
					"ELSE (%s) + (TIMESTAMP_DIFF(`%s`, (%s), SECOND) * CAST(((%s) - (%s)) AS FLOAT64) / TIMESTAMP_DIFF((%s), (%s), SECOND)) END",
				vc.name, vc.name,
				prevV, nextV,
				nextT, prevT, prevV,
				prevV, ts, prevT,
				nextV, prevV, nextT, prevT,
			)
			projCols = append(projCols, fmt.Sprintf("%s AS `%s`", interp, vc.name))
		case "null", "":
			projCols = append(projCols, fmt.Sprintf("`%s`", vc.name))
		default:
			return "", fmt.Errorf("GAP_FILL: unknown method %q for column %q", vc.method, vc.name)
		}
	}

	out := fmt.Sprintf(
		"(SELECT %s FROM (%s) AS `__gf_g` ORDER BY `%s`)",
		strings.Join(projCols, ", "),
		grouped, ts,
	)
	return out, nil
}
