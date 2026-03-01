package cli

import (
	"github.com/lgbarn/pdf-cli/internal/logging"
	"github.com/spf13/cobra"
)

// Logging flag variables.
var (
	logLevel  string
	logFormat string
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

// AddPasswordFileFlag adds the --password-file flag to a command.
func AddPasswordFileFlag(cmd *cobra.Command, usage string) {
	if usage == "" {
		usage = "Read password from file (more secure than --password)"
	}
	cmd.Flags().String("password-file", "", usage)
}

// AddAllowInsecurePasswordFlag adds the --allow-insecure-password flag to a command.
func AddAllowInsecurePasswordFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("allow-insecure-password", false, "Allow use of insecure --password flag")
}

// GetAllowInsecurePassword gets the allow-insecure-password flag value.
func GetAllowInsecurePassword(cmd *cobra.Command) bool {
	allow, _ := cmd.Flags().GetBool("allow-insecure-password")
	return allow
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

// GetPasswordSecure reads password securely from multiple sources.
func GetPasswordSecure(cmd *cobra.Command, promptMsg string) (string, error) {
	return ReadPassword(cmd, promptMsg)
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

// AddLoggingFlags adds --log-level and --log-format persistent flags to a command.
func AddLoggingFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", string(logging.LevelError), "Log level (debug, info, warn, error, silent)")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", string(logging.FormatText), "Log format (text, json)")
}

// GetLogLevel returns the log level flag value.
func GetLogLevel() string {
	return logLevel
}

// GetLogFormat returns the log format flag value.
func GetLogFormat() string {
	return logFormat
}

// InitLogging initializes the logging system from flags.
func InitLogging() {
	logging.Init(logging.ParseLevel(logLevel), logging.ParseFormat(logFormat))
}
