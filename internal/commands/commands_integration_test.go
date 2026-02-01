package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompressCommand_WithOutput(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "compressed.pdf")
	if err := executeCommand("compress", samplePDF(), "-o", output); err != nil {
		t.Fatalf("compress command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("compress did not create output file")
	}
}

func TestCompressCommand_ForceOverwrite(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "compressed.pdf")
	// Create existing file
	if err := os.WriteFile(output, []byte("existing"), 0600); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	if err := executeCommand("compress", samplePDF(), "-o", output, "-f"); err != nil {
		t.Fatalf("compress command with -f failed: %v", err)
	}

	// Verify it was overwritten (file size should be different)
	info, _ := os.Stat(output)
	if info.Size() == 8 { // "existing" is 8 bytes
		t.Error("compress did not overwrite existing file")
	}
}

func TestCompressCommand_MultipleFiles(t *testing.T) {
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

	// Batch compress
	if err := executeCommand("compress", pdf1, pdf2); err != nil {
		t.Fatalf("compress command with multiple files failed: %v", err)
	}
}

func TestRotateCommand_ValidAngles(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	for _, angle := range []string{"90", "180", "270"} {
		t.Run(angle+"_degrees", func(t *testing.T) {
			resetFlags(t)
			tmpDir, err := os.MkdirTemp("", "pdf-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			output := filepath.Join(tmpDir, "rotated.pdf")
			if err := executeCommand("rotate", samplePDF(), "-a", angle, "-o", output); err != nil {
				t.Fatalf("rotate command failed: %v", err)
			}
			if _, err := os.Stat(output); os.IsNotExist(err) {
				t.Error("rotate did not create output file")
			}
		})
	}
}

func TestRotateCommand_InvalidAngle(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = executeCommand("rotate", samplePDF(), "-a", "45", "-o", output)
	if err == nil {
		t.Error("rotate with invalid angle should fail")
	}
}

func TestRotateCommand_WithPageSelection(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	if err := executeCommand("rotate", samplePDF(), "-a", "90", "-p", "1", "-o", output); err != nil {
		t.Fatalf("rotate command with page selection failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("rotate did not create output file")
	}
}

func TestRotateCommand_MultipleFiles(t *testing.T) {
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

	// Batch rotate
	if err := executeCommand("rotate", pdf1, pdf2, "-a", "90"); err != nil {
		t.Fatalf("rotate command with multiple files failed: %v", err)
	}
}
