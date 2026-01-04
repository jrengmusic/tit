	// Handle character input in input modes
	if a.isInputMode() && len(keyStr) == 1 && keyStr[0] >= 32 && keyStr[0] <= 126 {
		// Insert character at cursor position
		if a.mode == ModeInitializeBranch {
			// Branch input mode uses initBranchName field
			a.initBranchName = a.initBranchName[:a.inputCursorPosition] + keyStr + a.initBranchName[a.inputCursorPosition:]
			a.inputCursorPosition++
		} else {
			// Generic input mode
			a.inputValue = a.inputValue[:a.inputCursorPosition] + keyStr + a.inputValue[a.inputCursorPosition:]
			a.inputCursorPosition++
			a.validateCloneURLInput()
		}
		return a, nil
	}

	// Handle backspace in input modes
	if a.isInputMode() && keyStr == "backspace" {
		if a.mode == ModeInitializeBranch {
			// Delete from active field
			if a.inputCursorPosition > 0 {
				a.initBranchName = a.initBranchName[:a.inputCursorPosition-1] + a.initBranchName[a.inputCursorPosition:]
				a.inputCursorPosition--
			}
		} else {
			// Generic input mode
			if a.inputCursorPosition > 0 {
				a.inputValue = a.inputValue[:a.inputCursorPosition-1] + a.inputValue[a.inputCursorPosition:]
				a.inputCursorPosition--
				a.validateCloneURLInput()
			}
		}
		return a, nil
	}
