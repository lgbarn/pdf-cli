package ocr

// KnownChecksums maps language codes to SHA256 checksums for tessdata_fast files.
//
// To add a new language:
// 1. curl -sL "https://github.com/tesseract-ocr/tessdata_fast/raw/main/LANG.traineddata" -o /tmp/LANG.traineddata
// 2. sha256sum /tmp/LANG.traineddata (or shasum -a 256 on macOS)
// 3. Add entry below
//
// TODO: The 'eng' checksum below is a placeholder and must be updated with the real checksum.
// To compute the real checksum, run:
//
//	curl -sL "https://github.com/tesseract-ocr/tessdata_fast/raw/main/eng.traineddata" | shasum -a 256
var KnownChecksums = map[string]string{
	// Placeholder checksum - must be updated with real value
	"eng": "0000000000000000000000000000000000000000000000000000000000000000",
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
