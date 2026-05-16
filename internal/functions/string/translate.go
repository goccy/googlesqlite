package string

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func TRANSLATE(expr, source, target value.Value) (value.Value, error) {
	switch expr.(type) {
	case value.StringValue:
		if _, ok := source.(value.StringValue); !ok {
			return nil, fmt.Errorf("TRANSLATE: source characters must be STRING type")
		}
		if _, ok := target.(value.StringValue); !ok {
			return nil, fmt.Errorf("TRANSLATE: target characters must be STRING type")
		}
		e, err := expr.ToString()
		if err != nil {
			return nil, err
		}
		s, err := source.ToString()
		if err != nil {
			return nil, err
		}
		t, err := target.ToString()
		if err != nil {
			return nil, err
		}
		// Iterate by Unicode code point (rune) — STRING semantics. A
		// byte-level walk breaks multi-byte source characters (CJK,
		// emoji, surrogate pairs); the only thing that survives is
		// the ASCII subset.
		srcRunes := []rune(s)
		tgtRunes := []rune(t)
		mapping := make(map[rune]rune, len(srcRunes))
		delete := make(map[rune]struct{})
		for i, sr := range srcRunes {
			if _, exists := mapping[sr]; exists {
				return nil, fmt.Errorf("TRANSLATE: found duplicated source character: %c", sr)
			}
			if _, exists := delete[sr]; exists {
				return nil, fmt.Errorf("TRANSLATE: found duplicated source character: %c", sr)
			}
			if i < len(tgtRunes) {
				mapping[sr] = tgtRunes[i]
			} else {
				delete[sr] = struct{}{}
			}
		}
		var b strings.Builder
		b.Grow(len(e))
		for _, r := range e {
			if _, drop := delete[r]; drop {
				continue
			}
			if v, ok := mapping[r]; ok {
				b.WriteRune(v)
				continue
			}
			b.WriteRune(r)
		}
		return value.StringValue(b.String()), nil
	case value.BytesValue:
		if _, ok := source.(value.BytesValue); !ok {
			return nil, fmt.Errorf("TRANSLATE: source characters must be BYTES type")
		}
		if _, ok := target.(value.BytesValue); !ok {
			return nil, fmt.Errorf("TRANSLATE: target characters must be BYTES type")
		}
		e, err := expr.ToBytes()
		if err != nil {
			return nil, err
		}
		s, err := source.ToBytes()
		if err != nil {
			return nil, err
		}
		t, err := target.ToBytes()
		if err != nil {
			return nil, err
		}
		evaluatedByte := map[byte]struct{}{}
		for i := range s {
			if _, exists := evaluatedByte[s[i]]; exists {
				return nil, fmt.Errorf("TRANSLATE: found duplicated source character: %c", s[i])
			}
			if len(t) > i {
				e = bytes.ReplaceAll(e, []byte{s[i]}, []byte{t[i]})
			} else {
				e = bytes.ReplaceAll(e, []byte{s[i]}, []byte{})
			}
			evaluatedByte[s[i]] = struct{}{}
		}
		return value.BytesValue(e), nil
	}
	return nil, fmt.Errorf("TRANSLATE: expression type is must be STRING or BYTES type")
}

var BindTranslate = helper.Scalar3(TRANSLATE)
