package internal

import (
	"errors"
	"strings"

	sqlite3 "github.com/ncruces/go-sqlite3"
)

// unwrapSQLiteUserError peels the ncruces "sqlite3: <code>: " prefix
// off errors that originate from a custom function's ResultError. The
// predecessor surface returned the user-set message verbatim; many
// parity tests compare on exact strings, so we restore that contract.
//
// Only errors that are *sqlite3.Error with code SQLITE_ERROR (the
// generic "SQL logic error" code) are stripped. OS / IO / constraint
// errors that legitimately carry a category prefix are passed through
// unchanged.
func unwrapSQLiteUserError(err error) error {
	if err == nil {
		return nil
	}
	var serr *sqlite3.Error
	if !errors.As(err, &serr) {
		return err
	}
	if serr.Code() != sqlite3.ERROR {
		return err
	}
	full := serr.Error()
	prefix := sqlite3.ERROR.Error() + ": "
	if !strings.HasPrefix(full, prefix) {
		return err
	}
	return errors.New(strings.TrimPrefix(full, prefix))
}
