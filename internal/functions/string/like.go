package string

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LIKE(a, b value.Value) (value.Value, error) {
	var va, vb string
	if ba, ok := a.(value.BytesValue); ok {
		bb, err := b.ToBytes()
		if err != nil {
			return nil, err
		}
		va, vb = latin1(ba), latin1(bb)
	} else {
		var err error
		if va, err = a.ToString(); err != nil {
			return nil, err
		}
		if vb, err = b.ToString(); err != nil {
			return nil, err
		}
	}
	re, err := likePatternToRegexp(vb)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(re.MatchString(va)), nil
}

// latin1 maps each byte to one rune so BYTES operands match byte-wise.
func latin1(bs []byte) string {
	rs := make([]rune, len(bs))
	for i, c := range bs {
		rs[i] = rune(c)
	}
	return string(rs)
}

func likePatternToRegexp(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString("(?s)^")
	for i := 0; i < len(pattern); {
		r, n := utf8.DecodeRuneInString(pattern[i:])
		switch r {
		case '%':
			b.WriteString(".*")
		case '_':
			b.WriteString(".")
		case '\\':
			if i+n == len(pattern) {
				return nil, fmt.Errorf("LIKE pattern ends with a backslash")
			}
			i += n
			r, n = utf8.DecodeRuneInString(pattern[i:])
			b.WriteString(regexp.QuoteMeta(string(r)))
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
		i += n
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}

// BindLike: per GoogleSQL three-valued logic, LIKE with a NULL operand
// returns NULL, not FALSE — Scalar2 propagates that.
var BindLike = helper.Scalar2(LIKE)
