package string

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/goccy/googlesqlite/internal/functions/helper"
	"github.com/goccy/googlesqlite/internal/value"
)

func LIKE(a, b value.Value) (value.Value, error) {
	va, err := a.ToString()
	if err != nil {
		return nil, err
	}
	vb, err := b.ToString()
	if err != nil {
		return nil, err
	}
	wildcard := strings.ReplaceAll(regexp.QuoteMeta(vb), "%", ".*")
	matchLimits := fmt.Sprintf("^%s$", wildcard)
	re, err := regexp.Compile(matchLimits)
	if err != nil {
		return nil, err
	}
	return value.BoolValue(re.MatchString(va)), nil
}

// BindLike: per GoogleSQL three-valued logic, LIKE with a NULL operand
// returns NULL, not FALSE — Scalar2 propagates that.
var BindLike = helper.Scalar2(LIKE)
