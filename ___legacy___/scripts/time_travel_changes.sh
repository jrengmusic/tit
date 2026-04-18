#!/bin/bash
# Run this WHILE in time travel mode (detached HEAD at commit 4)
# This creates changes that will conflict with the stash

cd /var/tmp/test_repo

echo "=== Creating TIME TRAVEL changes (will conflict with stash) ==="

# Check current state
echo "Current main.txt in time travel:"
cat main.txt
echo ""

# Modify Main 1 (this will conflict with STASH MODIFIED LINE 2 from stash)
perl -pi -e 's/Main 1/TIME TRAVEL MODIFIED LINE 1/' main.txt

# Add a new line
echo "TIME TRAVEL NEW LINE" >> main.txt

echo "After time travel modifications:"
cat main.txt
echo ""
echo "=== Changes created! Return to main to trigger conflicts! ==="
