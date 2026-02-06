package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ReadPassword reads a password securely from multiple sources with priority:
// 1. --password-file flag (if present)
// 2. PDF_CLI_PASSWORD environment variable (if set)
// 3. --password flag (deprecated, shows warning)
// 4. Interactive terminal prompt (if terminal and not CI/batch mode)
// Returns empty string if no password source available.
func ReadPassword(cmd *cobra.Command, promptMsg string) (string, error) {
	// 1. Check --password-file flag
	if cmd.Flags().Lookup("password-file") != nil {
		passwordFile, _ := cmd.Flags().GetString("password-file")
		if passwordFile != "" {
			// Sanitize password file path against directory traversal
			for _, part := range strings.Split(passwordFile, "/") {
				if part == ".." {
					return "", fmt.Errorf("invalid password file path: contains directory traversal")
				}
			}
			passwordFile = filepath.Clean(passwordFile)
			data, err := os.ReadFile(passwordFile) // #nosec G304 -- path sanitized above
			if err != nil {
				return "", fmt.Errorf("failed to read password file: %w", err)
			}
			if len(data) > 1024 {
				return "", fmt.Errorf("password file exceeds 1KB size limit")
			}
			return strings.TrimSpace(string(data)), nil
		}
	}

	// 2. Check PDF_CLI_PASSWORD env var
	if envPass := os.Getenv("PDF_CLI_PASSWORD"); envPass != "" {
		return envPass, nil
	}

	// 3. Check --password flag (insecure, requires opt-in)
	if cmd.Flags().Lookup("password") != nil {
		password, _ := cmd.Flags().GetString("password")
		if password != "" {
			// Check if opt-in flag is present and set
			allowInsecure := false
			if cmd.Flags().Lookup("allow-insecure-password") != nil {
				allowInsecure, _ = cmd.Flags().GetBool("allow-insecure-password")
			}

			if !allowInsecure {
				return "", fmt.Errorf(`--password flag is insecure and disabled by default.
Use one of these secure alternatives:
  1. --password-file <path>        (recommended for automation)
  2. PDF_CLI_PASSWORD env var      (recommended for CI/scripts)
  3. Interactive prompt            (recommended for manual use)

To use --password anyway (not recommended), add --allow-insecure-password`)
			}

			fmt.Fprintln(os.Stderr, "WARNING: --password flag exposes passwords in process listings. Use --password-file, PDF_CLI_PASSWORD, or interactive prompt instead.")
			return password, nil
		}
	}

	// 4. Interactive terminal prompt
	if promptMsg != "" && isInteractiveTerminal() {
		fmt.Fprint(os.Stderr, promptMsg)
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // newline after password input
		if err != nil {
			return "", fmt.Errorf("failed to read password from terminal: %w", err)
		}
		return string(passwordBytes), nil
	}

	return "", nil
}

// isInteractiveTerminal returns true if stdin is an interactive terminal and not in CI/batch mode.
func isInteractiveTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && os.Getenv("CI") == "" && os.Getenv("PDF_CLI_BATCH") == ""
}
