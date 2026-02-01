package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("password", "", "password")
	cmd.Flags().String("password-file", "", "password file")
	return cmd
}

func TestReadPassword_PasswordFile(t *testing.T) {
	tmpDir := t.TempDir()
	pwdFile := filepath.Join(tmpDir, "pwd.txt")
	if err := os.WriteFile(pwdFile, []byte("filepassword\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCmd()
	cmd.Flags().Set("password-file", pwdFile)

	got, err := ReadPassword(cmd, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "filepassword" {
		t.Errorf("got %q, want %q", got, "filepassword")
	}
}

func TestReadPassword_PasswordFileTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	pwdFile := filepath.Join(tmpDir, "pwd.txt")
	data := make([]byte, 1025)
	if err := os.WriteFile(pwdFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCmd()
	cmd.Flags().Set("password-file", pwdFile)

	_, err := ReadPassword(cmd, "")
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
}

func TestReadPassword_PasswordFileMissing(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Set("password-file", "/nonexistent/path")

	_, err := ReadPassword(cmd, "")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadPassword_EnvVar(t *testing.T) {
	t.Setenv("PDF_CLI_PASSWORD", "envpassword")

	cmd := newTestCmd()

	got, err := ReadPassword(cmd, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "envpassword" {
		t.Errorf("got %q, want %q", got, "envpassword")
	}
}

func TestReadPassword_DeprecatedFlag(t *testing.T) {
	// Ensure env var is not set
	t.Setenv("PDF_CLI_PASSWORD", "")

	cmd := newTestCmd()
	cmd.Flags().Set("password", "flagpassword")

	got, err := ReadPassword(cmd, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "flagpassword" {
		t.Errorf("got %q, want %q", got, "flagpassword")
	}
}

func TestReadPassword_Priority_FileOverEnv(t *testing.T) {
	tmpDir := t.TempDir()
	pwdFile := filepath.Join(tmpDir, "pwd.txt")
	if err := os.WriteFile(pwdFile, []byte("filepass"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PDF_CLI_PASSWORD", "envpass")

	cmd := newTestCmd()
	cmd.Flags().Set("password-file", pwdFile)
	cmd.Flags().Set("password", "flagpass")

	got, err := ReadPassword(cmd, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "filepass" {
		t.Errorf("got %q, want %q (file should have priority)", got, "filepass")
	}
}

func TestReadPassword_Priority_EnvOverFlag(t *testing.T) {
	t.Setenv("PDF_CLI_PASSWORD", "envpass")

	cmd := newTestCmd()
	cmd.Flags().Set("password", "flagpass")

	got, err := ReadPassword(cmd, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "envpass" {
		t.Errorf("got %q, want %q (env should have priority over flag)", got, "envpass")
	}
}

func TestReadPassword_NoSource(t *testing.T) {
	t.Setenv("PDF_CLI_PASSWORD", "")
	t.Setenv("CI", "true") // Prevent interactive prompt

	cmd := newTestCmd()

	got, err := ReadPassword(cmd, "Enter password: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string when no source available", got)
	}
}
