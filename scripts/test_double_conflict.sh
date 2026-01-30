#!/bin/bash
set -e

TEST_REPO="/var/tmp/test_repo"

echo "=== Setting up double-conflict scenario at $TEST_REPO ==="

cd "$TEST_REPO"

# Reset to clean state first
git checkout main 2>/dev/null || true
git reset --hard HEAD
git clean -fd
git stash clear 2>/dev/null || true

# Recreate main.txt to original state
echo "Main 1" > main.txt
echo "Main 2" >> main.txt
echo "Main 3" >> main.txt

echo ""
echo "=== Current state ==="
git log --oneline -3
echo ""
cat main.txt

echo ""
echo "=== Step 1: Creating uncommitted changes (STASH will have these) ==="
# Modify Main 2 line and add a new line using ed or perl for macOS compatibility
perl -pi -e 's/Main 2/STASH MODIFIED LINE 2/' main.txt
echo "STASH NEW LINE" >> main.txt

echo "Uncommitted changes:"
cat main.txt
echo ""

echo "=== Step 2: Starting tit and time traveling ==="
echo ""
echo "IN TIT:"
echo "1. History mode (h)"
echo "2. Select commit '4. Add main file'"
echo "3. Enter > Yes"
echo ""
echo "THEN in another terminal (while in time travel mode):"
echo "   perl -pi -e 's/Main 1/TIME TRAVEL MODIFIED LINE 1/' main.txt"
echo "   echo 'TIME TRAVEL NEW LINE' >> main.txt"
echo ""
echo "BACK IN TIT:"
echo "4. ESC > Return to main > Yes"
echo "5. Should trigger FIRST CONFLICT RESOLVER!"
echo "   - LOCAL: TIME TRAVEL MODIFIED LINE 1 (your time travel change)"
echo "   - REMOTE: Main 1 (original from main)"
echo "6. Choose one version or merge both"
echo "7. After resolving, should trigger SECOND conflict (stash apply)!"
echo "   - LOCAL: STASH MODIFIED LINE 2 (your uncommitted changes)"
echo "   - REMOTE: The resolved version from step 6"
echo "8. Resolve stash conflict"
echo "9. ESC > back to menu"
echo ""
echo "=== You should end up DIRTY with all changes preserved! ==="
echo ""
echo "Current uncommitted changes (will be stashed during time travel):"
cat main.txt
