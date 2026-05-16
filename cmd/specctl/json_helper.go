package main

import (
	"encoding/json"
	"io"
)

// jsonEncoder builds an indented JSON encoder so coverage --json
// output is human-readable in addition to machine-readable.
func jsonEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc
}
