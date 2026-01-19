package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// SetVersion sets the version info from build flags
func SetVersion(v, c, d string) {
	version = v
	commit = c
	buildDate = d
}

var rootCmd = &cobra.Command{
	Use:   "pdf",
	Short: "A powerful CLI tool for PDF manipulation",
	Long: `pdf-cli is a fast, single-binary CLI tool for common PDF operations.

It supports merging, splitting, extracting pages, rotating, compressing,
encrypting, decrypting, extracting text/images, managing metadata, and
adding watermarks.

Examples:
  pdf info document.pdf
  pdf merge -o combined.pdf file1.pdf file2.pdf
  pdf split input.pdf -o output/
  pdf extract input.pdf -p 1-5,10 -o selected.pdf
  pdf rotate input.pdf -a 90 -o rotated.pdf
  pdf compress input.pdf -o smaller.pdf
  pdf encrypt input.pdf -o secure.pdf --password secret
  pdf decrypt secure.pdf -o unlocked.pdf --password secret
  pdf text document.pdf
  pdf images document.pdf -o images/
  pdf meta document.pdf
  pdf watermark input.pdf -t "CONFIDENTIAL" -o marked.pdf`,
	Version: version,
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("pdf-cli version %s\ncommit: %s\nbuilt: %s\n", version, commit, buildDate))
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("force", "f", false, "Overwrite existing files without prompting")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// GetRootCmd returns the root command for testing
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// AddCommand adds a subcommand to the root command
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

// Verbose returns whether verbose mode is enabled
func Verbose() bool {
	v, _ := rootCmd.PersistentFlags().GetBool("verbose")
	return v
}

// Force returns whether force mode is enabled
func Force() bool {
	f, _ := rootCmd.PersistentFlags().GetBool("force")
	return f
}

// PrintVerbose prints a message if verbose mode is enabled
func PrintVerbose(format string, args ...interface{}) {
	if Verbose() {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// PrintStatus prints a status message to stderr (always shown)
func PrintStatus(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintProgress prints a progress message (shown in verbose mode or for long operations)
func PrintProgress(operation string) {
	if Verbose() {
		fmt.Fprintf(os.Stderr, "Processing: %s...\n", operation)
	}
}
