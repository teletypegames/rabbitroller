# -----------------------------------------
# Makefile – Ebitengine project builder
# -----------------------------------------

PROJECT = rabbitroller

# Our own modules live on the forge, not on proxy.golang.org: fetch them
# directly and skip the public checksum database.
export GOPRIVATE = git.teletypegames.org

BIN_DIR     = bin
DIST_DIR    = dist
WASM_NAME   = game.wasm
OUTPUT_BIN  = $(BIN_DIR)/$(PROJECT)
OUTPUT_WASM = $(DIST_DIR)/$(WASM_NAME)
OUTPUT_JS   = $(DIST_DIR)/wasm_exec.js
OUTPUT_ZIP  = $(PROJECT)-$(VERSION).html.zip

INDEX_HTML_URL = https://git.teletypegames.org/build/ebitengine-tools/raw/branch/master/web/index.html

# mac-release: the catalog and the WarpEngine CLI; WARP_TOKEN is an
# application token with the update + upload scopes
WARP_SERVER ?= https://teletypegames.org
WARP_ENGINE ?= warp-engine
export WARP_SERVER WARP_TOKEN

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(OUTPUT_BIN) .

run: build
	./$(OUTPUT_BIN)

check:
	go vet ./...

web:
	@mkdir -p $(DIST_DIR)
	GOOS=js GOARCH=wasm go build -o $(OUTPUT_WASM) .
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(OUTPUT_JS)
	curl -sSL $(INDEX_HTML_URL) -o $(DIST_DIR)/index.html

export: web
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION not set!"; exit 1; \
	fi
	@echo "==> Packaging HTML/WASM for $(VERSION)"
	zip -r $(OUTPUT_ZIP) -j $(DIST_DIR)/$(WASM_NAME) $(DIST_DIR)/wasm_exec.js $(DIST_DIR)/index.html
	@echo "==> Cleaning temporary files"
	rm -f $(DIST_DIR)/$(WASM_NAME) $(DIST_DIR)/wasm_exec.js $(DIST_DIR)/index.html

# --- native binaries (lásd: teletypegames/NOTES_BINARY_PLAN.md) -------------
# win-x86 / win-x64: pure Go cross-compile (Windowson nem kell cgo)
# linux-x64: cgo build, linux/amd64 hoston fut (builder image, X11/GL dev libekkel)
# a CI (linux builder) ezt a hármat építi:
BINARY_TARGETS = win-x86 win-x64 linux-x64
# mac-x64 / mac-arm64: az Ebitengine darwinon cgo-t igényel (GLFW/Metal),
# ezért ezek CSAK mac hoston építhetők: `make binaries-mac VERSION=x.y.z`,
# upload + publish: `WARP_TOKEN=<application token: update + upload> make mac-release VERSION=x.y.z`
MAC_TARGETS = mac-x64 mac-arm64

binaries:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION not set!"; exit 1; \
	fi
	@for target in $(BINARY_TARGETS); do \
		$(MAKE) binary-$$target VERSION=$(VERSION) || exit 1; \
	done

binaries-mac:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION not set!"; exit 1; \
	fi
	@if [ "$$(uname -s)" != "Darwin" ]; then \
		echo "ERROR: a mac targetek csak mac hoston építhetők (cgo + SDK kell)"; exit 1; \
	fi
	@for target in $(MAC_TARGETS); do \
		$(MAKE) binary-$$target VERSION=$(VERSION) || exit 1; \
	done

binary-win-x86:
	@$(MAKE) binary-build VERSION=$(VERSION) B_GOOS=windows B_GOARCH=386 B_CGO=0 B_EXT=.exe B_TARGET=win-x86

binary-win-x64:
	@$(MAKE) binary-build VERSION=$(VERSION) B_GOOS=windows B_GOARCH=amd64 B_CGO=0 B_EXT=.exe B_TARGET=win-x64

binary-linux-x64:
	@$(MAKE) binary-build VERSION=$(VERSION) B_GOOS=linux B_GOARCH=amd64 B_CGO=1 B_EXT= B_TARGET=linux-x64

binary-mac-x64:
	@$(MAKE) binary-build-mac VERSION=$(VERSION) B_GOARCH=amd64 B_TARGET=mac-x64

binary-mac-arm64:
	@$(MAKE) binary-build-mac VERSION=$(VERSION) B_GOARCH=arm64 B_TARGET=mac-arm64

