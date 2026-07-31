#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# build.sh — produces in a single step everything in the dist/ folder that the
# Makefile (export + ci-upload) would scp up: the HTML/WASM zip and the
# metadata.json. It does NOT perform the upload itself.
#
#   ./build.sh            # version from metadata.json (1.0.0)
#   ./build.sh 1.2.3      # override version via argument
# -----------------------------------------------------------------------------
set -euo pipefail

cd "$(dirname "$0")"

PROJECT="rabbitroller"
DIST_DIR="dist"
WASM_NAME="game.wasm"
INDEX_HTML_URL="https://git.teletypegames.org/tools/ebitengine-tools/raw/branch/master/web/index.html"

# --- determine version (arg > metadata.json) --------------------------------
if [ "${1:-}" != "" ]; then
	VERSION="$1"
elif command -v jq >/dev/null 2>&1 && [ -f metadata.json ]; then
	VERSION="$(jq -r '.version' metadata.json)"
else
	echo "ERROR: no version (pass an argument, or install jq to read metadata.json)" >&2
	exit 1
fi
echo "==> Version: $VERSION"

ZIP_NAME="$PROJECT-$VERSION.html.zip"
META_DST="$PROJECT-$VERSION.metadata.json"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# --- WASM build -------------------------------------------------------------
echo "==> WASM build ($WASM_NAME)"
GOOS=js GOARCH=wasm go build -o "$DIST_DIR/$WASM_NAME" .

# --- wasm_exec.js from the Go distribution ----------------------------------
echo "==> copying wasm_exec.js"
GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
	cp "$GOROOT/lib/wasm/wasm_exec.js" "$DIST_DIR/wasm_exec.js"
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
	cp "$GOROOT/misc/wasm/wasm_exec.js" "$DIST_DIR/wasm_exec.js"
else
	echo "ERROR: wasm_exec.js not found under GOROOT ($GOROOT)" >&2
	exit 1
fi

# --- download index.html ----------------------------------------------------
echo "==> downloading index.html"
curl -sSL "$INDEX_HTML_URL" -o "$DIST_DIR/index.html"

# --- package (flat zip, like Makefile -j) -----------------------------------
echo "==> packaging: $DIST_DIR/$ZIP_NAME"
zip -j -r "$DIST_DIR/$ZIP_NAME" \
	"$DIST_DIR/$WASM_NAME" \
	"$DIST_DIR/wasm_exec.js" \
	"$DIST_DIR/index.html" >/dev/null

# --- metadata.json with the upload name -------------------------------------
echo "==> metadata.json -> $DIST_DIR/$META_DST"
cp metadata.json "$DIST_DIR/$META_DST"

# --- clean up intermediate files (the zip already contains them) ------------
rm -f "$DIST_DIR/$WASM_NAME" "$DIST_DIR/wasm_exec.js" "$DIST_DIR/index.html"

echo ""
echo "==> Done. Upload-ready artifacts in the $DIST_DIR/ folder:"
ls -la "$DIST_DIR"
