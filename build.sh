#!/bin/bash
set -e

APP_NAME="tit"

# Resolve GOBIN (fallback to $GOPATH/bin)
GOBIN=$(go env GOBIN)
if [ -z "$GOBIN" ]; then
  GOBIN="$(go env GOPATH)/bin"
fi

# Version from git tag
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")

echo "Installing $APP_NAME ($VERSION) to $GOBIN..."
go install -ldflags="-s -w -X github.com/jrengmusic/tit/internal.AppVersion=$VERSION" ./cmd/tit

echo "✓ Installed: $GOBIN/$APP_NAME"
