package ocr

import (
	"context"
	"errors"
	"sync/atomic"
)

// mockBackend is a test implementation of the Backend interface.
type mockBackend struct {
	name         string
	available    bool
	processOut   string
	processErr   error
	closeErr     error
	processCalls int32 // atomic counter for thread safety
}

func (m *mockBackend) Name() string {
	return m.name
}

func (m *mockBackend) Available() bool {
	return m.available
}

func (m *mockBackend) ProcessImage(ctx context.Context, imagePath, lang string) (string, error) {
	atomic.AddInt32(&m.processCalls, 1)
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return m.processOut, m.processErr
}

func (m *mockBackend) Close() error {
	return m.closeErr
}

// newMockBackend creates a new mock backend for testing.
func newMockBackend(name string, available bool) *mockBackend {
	return &mockBackend{
		name:      name,
		available: available,
	}
}

// withOutput sets the output text for ProcessImage.
func (m *mockBackend) withOutput(text string) *mockBackend {
	m.processOut = text
	return m
}

// withError sets the error for ProcessImage.
func (m *mockBackend) withError(err error) *mockBackend {
	m.processErr = err
	return m
}

// withCloseError sets the error for Close.
func (m *mockBackend) withCloseError(err error) *mockBackend {
	m.closeErr = err
	return m
}

// Common test errors
var (
	errTestProcess = errors.New("test process error")
	errTestClose   = errors.New("test close error")
)
