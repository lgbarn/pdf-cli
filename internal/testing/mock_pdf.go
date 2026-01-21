package testing

import "errors"

// MockPDFOps provides mock PDF operations for testing commands
// without requiring actual PDF files.
type MockPDFOps struct {
	// PageCountResult is returned by PageCount
	PageCountResult int
	// PageCountError is returned by PageCount if set
	PageCountError error
	// CompressError is returned by Compress if set
	CompressError error
	// MergeError is returned by Merge if set
	MergeError error
	// SplitError is returned by Split if set
	SplitError error
	// ExtractTextResult is returned by ExtractText
	ExtractTextResult string
	// ExtractTextError is returned by ExtractText if set
	ExtractTextError error
	// Calls tracks which methods were called
	Calls []string
}

// NewMockPDFOps creates a MockPDFOps with sensible defaults.
func NewMockPDFOps() *MockPDFOps {
	return &MockPDFOps{
		PageCountResult:   10,
		ExtractTextResult: "Sample extracted text",
		Calls:             make([]string, 0),
	}
}

// PageCount mocks pdf.PageCount.
func (m *MockPDFOps) PageCount(file, _ string) (int, error) {
	m.Calls = append(m.Calls, "PageCount:"+file)
	if m.PageCountError != nil {
		return 0, m.PageCountError
	}
	return m.PageCountResult, nil
}

// Compress mocks pdf.Compress.
func (m *MockPDFOps) Compress(input, output, _ string) error {
	m.Calls = append(m.Calls, "Compress:"+input+"->"+output)
	return m.CompressError
}

// Merge mocks pdf.Merge.
func (m *MockPDFOps) Merge(_ []string, output, _ string) error {
	m.Calls = append(m.Calls, "Merge:"+output)
	return m.MergeError
}

// Split mocks pdf.Split.
func (m *MockPDFOps) Split(input, _, _ string) error {
	m.Calls = append(m.Calls, "Split:"+input)
	return m.SplitError
}

// ExtractText mocks pdf.ExtractText.
func (m *MockPDFOps) ExtractText(input, _ string, _ []int) (string, error) {
	m.Calls = append(m.Calls, "ExtractText:"+input)
	if m.ExtractTextError != nil {
		return "", m.ExtractTextError
	}
	return m.ExtractTextResult, nil
}

// Reset clears the call history.
func (m *MockPDFOps) Reset() {
	m.Calls = make([]string, 0)
}

// AssertCalled checks if a method was called with the given prefix.
func (m *MockPDFOps) AssertCalled(prefix string) bool {
	for _, call := range m.Calls {
		if len(call) >= len(prefix) && call[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// ErrMockPasswordRequired is a mock error for password-protected files.
var ErrMockPasswordRequired = errors.New("mock: password required")

// ErrMockCorrupted is a mock error for corrupted files.
var ErrMockCorrupted = errors.New("mock: file corrupted")
