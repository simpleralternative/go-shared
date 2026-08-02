package aes

import (
	"crypto/aes"
	"crypto/cipher"
)

func prep(key []byte) (cipher.AEAD, error) {
	cypher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// https://en.wikipedia.org/wiki/Galois/Counter_Mode
	return cipher.NewGCM(cypher)
}
