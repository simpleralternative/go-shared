package aes

import (
	"crypto/rand"
	"io"
)

// Encrypt will attempt to use the provided key to encrypt the provided
// input bytes.
//
// From aes.NewCypher:
// The key argument must be the AES key, either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256.
func Encrypt(key, input []byte) ([]byte, error) {
	gcm, err := prep(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	//populates our nonce with a cryptographically secure random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		// this shouldn't ever happen because the nonce's size prevents an
		// EOF error
		return nil, err
	}

	return gcm.Seal(nonce, nonce, input, nil), nil
}
