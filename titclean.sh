#!/bin/bash
# Emergency cleanup for broken test_repo state

cd /Users/jreng/Documents/Poems/inf/tit_test_repo

echo "Cleaning up broken git state..."

# Abort any ongoing operations
git rebase --abort 2>/dev/null
git merge --abort 2>/dev/null
git cherry-pick --abort 2>/dev/null

# Hard reset to clean state
git reset --hard HEAD

# Clean untracked files
git clean -fd

# Drop all stashes
git stash clear

# Return to main branch
git checkout main

# Show final status
echo ""
echo "Cleanup complete. Current status:"
git status

echo ""
echo "✓ test_repo is now clean"
echo "Run: /Users/jreng/Documents/Poems/inf/tit_test_setup.sh"
echo "to set up test scenarios"
