PKG  := github.com/goccy/googlesqlite
BIN  := $(CURDIR)/bin
LINT := go tool -modfile=tools/go.mod golangci-lint

.PHONY: all build test lint lint/fix validate spec/check spec/check/ci spec/coverage spec/upstream-sync bench identity-check clean

all: build

$(BIN):
	@mkdir -p $(BIN)

build: | $(BIN)
	go build -o $(BIN)/googlesqlite ./cmd/googlesqlite
	go build -o $(BIN)/specctl ./cmd/specctl

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
