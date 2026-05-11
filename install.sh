#!/usr/bin/env sh
set -e

BINARY="gflow"
BUILD_DIR="bin"
INSTALL_DIR="/usr/local/bin"

cd "$(dirname "$0")"

echo "Building $BINARY..."
go build -ldflags "-s -w" -o "$BUILD_DIR/$BINARY" ./cmd/gflow

echo "Installing to $INSTALL_DIR/$BINARY..."
cp "$BUILD_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

echo "Done. $(gflow --version 2>/dev/null || echo "$BINARY installed")"
