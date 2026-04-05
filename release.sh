#!/bin/bash
# Usage: bash release.sh v0.0.1 "Optional commit message"
# Release notes come from RELEASE_NOTES.md (update before releasing)

TAG="${1:?Usage: release.sh <tag> [message]}"
MSG="${2:-$TAG}"

# Delete existing tag if present
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "Tag $TAG exists — removing local and remote"
    git tag -d "$TAG"
    git push origin ":refs/tags/$TAG" 2>/dev/null
fi

git add -A
git commit -m "$MSG"
git tag "$TAG"
git push origin main "$TAG"

GITHUB_TOKEN=$(gh auth token) goreleaser release --clean --release-notes=RELEASE_NOTES.md
