//go:build js && wasm

package main

// On js/wasm the host filesystem is asynchronous, so a synchronous
// call into the engine (the Playground's exec) cannot block on
// time.LoadLocation reading /usr/share/zoneinfo — doing so deadlocks
// the Go scheduler. Embedding the IANA timezone database makes
// time.LoadLocation resolve zones from memory instead, with no
// filesystem access. This blank import is js/wasm-only; native builds
// keep using the operating system's zoneinfo.
import _ "time/tzdata"
