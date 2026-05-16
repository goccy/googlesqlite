package string

import (
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func INITCAP(val string, delimiters []rune) (value.Value, error) {
	if delimiters == nil {
		delimiters = defaultInitcapDelimiters
	}
	src := []rune(val)
	dst := make([]rune, 0, len(src))
	for i := 0; i < len(src); i++ {
		r := src[i]
		isCurDelim := isDelim(r, delimiters)
		switch {
		case i == 0:
			// first character is upper case.
			dst = append(dst, []rune(strings.ToUpper(string([]rune{r})))...)
		case isCurDelim:
			// if current character is delimiter, add it as is.
			dst = append(dst, r)
		default:
			// if other characters, add it as lower case.
			dst = append(dst, []rune(strings.ToLower(string([]rune{r})))...)
		}
		// break if current character is last
		if i+1 == len(src) {
			continue
		}
		// if next character is delimiter, skip current character.
		if isDelim(src[i+1], delimiters) {
			continue
		}
		if isCurDelim {
			// if current character is delimiter, add next character as upper case character and skip next character.
			dst = append(dst, []rune(strings.ToUpper(string([]rune{src[i+1]})))...)
			i++
		}
	}
	return value.StringValue(string(dst)), nil
}

var BindInitcap = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("INITCAP: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	var delimiters []rune
	if len(args) == 2 {
		v, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		delimiters = []rune{}
		for _, vv := range v {
			delimiters = append(delimiters, vv)
		}
	}
	value, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return INITCAP(value, delimiters)
})
