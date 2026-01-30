#!/bin/bash
set -e

TEST_REPO="/var/tmp/test_repo_conflict"
CANON_BRANCH="main"

echo "=== Creating time travel conflict test repo at $TEST_REPO ==="

# Remove existing repo
rm -rf "$TEST_REPO"
mkdir -p "$TEST_REPO"
cd "$TEST_REPO"

# Initialize repo
git init
git config user.email "test@example.com"
git config user.name "Test User"

# Create initial commit
echo "Line 1" > file.txt
git add file.txt
git commit -m "1. Initial commit"

# Create second commit
echo "Line 2" >> file.txt
git add file.txt
git commit -m "2. Add line 2"

# Create third commit
echo "Line 3" >> file.txt
git add file.txt
git commit -m "3. Add line 3"

echo ""
echo "=== Initial state ==="
echo "file.txt:"
cat file.txt
echo ""
echo "Commit history:"
git log --oneline

echo ""
echo "=== Setting up conflict scenario ==="
echo "Creating uncommitted changes that will be stashed during time travel..."

# Create uncommitted changes (simulating user edits before time traveling)
echo "UNCOMMITTED CHANGE 1" >> file.txt
echo "UNCOMMITTED CHANGE 2" >> file.txt

echo ""
echo "file.txt with uncommitted changes:"
cat file.txt
echo ""

echo "Uncommitted changes staged:"
git status

echo ""
echo "✓ Test repo ready at $TEST_REPO"
echo ""
echo "=== HOW TO TRIGGER CONFLICT ==="
echo "1. Run: cd $TEST_REPO && tit"
echo "2. Enter History mode (press 'h')"
echo "3. Select commit '2. Add line 2' (second commit in list)"
echo "4. Press Enter, confirm time travel with 'y'"
echo "   → Your uncommitted changes will be stashed"
echo "5. Now you're time traveling with a clean working tree"
echo "6. Modify file.txt during time travel:"
echo "   echo 'TIME TRAVEL MODIFICATION' >> file.txt"
echo "7. Press ESC, select 'Return to main', confirm with 'y'"
echo "8. Should trigger conflict resolver!"
echo ""
echo "Conflict reason:"
echo "- Stash contains: UNCOMMITTED CHANGE 1, UNCOMMITTED CHANGE 2"
echo "- Time travel has: TIME TRAVEL MODIFICATION"
echo "- Both modify file.txt differently → CONFLICT"
