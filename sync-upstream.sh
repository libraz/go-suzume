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
    cp -R "$LOCAL_SRC/include" "$DEST/include"
    cp -R "$LOCAL_SRC/data" "$DEST/data"
    cp "$LOCAL_SRC/CMakeLists.txt" "$DEST/"
else
    echo "Cloning from $REPO ..."
    rm -rf "$DEST"
    git clone --depth 1 "$REPO" "$DEST"
    # Keep only the C++ sources, public headers, dictionaries, and the top-level
    # CMake project; drop bindings, tests, tooling, and CI scaffolding.
    rm -rf "$DEST/.git" "$DEST/.github" "$DEST/bindings" "$DEST/tests" \
           "$DEST/benchmarks" "$DEST/docs" "$DEST/examples" "$DEST/scripts" \
           "$DEST/cmake" "$DEST/node_modules" "$DEST/.gitignore" \
           "$DEST/.clang-format" "$DEST/Makefile"
fi

echo "Done. Run 'make lib' to build."
