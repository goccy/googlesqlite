//go:build !js && windows

package main

import "io"

// interactiveInput chooses the terminal the REPL reads from. On
// Windows there is no /dev/tty equivalent wired up here, so a piped
// stdin means the script has run and the program stops; a real
// console stdin uses readline's default os.Stdin.
func interactiveInput(stdinPiped bool) (reader io.Reader, ok bool) {
	if stdinPiped {
		return nil, false
	}
	return nil, true
}
