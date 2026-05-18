# Makefile for the googlesqlite Playground.
#
# Typical first run:
#   make deps           # install npm dependencies
#   make dev            # build the wasm, then start the Vite dev server
#
# `make build` produces the static site under dist/.
# `make build/release` produces the wasm-opt-optimised site that the
# release workflow deploys to GitHub Pages.

# Repository root, relative to this directory.
ROOT   := ../..
PUBLIC := public
WASM   := $(PUBLIC)/googlesqlite.wasm

# Binaryen provides wasm-opt. It is a C++ tool, so it cannot live in
# tools/go.mod alongside the Go tools; instead a pinned release is
# downloaded on demand by the $(WASMOPT) rule below. The version is
# part of the install directory name, so bumping BINARYEN_VERSION
# triggers a fresh download while an existing install is reused as-is.
# This is the single mechanism shared by local builds and CI — there is
# no separate install step.
BINARYEN_VERSION := version_129

# Map the host to Binaryen's release-asset naming. uname -m reports
# x86_64/amd64 or arm64/aarch64; Binaryen labels its arm assets arm64
# on macOS but aarch64 on Linux.
binaryen_os   := $(if $(filter Darwin,$(shell uname -s)),macos,linux)
binaryen_arch := $(if $(filter x86_64 amd64,$(shell uname -m)),x86_64,$(if $(filter macos,$(binaryen_os)),arm64,aarch64))

BINARYEN_DIR     := $(ROOT)/tools/binaryen-$(BINARYEN_VERSION)
WASMOPT          := $(BINARYEN_DIR)/bin/wasm-opt
binaryen_tarball := binaryen-$(BINARYEN_VERSION)-$(binaryen_arch)-$(binaryen_os).tar.gz
binaryen_baseurl := https://github.com/WebAssembly/binaryen/releases/download/$(BINARYEN_VERSION)

.PHONY: build build/release build/engine build/wasm build/assets build/page \
        deps dev preview optimize clean

build: build/wasm build/assets build/page

# build/engine produces just the optimised engine binary
# (public/googlesqlite.wasm): compile, then shrink it with wasm-opt. It
# carries no compressed copy, build id or bundled site. The release
# workflow builds the engine and the site in separate jobs, so the two
# halves are kept separable here too.
build/engine:
	$(MAKE) build/wasm
	$(MAKE) optimize

# build/release produces the full optimised static site for a local
# deployment build: the optimised engine, then the derived assets and
# the bundled site. The sub-makes run in sequence so the build id and
# the .gz always describe the exact binary that ships.
build/release:
	$(MAKE) build/engine
	$(MAKE) build/assets
	$(MAKE) build/page

# build/wasm compiles the js/wasm Playground entrypoint
# (cmd/googlesqlite, the `js && wasm` build) to public/googlesqlite.wasm
# and copies the matching Go wasm_exec.js loader next to it. Everything
# under public/ is a generated artifact and is gitignored.
build/wasm:
	mkdir -p $(PUBLIC)
	cd $(ROOT) && GOOS=js GOARCH=wasm go build \
	  -o docs/playground/$(WASM) ./cmd/googlesqlite
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(PUBLIC)/wasm_exec.js

# build/assets derives the artifacts that depend on the final binary:
# the packed timezone database, the content-hash build id, and the
# pre-compressed .gz copy the server delivers (mirroring a deploy that
# ships the compressed artifact). Run it after build/wasm — and after
# optimize, when optimising — so the build id and the .gz describe the
# binary that actually ships.
build/assets:
	node scripts/pack-zoneinfo.mjs
	node scripts/build-id.mjs
	gzip -kf -6 $(WASM)
	@ls -lh $(WASM) $(WASM).gz | awk '{print "build/assets: " $$9 " (" $$5 ")"}'

# build/page bundles the static site into dist/. The deployed site
# fetches the gzip-compressed engine and decompresses it in the worker,
# so the uncompressed copy Vite mirrors from public/ is dead weight on
# the host (and large enough to trip GitHub Pages' per-file limit);
# drop it to keep the published site small.
build/page:
	npm run build
	rm -f dist/googlesqlite.wasm

# deps installs the npm dependency tree.
deps:
	npm install

# dev builds the wasm and starts the Vite dev server. The page is
# served at http://localhost:5173/googlesqlite/.
dev: build/wasm build/assets
	npm run dev

# preview serves the production build from dist/.
preview:
	npm run preview

# $(WASMOPT) downloads and checksum-verifies the pinned Binaryen
# release. As a file target it runs only when wasm-opt is not already
# installed; the version-stamped directory name makes a BINARYEN_VERSION
# bump re-download. The tarball unpacks to binaryen-$(BINARYEN_VERSION)/,
# which is $(BINARYEN_DIR).
$(WASMOPT):
	mkdir -p $(ROOT)/tools
	cd $(ROOT)/tools && \
	  curl -fsSL "$(binaryen_baseurl)/$(binaryen_tarball)" -o "$(binaryen_tarball)" && \
	  curl -fsSL "$(binaryen_baseurl)/$(binaryen_tarball).sha256" -o "$(binaryen_tarball).sha256" && \
	  shasum -a 256 -c "$(binaryen_tarball).sha256" && \
	  tar -xzf "$(binaryen_tarball)" && \
	  rm -f "$(binaryen_tarball)" "$(binaryen_tarball).sha256"

# optimize runs wasm-opt over the built binary to shrink the code
# section and speed up startup. wasm-opt comes from a pinned Binaryen
# release (the $(WASMOPT) prerequisite), so no manual install is
# needed. Note: most of the binary is embedded wasm module data (the
# SQLite and GoogleSQL runtimes), which wasm-opt cannot compress —
# expect a modest reduction.
optimize: $(WASMOPT)
	$(WASMOPT) -Oz \
	  --enable-bulk-memory \
	  --enable-nontrapping-float-to-int \
	  --enable-sign-ext \
	  --enable-mutable-globals \
	  $(WASM) -o $(WASM)
	@ls -lh $(WASM) | awk '{print "optimize: " $$9 " (" $$5 ")"}'

clean:
	rm -rf dist node_modules $(WASM) $(WASM).gz $(PUBLIC)/wasm_exec.js \
	  $(PUBLIC)/zoneinfo.json $(PUBLIC)/build-id.json
