package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

// testdataDir returns the path to the testdata directory
func testdataDir() string {
	return filepath.Join("..", "..", "testdata")
}

// samplePDF returns the path to the sample PDF file
func samplePDF() string {
	return filepath.Join(testdataDir(), "sample.pdf")
}

func TestPagesToStrings(t *testing.T) {
	tests := []struct {
		name  string
		pages []int
		want  []string
	}{
		{
			name:  "empty slice",
			pages: []int{},
			want:  nil,
		},
		{
			name:  "nil slice",
			pages: nil,
			want:  nil,
		},
		{
			name:  "single page",
			pages: []int{1},
			want:  []string{"1"},
		},
		{
			name:  "multiple pages",
			pages: []int{1, 3, 5, 10},
			want:  []string{"1", "3", "5", "10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pagesToStrings(tt.pages)
			if len(got) != len(tt.want) {
				t.Errorf("pagesToStrings() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("pagesToStrings()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "no password",
			password: "",
		},
		{
			name:     "with password",
			password: "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := NewConfig(tt.password)
			if conf == nil {
				t.Fatal("newConfig() returned nil")
			}
			if tt.password != "" {
				if conf.UserPW != tt.password {
					t.Errorf("newConfig() UserPW = %v, want %v", conf.UserPW, tt.password)
				}
				if conf.OwnerPW != tt.password {
					t.Errorf("newConfig() OwnerPW = %v, want %v", conf.OwnerPW, tt.password)
				}
			}
		})
	}
}

func TestGetInfo(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	info, err := GetInfo(pdf, "")
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}

	if info == nil {
		t.Fatal("GetInfo() returned nil")
		return
	}

	if info.Pages < 1 {
		t.Errorf("GetInfo() Pages = %d, want >= 1", info.Pages)
	}

	if info.FileSize <= 0 {
		t.Errorf("GetInfo() FileSize = %d, want > 0", info.FileSize)
	}

	if info.Version == "" {
		t.Error("GetInfo() Version is empty")
	}
}

func TestGetInfoNonExistent(t *testing.T) {
	_, err := GetInfo("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("GetInfo() expected error for non-existent file")
	}
}

func TestPageCount(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	count, err := PageCount(pdf, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}

	if count < 1 {
		t.Errorf("PageCount() = %d, want >= 1", count)
	}
}

func TestPageCountNonExistent(t *testing.T) {
	_, err := PageCount("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("PageCount() expected error for non-existent file")
	}
}

func TestValidate(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	err := Validate(pdf, "")
	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestValidateNonExistent(t *testing.T) {
	err := Validate("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("Validate() expected error for non-existent file")
	}
}

func TestValidateValid(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	err := Validate(pdfFile, "")
	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestValidateToBuffer(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	data, err := os.ReadFile(pdfFile)
	if err != nil {
		t.Fatalf("Failed to read sample PDF: %v", err)
	}

	err = ValidateToBuffer(data)
	if err != nil {
		t.Errorf("ValidateToBuffer() error = %v", err)
	}
}

func TestValidateToBufferInvalid(t *testing.T) {
	// Invalid PDF data
	invalidData := []byte("This is not a valid PDF")
	err := ValidateToBuffer(invalidData)
	if err == nil {
		t.Error("ValidateToBuffer() expected error for invalid data")
	}
}

func TestValidateToBufferEmpty(t *testing.T) {
	err := ValidateToBuffer([]byte{})
	if err == nil {
		t.Error("ValidateToBuffer() expected error for empty data")
	}
}
