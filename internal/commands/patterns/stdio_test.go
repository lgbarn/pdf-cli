package patterns

import (
	"os"
	"strings"
	"testing"
)

func TestStdioHandler_Setup_WithFile(t *testing.T) {
	// Create a temp file to use as input
	tmpFile, err := os.CreateTemp("", "test-input-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	handler := &StdioHandler{
		InputArg:       tmpFile.Name(),
		ExplicitOutput: "",
		ToStdout:       false,
		DefaultSuffix:  "_processed",
		Operation:      "test",
	}
	defer handler.Cleanup()

	input, output, err := handler.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if input != tmpFile.Name() {
		t.Errorf("Expected input %s, got %s", tmpFile.Name(), input)
	}

	expectedSuffix := "_processed.pdf"
	if !strings.HasSuffix(output, expectedSuffix) {
		t.Errorf("Expected output to have suffix %s, got %s", expectedSuffix, output)
	}
}

func TestStdioHandler_Setup_WithExplicitOutput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-input-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	handler := &StdioHandler{
		InputArg:       tmpFile.Name(),
		ExplicitOutput: "/tmp/explicit-output.pdf",
		ToStdout:       false,
		DefaultSuffix:  "_processed",
		Operation:      "test",
	}
	defer handler.Cleanup()

	_, output, err := handler.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if output != "/tmp/explicit-output.pdf" {
		t.Errorf("Expected output /tmp/explicit-output.pdf, got %s", output)
	}
}

func TestStdioHandler_Setup_WithStdout(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-input-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	handler := &StdioHandler{
		InputArg:       tmpFile.Name(),
		ExplicitOutput: "",
		ToStdout:       true,
		DefaultSuffix:  "_processed",
		Operation:      "compress",
	}
	defer handler.Cleanup()

	_, output, err := handler.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Should create a temp file for stdout
	if !strings.Contains(output, "pdf-cli-compress") {
		t.Errorf("Expected temp output file, got %s", output)
	}

	// Temp file should exist
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Temp output file should exist")
	}
}

func TestStdioHandler_Cleanup(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-input-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	handler := &StdioHandler{
		InputArg:  tmpFile.Name(),
		ToStdout:  true,
		Operation: "test",
	}

	_, output, err := handler.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Temp file should exist
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Fatal("Temp output file should exist before cleanup")
	}

	handler.Cleanup()

	// Temp file should be removed
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Error("Temp output file should be removed after cleanup")
	}
}

func TestStdioHandler_OutputPath(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-input-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	handler := &StdioHandler{
		InputArg:       tmpFile.Name(),
		ExplicitOutput: "/custom/path.pdf",
		Operation:      "test",
	}
	defer handler.Cleanup()

	handler.Setup()

	if handler.OutputPath() != "/custom/path.pdf" {
		t.Errorf("Expected /custom/path.pdf, got %s", handler.OutputPath())
	}
}
