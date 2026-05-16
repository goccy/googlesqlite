package string

import (
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
	"golang.org/x/text/unicode/norm"
)

func NORMALIZE_AND_CASEFOLD(v, mode string) (value.Value, error) {
	v = strings.ToLower(v)
	switch mode {
	case "NFC":
		return value.StringValue(norm.NFC.String(v)), nil
	case "NFD":
		return value.StringValue(norm.NFD.String(v)), nil
	case "NFKC":
		return value.StringValue(norm.NFKC.String(v)), nil
	case "NFKD":
		return value.StringValue(norm.NFKD.String(v)), nil
	}
	return nil, fmt.Errorf("unexpected normalize mode %s", mode)
}

var BindNormalizeAndCasefold = helper.ScalarN(func(args ...value.Value) (value.Value, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, fmt.Errorf("NORMALIZE_AND_CASEFOLD: invalid number of arguments: got %d, want 1 or 2", len(args))
	}
	mode := "NFC"
	if len(args) == 2 {
		v, err := args[1].ToString()
		if err != nil {
			return nil, err
		}
		mode = v
	}
	v, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return NORMALIZE_AND_CASEFOLD(v, mode)
})
