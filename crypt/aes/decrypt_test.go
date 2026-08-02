package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecrypt(t *testing.T) {
	t.Run("nil keys will fail", func(t *testing.T) {
		_, err := Decrypt(nil, input1)
		require.Error(t, err)
	})

	t.Run("nil inputs will fail", func(t *testing.T) {
		_, err := Decrypt(key1, nil)
		require.Error(t, err)
	})
}
