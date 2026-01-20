package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddFlags(t *testing.T) {
	tests := []struct {
		name        string
		addFlag     func(*cobra.Command, string)
		flagName    string
		shorthand   string
		customUsage string
		wantUsage   string
		wantDefault string
	}{
		{
			name:        "output flag default usage",
			addFlag:     AddOutputFlag,
			flagName:    "output",
			shorthand:   "o",
			customUsage: "",
			wantUsage:   "Output file path",
			wantDefault: "",
		},
		{
			name:        "output flag custom usage",
			addFlag:     AddOutputFlag,
			flagName:    "output",
			shorthand:   "o",
			customUsage: "Custom output path",
			wantUsage:   "Custom output path",
			wantDefault: "",
		},
		{
			name:        "pages flag default usage",
			addFlag:     AddPagesFlag,
			flagName:    "pages",
			shorthand:   "p",
			customUsage: "",
			wantUsage:   "Page range (e.g., 1-5,7,10-12)",
			wantDefault: "",
		},
		{
			name:        "pages flag custom usage",
			addFlag:     AddPagesFlag,
			flagName:    "pages",
			shorthand:   "p",
			customUsage: "Pages to process",
			wantUsage:   "Pages to process",
			wantDefault: "",
		},
		{
			name:        "password flag default usage",
			addFlag:     AddPasswordFlag,
			flagName:    "password",
			shorthand:   "",
			customUsage: "",
			wantUsage:   "Password for encryption/decryption",
			wantDefault: "",
		},
		{
			name:        "password flag custom usage",
			addFlag:     AddPasswordFlag,
			flagName:    "password",
			shorthand:   "",
			customUsage: "PDF password",
			wantUsage:   "PDF password",
			wantDefault: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			tt.addFlag(cmd, tt.customUsage)

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.flagName)
			}
			if flag.Usage != tt.wantUsage {
				t.Errorf("usage = %q, want %q", flag.Usage, tt.wantUsage)
			}
			if flag.DefValue != tt.wantDefault {
				t.Errorf("default = %q, want %q", flag.DefValue, tt.wantDefault)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("shorthand = %q, want %q", flag.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestAddFormatFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddFormatFlag(cmd)

	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("format flag not found")
	}
	if flag.DefValue != "" {
		t.Errorf("default = %q, want empty", flag.DefValue)
	}
	if flag.Usage != "Output format: json, csv, tsv (default: human-readable)" {
		t.Errorf("usage = %q", flag.Usage)
	}
}

func TestAddStdoutFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddStdoutFlag(cmd)

	flag := cmd.Flags().Lookup("stdout")
	if flag == nil {
		t.Fatal("stdout flag not found")
	}
	if flag.DefValue != "false" {
		t.Errorf("default = %q, want %q", flag.DefValue, "false")
	}
	if flag.Usage != "Write binary output to stdout" {
		t.Errorf("usage = %q", flag.Usage)
	}
}

func TestGetStringFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		addFlag  func(*cobra.Command)
		getFlag  func(*cobra.Command) string
		value    string
	}{
		{"output empty", "output", func(c *cobra.Command) { AddOutputFlag(c, "") }, GetOutput, ""},
		{"output set", "output", func(c *cobra.Command) { AddOutputFlag(c, "") }, GetOutput, "out.pdf"},
		{"output path", "output", func(c *cobra.Command) { AddOutputFlag(c, "") }, GetOutput, "/path/to/out.pdf"},
		{"pages empty", "pages", func(c *cobra.Command) { AddPagesFlag(c, "") }, GetPages, ""},
		{"pages range", "pages", func(c *cobra.Command) { AddPagesFlag(c, "") }, GetPages, "1-3,5,7-10"},
		{"password empty", "password", func(c *cobra.Command) { AddPasswordFlag(c, "") }, GetPassword, ""},
		{"password special", "password", func(c *cobra.Command) { AddPasswordFlag(c, "") }, GetPassword, "p@ss!w0rd#$%"},
		{"format empty", "format", func(c *cobra.Command) { AddFormatFlag(c) }, GetFormat, ""},
		{"format json", "format", func(c *cobra.Command) { AddFormatFlag(c) }, GetFormat, "json"},
		{"format csv", "format", func(c *cobra.Command) { AddFormatFlag(c) }, GetFormat, "csv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			tt.addFlag(cmd)
			if tt.value != "" {
				if err := cmd.Flags().Set(tt.flagName, tt.value); err != nil {
					t.Fatalf("Failed to set flag: %v", err)
				}
			}
			if got := tt.getFlag(cmd); got != tt.value {
				t.Errorf("got %q, want %q", got, tt.value)
			}
		})
	}
}

func TestGetStdout(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"", false},
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			AddStdoutFlag(cmd)
			if tt.value != "" {
				if err := cmd.Flags().Set("stdout", tt.value); err != nil {
					t.Fatalf("Failed to set flag: %v", err)
				}
			}
			if got := GetStdout(cmd); got != tt.want {
				t.Errorf("GetStdout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlagsCombined(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddOutputFlag(cmd, "")
	AddPagesFlag(cmd, "")
	AddPasswordFlag(cmd, "")
	AddFormatFlag(cmd)
	AddStdoutFlag(cmd)

	flags := map[string]string{
		"output":   "out.pdf",
		"pages":    "1-5",
		"password": "secret",
		"format":   "json",
		"stdout":   "true",
	}
	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("Failed to set %s: %v", name, err)
		}
	}

	if got := GetOutput(cmd); got != "out.pdf" {
		t.Errorf("GetOutput() = %q, want %q", got, "out.pdf")
	}
	if got := GetPages(cmd); got != "1-5" {
		t.Errorf("GetPages() = %q, want %q", got, "1-5")
	}
	if got := GetPassword(cmd); got != "secret" {
		t.Errorf("GetPassword() = %q, want %q", got, "secret")
	}
	if got := GetFormat(cmd); got != "json" {
		t.Errorf("GetFormat() = %q, want %q", got, "json")
	}
	if !GetStdout(cmd) {
		t.Error("GetStdout() = false, want true")
	}
}
