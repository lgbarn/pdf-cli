package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func TestVerbose(t *testing.T) {
	cmd := GetRootCmd()

	// Reset verbose flag
	if err := cmd.PersistentFlags().Set("verbose", "false"); err != nil {
		t.Fatalf("Failed to reset verbose flag: %v", err)
	}

	// Test default (off)
	if Verbose() {
		t.Error("Verbose() should be false by default")
	}

	// Test verbose on
	if err := cmd.PersistentFlags().Set("verbose", "true"); err != nil {
		t.Fatalf("Failed to set verbose flag: %v", err)
	}
	if !Verbose() {
		t.Error("Verbose() should be true after setting flag")
	}

	// Reset for other tests
	_ = cmd.PersistentFlags().Set("verbose", "false")
}

func TestForce(t *testing.T) {
	cmd := GetRootCmd()

	// Reset force flag
	if err := cmd.PersistentFlags().Set("force", "false"); err != nil {
		t.Fatalf("Failed to reset force flag: %v", err)
	}

	// Test default (off)
	if Force() {
		t.Error("Force() should be false by default")
	}

	// Test force on
	if err := cmd.PersistentFlags().Set("force", "true"); err != nil {
		t.Fatalf("Failed to set force flag: %v", err)
	}
	if !Force() {
		t.Error("Force() should be true after setting flag")
	}

	// Reset for other tests
	_ = cmd.PersistentFlags().Set("force", "false")
}

func TestProgress(t *testing.T) {
	cmd := GetRootCmd()

	// Test that Progress() returns flag value correctly
	// Default is true from config.Defaults.ShowProgress
	if err := cmd.PersistentFlags().Set("progress", "false"); err != nil {
		t.Fatalf("Failed to set progress flag: %v", err)
	}
	if Progress() {
		t.Error("Progress() should be false after setting flag to false")
	}

	// Test progress on
	if err := cmd.PersistentFlags().Set("progress", "true"); err != nil {
		t.Fatalf("Failed to set progress flag: %v", err)
	}
	if !Progress() {
		t.Error("Progress() should be true after setting flag")
	}

	// Reset for other tests
	_ = cmd.PersistentFlags().Set("progress", "false")
}

func TestProgressFlagExists(t *testing.T) {
	cmd := GetRootCmd()

	progressFlag := cmd.PersistentFlags().Lookup("progress")
	if progressFlag == nil {
		t.Fatal("progress flag not found")
	}

	// Default comes from config.Defaults.ShowProgress which is true
	if progressFlag.DefValue != "true" {
		t.Errorf("progress flag default = %q, want %q", progressFlag.DefValue, "true")
	}
}

func TestPrintVerbose(t *testing.T) {
	cmd := GetRootCmd()

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test with verbose off
	_ = cmd.PersistentFlags().Set("verbose", "false")
	PrintVerbose("test message %s", "arg")

	// Close write end and read output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()
	if output != "" {
		t.Errorf("PrintVerbose() with verbose=false should produce no output, got %q", output)
	}

	// Test with verbose on
	r, w, _ = os.Pipe()
	os.Stderr = w

	_ = cmd.PersistentFlags().Set("verbose", "true")
	PrintVerbose("test message %s", "arg")

	w.Close()
	buf.Reset()
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output = buf.String()
	if !strings.Contains(output, "test message arg") {
		t.Errorf("PrintVerbose() with verbose=true should output message, got %q", output)
	}

	// Reset for other tests
	_ = cmd.PersistentFlags().Set("verbose", "false")
}

func TestPrintStatus(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintStatus("status message %d", 42)

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()
	if !strings.Contains(output, "status message 42") {
		t.Errorf("PrintStatus() should output message, got %q", output)
	}
}

func TestPrintProgress(t *testing.T) {
	cmd := GetRootCmd()

	// Test with verbose off - should produce no output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	_ = cmd.PersistentFlags().Set("verbose", "false")
	PrintProgress("operation")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()
	if output != "" {
		t.Errorf("PrintProgress() with verbose=false should produce no output, got %q", output)
	}

	// Test with verbose on
	r, w, _ = os.Pipe()
	os.Stderr = w

	_ = cmd.PersistentFlags().Set("verbose", "true")
	PrintProgress("operation")

	w.Close()
	buf.Reset()
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output = buf.String()
	if !strings.Contains(output, "Processing: operation") {
		t.Errorf("PrintProgress() with verbose=true should output message, got %q", output)
	}

	// Reset for other tests
	_ = cmd.PersistentFlags().Set("verbose", "false")
}

