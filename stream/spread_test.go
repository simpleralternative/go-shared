package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSpread(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		batched := Stream(func(_ context.Context, output chan<- []int) {
			for i := range 3 {
				output <- []int{i, i + 1, i + 2}
			}
		})
		unbatched := Spread(batched)

		require.Equal(t, 0, <-unbatched)
		require.Equal(t, 1, <-unbatched)
		require.Equal(t, 2, <-unbatched)

		require.Equal(t, 1, <-unbatched)
		require.Equal(t, 2, <-unbatched)
		require.Equal(t, 3, <-unbatched)

		require.Equal(t, 2, <-unbatched)
		require.Equal(t, 3, <-unbatched)
		require.Equal(t, 4, <-unbatched)

		value, ok := <-unbatched
		require.False(t, ok)
		require.Equal(t, 0, value)
	})

	t.Run("cancelable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		batched := Stream(func(ctx context.Context, output chan<- []int) {
			for i := range 3 {
				select {
				case <-ctx.Done():
					return
				case output <- []int{i, i + 1, i + 2}:
				}
			}
		}, WithContext(ctx), WithBufferSize(0))
		unbatched := Spread(batched, WithContext(ctx), WithBufferSize(0))

		require.Equal(t, 0, <-unbatched)
		require.Equal(t, 1, <-unbatched)
		require.Equal(t, 2, <-unbatched)

		cancel()
		time.Sleep(1 * time.Millisecond)

		// batch already in pipeline
		require.Equal(t, 1, <-unbatched)
		require.Equal(t, 2, <-unbatched)
		require.Equal(t, 3, <-unbatched)

		// the third batch isn't sent
		value, ok := <-unbatched
		require.False(t, ok)
		require.Equal(t, 0, value)
	})
}
