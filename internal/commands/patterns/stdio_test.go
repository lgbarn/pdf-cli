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

func TestStdioHandler_InputPath(t *testing.T) {
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

	if handler.InputPath() != tmpFile.Name() {
		t.Errorf("Expected %s, got %s", tmpFile.Name(), handler.InputPath())
	}
}

func TestStdioHandler_Finalize_NoStdout(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-input-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	handler := &StdioHandler{
		InputArg:       tmpFile.Name(),
		ExplicitOutput: "/tmp/output.pdf",
		ToStdout:       false,
		Operation:      "test",
	}
	defer handler.Cleanup()

	_, _, err = handler.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Finalize should be a no-op when ToStdout is false
	err = handler.Finalize()
	if err != nil {
		t.Errorf("Finalize without stdout should not error: %v", err)
	}
}

func TestStdioHandler_Finalize_WithStdout(t *testing.T) {
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
	defer handler.Cleanup()

	_, output, err := handler.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Write some content to the temp output file
	if err := os.WriteFile(output, []byte("test content"), 0600); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}

	// Finalize should write to stdout (which we can't easily capture in a test)
	// But we can verify it doesn't panic or error
	err = handler.Finalize()
	if err != nil {
		t.Errorf("Finalize with stdout returned error: %v", err)
	}
}

func TestStdioHandler_Cleanup_MultipleCalls(t *testing.T) {
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

	_, _, err = handler.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Call Cleanup multiple times - should be safe
	handler.Cleanup()
	handler.Cleanup()
	handler.Cleanup()
	// No panic = pass
}

func TestStdioHandler_Setup_NonexistentInput(t *testing.T) {
	// Note: StdioHandler.Setup() does not validate that the input file exists
	// It only resolves the path. File validation happens in the command handler.
	// This test verifies the handler correctly returns the nonexistent path.
	handler := &StdioHandler{
		InputArg:  "/nonexistent/file.pdf",
		ToStdout:  false,
		Operation: "test",
	}
	defer handler.Cleanup()

	input, _, err := handler.Setup()
	if err != nil {
		t.Errorf("Setup should not fail for nonexistent file: %v", err)
	}
	if input != "/nonexistent/file.pdf" {
		t.Errorf("Expected input path to be /nonexistent/file.pdf, got %s", input)
	}
}
