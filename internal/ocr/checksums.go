package ocr

// KnownChecksums maps language codes to SHA256 checksums for tessdata_fast files.
//
// To add a new language:
// 1. curl -sL "https://github.com/tesseract-ocr/tessdata_fast/raw/main/LANG.traineddata" -o /tmp/LANG.traineddata
// 2. sha256sum /tmp/LANG.traineddata (or shasum -a 256 on macOS)
// 3. Add entry below
var KnownChecksums = map[string]string{
	"eng": "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2",
}

// GetChecksum returns the known SHA256 checksum for a language, or empty string if unknown.
func GetChecksum(lang string) string {
	return KnownChecksums[lang]
}

// HasChecksum returns true if a checksum is known for the given language.
func HasChecksum(lang string) bool {
	_, ok := KnownChecksums[lang]
	return ok
}
