#!/bin/bash
# Sync suzume C++ source for CGO builds.
#
# Usage:
#   ./sync-upstream.sh              # Clone from GitHub
#   ./sync-upstream.sh --local      # Copy from ../suzume (dev only)

set -euo pipefail

DEST="csuzume"
REPO="https://github.com/libraz/suzume.git"

if [ "${1:-}" = "--local" ]; then
    LOCAL_SRC="${2:-../suzume}"
    if [ ! -d "$LOCAL_SRC/src" ]; then
        echo "error: $LOCAL_SRC/src not found" >&2
        exit 1
    fi
    echo "Syncing from $LOCAL_SRC ..."
    rm -rf "$DEST"
    mkdir -p "$DEST"
    cp -R "$LOCAL_SRC/src" "$DEST/src"
    cp -R "$LOCAL_SRC/data" "$DEST/data"
    cp "$LOCAL_SRC/CMakeLists.txt" "$DEST/"
    cp "$LOCAL_SRC/package.json" "$DEST/"
else
    echo "Cloning from $REPO ..."
    rm -rf "$DEST"
    git clone --depth 1 "$REPO" "$DEST"
    # Remove unnecessary files
    rm -rf "$DEST/.git" "$DEST/.github" "$DEST/js" "$DEST/tests" \
           "$DEST/benchmarks" "$DEST/docs" "$DEST/node_modules" \
           "$DEST/.gitignore" "$DEST/.clang-format" "$DEST/Makefile" \
           "$DEST/vitest*" "$DEST/tsconfig*" "$DEST/biome*" \
           "$DEST/binding.gyp" "$DEST/wasm" "$DEST/native"
fi

echo "Done. Run 'make lib' to build."
