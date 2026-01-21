package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

// TestDryRunMode tests dry-run mode for various commands
// These tests verify that the dry-run mode executes the code paths without making changes

func TestCompressDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"compress", samplePDF(), "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("compress --dry-run failed: %v", err)
	}
}

func TestCompressDryRun_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"compress", samplePDF(), samplePDF(), "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("compress --dry-run with multiple files failed: %v", err)
	}
}

func TestCompressDryRun_WithOutput(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"compress", samplePDF(), "-o", "/tmp/test.pdf", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("compress --dry-run with output failed: %v", err)
	}
}

func TestDecryptDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"decrypt", samplePDF(), "--password", "test", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Decrypt on unencrypted file will report "unable to read info" path
	_ = rootCmd.Execute()
}

func TestDecryptDryRun_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"decrypt", samplePDF(), samplePDF(), "--password", "test", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	_ = rootCmd.Execute()
}

func TestEncryptDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"encrypt", samplePDF(), "--password", "test", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("encrypt --dry-run failed: %v", err)
	}
}

func TestEncryptDryRun_WithOwnerPassword(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"encrypt", samplePDF(), "--password", "user", "--owner-password", "owner", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("encrypt --dry-run with owner password failed: %v", err)
	}
}

func TestEncryptDryRun_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"encrypt", samplePDF(), samplePDF(), "--password", "test", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("encrypt --dry-run with multiple files failed: %v", err)
	}
}

func TestRotateDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"rotate", samplePDF(), "-a", "90", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rotate --dry-run failed: %v", err)
	}
}

func TestRotateDryRun_WithPages(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"rotate", samplePDF(), "-a", "180", "-p", "1", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rotate --dry-run with pages failed: %v", err)
	}
}

func TestRotateDryRun_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"rotate", samplePDF(), samplePDF(), "-a", "90", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rotate --dry-run with multiple files failed: %v", err)
	}
}

func TestWatermarkDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"watermark", samplePDF(), "-t", "DRAFT", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("watermark --dry-run failed: %v", err)
	}
}

func TestWatermarkDryRun_WithPages(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"watermark", samplePDF(), "-t", "CONFIDENTIAL", "-p", "1", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("watermark --dry-run with pages failed: %v", err)
	}
}

func TestWatermarkDryRun_ImageWatermark(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"watermark", samplePDF(), "-i", testImage, "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("watermark --dry-run with image failed: %v", err)
	}
}

func TestExtractDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"extract", samplePDF(), "-p", "1", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("extract --dry-run failed: %v", err)
	}
}

func TestExtractDryRun_WithOutput(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"extract", samplePDF(), "-p", "1-2", "-o", "/tmp/extracted.pdf", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("extract --dry-run with output failed: %v", err)
	}
}

func TestReorderDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"reorder", samplePDF(), "-s", "3,2,1", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("reorder --dry-run failed: %v", err)
	}
}

func TestReorderDryRun_InvalidSequence(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"reorder", samplePDF(), "-s", "invalid", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should still execute (dry-run shows what would happen, even with invalid input)
	_ = rootCmd.Execute()
}

func TestMergeDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"merge", samplePDF(), samplePDF(), "-o", "/tmp/merged.pdf", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("merge --dry-run failed: %v", err)
	}
}

func TestSplitDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"split", samplePDF(), "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("split --dry-run failed: %v", err)
	}
}

func TestSplitDryRun_WithPagesPerFile(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"split", samplePDF(), "-n", "2", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("split --dry-run with pages per file failed: %v", err)
	}
}

func TestMetaSetDryRun(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), "--title", "Test Title", "--dry-run"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta set --dry-run failed: %v", err)
	}
}

func TestMetaSetDryRun_AllFields(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{
		"meta", samplePDF(),
		"--title", "Test Title",
		"--author", "Test Author",
		"--subject", "Test Subject",
		"--keywords", "test,keywords",
		"--creator", "Test Creator",
		"--dry-run",
	})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta set --dry-run with all fields failed: %v", err)
	}
}
