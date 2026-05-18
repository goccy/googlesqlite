PKG  := github.com/goccy/googlesqlite
BIN  := $(CURDIR)/bin

# Project tools run through tools/go.mod (Go tools) or a pinned-install
# script (non-Go tools), so a contributor and a CI runner use byte-for-
# byte the same versions. Never invoke these through an ambient PATH.
LINT       := go tool -modfile=tools/go.mod golangci-lint
GORELEASER := go tool -modfile=tools/go.mod goreleaser

.PHONY: all build build-wasm test lint lint/fix validate spec/check spec/check/ci spec/coverage spec/upstream-sync bench release release/check identity-check clean

all: build

$(BIN):
	@mkdir -p $(BIN)

build: | $(BIN)
	go build -o $(BIN)/googlesqlite ./cmd/googlesqlite
	go build -o $(BIN)/specctl ./cmd/specctl

# build-wasm compiles the Playground entrypoint (cmd/googlesqlite, the
# js && wasm build) to WebAssembly and copies the matching Go
# wasm_exec.js loader next to it. The two files under $(BIN) are what
# the gh-pages Playground page loads to run GoogleSQL in the browser.
build-wasm: | $(BIN)
	GOOS=js GOARCH=wasm go build -o $(BIN)/googlesqlite.wasm ./cmd/googlesqlite
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(BIN)/wasm_exec.js
	@echo "build-wasm: wrote $(BIN)/googlesqlite.wasm and $(BIN)/wasm_exec.js"

test:
	go test -race -timeout 30m ./...

lint:
	$(LINT) run --timeout 30m

lint/fix:
	$(LINT) run --fix --timeout 30m

# validate runs the non-test repository invariants: predecessor-brand
# check and spec-index consistency.
validate: identity-check spec/check/ci

spec/check:
	go run ./cmd/specctl check

spec/check/ci:
	go run ./cmd/specctl check --check

spec/coverage:
	go run ./cmd/specctl coverage

spec/upstream-sync:
	go run ./cmd/specctl upstream-sync

BENCH_OUT := bench/bench.out
# bench/ has its own go.mod; it benchmarks the googlesqlite driver to
# detect performance regressions across revisions.
bench:
	cd bench && go test -bench=. -benchmem -benchtime=200ms -run=^$$ \
	  -timeout 1800s . | tee bench.out
	cd bench && go run -tags=render_results render_results.go bench.out > RESULTS.md
	@echo "bench: results written to bench/RESULTS.md"

# GoReleaser builds each target concurrently; its default is one build
# per CPU. cmd/googlesqlite links the embedded GoogleSQL analyzer wasm
# module into every binary, so a single build is already heavy —
# running one per CPU at once exhausts a CI runner. Serialise the
# builds instead; the release is infrequent, so the extra wall-clock
# time is fine.
GORELEASER_PARALLELISM := 1

# release runs GoReleaser to build and publish the cross-platform CLI
# release. The js/wasm Playground engine must already exist at
# docs/playground/public/googlesqlite.wasm — GoReleaser attaches it to
# the release. The release workflow builds it in a dedicated `wasm`
# job and hands it to the release job as an artifact; to publish from
# a local checkout, run `make -C docs/playground build/wasm` first.
# GITHUB_TOKEN must be set in the environment.
release:
	$(GORELEASER) release --clean --parallelism $(GORELEASER_PARALLELISM)

# release/check validates the GoReleaser configuration and runs a
# publish-free dry-run build. The release-test workflow runs this; it
# does not build the Playground engine (release.extra_files is only
# resolved by the publishing step, which the dry-run skips).
release/check:
	$(GORELEASER) check
	$(GORELEASER) release --snapshot --skip=publish --clean \
	  --parallelism $(GORELEASER_PARALLELISM)

# identity-check fails the build if any predecessor branding leaked in.
# Exempt directory:
#   docs/third_party/  — upstream-licensed snapshot
identity-check:
	@if git grep -nE 'zetasql|zetasqlite|ZetaSQL|ZetaSQLite' -- \
	    ':!docs/third_party' ':!Makefile' ':!.claude/plans'; then \
		echo "identity-check: predecessor branding found in tracked files" >&2; \
		exit 1; \
	fi

clean:
	rm -rf $(BIN)
