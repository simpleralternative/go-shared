package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncrypt(t *testing.T) {
	t.Run("nil keys will fail", func(t *testing.T) {
		_, err := Encrypt(nil, input1)
		require.Error(t, err)
	})

	t.Run("invalid length will fail", func(t *testing.T) {
		_, err := Encrypt(([]byte("random key length")), input1)
		require.Error(t, err, err)
	})

	t.Run("nil inputs will not fail", func(t *testing.T) {
		encoded, err := Encrypt(key1, nil)
		require.NoError(t, err)
		require.Equal(t, 28, len(encoded)) // the encoded variable nonce
	})
}