func TestAddCommand(t *testing.T) {
	cmd := GetRootCmd()

	// Create a test command
	testCmd := &cobra.Command{
		Use:   "testcmd",
		Short: "A test command",
	}

	// Count commands before
	beforeCount := len(cmd.Commands())

	// Add the test command
	AddCommand(testCmd)

	// Count commands after
	afterCount := len(cmd.Commands())

	if afterCount != beforeCount+1 {
		t.Errorf("AddCommand() should increase command count by 1, got %d -> %d", beforeCount, afterCount)
	}

	// Verify the command was added
	found := false
	for _, c := range cmd.Commands() {
		if c.Use == "testcmd" {
			found = true
			break
		}
	}
	if !found {
		t.Error("AddCommand() did not add the command")
	}
}

func TestExecute(t *testing.T) {
	cmd := GetRootCmd()

	// Test with help flag - should not error
	cmd.SetArgs([]string{"--help"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := Execute()
	if err != nil {
		t.Errorf("Execute() with --help returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "pdf") {
		t.Error("Execute() --help should contain 'pdf'")
	}
}

func TestExecuteWithInvalidCommand(t *testing.T) {
	cmd := GetRootCmd()

	// Test with invalid command
	cmd.SetArgs([]string{"nonexistentcommand"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := Execute()
	if err == nil {
		t.Error("Execute() with invalid command should return error")
	}
}

func TestRootCommandDescription(t *testing.T) {
	cmd := GetRootCmd()

	if cmd.Short == "" {
		t.Error("Root command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Root command should have a long description")
	}

	// Verify description mentions key features
	if !strings.Contains(cmd.Long, "merge") {
		t.Error("Root command description should mention 'merge'")
	}
	if !strings.Contains(cmd.Long, "split") {
		t.Error("Root command description should mention 'split'")
	}
}

func TestVerboseFlagShorthand(t *testing.T) {
	cmd := GetRootCmd()

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Fatal("verbose flag not found")
	}

	if verboseFlag.Shorthand != "v" {
		t.Errorf("verbose flag shorthand = %q, want %q", verboseFlag.Shorthand, "v")
	}
}

func TestForceFlagShorthand(t *testing.T) {
	cmd := GetRootCmd()

	forceFlag := cmd.PersistentFlags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("force flag not found")
	}

	if forceFlag.Shorthand != "f" {
		t.Errorf("force flag shorthand = %q, want %q", forceFlag.Shorthand, "f")
	}
}

func TestIsDryRun(t *testing.T) {
	cmd := GetRootCmd()

	// Reset dry-run flag
	if err := cmd.PersistentFlags().Set("dry-run", "false"); err != nil {
		t.Fatalf("Failed to reset dry-run flag: %v", err)
	}

	// Test default (off)
	if IsDryRun() {
		t.Error("IsDryRun() should be false by default")
	}

	// Test dry-run on
	if err := cmd.PersistentFlags().Set("dry-run", "true"); err != nil {
		t.Fatalf("Failed to set dry-run flag: %v", err)
	}
	if !IsDryRun() {
		t.Error("IsDryRun() should be true after setting flag")
	}

	// Reset for other tests
	_ = cmd.PersistentFlags().Set("dry-run", "false")
}

func TestDryRunFlagExists(t *testing.T) {
	cmd := GetRootCmd()

	dryRunFlag := cmd.PersistentFlags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("dry-run flag not found")
	}

	if dryRunFlag.DefValue != "false" {
		t.Errorf("dry-run flag default = %q, want %q", dryRunFlag.DefValue, "false")
	}

	if dryRunFlag.Usage != "Show what would be done without executing" {
		t.Errorf("dry-run flag usage = %q", dryRunFlag.Usage)
	}
}

func TestDryRunPrint(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	DryRunPrint("would do %s with %d items", "action", 5)

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("DryRunPrint() should contain [dry-run] prefix, got %q", output)
	}
	if !strings.Contains(output, "would do action with 5 items") {
		t.Errorf("DryRunPrint() should contain formatted message, got %q", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("DryRunPrint() should end with newline, got %q", output)
	}
}
