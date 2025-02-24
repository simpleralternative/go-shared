package stream

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResult(t *testing.T) {
	val := 1
	errTest := errors.New("test error")
	result := NewResult(1, errTest)
	require.Equal(t, &Result[int]{Value: val, Error: errTest}, result)
	value, err := result.Destructure()
	require.Equal(t, errTest, err)
	require.Equal(t, val, value)
}
