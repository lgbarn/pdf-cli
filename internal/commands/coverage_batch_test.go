package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

func TestInfoCommand_BatchWithErrors(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Mix valid and invalid files to test error handling in batch mode
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), "/nonexistent/file.pdf"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should complete without fatal error even with invalid files
	_ = rootCmd.Execute()
}

func TestInfoCommand_BatchCSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), samplePDF(), "--format", "csv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info command with csv format failed: %v", err)
	}
}

func TestInfoCommand_BatchTSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), samplePDF(), "--format", "tsv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info command with tsv format failed: %v", err)
	}
}

func TestInfoCommand_BatchJSONFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), samplePDF(), "--format", "json"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info command batch with json format failed: %v", err)
	}
}
func TestMetaCommand_BatchWithErrors(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Mix valid and invalid files to test error handling in batch mode
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), "/nonexistent/file.pdf"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should complete without fatal error even with invalid files
	_ = rootCmd.Execute()
}

func TestMetaCommand_BatchCSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--format", "csv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta command with csv format failed: %v", err)
	}
}

func TestMetaCommand_BatchTSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--format", "tsv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta command with tsv format failed: %v", err)
	}
}

func TestMetaCommand_BatchJSONFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--format", "json"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta command batch with json format failed: %v", err)
	}
}

func TestMetaCommand_SetMultipleOnBatch(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Setting metadata on multiple files should fail
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--title", "Test"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("meta --title on multiple files should fail")
	}
}
func TestCompressCommand_BatchWithOutputFlag(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy sample.pdf twice
	pdf1 := filepath.Join(tmpDir, "test1.pdf")
	pdf2 := filepath.Join(tmpDir, "test2.pdf")
	input, _ := os.ReadFile(samplePDF())
	os.WriteFile(pdf1, input, 0644)
	os.WriteFile(pdf2, input, 0644)

	// Batch compress with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"compress", pdf1, pdf2, "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("compress batch with -o flag should fail")
	}
}

func TestEncryptCommand_BatchWithOutputFlag(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy sample.pdf twice
	pdf1 := filepath.Join(tmpDir, "test1.pdf")
	pdf2 := filepath.Join(tmpDir, "test2.pdf")
	input, _ := os.ReadFile(samplePDF())
	os.WriteFile(pdf1, input, 0644)
	os.WriteFile(pdf2, input, 0644)

	// Batch encrypt with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"encrypt", pdf1, pdf2, "--password", "test", "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("encrypt batch with -o flag should fail")
	}
}

func TestRotateCommand_BatchWithOutputFlag(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy sample.pdf twice
	pdf1 := filepath.Join(tmpDir, "test1.pdf")
	pdf2 := filepath.Join(tmpDir, "test2.pdf")
	input, _ := os.ReadFile(samplePDF())
	os.WriteFile(pdf1, input, 0644)
	os.WriteFile(pdf2, input, 0644)

	// Batch rotate with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"rotate", pdf1, pdf2, "-a", "90", "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("rotate batch with -o flag should fail")
	}
}

func TestWatermarkCommand_BatchWithOutputFlag(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy sample.pdf twice
	pdf1 := filepath.Join(tmpDir, "test1.pdf")
	pdf2 := filepath.Join(tmpDir, "test2.pdf")
	input, _ := os.ReadFile(samplePDF())
	os.WriteFile(pdf1, input, 0644)
	os.WriteFile(pdf2, input, 0644)

	// Batch watermark with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"watermark", pdf1, pdf2, "-t", "DRAFT", "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("watermark batch with -o flag should fail")
	}
}
