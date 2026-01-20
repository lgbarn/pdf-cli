package cli

import (
	"github.com/spf13/cobra"
)

// AddOutputFlag adds the -o/--output flag to a command
func AddOutputFlag(cmd *cobra.Command, usage string) {
	if usage == "" {
		usage = "Output file path"
	}
	cmd.Flags().StringP("output", "o", "", usage)
}

// AddPagesFlag adds the -p/--pages flag to a command
func AddPagesFlag(cmd *cobra.Command, usage string) {
	if usage == "" {
		usage = "Page range (e.g., 1-5,7,10-12)"
	}
	cmd.Flags().StringP("pages", "p", "", usage)
}

// AddPasswordFlag adds the --password flag to a command
func AddPasswordFlag(cmd *cobra.Command, usage string) {
	if usage == "" {
		usage = "Password for encryption/decryption"
	}
	cmd.Flags().String("password", "", usage)
}

// GetOutput gets the output flag value
func GetOutput(cmd *cobra.Command) string {
	output, _ := cmd.Flags().GetString("output")
	return output
}

// GetPages gets the pages flag value
func GetPages(cmd *cobra.Command) string {
	pages, _ := cmd.Flags().GetString("pages")
	return pages
}

// GetPassword gets the password flag value
func GetPassword(cmd *cobra.Command) string {
	password, _ := cmd.Flags().GetString("password")
	return password
}

// AddFormatFlag adds the --format flag to a command for structured output.
func AddFormatFlag(cmd *cobra.Command) {
	cmd.Flags().String("format", "", "Output format: json, csv, tsv (default: human-readable)")
}

// GetFormat gets the format flag value.
func GetFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("format")
	return format
}

// AddStdoutFlag adds the --stdout flag to a command for binary stdout output.
func AddStdoutFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("stdout", false, "Write binary output to stdout")
}

// GetStdout gets the stdout flag value.
func GetStdout(cmd *cobra.Command) bool {
	stdout, _ := cmd.Flags().GetBool("stdout")
	return stdout
}
