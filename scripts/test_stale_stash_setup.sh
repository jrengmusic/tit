#!/bin/bash
# Automated setup for stale stash test - runs everything up to the point
# where user needs to interact with TIT

TEST_REPO="/var/tmp/test_repo"

echo "=== STALE STASH TEST SETUP ==="

# Reset repo to clean state
cd "$TEST_REPO"
git checkout main 2>/dev/null || true
git reset --hard HEAD
git clean -fd
git stash clear 2>/dev/null || true

# Clean up tit config
rm -f ~/.config/tit/stash/list.toml
mkdir -p ~/.config/tit/stash

# Recreate main.txt
echo "Main 1" > main.txt
echo "Main 2" >> main.txt
echo "Main 3" >> main.txt

echo "Created clean repo at $TEST_REPO"

# Create a fake stash entry in config (simulating what would happen after time travel)
# This represents a stash that existed but was manually dropped
cat > ~/.config/tit/stash/list.toml << 'EOF'
[[stash]]
  operation = "time_travel"
  stash_hash = "46885675f1690000ef66e5eb7071790856256f3e"
  created_at = 2026-01-30T11:39:43.679496+07:00
  repo_path = "/var/tmp/test_repo"
  original_branch = "main"
  commit_hash = "a34e38f894fe39db9723996d610b3c6b29cef321"
EOF

echo "Created fake stash entry in config (simulating manually dropped stash)"
echo ""
echo "To test:"
echo "1. Start tit: cd $TEST_REPO && tit"
echo "2. History mode > commit 4 > Enter > Yes"
echo "3. ESC > Return to main > Yes"
echo "4. Should see 'Stash Not Found' dialog!"
echo ""
echo "The stash hash in config doesn't exist in git (manually dropped scenario)"
