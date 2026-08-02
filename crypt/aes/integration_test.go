package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	key1   = []byte("supersecret phrase example key 1") // 32B
	input1 = []byte("this is an example string")
)

func TestXCrypt(t *testing.T) {
	t.Run("an encrypted value can be decrypted", func(t *testing.T) {
		encoded, err := Encrypt(key1, input1)
		require.NoError(t, err)

		original, err := Decrypt(key1, encoded)
		require.NoError(t, err)
		require.Equal(t, string(input1), string(original))
	})
}
