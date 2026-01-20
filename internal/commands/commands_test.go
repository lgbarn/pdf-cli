package commands

import (
	"os"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

func TestOutputOrDefault(t *testing.T) {
	tests := []struct {
		output, inputFile, suffix, want string
	}{
		{"custom.pdf", "input.pdf", "_modified", "custom.pdf"},
		{"", "document.pdf", "_compressed", "document_compressed.pdf"},
		{"", "/path/to/file.pdf", "_rotated", "/path/to/file_rotated.pdf"},
		{"", "document.pdf", "", "document.pdf"},
		{"myfile.pdf", "document.pdf", "_ignored", "myfile.pdf"},
		{"", "./docs/file.pdf", "_out", "./docs/file_out.pdf"},
	}

	for _, tt := range tests {
		if got := outputOrDefault(tt.output, tt.inputFile, tt.suffix); got != tt.want {
			t.Errorf("outputOrDefault(%q, %q, %q) = %q, want %q", tt.output, tt.inputFile, tt.suffix, got, tt.want)
		}
	}
}

func TestCheckOutputFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.pdf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{
			name:    "non-existent file",
			output:  "/nonexistent/path/file.pdf",
			wantErr: false,
		},
		{
			name:    "existing file without force",
			output:  tmpPath,
			wantErr: true, // Should error because file exists and force is not enabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOutputFile(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkOutputFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAllCommandsRegistered(t *testing.T) {
	// Get the root command
	rootCmd := cli.GetRootCmd()
	if rootCmd == nil {
		t.Fatal("GetRootCmd() returned nil")
	}

	// Expected commands
	expectedCommands := []string{
		"info",
		"merge",
		"split",
		"extract",
		"reorder",
		"rotate",
		"compress",
		"encrypt",
		"decrypt",
		"text",
		"images",
		"meta",
		"watermark",
		"pdfa",
		"completion",
	}

	// Get all registered subcommands
	subCommands := rootCmd.Commands()
	registeredCommands := make(map[string]bool)
	for _, cmd := range subCommands {
		registeredCommands[cmd.Name()] = true
	}

	// Check each expected command is registered
	for _, expected := range expectedCommands {
		if !registeredCommands[expected] {
			t.Errorf("Expected command %q not registered", expected)
		}
	}
}

func TestCommandsExist(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	commands := []string{"info", "merge", "split", "extract", "rotate", "compress", "encrypt", "decrypt", "text", "images", "meta", "watermark", "completion"}

	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{name})
			if err != nil {
				t.Fatalf("Failed to find %s command: %v", name, err)
			}
			if cmd == nil {
				t.Fatalf("%s command is nil", name)
			}
		})
	}
}

func TestCommandFlags(t *testing.T) {
	tests := []struct {
		command string
		flags   []string
	}{
		{"info", []string{"format", "password"}},
		{"merge", []string{"output", "password"}},
		{"split", []string{"output", "pages"}},
		{"extract", []string{"output", "pages", "stdout"}},
		{"rotate", []string{"output", "angle", "pages"}},
		{"compress", []string{"output", "stdout"}},
		{"encrypt", []string{"output", "password", "owner-password"}},
		{"decrypt", []string{"output", "password", "stdout"}},
		{"text", []string{"pages"}},
		{"meta", []string{"format"}},
		{"watermark", []string{"text", "image", "pages"}},
		{"reorder", []string{"output", "stdout"}},
	}

	rootCmd := cli.GetRootCmd()
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{tt.command})
			if err != nil {
				t.Fatalf("Failed to find %s command: %v", tt.command, err)
			}
			for _, flag := range tt.flags {
				if cmd.Flags().Lookup(flag) == nil {
					t.Errorf("%s command should have --%s flag", tt.command, flag)
				}
			}
		})
	}
}

func TestValidateBatchOutput(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		output  string
		wantErr bool
	}{
		{"single file with output", []string{"file1.pdf"}, "out.pdf", false},
		{"multiple files with output", []string{"file1.pdf", "file2.pdf"}, "out.pdf", true},
		{"multiple files without output", []string{"file1.pdf", "file2.pdf"}, "", false},
		{"empty files with output", []string{}, "out.pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateBatchOutput(tt.files, tt.output, "_modified"); (err != nil) != tt.wantErr {
				t.Errorf("validateBatchOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessBatch(t *testing.T) {
	// Test with all successful
	files := []string{"file1", "file2", "file3"}
	successProcessor := func(file string) error {
		return nil
	}

	err := processBatch(files, successProcessor)
	if err != nil {
		t.Errorf("processBatch() with all success should not error, got %v", err)
	}

	// Test with some failures
	failedCount := 0
	failProcessor := func(file string) error {
		failedCount++
		if failedCount <= 2 {
			return nil
		}
		return os.ErrNotExist
	}

	err = processBatch(files, failProcessor)
	if err == nil {
		t.Error("processBatch() with failure should return error")
	}
}
