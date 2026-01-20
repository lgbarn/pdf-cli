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

func TestReorderFlags(t *testing.T) {
	t.Run("sequence flag", func(t *testing.T) {
		flag := reorderCmd.Flags().Lookup("sequence")
		if flag == nil {
			t.Fatal("reorder should have --sequence flag")
		}
		if flag.Shorthand != "s" {
			t.Errorf("--sequence shorthand = %q, want %q", flag.Shorthand, "s")
		}

		// Check if marked as required
		required, ok := flag.Annotations["cobra_annotation_bash_completion_one_required_flag"]
		if !ok || len(required) == 0 {
			t.Error("--sequence flag should be marked as required")
		}
	})

	t.Run("output flag", func(t *testing.T) {
		if flag := reorderCmd.Flags().Lookup("output"); flag == nil {
			t.Error("reorder should have --output flag")
		}
	})

	t.Run("password flag", func(t *testing.T) {
		if flag := reorderCmd.Flags().Lookup("password"); flag == nil {
			t.Error("reorder should have --password flag")
		}
	})
}

func TestReorderCommandDescription(t *testing.T) {
	if reorderCmd.Short == "" {
		t.Error("reorder command should have a short description")
	}
	if reorderCmd.Long == "" {
		t.Error("reorder command should have a long description")
	}
}
