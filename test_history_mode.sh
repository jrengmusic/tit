#!/bin/bash

# Test script for History mode implementation
# This script creates a test git repository and verifies that History mode works

echo "🧪 Testing History Mode Implementation"
echo "===================================="

# Build the application
echo "1. Building application..."
cd /Users/jreng/Documents/Poems/inf/tit
./build.sh
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"

# Create a test repository
echo "2. Creating test repository..."
TEST_DIR="/tmp/tit_history_test"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

git init
git config user.name "Test User"
git config user.email "test@example.com"

# Create some commits
echo "   Creating test commits..."
for i in {1..5}; do
    echo "Test content $i" > "file$i.txt"
    git add "file$i.txt"
    git commit -m "Test commit $i"
done

echo "✅ Test repository created with 5 commits"

# Test the application
echo "3. Testing History mode..."
echo "   (This will require manual verification)"
echo "   Run: cd $TEST_DIR && /Users/jreng/Documents/Poems/inf/tit/tit_x64"
echo "   Then press 'l' to enter History mode"
echo "   Expected: Should show commit list and details pane"
echo "   Navigation: ↑↓ to move, TAB to switch panes, ESC to exit"

echo ""
echo "📋 Test Checklist:"
echo "- [ ] History menu item appears in main menu (shortcut 'l')"
echo "- [ ] Pressing 'l' enters History mode"
echo "- [ ] Shows split-pane layout (commits left, details right)"
echo "- [ ] Up/Down navigates commit list"
echo "- [ ] TAB switches focus between panes"
echo "- [ ] ESC returns to main menu"
echo "- [ ] Footer shows correct hints"

echo ""
echo "🎯 Phase 4 Implementation Status:"
echo "✅ UI rendering (Phase 3)"
echo "✅ Menu integration"
echo "✅ Keyboard handlers"
echo "✅ Dispatcher"
echo "✅ Cache loading"
echo "✅ Navigation logic"

echo ""
echo "🔧 Next Steps:"
echo "- Manual testing of History mode"
echo "- Verify cache loading works correctly"
echo "- Test navigation and pane switching"
echo "- Phase 5: File(s) History UI"

echo ""
echo "Test repository location: $TEST_DIR"