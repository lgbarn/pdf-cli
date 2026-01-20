package commands

import (
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

func TestPdfaCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"pdfa"})
	if err != nil {
		t.Fatalf("Failed to find pdfa command: %v", err)
	}
	if cmd == nil {
		t.Fatal("pdfa command is nil")
	}
	if cmd.Use != "pdfa" {
		t.Errorf("pdfa command Use = %q, want %q", cmd.Use, "pdfa")
	}
}

func TestPdfaCommandStructure(t *testing.T) {
	// Verify pdfa has subcommands
	if !pdfaCmd.HasSubCommands() {
		t.Error("pdfa command should have subcommands")
	}

	subCmds := pdfaCmd.Commands()
	names := make(map[string]bool)
	for _, cmd := range subCmds {
		names[cmd.Name()] = true
	}

	if !names["validate"] {
		t.Error("pdfa should have 'validate' subcommand")
	}
	if !names["convert"] {
		t.Error("pdfa should have 'convert' subcommand")
	}
}

func TestPdfaValidateSubcommand(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"pdfa", "validate"})
	if err != nil {
		t.Fatalf("Failed to find pdfa validate command: %v", err)
	}
	if cmd == nil {
		t.Fatal("pdfa validate command is nil")
	}
	if cmd.Use != "validate <file.pdf>" {
		t.Errorf("pdfa validate command Use = %q, want %q", cmd.Use, "validate <file.pdf>")
	}
}

func TestPdfaConvertSubcommand(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"pdfa", "convert"})
	if err != nil {
		t.Fatalf("Failed to find pdfa convert command: %v", err)
	}
	if cmd == nil {
		t.Fatal("pdfa convert command is nil")
	}
	if cmd.Use != "convert <file.pdf>" {
		t.Errorf("pdfa convert command Use = %q, want %q", cmd.Use, "convert <file.pdf>")
	}
}

func TestPdfaValidateFlags(t *testing.T) {
	requiredFlags := []string{"level", "password"}

	for _, flagName := range requiredFlags {
		t.Run(flagName+" flag", func(t *testing.T) {
			if flag := pdfaValidateCmd.Flags().Lookup(flagName); flag == nil {
				t.Errorf("pdfa validate should have --%s flag", flagName)
			}
		})
	}
}

func TestPdfaConvertFlags(t *testing.T) {
	requiredFlags := []string{"level", "output", "password"}

	for _, flagName := range requiredFlags {
		t.Run(flagName+" flag", func(t *testing.T) {
			if flag := pdfaConvertCmd.Flags().Lookup(flagName); flag == nil {
				t.Errorf("pdfa convert should have --%s flag", flagName)
			}
		})
	}
}

func TestPdfaConvertLevelDefault(t *testing.T) {
	flag := pdfaConvertCmd.Flags().Lookup("level")
	if flag == nil {
		t.Fatal("pdfa convert should have --level flag")
	}
	if flag.DefValue != "2b" {
		t.Errorf("pdfa convert --level default = %q, want %q", flag.DefValue, "2b")
	}
}

func TestPdfaCommandDescriptions(t *testing.T) {
	if pdfaCmd.Short == "" {
		t.Error("pdfa command should have a short description")
	}
	if pdfaCmd.Long == "" {
		t.Error("pdfa command should have a long description")
	}
	if pdfaValidateCmd.Short == "" {
		t.Error("pdfa validate command should have a short description")
	}
	if pdfaConvertCmd.Short == "" {
		t.Error("pdfa convert command should have a short description")
	}
}
