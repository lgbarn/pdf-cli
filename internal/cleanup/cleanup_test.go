package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func setup(t *testing.T) {
	t.Helper()
	Reset()
	t.Cleanup(func() { Reset() })
}

func TestRegisterAndRun(t *testing.T) {
	setup(t)

	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.tmp")
	f2 := filepath.Join(dir, "b.tmp")

	if err := os.WriteFile(f1, []byte("a"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("b"), 0600); err != nil {
		t.Fatal(err)
	}

	Register(f1)
	Register(f2)

	if err := Run(); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Errorf("expected f1 to be removed, got err=%v", err)
	}
	if _, err := os.Stat(f2); !os.IsNotExist(err) {
		t.Errorf("expected f2 to be removed, got err=%v", err)
	}
}

func TestUnregister(t *testing.T) {
	setup(t)

	dir := t.TempDir()
	f1 := filepath.Join(dir, "keep.tmp")

	if err := os.WriteFile(f1, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}

	unregister := Register(f1)
	unregister() // should mark as skipped

	if err := Run(); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	// File should still exist because we unregistered it.
	if _, err := os.Stat(f1); err != nil {
		t.Errorf("expected f1 to still exist, got err=%v", err)
	}
}

func TestConcurrentRegister(t *testing.T) {
	setup(t)

	dir := t.TempDir()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			p := filepath.Join(dir, fmt.Sprintf("file-%d.tmp", n))
			if err := os.WriteFile(p, []byte("x"), 0600); err != nil {
				return
			}
			Register(p)
		}(i)
	}
	wg.Wait()

	if err := Run(); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected dir to be empty after Run, found %d entries", len(entries))
	}
}

func TestRunIdempotent(t *testing.T) {
	setup(t)

	dir := t.TempDir()
	f1 := filepath.Join(dir, "idem.tmp")

	if err := os.WriteFile(f1, []byte("idem"), 0600); err != nil {
		t.Fatal(err)
	}

	Register(f1)

	if err := Run(); err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}
	if err := Run(); err != nil {
		t.Fatalf("second Run returned error: %v", err)
	}

	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Errorf("expected f1 to be removed, got err=%v", err)
	}
}
