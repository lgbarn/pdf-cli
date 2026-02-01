package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRotate(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = Rotate(pdf, output, 90, nil, "")
	if err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Rotate() did not create output file")
	}
}

func TestRotateWithPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = Rotate(pdfFile, output, 90, []int{1}, "")
	if err != nil {
		t.Fatalf("Rotate() with pages error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Rotate() with pages did not create output file")
	}
}

func TestRotateAllAngles(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	angles := []int{0, 90, 180, 270}

	for _, angle := range angles {
		t.Run(fmt.Sprintf("%d_degrees", angle), func(t *testing.T) {
			output := filepath.Join(tmpDir, fmt.Sprintf("rotated_%d.pdf", angle))
			err := Rotate(pdfFile, output, angle, nil, "")
			if err != nil {
				t.Fatalf("Rotate() angle=%d error = %v", angle, err)
			}
			if _, err := os.Stat(output); os.IsNotExist(err) {
				t.Errorf("Rotate() did not create output file")
			}
		})
	}
}

func TestRotateNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = Rotate("/nonexistent/file.pdf", output, 90, nil, "")
	if err == nil {
		t.Error("Rotate() expected error for non-existent file")
	}
}

func TestCompress(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "compressed.pdf")
	err = Compress(pdf, output, "")
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Compress() did not create output file")
	}
}

func TestCompressNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "compressed.pdf")
	err = Compress("/nonexistent/file.pdf", output, "")
	if err == nil {
		t.Error("Compress() expected error for non-existent file")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Encrypt
	encrypted := filepath.Join(tmpDir, "encrypted.pdf")
	password := "testpassword123"
	err = Encrypt(pdf, encrypted, password, "")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Verify encrypted file exists
	if _, err := os.Stat(encrypted); os.IsNotExist(err) {
		t.Fatal("Encrypt() did not create output file")
	}

	// Decrypt
	decrypted := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt(encrypted, decrypted, password)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Verify decrypted file exists
	if _, err := os.Stat(decrypted); os.IsNotExist(err) {
		t.Error("Decrypt() did not create output file")
	}
}

func TestEncryptNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "encrypted.pdf")
	err = Encrypt("/nonexistent/file.pdf", output, "password", "")
	if err == nil {
		t.Error("Encrypt() expected error for non-existent file")
	}
}

func TestDecryptNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt("/nonexistent/file.pdf", output, "password")
	if err == nil {
		t.Error("Decrypt() expected error for non-existent file")
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First encrypt the file
	encrypted := filepath.Join(tmpDir, "encrypted.pdf")
	err = Encrypt(pdfFile, encrypted, "correctpassword", "")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with wrong password
	decrypted := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt(encrypted, decrypted, "wrongpassword")
	if err == nil {
		t.Error("Decrypt() should return error for wrong password")
	}
}

func TestEncryptSpecialPassword(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with special characters in password
	passwords := []string{
		"p@ss!w0rd#$%",
		"pass with spaces",
		"unicode:日本語",
	}

	for _, pw := range passwords {
		t.Run(pw, func(t *testing.T) {
			encrypted := filepath.Join(tmpDir, "encrypted_special.pdf")
			err := Encrypt(pdfFile, encrypted, pw, "")
			if err != nil {
				t.Fatalf("Encrypt() with special password error = %v", err)
			}
		})
	}
}

func TestEncryptWithOwnerPassword(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	encrypted := filepath.Join(tmpDir, "encrypted_owner.pdf")
	userPW := "user123"
	ownerPW := "owner456"

	// Encrypt with both user and owner passwords
	err = Encrypt(pdfFile, encrypted, userPW, ownerPW)
	if err != nil {
		t.Fatalf("Encrypt() with owner password error = %v", err)
	}

	// Verify encrypted file exists
	if _, err := os.Stat(encrypted); os.IsNotExist(err) {
		t.Fatal("Encrypt() did not create output file")
	}

	// Verify we can decrypt with owner password
	decrypted := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt(encrypted, decrypted, ownerPW)
	if err != nil {
		t.Fatalf("Decrypt() with owner password error = %v", err)
	}
}

func TestExtractPages(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "extracted.pdf")
	err = ExtractPages(pdf, output, []int{1}, "")
	if err != nil {
		t.Fatalf("ExtractPages() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("ExtractPages() did not create output file")
	}

	// Verify extracted file has 1 page
	count, err := PageCount(output, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Extracted PDF has %d pages, want 1", count)
	}
}

func TestExtractPagesEmptyList(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "extracted.pdf")
	// Empty pages list - behavior depends on implementation
	err = ExtractPages(pdfFile, output, []int{}, "")
	// Just verify it doesn't panic
	_ = err
}

func TestExtractPagesNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "extracted.pdf")
	err = ExtractPages("/nonexistent/file.pdf", output, []int{1}, "")
	if err == nil {
		t.Error("ExtractPages() expected error for non-existent file")
	}
}
