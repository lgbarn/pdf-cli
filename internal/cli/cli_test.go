package cli

import (
	"testing"
)

func TestSetVersion(t *testing.T) {
	SetVersion("1.0.0", "abc123", "2024-01-01")

	if version != "1.0.0" {
		t.Errorf("SetVersion() version = %v, want %v", version, "1.0.0")
	}
	if commit != "abc123" {
		t.Errorf("SetVersion() commit = %v, want %v", commit, "abc123")
	}
	if buildDate != "2024-01-01" {
		t.Errorf("SetVersion() buildDate = %v, want %v", buildDate, "2024-01-01")
	}
}

func TestGetRootCmd(t *testing.T) {
	cmd := GetRootCmd()
	if cmd == nil {
		t.Fatal("GetRootCmd() returned nil")
	}
	if cmd.Use != "pdf" {
		t.Errorf("GetRootCmd() Use = %v, want %v", cmd.Use, "pdf")
	}
}

func TestRootCommandFlags(t *testing.T) {
	cmd := GetRootCmd()

	// Check verbose flag exists
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("verbose flag not found")
	}

	// Check force flag exists
	forceFlag := cmd.PersistentFlags().Lookup("force")
	if forceFlag == nil {
		t.Error("force flag not found")
	}
}
