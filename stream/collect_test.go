package stream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCollect(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	t.Run("valid", func(t *testing.T) {
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for _, value := range input {
				output <- NewResult(value, nil)
			}
		})
		collected := Collect(src)

		value, ok := <-collected
		require.True(t, ok)
		require.Equal(t, NewResult(input, nil), value)
	})

	t.Run("cancelable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		// intentionally not providing the context to the source stream
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for i := range 10 {
				output <- NewResult(i, nil)
				time.Sleep(10 * time.Millisecond)
			}
		})
		collected := Collect(src, WithContext(ctx))

		time.Sleep(1 * time.Millisecond)
		cancel()

		value, ok := <-collected
		require.False(t, ok)
		require.Nil(t, value)
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("general error")
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			output <- NewResult(1, nil)
			output <- NewResult(2, nil)
			output <- NewResult(3, nil)
			output <- NewResult(0, err)
			output <- NewResult(4, nil)
			output <- NewResult(5, nil)
			output <- NewResult(6, nil)
		})
		collected := Collect(src)

		value, ok := <-collected
		require.True(t, ok)
		require.Equal(t, NewResult[[]int](nil, err), value)
	})
}
