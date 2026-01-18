#!/bin/bash

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH_SUFFIX="x64" ;;
  arm64)   ARCH_SUFFIX="arm64" ;;
  *)       ARCH_SUFFIX="$ARCH" ;;
esac

BINARY_NAME="tit_${ARCH_SUFFIX}"
DEST_DIR="$HOME/.local/bin/tit"

# Build
echo "Building $BINARY_NAME..."
go build -o "$BINARY_NAME" ./cmd/tit || exit 1

# Post-build: copy to automation
mkdir -p "$DEST_DIR"
cp "$BINARY_NAME" "$DEST_DIR/"
chmod +x "$DEST_DIR/$BINARY_NAME"

echo "✓ Built: $BINARY_NAME"
echo "✓ Copied: $DEST_DIR/$BINARY_NAME"
