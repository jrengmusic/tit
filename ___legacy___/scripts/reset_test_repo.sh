#!/bin/bash
set -e

TEST_REPO="/var/tmp/test_repo"
CANON_BRANCH="main"

echo "=== Resetting test repo at $TEST_REPO ==="

# Remove existing repo
rm -rf "$TEST_REPO"
mkdir -p "$TEST_REPO"
cd "$TEST_REPO"

# Initialize repo
git init
git config user.email "test@example.com"
git config user.name "Test User"

# Create initial commit
echo "README" > README.md
git add README.md
git commit -m "1. Initial commit"

# Create feature branch with commits
git checkout -b feature
echo "Feature 1" > feature.txt
git add feature.txt
git commit -m "2. Add feature file"

echo "Feature 2" >> feature.txt
git add feature.txt
git commit -m "3. Update feature file"

# Switch back to main and add more commits
git checkout main
echo "Main 1" > main.txt
git add main.txt
git commit -m "4. Add main file"

echo "Main 2" >> main.txt
git add main.txt
git commit -m "5. Update main file"

echo "Main 3" >> main.txt
git add main.txt
git commit -m "6. Final main update"

# Create another branch with different content
git checkout -b experiment
echo "Experimental" > experiment.txt
git add experiment.txt
git commit -m "7. Add experiment"

git checkout main

# Create uncommitted changes for conflict testing
echo "UNCOMMITTED CHANGE 1" >> main.txt
echo "UNCOMMITTED CHANGE 2" >> main.txt

# Show the commit graph
echo ""
echo "=== Commit history ==="
git log --oneline --all --graph

echo ""
echo "=== Branches ==="
git branch -a

echo ""
echo "✓ Test repo ready at $TEST_REPO"
echo "Branches: main, feature, experiment"
echo ""
echo "UNCOMMITTED changes in main.txt for conflict testing!"
