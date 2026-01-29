# SURGEON Fix Summary

**Issue:** "Discard all changes" menu item wasn't actually restoring working tree to clean state.

**Root Cause:** `cmdHardReset()` in `internal/app/op_pull.go` only performed `git reset --hard origin/<branch>` which resets tracked files but does NOT remove untracked files/directories. The `.carol/` directory remained because it's untracked.

**Fix Applied:** Added `git clean -fd` after successful reset to remove untracked files, matching the behavior of `discardWorkingTreeChanges()`.

**Files Changed:**
- `internal/app/op_pull.go` - Added untracked file cleanup to hard reset operation

**Testing:** "Discard all changes" now properly removes untracked directories like `.carol/`.
