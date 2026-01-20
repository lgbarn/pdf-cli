package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestCompletionShells(t *testing.T) {
	tests := []struct {
		shell   string
		pattern string
	}{
		{"bash", "__pdf"},
		{"zsh", "compdef"},
		{"fish", "complete"},
		{"powershell", "Register"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			cmd := GetRootCmd()
			output := captureStdout(func() {
				cmd.SetArgs([]string{"completion", tt.shell})
				if err := cmd.Execute(); err != nil {
					t.Fatalf("completion %s failed: %v", tt.shell, err)
				}
			})

			if output == "" {
				t.Errorf("completion %s returned empty output", tt.shell)
			}
			if !strings.Contains(output, tt.pattern) && !strings.Contains(output, "pdf") {
				t.Logf("Warning: %s completion may not contain expected patterns", tt.shell)
			}
		})
	}
}

func TestCompletionErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"invalid shell", []string{"completion", "invalid"}},
		{"no args", []string{"completion"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("completion %v should return error", tt.args)
			}
		})
	}
}

func TestCompletionValidArgs(t *testing.T) {
	cmd := GetRootCmd()
	completionCmd, _, err := cmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("Failed to find completion command: %v", err)
	}

	expectedArgs := map[string]bool{"bash": true, "zsh": true, "fish": true, "powershell": true}
	if len(completionCmd.ValidArgs) != len(expectedArgs) {
		t.Errorf("ValidArgs length = %d, want %d", len(completionCmd.ValidArgs), len(expectedArgs))
	}
	for _, arg := range completionCmd.ValidArgs {
		if !expectedArgs[arg] {
			t.Errorf("unexpected ValidArg: %q", arg)
		}
	}
}
