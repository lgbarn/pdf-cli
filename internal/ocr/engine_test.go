package ocr

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestEngineBackendName(t *testing.T) {
	tests := []struct {
		name    string
		backend Backend
		want    string
	}{
		{"with backend", newMockBackend("test", true), "test"},
		{"nil backend", nil, "none"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{backend: tt.backend}
			if got := e.BackendName(); got != tt.want {
				t.Errorf("Engine.BackendName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngineClose(t *testing.T) {
	tests := []struct {
		name    string
		backend Backend
		wantErr bool
	}{
		{"nil backend", nil, false},
		{"backend without error", newMockBackend("test", true), false},
		{"backend with close error", newMockBackend("test", true).withCloseError(errTestClose), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{backend: tt.backend}
			err := e.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Engine.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockBackend(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		m := newMockBackend("test", true).
			withOutput("extracted text").
			withError(nil)

		if m.Name() != "test" {
			t.Errorf("Name() = %q, want %q", m.Name(), "test")
		}
		if !m.Available() {
			t.Error("Available() = false, want true")
		}

		text, err := m.ProcessImage(context.Background(), "test.png", "eng")
		if err != nil {
			t.Errorf("ProcessImage() error = %v", err)
		}
		if text != "extracted text" {
			t.Errorf("ProcessImage() = %q, want %q", text, "extracted text")
		}
		if calls := atomic.LoadInt32(&m.processCalls); calls != 1 {
			t.Errorf("processCalls = %d, want 1", calls)
		}
	})

	t.Run("with error", func(t *testing.T) {
		m := newMockBackend("test", true).withError(errTestProcess)

		_, err := m.ProcessImage(context.Background(), "test.png", "eng")
		if err != errTestProcess {
			t.Errorf("ProcessImage() error = %v, want %v", err, errTestProcess)
		}
	})

	t.Run("unavailable backend", func(t *testing.T) {
		m := newMockBackend("test", false)
		if m.Available() {
			t.Error("Available() = true, want false")
		}
	})
}
