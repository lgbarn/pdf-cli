package pdf

import (
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Encrypt adds password protection to a PDF
func Encrypt(input, output, userPW, ownerPW string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPW
	if ownerPW != "" {
		conf.OwnerPW = ownerPW
	} else {
		conf.OwnerPW = userPW
	}
	return api.EncryptFile(input, output, conf)
}

// Decrypt removes password protection from a PDF
func Decrypt(input, output, password string) error {
	return api.DecryptFile(input, output, NewConfig(password))
}
