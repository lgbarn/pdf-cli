// Package cleanup provides a thread-safe registry for temporary file paths
// that should be removed on program exit or signal interruption.
package cleanup

import (
	"os"
	"sync"
)

var (
	mu     sync.Mutex
	paths  []string
	hasRun bool
)

// Register adds a path to the cleanup registry and returns an unregister
// function. Call the returned function when the path has been cleaned up
// normally (e.g., via defer os.Remove) so that Run does not attempt to
// remove it again.
func Register(path string) func() {
	mu.Lock()
	defer mu.Unlock()

	idx := len(paths)
	paths = append(paths, path)

	return func() {
		mu.Lock()
		defer mu.Unlock()
		if idx < len(paths) {
			paths[idx] = "" // mark as unregistered
		}
	}
}

// Run removes all registered paths in reverse order (LIFO). It is
// idempotent: subsequent calls after the first are no-ops.
func Run() error {
	mu.Lock()
	defer mu.Unlock()

	if hasRun {
		return nil
	}
	hasRun = true

	var firstErr error
	for i := len(paths) - 1; i >= 0; i-- {
		p := paths[i]
		if p == "" {
			continue
		}
		if err := os.RemoveAll(p); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	paths = nil
	return firstErr
}

// Reset clears all registered paths without removing them. Intended for
// use in tests only.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	paths = nil
	hasRun = false
}
