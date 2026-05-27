#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# build.sh — egy lépésben előállítja a dist/ mappába mindazt, amit a Makefile
# (export + ci-upload) felscp-zne: a HTML/WASM zip-et és a metadata.json-t.
# Magát a feltöltést NEM végzi el.
#
#   ./build.sh            # verzió a metadata.json-ból (1.0.0)
#   ./build.sh 1.2.3      # verzió felülírása argumentummal
# -----------------------------------------------------------------------------
set -euo pipefail

cd "$(dirname "$0")"

PROJECT="rabbitroller"
DIST_DIR="dist"
WASM_NAME="game.wasm"
INDEX_HTML_URL="https://git.teletype.hu/tools/ebitengine-tools/raw/branch/master/web/index.html"

# --- verzió meghatározása (arg > metadata.json) -----------------------------
if [ "${1:-}" != "" ]; then
	VERSION="$1"
elif command -v jq >/dev/null 2>&1 && [ -f metadata.json ]; then
	VERSION="$(jq -r '.version' metadata.json)"
else
	echo "ERROR: nincs verzió (adj meg argumentumot, vagy telepíts jq-t a metadata.json olvasásához)" >&2
	exit 1
fi
echo "==> Verzió: $VERSION"

ZIP_NAME="$PROJECT-$VERSION.html.zip"
META_DST="$PROJECT-$VERSION.metadata.json"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# --- WASM build -------------------------------------------------------------
echo "==> WASM build ($WASM_NAME)"
GOOS=js GOARCH=wasm go build -o "$DIST_DIR/$WASM_NAME" .

# --- wasm_exec.js a Go disztribúcióból --------------------------------------
echo "==> wasm_exec.js másolása"
GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
	cp "$GOROOT/lib/wasm/wasm_exec.js" "$DIST_DIR/wasm_exec.js"
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
	cp "$GOROOT/misc/wasm/wasm_exec.js" "$DIST_DIR/wasm_exec.js"
else
	echo "ERROR: wasm_exec.js nem található a GOROOT alatt ($GOROOT)" >&2
	exit 1
fi

# --- index.html letöltése ---------------------------------------------------
echo "==> index.html letöltése"
curl -sSL "$INDEX_HTML_URL" -o "$DIST_DIR/index.html"

# --- csomagolás (lapos zip, mint a Makefile -j) -----------------------------
echo "==> Csomagolás: $DIST_DIR/$ZIP_NAME"
zip -j -r "$DIST_DIR/$ZIP_NAME" \
	"$DIST_DIR/$WASM_NAME" \
	"$DIST_DIR/wasm_exec.js" \
	"$DIST_DIR/index.html" >/dev/null

# --- metadata.json a feltöltési névvel --------------------------------------
echo "==> metadata.json -> $DIST_DIR/$META_DST"
cp metadata.json "$DIST_DIR/$META_DST"

# --- ideiglenes fájlok takarítása (a zip már tartalmazza őket) ---------------
rm -f "$DIST_DIR/$WASM_NAME" "$DIST_DIR/wasm_exec.js" "$DIST_DIR/index.html"

echo ""
echo "==> Kész. Feltöltésre kész artifactok a(z) $DIST_DIR/ mappában:"
ls -la "$DIST_DIR"
