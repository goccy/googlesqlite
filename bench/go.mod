module github.com/goccy/googlesqlite/bench

go 1.25.0

// The bench module benchmarks the googlesqlite driver to detect
// performance regressions across revisions.

require github.com/goccy/googlesqlite v0.0.0

require (
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/DataDog/go-hll v1.0.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/dop251/goja v0.0.0-20260311135729-065cd970411c // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/goccy/go-googlesql v0.3.0 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/goccy/googlesqlwasm2go v0.1.0 // indirect
	github.com/golang/geo v0.0.0-20260505155700-1c5af9662e82 // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-sqlite3 v0.34.0 // indirect
	github.com/ncruces/go-sqlite3-wasm/v2 v2.1.35300 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/pkg/errors v0.8.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/genproto v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	modernc.org/libc v1.72.3 // indirect
	modernc.org/libquickjs v0.12.8 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/quickjs v0.18.2 // indirect
)

// googlesqlite is the module under development; benchmark the local
// checkout rather than a published version.
replace github.com/goccy/googlesqlite => ../
