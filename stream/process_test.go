package stream

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	srcChannels := Distribute(
		Stream(func(_ context.Context, output chan<- *Result[int]) {
			for i := range 100 {
				output <- NewResult(i, nil)
			}
		}),
		2,
		WithBufferSize(0),
	)

	destChannels := Processor(
		srcChannels,
		func(id int) func(ctx context.Context, input int) ([]int, error) {
			return func(ctx context.Context, input int) ([]int, error) {
				time.Sleep(1 * time.Millisecond) // synthetic load
				return []int{id, input}, nil
			}
		},
	)

	var first atomic.Int64
	var second atomic.Int64
	for value := range Multiplex(destChannels) {
		require.NoError(t, value.Error)
		switch value.Value[0] {
		case 0:
			first.Add(1)
		case 1:
			second.Add(1)
		}
	}

	require.Less(t, int64(30), first.Load())
	require.Less(t, int64(30), second.Load())
}
