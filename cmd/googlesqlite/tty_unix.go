//go:build !js && !windows

package main

import (
	"io"
	"os"
)

// interactiveInput chooses the terminal the REPL reads from.
//
// When stdin was not piped it is already a terminal, so a nil reader
// is returned to let readline use os.Stdin directly. When stdin was
// piped (and has now been consumed as a script), the controlling
// terminal /dev/tty is reopened so the REPL can still run; if there
// is no controlling terminal (CI, cron) ok is false and the caller
// stops after the script.
func interactiveInput(stdinPiped bool) (reader io.Reader, ok bool) {
	if !stdinPiped {
		return nil, true
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, false
	}
	return tty, true
}