warp-engine-cli:
	@command -v $(WARP_ENGINE) >/dev/null 2>&1 || { echo "ERROR: $(WARP_ENGINE) (the WarpEngine CLI) is not installed: curl -fsSL https://git.teletypegames.org/engines/warp-engine-cli/raw/branch/master/install.sh | sh"; exit 1; }

# run on a Mac: builds the mac zips, uploads them into the release the CI
# published and publishes it again (warp-engine, the WarpEngine CLI)
mac-release: warp-engine-cli binaries-mac
	$(WARP_ENGINE) release upload --name $(PROJECT) --version $(VERSION) $(foreach target,$(MAC_TARGETS),$(PROJECT)-$(VERSION)-$(target).zip)
	$(WARP_ENGINE) release publish --name $(PROJECT) --version $(VERSION) --platform ebitengine

# belső helper: egy target buildje + zip egyetlen gyökérmappával
# (a zipet unixos zip készíti, így a végrehajtási bit megmarad)
binary-build:
	@set -e; \
	PKG_DIR="$(PROJECT)-$(VERSION)-$(B_TARGET)"; \
	echo "==> Building $$PKG_DIR"; \
	rm -rf "$$PKG_DIR" "$$PKG_DIR.zip"; \
	mkdir -p "$$PKG_DIR"; \
	CGO_ENABLED=$(B_CGO) GOOS=$(B_GOOS) GOARCH=$(B_GOARCH) go build -o "$$PKG_DIR/$(PROJECT)$(B_EXT)" .; \
	if [ -f LICENSE ]; then cp LICENSE "$$PKG_DIR/"; fi; \
	if [ -f README.md ]; then cp README.md "$$PKG_DIR/"; fi; \
	zip -r "$$PKG_DIR.zip" "$$PKG_DIR" >/dev/null; \
	rm -rf "$$PKG_DIR"; \
	echo "==> $$PKG_DIR.zip kesz"

# mac helper: .app bundle Info.plist-tel, ad-hoc aláírással, egy gyökérmappában
binary-build-mac:
	@set -e; \
	PKG_DIR="$(PROJECT)-$(VERSION)-$(B_TARGET)"; \
	APP_DIR="$$PKG_DIR/$(PROJECT).app"; \
	echo "==> Building $$PKG_DIR"; \
	rm -rf "$$PKG_DIR" "$$PKG_DIR.zip"; \
	mkdir -p "$$APP_DIR/Contents/MacOS"; \
	CGO_ENABLED=1 GOOS=darwin GOARCH=$(B_GOARCH) go build -o "$$APP_DIR/Contents/MacOS/$(PROJECT)" .; \
	{ \
		echo '<?xml version="1.0" encoding="UTF-8"?>'; \
		echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">'; \
		echo '<plist version="1.0">'; \
		echo '<dict>'; \
		echo '  <key>CFBundleExecutable</key><string>$(PROJECT)</string>'; \
		echo '  <key>CFBundleIdentifier</key><string>org.teletypegames.$(PROJECT)</string>'; \
		echo '  <key>CFBundleName</key><string>$(PROJECT)</string>'; \
		echo '  <key>CFBundlePackageType</key><string>APPL</string>'; \
		echo '  <key>CFBundleShortVersionString</key><string>$(VERSION)</string>'; \
		echo '  <key>CFBundleVersion</key><string>$(VERSION)</string>'; \
		echo '  <key>LSMinimumSystemVersion</key><string>11.0</string>'; \
		echo '  <key>NSHighResolutionCapable</key><true/>'; \
		echo '</dict>'; \
		echo '</plist>'; \
	} > "$$APP_DIR/Contents/Info.plist"; \
	codesign --force --deep -s - "$$APP_DIR"; \
	if [ -f LICENSE ]; then cp LICENSE "$$PKG_DIR/"; fi; \
	if [ -f README.md ]; then cp README.md "$$PKG_DIR/"; fi; \
	zip -r "$$PKG_DIR.zip" "$$PKG_DIR" >/dev/null; \
	rm -rf "$$PKG_DIR"; \
	echo "==> $$PKG_DIR.zip kesz"

watch:
	fswatch -o . | while read; do make build; done

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)

.PHONY: all build run check web export watch clean binaries binaries-mac warp-engine-cli mac-release binary-win-x86 binary-win-x64 binary-linux-x64 binary-mac-x64 binary-mac-arm64 binary-build binary-build-mac
