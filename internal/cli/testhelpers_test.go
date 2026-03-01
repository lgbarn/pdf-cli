package cli

import (
	"io"
	"os"
	"testing"
)

// captureStderr runs fn and returns everything written to os.Stderr during its execution.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	fn()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = oldStderr
	return string(out)
}
