package stream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSpread(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		batched := Stream(
			func(_ context.Context, output chan<- *Result[[]int]) {
				for i := range 3 {
					output <- NewResult([]int{i, i + 1, i + 2}, nil)
				}
			},
		)
		unbatched := Spread(batched)

		require.Equal(t, NewResult(0, nil), <-unbatched)
		require.Equal(t, NewResult(1, nil), <-unbatched)
		require.Equal(t, NewResult(2, nil), <-unbatched)

		require.Equal(t, NewResult(1, nil), <-unbatched)
		require.Equal(t, NewResult(2, nil), <-unbatched)
		require.Equal(t, NewResult(3, nil), <-unbatched)

		require.Equal(t, NewResult(2, nil), <-unbatched)
		require.Equal(t, NewResult(3, nil), <-unbatched)
		require.Equal(t, NewResult(4, nil), <-unbatched)

		value, ok := <-unbatched
		require.False(t, ok)
		require.Nil(t, value)
	})

	t.Run("cancelable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		batched := Stream(
			func(ctx context.Context, output chan<- *Result[[]int]) {
				for i := range 3 {
					select {
					case <-ctx.Done():
						return
					case output <- NewResult([]int{i, i + 1, i + 2}, nil):
					}
				}
			},
			WithContext(ctx),
			WithBufferSize(0),
		)
		unbatched := Spread(batched, WithContext(ctx), WithBufferSize(0))

		require.Equal(t, NewResult(0, nil), <-unbatched)
		require.Equal(t, NewResult(1, nil), <-unbatched)
		require.Equal(t, NewResult(2, nil), <-unbatched)

		cancel()
		time.Sleep(1 * time.Millisecond)

		// batch already in pipeline
		require.Equal(t, NewResult(1, nil), <-unbatched)
		require.Equal(t, NewResult(2, nil), <-unbatched)
		require.Equal(t, NewResult(3, nil), <-unbatched)

		// the third batch isn't sent
		value, ok := <-unbatched
		require.False(t, ok)
		require.Nil(t, value)
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("generic error")
		batched := Stream(
			func(_ context.Context, output chan<- *Result[[]int]) {
				output <- NewResult([]int{1, 2, 3}, nil)
				output <- NewResult[[]int](nil, err)
				output <- NewResult([]int{4, 5, 6}, nil)
			},
		)
		unbatched := Spread(batched)

		require.Equal(t, NewResult(1, nil), <-unbatched)
		require.Equal(t, NewResult(2, nil), <-unbatched)
		require.Equal(t, NewResult(3, nil), <-unbatched)

		require.Equal(t, NewResult[int](0, err), <-unbatched)

		require.Equal(t, NewResult(4, nil), <-unbatched)
		require.Equal(t, NewResult(5, nil), <-unbatched)
		require.Equal(t, NewResult(6, nil), <-unbatched)

		value, ok := <-unbatched
		require.False(t, ok)
		require.Nil(t, value)
	})
}
