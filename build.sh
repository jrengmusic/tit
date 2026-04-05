#!/bin/bash
set -e

# Detect architecture ($MSYSTEM is authoritative on MSYS2, uname -m lies)
case "$MSYSTEM" in
  CLANGARM64) ARCH_SUFFIX="arm64" ;;
  MINGW64|UCRT64|MSYS) ARCH_SUFFIX="x64" ;;
  *)
    ARCH=$(uname -m)
    case "$ARCH" in
      x86_64) ARCH_SUFFIX="x64" ;;
      arm64|aarch64) ARCH_SUFFIX="arm64" ;;
      *) ARCH_SUFFIX="$ARCH" ;;
    esac
    ;;
esac

APP_NAME="tit"
BINARY_NAME="${APP_NAME}_${ARCH_SUFFIX}"

INSTALL_ROOT="$HOME/.${APP_NAME}"
BIN_DIR="$INSTALL_ROOT/bin"
SYMLINK_DIR="$HOME/.local/bin"
SYMLINK_PATH="$SYMLINK_DIR/$APP_NAME"

# Version from git tag
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")

# Build
echo "Building $BINARY_NAME ($VERSION)..."
go build -ldflags="-s -w -X github.com/jrengmusic/tit/internal.AppVersion=$VERSION" -o "$BINARY_NAME" ./cmd/tit

# Install
mkdir -p "$BIN_DIR" "$SYMLINK_DIR"
mv "$BINARY_NAME" "$BIN_DIR/"
chmod +x "$BIN_DIR/$BINARY_NAME"

# Symlink (atomic replace)
ln -sfn "$BIN_DIR/$BINARY_NAME" "$SYMLINK_PATH"

echo "✓ Installed: $BIN_DIR/$BINARY_NAME"
echo "✓ Symlinked: $SYMLINK_PATH -> $BIN_DIR/$BINARY_NAME"

