#!/bin/bash
set -e

# Called by goreleaser after each build. Signs and notarizes macOS binaries.
# Env vars set by goreleaser: HOOK_TARGET (e.g. darwin_arm64), HOOK_PATH (binary path)

[[ "$(uname)" == "Darwin" ]] || exit 0

case "$HOOK_TARGET" in
  darwin_*)
    echo "Signing $HOOK_PATH..."
    codesign --force --options runtime \
      --entitlements ./entitlements.plist \
      --sign "Developer ID Application: Bayu Ardianto (9BDSN9TDX3)" \
      "$HOOK_PATH"
    codesign --verify --verbose "$HOOK_PATH"

    echo "Notarizing $HOOK_PATH..."
    ZIPDIR=$(mktemp -d)
    ZIP="$ZIPDIR/$(basename "$HOOK_PATH").zip"
    ditto -c -k --keepParent "$HOOK_PATH" "$ZIP"
    xcrun notarytool submit "$ZIP" \
      --keychain-profile notary \
      --wait
    rm -rf "$ZIPDIR"
    ;;
esac
