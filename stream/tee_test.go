package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTee(t *testing.T) {
	t.Run("simple value", func(t *testing.T) {
		src := Stream(func(ctx context.Context, output chan<- int) {
			for id := range 10 {
				output <- id
			}
		})

		out1, out2 := Tee(src)
		for expected := range 10 {
			require.Equal(t, expected, <-out1)
			require.Equal(t, expected, <-out2)
		}
		value, ok := <-out1
		require.False(t, ok)
		require.Equal(t, 0, value)

		value, ok = <-out2
		require.False(t, ok)
		require.Equal(t, 0, value)
	})

	t.Run("simple pointer", func(t *testing.T) {
		src := Stream(func(ctx context.Context, output chan<- *int) {
			for id := range 10 {
				output <- &id
			}
		})

		out1, out2 := Tee(src)
		for expected := range 10 {
			require.Equal(t, &expected, <-out1)
			require.Equal(t, &expected, <-out2)
		}
		value, ok := <-out1
		require.False(t, ok)
		require.Nil(t, value)

		value, ok = <-out2
		require.False(t, ok)
		require.Nil(t, value)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(
			context.Background(),
			time.Now().Add(70*time.Millisecond),
		)
		defer cancel()

		src := Stream(
			func(ctx context.Context, output chan<- int) {
				for id := range 10 {
					output <- id + 10
					time.Sleep(50 * time.Millisecond)
				}
			},
			WithContext(ctx),
		)

		out1, out2 := Tee(src, WithContext(ctx))

		value, ok := <-out1
		require.True(t, ok)
		require.Equal(t, 10, value)

		value, ok = <-out1
		require.True(t, ok)
		require.Equal(t, 11, value)

		// buffered records
		value, ok = <-out2
		require.True(t, ok)
		require.Equal(t, 10, value)

		value, ok = <-out2
		require.True(t, ok)
		require.Equal(t, 11, value)

		// cancel occured
		value, ok = <-out1
		require.False(t, ok)
		require.Equal(t, 0, value)

		value, ok = <-out2
		require.False(t, ok)
		require.Equal(t, 0, value)
	})
}
