package ocr

import "testing"

func TestGetChecksum(t *testing.T) {
	if checksum := GetChecksum("eng"); checksum == "" {
		t.Error("Expected checksum for 'eng', got empty string")
	}

	if checksum := GetChecksum("xyz_nonexistent"); checksum != "" {
		t.Errorf("Expected empty checksum for unknown language, got: %s", checksum)
	}
}

func TestHasChecksum(t *testing.T) {
	if !HasChecksum("eng") {
		t.Error("Expected HasChecksum('eng') to be true")
	}

	if HasChecksum("xyz_nonexistent") {
		t.Error("Expected HasChecksum('xyz_nonexistent') to be false")
	}
}

func TestAllChecksumsValidFormat(t *testing.T) {
	for lang, checksum := range KnownChecksums {
		if len(checksum) != 64 {
			t.Errorf("Invalid checksum length for %s: got %d, want 64", lang, len(checksum))
		}
		for _, c := range checksum {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				t.Errorf("Invalid hex character in checksum for %s: %c", lang, c)
				break
			}
		}
	}
}
