#!/bin/bash
# Automated test for double-conflict scenario
# This script sets up the repo and shows expected behavior at each step

TEST_REPO="/var/tmp/test_repo"
cd "$TEST_REPO"

echo "=============================================="
echo "  DOUBLE CONFLICT TEST - STARTING"
echo "=============================================="
echo ""

# Reset repo
git checkout main 2>/dev/null || true
git reset --hard HEAD
git clean -fd
git stash clear 2>/dev/null || true

# Recreate main.txt
echo "Main 1" > main.txt
echo "Main 2" >> main.txt
echo "Main 3" >> main.txt

echo "=== STEP 1: Clean state ==="
echo "main.txt:"
cat main.txt
echo ""

echo "=== STEP 2: Create UNCOMMITTED changes ==="
echo "Running: perl -pi -e 's/Main 2/STASH MODIFIED LINE 2/' main.txt"
echo "Running: echo 'STASH NEW LINE' >> main.txt"
perl -pi -e 's/Main 2/STASH MODIFIED LINE 2/' main.txt
echo "STASH NEW LINE" >> main.txt
echo ""
echo "Uncommitted changes (will be stashed):"
cat main.txt
echo ""

echo "=============================================="
echo "  NOW START TIT AND DO THE FOLLOWING:"
echo "=============================================="
echo ""
echo "1. In tit: History mode (h)"
echo "2. Select commit '4. Add main file'"
echo "3. Press Enter, confirm with y"
echo "   → Uncommitted changes STASHED"
echo "   → Now in time travel mode (detached HEAD)"
echo ""
echo "4. In another terminal, RUN:"
echo "   ./scripts/time_travel_changes.sh"
echo "   → Creates TIME TRAVEL changes"
echo "   → main.txt now has TIME TRAVEL MODIFIED LINE 1"
echo ""
echo "5. BACK IN TIT:"
echo "   - Press ESC"
echo "   - Select 'Return to main'"
echo "   - Press Enter, confirm with y"
echo ""
echo "6. FIRST CONFLICT RESOLVER should appear!"
echo "   - LOCAL: TIME TRAVEL MODIFIED LINE 1"
echo "   - REMOTE: Main 1"
echo "   → Choose one (or merge both)"
echo ""
echo "7. After resolving first conflict,"
echo "   SECOND CONFLICT RESOLVER should appear!"
echo "   (stash apply conflicts with your merge result)"
echo "   → LOCAL: STASH MODIFIED LINE 2"
echo "   - REMOTE: Your resolved version from step 6"
echo "   → Choose one (or merge both)"
echo ""
echo "8. After resolving both conflicts:"
echo "   - ESC back to menu"
echo "   - You should be on main, DIRTY"
echo "   - All changes preserved!"
echo ""
echo "=============================================="
echo "  EXPECTED OUTCOME:"
echo "=============================================="
echo "main.txt should have BOTH sets of changes:"
echo "  - TIME TRAVEL changes (from step 4)"
echo "  - STASH changes (from step 2)"
echo ""
echo "Working tree should be DIRTY (uncommitted)"
echo "No stash should remain (all applied)"
echo ""
echo "=============================================="
echo "  START TIT NOW!"
echo "=============================================="
