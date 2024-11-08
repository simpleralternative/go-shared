package stream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBatch(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	t.Run("basic batching 2", func(t *testing.T) {
		grouping := 2
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for _, value := range input {
				output <- NewResult(value, nil)
			}
		})
		batched := Batch(src, grouping)

		for value := range 5 {
			output, ok := <-batched
			require.True(t, ok)
			require.Equal(t,
				&Result[[]int]{[]int{
					value*grouping + 1,
					value*grouping + 2,
				}, nil},
				output,
			)
		}
		output, ok := <-batched
		require.False(t, ok)
		require.Nil(t, output)
	})

	t.Run("basic batching 3", func(t *testing.T) {
		grouping := 3
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for _, value := range input {
				output <- NewResult(value, nil)
			}
		})
		batched := Batch(src, grouping)

		for value := range 3 {
			output, ok := <-batched
			require.True(t, ok)
			require.Equal(t,
				&Result[[]int]{[]int{
					value*grouping + 1,
					value*grouping + 2,
					value*grouping + 3,
				}, nil},
				output,
			)
		}

		output, ok := <-batched
		require.True(t, ok)
		require.Equal(t, &Result[[]int]{[]int{10}, nil}, output)

		output, ok = <-batched
		require.False(t, ok)
		require.Nil(t, output)
	})

	t.Run("cancel mid stream", func(t *testing.T) {
		grouping := 2
		ctx, cancel := context.WithCancel(context.Background())
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for _, value := range input {
				output <- NewResult(value, nil)
			}
		}, WithBufferSize(0))
		batched := Batch(src, grouping, WithContext(ctx), WithBufferSize(0))

		output, ok := <-batched
		require.True(t, ok)
		require.Equal(t, &Result[[]int]{[]int{1, 2}, nil}, output)

		cancel()
		time.Sleep(1 * time.Millisecond)

		output, ok = <-batched
		require.False(t, ok)
		require.Nil(t, output)
	})

	t.Run("cancel at the last batch", func(t *testing.T) {
		grouping := 2
		ctx, cancel := context.WithCancel(context.Background())
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for _, value := range input {
				output <- NewResult(value, nil)
			}
		}, WithBufferSize(0))
		batched := Batch(src, grouping, WithContext(ctx), WithBufferSize(0))

		for value := range 4 {
			output, ok := <-batched
			require.True(t, ok)
			require.Equal(t,
				&Result[[]int]{[]int{
					value*grouping + 1,
					value*grouping + 2,
				}, nil},
				output,
			)
		}

		cancel()
		time.Sleep(1 * time.Millisecond)

		// the cancel signal doesn't prevent the final batch from being sent.
		output, ok := <-batched
		require.True(t, ok)
		require.Equal(t, &Result[[]int]{[]int{9, 10}, nil}, output)

		output, ok = <-batched
		require.False(t, ok)
		require.Nil(t, output)
	})

	t.Run("error", func(t *testing.T) {
		grouping := 2
		err := errors.New("generic error")
		src := Stream(func(_ context.Context, output chan<- *Result[int]) {
			output <- NewResult(1, nil)
			output <- NewResult(2, nil)
			output <- NewResult(3, nil)
			output <- NewResult(0, errors.Join(err, errors.New("wrapper")))
			output <- NewResult(4, nil)
			output <- NewResult(5, nil)
			output <- NewResult(6, nil)
		})
		batched := Batch(src, grouping)

		require.Equal(t, &Result[[]int]{[]int{1, 2}, nil}, <-batched)
		result := <-batched
		require.Nil(t, result.Value)
		require.True(t, errors.Is(result.Error, err))
		// require.Equal(t, &Result[[]int]{nil, err}, <-batched)
		require.Equal(t, &Result[[]int]{[]int{3, 4}, nil}, <-batched)
		require.Equal(t, &Result[[]int]{[]int{5, 6}, nil}, <-batched)

		output, ok := <-batched
		require.False(t, ok)
		require.Nil(t, output)
	})
}
