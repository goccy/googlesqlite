package string

import (
	"bytes"
	"unicode"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func SOUNDEX(val string) (value.Value, error) {
	var (
		soundex      = [4]byte{' ', '0', '0', '0'}
		prevCode     byte
		soundexPoint int
	)
	runes := []rune(val)
	for i := range runes {
		r := runes[i]
		if !unicode.IsLetter(r) {
			continue
		}
		b := []byte(string(r))
		if len(b) != 1 {
			continue
		}
		c := bytes.ToUpper(b)[0]
		code := soundexMap[c]
		if soundexPoint == 0 {
			soundex[soundexPoint] = b[0]
			prevCode = code
			soundexPoint++
			continue
		}
		if code == prevCode || code == '0' {
			continue
		}
		soundex[soundexPoint] = code
		prevCode = code
		soundexPoint++
		if soundexPoint == 4 {
			break
		}
	}
	if soundexPoint == 0 {
		return value.StringValue(""), nil
	}
	return value.StringValue(string(soundex[:])), nil
}

var BindSoundex = helper.Scalar1(func(a value.Value) (value.Value, error) {
	v, err := a.ToString()
	if err != nil {
		return nil, err
	}
	return SOUNDEX(v)
})
