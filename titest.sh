#!/bin/bash
# TIT Dirty Pull Test Setup
# Run this from OUTSIDE tit_test_repo
# It will set up various dirty pull scenarios

TEST_REPO="/Users/jreng/Documents/Poems/inf/tit_test_repo"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

show_menu() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║        TIT DIRTY PULL TEST SCENARIOS               ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}0.${NC} Clean up and reset repo to fresh state"
    echo -e "${GREEN}1.${NC} Pull with conflicts (clean tree)"
    echo -e "${GREEN}2.${NC} Dirty pull - merge with conflicts"
    echo -e "${GREEN}3.${NC} Dirty pull - rebase with conflicts"
    echo -e "${GREEN}4.${NC} Dirty pull - stash apply conflicts after pull"
    echo -e "${GREEN}5.${NC} Dirty pull - clean pull (no conflicts)"
    echo -e "${GREEN}s.${NC} Show current status"
    echo -e "${RED}q.${NC} Quit"
    echo ""
}

cleanup_git_state() {
    cd "$TEST_REPO"

    echo -e "${YELLOW}Cleaning up any leftover git state...${NC}"

    # Abort any ongoing operations
    git rebase --abort 2>/dev/null
    git merge --abort 2>/dev/null
    git cherry-pick --abort 2>/dev/null

    # Clean working tree
    git reset --hard HEAD 2>/dev/null
    git clean -fd 2>/dev/null

    # Drop all stashes
    git stash clear 2>/dev/null

    echo -e "${GREEN}✓ Git state cleaned${NC}"
}

reset_repo() {
    echo -e "${YELLOW}Resetting test repo to fresh state...${NC}"

    cd "$TEST_REPO"

    # Clean up any ongoing operations
    cleanup_git_state

    # Reset to main branch
    git checkout main 2>/dev/null

    # Pull latest from remote
    git pull origin main

    echo -e "${GREEN}✓ Repo reset to clean state${NC}"
    git status
}

show_status() {
    cd "$TEST_REPO"
    echo -e "${BLUE}Current git status:${NC}"
    git status
    echo ""
    echo -e "${BLUE}Branch info:${NC}"
    git branch -vv
    echo ""
    echo -e "${BLUE}Stash list:${NC}"
    git stash list
    echo ""
    echo -e "${BLUE}Operation state:${NC}"
    if [ -d .git/rebase-merge ] || [ -d .git/rebase-apply ]; then
        echo -e "${RED}REBASE IN PROGRESS${NC}"
    elif [ -f .git/MERGE_HEAD ]; then
        echo -e "${RED}MERGE IN PROGRESS${NC}"
    elif [ -f .git/CHERRY_PICK_HEAD ]; then
        echo -e "${RED}CHERRY-PICK IN PROGRESS${NC}"
    else
        echo -e "${GREEN}No ongoing operations${NC}"
    fi
}

scenario_1() {
    echo -e "${YELLOW}Setting up: Pull with conflicts (clean tree)${NC}"
    cd "$TEST_REPO"

    cleanup_git_state

    # Make conflicting local commit
    echo "local change" >> conflict.txt
    git add conflict.txt
    git commit -m "Local conflicting change"

    echo -e "${GREEN}✓ Setup complete${NC}"
    echo -e "${BLUE}Now run TIT and select 'Pull from remote'${NC}"
    echo -e "${BLUE}Expected: Immediate conflict resolution UI${NC}"
}

scenario_2() {
    echo -e "${YELLOW}Setting up: Dirty pull - merge with conflicts${NC}"
    cd "$TEST_REPO"

    cleanup_git_state

    # Create WIP changes
    echo "wip work" >> wip.txt

    # Make conflicting local commit that's ahead of remote
    echo "local change" >> conflict.txt
    git add conflict.txt
    git commit -m "Local conflicting change"

    echo -e "${GREEN}✓ Setup complete${NC}"
    echo -e "${BLUE}Working tree: DIRTY (wip.txt)${NC}"
    echo -e "${BLUE}Now run TIT and select 'Merge local and remote (+ save WIP)'${NC}"
    echo -e "${BLUE}Expected: Stash → Merge → Conflict UI → Resolve → Stash apply${NC}"
}

scenario_3() {
    echo -e "${YELLOW}Setting up: Dirty pull - rebase with conflicts${NC}"
    cd "$TEST_REPO"

    cleanup_git_state

    # Create WIP changes
    echo "wip work" >> wip.txt

    # Make conflicting local commit
    echo "local change" >> conflict.txt
    git add conflict.txt
    git commit -m "Local conflicting change"

    echo -e "${GREEN}✓ Setup complete${NC}"
    echo -e "${BLUE}Working tree: DIRTY (wip.txt)${NC}"
    echo -e "${BLUE}Now run TIT and select 'Rebase onto remote (+ save WIP)'${NC}"
    echo -e "${BLUE}Expected: Stash → Rebase → Conflict UI → Resolve → Stash apply${NC}"
}

scenario_4() {
    echo -e "${YELLOW}Setting up: Dirty pull - stash apply conflicts after pull${NC}"
    cd "$TEST_REPO"

    cleanup_git_state

    # Create WIP that will conflict with incoming changes
    echo "local wip change that conflicts" >> conflict.txt

    # Local is behind remote (will pull cleanly)
    # But stash apply will conflict

    echo -e "${GREEN}✓ Setup complete${NC}"
    echo -e "${BLUE}Working tree: DIRTY (conflict.txt modified)${NC}"
    echo -e "${BLUE}Now run TIT and select 'Pull from remote (+ save WIP)'${NC}"
    echo -e "${BLUE}Expected: Stash → Pull succeeds → Stash apply conflicts → Conflict UI${NC}"
}

scenario_5() {
    echo -e "${YELLOW}Setting up: Dirty pull - clean pull (no conflicts)${NC}"
    cd "$TEST_REPO"

    cleanup_git_state

    # Create WIP that won't conflict
    echo "wip work in separate file" >> safe_wip.txt

    echo -e "${GREEN}✓ Setup complete${NC}"
    echo -e "${BLUE}Working tree: DIRTY (safe_wip.txt)${NC}"
    echo -e "${BLUE}Now run TIT and select 'Pull from remote (+ save WIP)'${NC}"
    echo -e "${BLUE}Expected: Stash → Pull → Stash pop → Success (auto-return to menu)${NC}"
}

# Main loop
while true; do
    show_menu
    read -p "Select scenario: " choice

    case $choice in
        0) reset_repo ;;
        1) scenario_1 ;;
        2) scenario_2 ;;
        3) scenario_3 ;;
        4) scenario_4 ;;
        5) scenario_5 ;;
        s|S) show_status ;;
        q|Q)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            ;;
    esac

    read -p "Press Enter to continue..."
done
