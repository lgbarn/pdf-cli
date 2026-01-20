package commands

import (
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

func TestReorderCommandExists(t *testing.T) {
	rootCmd := cli.GetRootCmd()
	cmd, _, err := rootCmd.Find([]string{"reorder"})
	if err != nil {
		t.Fatalf("Failed to find reorder command: %v", err)
	}
	if cmd == nil {
		t.Fatal("reorder command is nil")
	}
	if cmd.Use != "reorder <file.pdf>" {
		t.Errorf("reorder command Use = %q, want %q", cmd.Use, "reorder <file.pdf>")
	}
}

func TestReorderSequenceFlag(t *testing.T) {
	flag := reorderCmd.Flags().Lookup("sequence")
	if flag == nil {
		t.Fatal("reorder should have --sequence flag")
	}
	if flag.Shorthand != "s" {
		t.Errorf("--sequence shorthand = %q, want %q", flag.Shorthand, "s")
	}
}

func TestReorderOutputFlag(t *testing.T) {
	flag := reorderCmd.Flags().Lookup("output")
	if flag == nil {
		t.Error("reorder should have --output flag")
	}
}

func TestReorderPasswordFlag(t *testing.T) {
	flag := reorderCmd.Flags().Lookup("password")
	if flag == nil {
		t.Error("reorder should have --password flag")
	}
}

func TestReorderRequiredFlags(t *testing.T) {
	// The sequence flag should be required
	flag := reorderCmd.Flags().Lookup("sequence")
	if flag == nil {
		t.Fatal("reorder should have --sequence flag")
	}

	// Check if it's marked as required by checking the annotations
	required, ok := flag.Annotations["cobra_annotation_bash_completion_one_required_flag"]
	if !ok || len(required) == 0 {
		t.Error("--sequence flag should be marked as required")
	}
}

func TestReorderCommandDescription(t *testing.T) {
	if reorderCmd.Short == "" {
		t.Error("reorder command should have a short description")
	}
	if reorderCmd.Long == "" {
		t.Error("reorder command should have a long description")
	}
}
