package commands

import (
	"os"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

func TestOutputOrDefault(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		inputFile string
		suffix    string
		want      string
	}{
		{
			name:      "explicit output",
			output:    "custom.pdf",
			inputFile: "input.pdf",
			suffix:    "_modified",
			want:      "custom.pdf",
		},
		{
			name:      "default output with suffix",
			output:    "",
			inputFile: "document.pdf",
			suffix:    "_compressed",
			want:      "document_compressed.pdf",
		},
		{
			name:      "default output with path",
			output:    "",
			inputFile: "/path/to/file.pdf",
			suffix:    "_rotated",
			want:      "/path/to/file_rotated.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := outputOrDefault(tt.output, tt.inputFile, tt.suffix)
			if got != tt.want {
				t.Errorf("outputOrDefault() = %v, want %v", got, tt.want)
			}
		})
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
		"rotate",
		"compress",
		"encrypt",
		"decrypt",
		"text",
		"images",
		"meta",
		"watermark",
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

func TestInfoCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"info"})
	if err != nil {
		t.Fatalf("Failed to find info command: %v", err)
	}
	if cmd == nil {
		t.Fatal("info command is nil")
	}
	if cmd.Use != "info <file.pdf>" {
		t.Errorf("info command Use = %q, want %q", cmd.Use, "info <file.pdf>")
	}
}

func TestMergeCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"merge"})
	if err != nil {
		t.Fatalf("Failed to find merge command: %v", err)
	}
	if cmd == nil {
		t.Fatal("merge command is nil")
	}
}

func TestSplitCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"split"})
	if err != nil {
		t.Fatalf("Failed to find split command: %v", err)
	}
	if cmd == nil {
		t.Fatal("split command is nil")
	}
}

func TestExtractCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"extract"})
	if err != nil {
		t.Fatalf("Failed to find extract command: %v", err)
	}
	if cmd == nil {
		t.Fatal("extract command is nil")
	}
}

func TestRotateCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"rotate"})
	if err != nil {
		t.Fatalf("Failed to find rotate command: %v", err)
	}
	if cmd == nil {
		t.Fatal("rotate command is nil")
	}
}

func TestCompressCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"compress"})
	if err != nil {
		t.Fatalf("Failed to find compress command: %v", err)
	}
	if cmd == nil {
		t.Fatal("compress command is nil")
	}
}

func TestEncryptCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"encrypt"})
	if err != nil {
		t.Fatalf("Failed to find encrypt command: %v", err)
	}
	if cmd == nil {
		t.Fatal("encrypt command is nil")
	}
}

func TestDecryptCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"decrypt"})
	if err != nil {
		t.Fatalf("Failed to find decrypt command: %v", err)
	}
	if cmd == nil {
		t.Fatal("decrypt command is nil")
	}
}

func TestTextCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"text"})
	if err != nil {
		t.Fatalf("Failed to find text command: %v", err)
	}
	if cmd == nil {
		t.Fatal("text command is nil")
	}
}

func TestImagesCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"images"})
	if err != nil {
		t.Fatalf("Failed to find images command: %v", err)
	}
	if cmd == nil {
		t.Fatal("images command is nil")
	}
}

func TestMetaCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"meta"})
	if err != nil {
		t.Fatalf("Failed to find meta command: %v", err)
	}
	if cmd == nil {
		t.Fatal("meta command is nil")
	}
}

func TestWatermarkCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"watermark"})
	if err != nil {
		t.Fatalf("Failed to find watermark command: %v", err)
	}
	if cmd == nil {
		t.Fatal("watermark command is nil")
	}
}
