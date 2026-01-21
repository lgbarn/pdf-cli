package testing

import "context"

// MockOCRBackend provides a mock OCR backend for testing.
type MockOCRBackend struct {
	// NameResult is returned by Name
	NameResult string
	// AvailableResult is returned by Available
	AvailableResult bool
	// ProcessImageResult is returned by ProcessImage
	ProcessImageResult string
	// ProcessImageError is returned by ProcessImage if set
	ProcessImageError error
	// Calls tracks which methods were called
	Calls []string
}

// NewMockOCRBackend creates a MockOCRBackend with sensible defaults.
func NewMockOCRBackend() *MockOCRBackend {
	return &MockOCRBackend{
		NameResult:         "mock",
		AvailableResult:    true,
		ProcessImageResult: "Mock OCR extracted text",
		Calls:              make([]string, 0),
	}
}

// Name returns the backend name.
func (m *MockOCRBackend) Name() string {
	m.Calls = append(m.Calls, "Name")
	return m.NameResult
}

// Available returns whether the backend is available.
func (m *MockOCRBackend) Available() bool {
	m.Calls = append(m.Calls, "Available")
	return m.AvailableResult
}

// ProcessImage mocks OCR processing.
func (m *MockOCRBackend) ProcessImage(_ context.Context, imagePath, _ string) (string, error) {
	m.Calls = append(m.Calls, "ProcessImage:"+imagePath)
	if m.ProcessImageError != nil {
		return "", m.ProcessImageError
	}
	return m.ProcessImageResult, nil
}

// Close mocks backend cleanup.
func (m *MockOCRBackend) Close() error {
	m.Calls = append(m.Calls, "Close")
	return nil
}

// Reset clears the call history.
func (m *MockOCRBackend) Reset() {
	m.Calls = make([]string, 0)
}
