package app

import (
	"strings"
	"testing"
)

func TestGetModeMetadata_KnownModes(t *testing.T) {
	tests := []struct {
		name         string
		mode         AppMode
		wantNamePart string
	}{
		{"ModeMenu", ModeMenu, "menu"},
		{"ModeConsole", ModeConsole, "console"},
		{"ModeInput", ModeInput, "input"},
		{"ModeHistory", ModeHistory, "history"},
		{"ModeConfirmation", ModeConfirmation, "confirmation"},
		{"ModeConflictResolve", ModeConflictResolve, "conflict"},
		{"ModeSetupWizard", ModeSetupWizard, "setup"},
		{"ModeConfig", ModeConfig, "config"},
		{"ModeBranchPicker", ModeBranchPicker, "branch"},
		{"ModePreferences", ModePreferences, "preferences"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := GetModeMetadata(tt.mode)
			if !strings.Contains(meta.Name, tt.wantNamePart) {
				t.Errorf("GetModeMetadata(%v).Name = %q, want it to contain %q", tt.mode, meta.Name, tt.wantNamePart)
			}
		})
	}
}

func TestGetModeMetadata_UnknownMode(t *testing.T) {
	meta := GetModeMetadata(AppMode(999))
	if meta.Name != "unknown" {
		t.Errorf("unknown mode Name: got %q, want %q", meta.Name, "unknown")
	}
}

func TestAppMode_String_MatchesMetadata(t *testing.T) {
	modes := []AppMode{
		ModeMenu,
		ModeInput,
		ModeConsole,
		ModeConfirmation,
		ModeHistory,
		ModeConflictResolve,
		ModeInitializeLocation,
		ModeInitializeBranches,
		ModeCloneURL,
		ModeCloneLocation,
		ModeClone,
		ModeSelectBranch,
		ModeFileHistory,
		ModeSetupWizard,
		ModeConfig,
		ModeBranchPicker,
		ModePreferences,
	}
	for _, m := range modes {
		want := GetModeMetadata(m).Name
		got := m.String()
		if got != want {
			t.Errorf("AppMode(%d).String() = %q, want %q", int(m), got, want)
		}
	}
}

func TestSetupWizardStep_String_AllDefined(t *testing.T) {
	tests := []struct {
		step SetupWizardStep
		want string
	}{
		{SetupStepWelcome, "welcome"},
		{SetupStepPrerequisites, "prerequisites"},
		{SetupStepEmail, "email"},
		{SetupStepGenerate, "generate"},
		{SetupStepDisplayKey, "display_key"},
		{SetupStepComplete, "complete"},
		{SetupStepError, "error"},
	}
	for _, tt := range tests {
		got := tt.step.String()
		if got == "" {
			t.Errorf("SetupWizardStep(%d).String() is empty", int(tt.step))
		}
		if got == "unknown" {
			t.Errorf("SetupWizardStep(%d).String() = %q, should not be unknown", int(tt.step), got)
		}
		if got != tt.want {
			t.Errorf("SetupWizardStep(%d).String() = %q, want %q", int(tt.step), got, tt.want)
		}
	}
}

func TestSetupWizardStep_String_Undefined(t *testing.T) {
	got := SetupWizardStep(999).String()
	if got != "unknown" {
		t.Errorf("undefined step String() = %q, want %q", got, "unknown")
	}
}
