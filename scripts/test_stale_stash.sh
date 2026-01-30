#!/bin/bash
# Test script to verify stale stash handling
# This creates a scenario where:
# 1. User time travels with dirty changes (stash created)
# 2. User manually drops the stash via git stash drop
# 3. User tries to return from time travel
# 4. TIT should show confirmation dialog

TEST_REPO="/var/tmp/test_repo"

echo "=== Setting up STALE STASH test scenario ==="
echo ""

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

echo "=== STEP 1: Clean state ==="
echo "main.txt:"
cat main.txt
echo ""

echo "=== STEP 2: Create uncommitted changes ==="
echo "perl -pi -e 's/Main 2/STASH CONTENT/' main.txt"
perl -pi -e 's/Main 2/STASH CONTENT/' main.txt
echo "main.txt after uncommitted change:"
cat main.txt
echo ""

echo "=== STEP 3: Start TIT and time travel ==="
echo ""
echo "IN TIT:"
echo "1. History mode (h)"
echo "2. Select commit '4. Add main file'"
echo "3. Press Enter > Yes"
echo "   → This will stash your uncommitted changes"
echo "   → You're now in time travel mode"
echo ""
echo "=== STEP 4: Manually drop the stash (while in time travel) ==="
echo ""
echo "In another terminal, RUN:"
echo "   git stash drop"
echo ""
echo "This manually drops the stash that TIT created."
echo "Now the stash entry exists in config (~/.config/tit/stash/list.toml)"
echo "but the actual stash no longer exists in git."
echo ""
echo "=== STEP 5: Try to return from time travel ==="
echo ""
echo "BACK IN TIT:"
echo "1. Press ESC"
echo "2. Select 'Return to main'"
echo "3. Press Enter"
echo ""
echo "=== EXPECTED BEHAVIOR ==="
echo "TIT should show a dialog:"
echo "  Title: 'Stash Not Found'"
echo "  Explanation: 'Original stash [hash] was manually dropped. Continue without restoring stash?'"
echo "  Options: [Continue] [Cancel]"
echo ""
echo "If you click Continue → will proceed without restoring stash (clean config)"
echo "If you click Cancel → will cancel operation"
echo ""
echo "=============================================="
echo "  READY TO TEST - START TIT NOW!"
echo "=============================================="
