package ocr

// KnownChecksums maps language codes to SHA256 checksums for tessdata_fast files.
//
// To add a new language:
// 1. curl -sL "https://github.com/tesseract-ocr/tessdata_fast/raw/main/LANG.traineddata" -o /tmp/LANG.traineddata
// 2. sha256sum /tmp/LANG.traineddata (or shasum -a 256 on macOS)
// 3. Add entry below
var KnownChecksums = map[string]string{
	"ara":     "e3206d3dc87fd50c24a0fb9f01838615911d25168f4e64415244b67d2bb3e729",
	"ces":     "934bcaf97ef3348413263331131c9fa7f55f30db333c711929c124fb635f7e1b",
	"chi_sim": "a5fcb6f0db1e1d6d8522f39db4e848f05984669172e584e8d76b6b3141e1f730",
	"chi_tra": "529c5b5797d64b126065cd55f2bb4c7fd7b15790798091b1ff259941a829330b",
	"deu":     "19d219bbb6672c869d20a9636c6816a81eb9a71796cb93ebe0cb1530e2cdb22d",
	"eng":     "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2",
	"fra":     "ced037562e8c80c13122dece28dd477d399af80911a28791a66a63ac1e3445ca",
	"hin":     "4c73ffc59d497c186b19d1e90f5d721d678ea6b2e277b719bee4e2af12271825",
	"ita":     "b8f89e1e785118dac4d51ae042c029a64edb5c3ee42ef73027a6d412748d8827",
	"jpn":     "1f5de9236d2e85f5fdf4b3c500f2d4926f8d9449f28f5394472d9e8d83b91b4d",
	"kor":     "6b85e11d9bbf07863b97b3523b1b112844c43e713df8b66418a081fd1060b3b2",
	"nld":     "ced0e5e046a84c908a6aa7accbef9a232c4a5d9a8276691b81c6ee64d02963f6",
	"nor":     "0451eb4f8049ae78196806bf878a389a2f40f1386fe038568cf4441226ba6ef2",
	"pol":     "c4476cdbc0e33d898d32345122b7be1cbf85ace15f920f06c7714756e1ef79b2",
	"por":     "c4932b937207a9514b7514d518b931a99938c02a28a5a5a553f8599ed58b7deb",
	"rus":     "e16e5e036cce1d9ec2b00063cf8b54472625b9e14d893a169e2b0dedeb4df225",
	"spa":     "6f2e04d02774a18f01bed44b1111f2cd7f3ba7ac9dc4373cd3f898a40ea6b464",
	"swe":     "f7304988d41f833efebcc2d529df54b1903ecebbc3da1faabd19a0fddd4fe586",
	"tur":     "7393381111e1152420fc4092cb44eef4237580d21b92bf30d7d221aad192c6b7",
	"ukr":     "d59e53e2bded32f4445f124b4b00240fcac7e8044c003ab822ccb94f0b3db59b",
	"vie":     "79df64caf7bcfb2a27df5042ecb6121e196eada34da774956995747636d5bfa1",
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
