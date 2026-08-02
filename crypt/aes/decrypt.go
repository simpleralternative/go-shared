package aes

import "errors"

var errorDecryptInputInvalid = errors.New(
	"this input provided is invalid for decrypting",
)

// Decrypt uses the provided 32b key to attempt to decrypt and return the
// provided input bytes.
func Decrypt(key, input []byte) (output []byte, err error) {
	gcm, err := prep(key)
	if err != nil {
		return
	}

	nonceSize := gcm.NonceSize()

	if len(input) < nonceSize {
		return nil, errorDecryptInputInvalid
	}

	nonce, cipherText := input[:nonceSize], input[nonceSize:]
	return gcm.Open(nil, nonce, cipherText, nil)
}
