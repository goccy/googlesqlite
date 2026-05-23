package internal

import (
	"strings"
)

type ErrorGroup struct {
	errs []error
}

func (eg *ErrorGroup) HasError() bool {
	return len(eg.errs) != 0
}

func (eg *ErrorGroup) Add(e error) {
	if e == nil {
		return
	}
	eg.errs = append(eg.errs, e)
}

func (eg *ErrorGroup) Error() string {
	errs := []string{}
	for _, err := range eg.errs {
		errs = append(errs, err.Error())
	}
	if len(errs) != 0 {
		return strings.Join(errs, ",")
	}
	return ""
}

// Unwrap exposes the wrapped errors so errors.Is / errors.As can
// traverse them. Without this an ErrorGroup that happens to contain a
// sentinel like sql.ErrConnDone or driver.ErrBadConn is opaque to
// callers that use errors.Is for control flow — notably the driver's
// dead-connection translation in driver.go.
func (eg *ErrorGroup) Unwrap() []error {
	return eg.errs
}
