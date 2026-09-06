package json

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func PARSE_JSON(expr, mode string) (value.Value, error) {
	dec := json.NewDecoder(strings.NewReader(expr))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if rest, _ := io.ReadAll(dec.Buffered()); strings.TrimSpace(string(rest)) != "" {
		return nil, fmt.Errorf("PARSE_JSON: unexpected trailing content")
	}
	var buf bytes.Buffer
	if err := writeCanonicalJSON(&buf, v, mode == "round"); err != nil {
		return nil, err
	}
	return value.JsonValue(buf.String()), nil
}

func writeCanonicalJSON(buf *bytes.Buffer, v any, round bool) error {
	switch vv := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(vv))
		for k := range vv {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeJSONString(buf, k)
			buf.WriteByte(':')
			if err := writeCanonicalJSON(buf, vv[k], round); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case []any:
		buf.WriteByte('[')
		for i, e := range vv {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonicalJSON(buf, e, round); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case json.Number:
		s, err := canonicalJSONNumber(string(vv), round)
		if err != nil {
			return err
		}
		buf.WriteString(s)
	case string:
		writeJSONString(buf, vv)
	case bool:
		buf.WriteString(strconv.FormatBool(vv))
	case nil:
		buf.WriteString("null")
	default:
		return fmt.Errorf("PARSE_JSON: unexpected value %T", v)
	}
	return nil
}

func canonicalJSONNumber(s string, round bool) (string, error) {
	if !strings.ContainsAny(s, ".eE") {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return strconv.FormatInt(i, 10), nil
		}
		if u, err := strconv.ParseUint(s, 10, 64); err == nil {
			return strconv.FormatUint(u, 10), nil
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(f, 0) {
		return "", fmt.Errorf("PARSE_JSON: number overflow parsing '%s'", s)
	}
	out := formatJSONFloat(f)
	if !round {
		in, ok := new(big.Rat).SetString(s)
		back, _ := new(big.Rat).SetString(strconv.FormatFloat(f, 'g', -1, 64))
		if !ok || in.Cmp(back) != 0 {
			return "", fmt.Errorf("PARSE_JSON: input number: %s cannot round-trip through string representation", s)
		}
	}
	return out, nil
}

func formatJSONFloat(f float64) string {
	e := strconv.FormatFloat(f, 'e', -1, 64)
	exp, _ := strconv.Atoi(e[strings.LastIndexByte(e, 'e')+1:])
	if exp < -4 || exp >= 15 {
		return e
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

func writeJSONString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"' || r == '\\':
			buf.WriteByte('\\')
			buf.WriteRune(r)
		case r == '\n':
			buf.WriteString(`\n`)
		case r == '\r':
			buf.WriteString(`\r`)
		case r == '\t':
			buf.WriteString(`\t`)
		case r == '\b':
			buf.WriteString(`\b`)
		case r == '\f':
			buf.WriteString(`\f`)
		case r < 0x20:
			fmt.Fprintf(buf, `\u%04x`, r)
		default:
			buf.WriteRune(r)
		}
	}
	buf.WriteByte('"')
}

var BindParseJson = helper.Scalar2(func(a, b value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	mode, err := b.ToString()
	if err != nil {
		return nil, err
	}
	switch mode {
	case "exact", "round":
	default:
		return nil, fmt.Errorf("PARSE_JSON: unexpected wide_number_mode: %s", mode)
	}
	return PARSE_JSON(v, mode)
})
