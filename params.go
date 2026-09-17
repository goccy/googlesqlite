package googlesqlite

import (
	"fmt"
	"math/big"

	"github.com/goccy/googlesqlite/internal/value"
)

// Date binds a query parameter as GoogleSQL DATE.
//
// ValueFromGoValue (the reflection-based fallback every other bound
// value goes through) has no way to tell a DATE apart from a
// TIMESTAMP: both are naturally represented in Go as time.Time, so
// every time.Time parameter is treated as TIMESTAMP. Wrap the value
// in Date to bind it as DATE instead; only the year/month/day
// components are significant, any time-of-day component is discarded.
//
//	db.QueryContext(ctx, "SELECT @d", sql.Named("d", googlesqlite.Date(t)))
type Date = value.DateValue

// Numeric binds a query parameter as GoogleSQL NUMERIC (or BIGNUMERIC
// when IsBigNumeric is set).
//
// ValueFromGoValue has no way to tell a NUMERIC apart from a STRING:
// both are naturally represented in Go as a decimal-looking string,
// so every string parameter is treated as STRING. Construct a Numeric
// with NewNumeric (or NewBigNumeric) to bind it as NUMERIC /
// BIGNUMERIC instead.
//
//	n, err := googlesqlite.NewNumeric("123.456")
//	db.QueryContext(ctx, "SELECT @n", sql.Named("n", n))
//
// Methods are defined on *Numeric; always pass a pointer (NewNumeric /
// NewBigNumeric already return one).
type Numeric = value.NumericValue

// NewNumeric parses s (a base-10 decimal string, e.g. "123.456") into
// a Numeric query parameter.
func NewNumeric(s string) (*Numeric, error) {
	return newNumeric(s, false)
}

// NewBigNumeric parses s (a base-10 decimal string) into a BIGNUMERIC
// query parameter.
func NewBigNumeric(s string) (*Numeric, error) {
	return newNumeric(s, true)
}

func newNumeric(s string, isBig bool) (*Numeric, error) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		kind := "NUMERIC"
		if isBig {
			kind = "BIGNUMERIC"
		}
		return nil, fmt.Errorf("googlesqlite: invalid %s decimal string %q", kind, s)
	}
	return &Numeric{Rat: r, IsBigNumeric: isBig}, nil
}
