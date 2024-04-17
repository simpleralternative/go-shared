package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validateChannel[T any](t *testing.T, expected T, okay bool, src <-chan T) {
	actual, ok := <-src
	require.Equal(t, okay, ok)
	require.Equal(t, expected, actual)
}

func TestStream(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		output := Stream(func(output chan<- int) {
			output <- 1
			output <- 2
			output <- 3
		})

		validateChannel(t, 1, true, output)
		validateChannel(t, 2, true, output)
		validateChannel(t, 3, true, output)
		validateChannel(t, 0, false, output)
	})

	t.Run("loop", func(t *testing.T) {
		output := Stream(func(output chan<- int) {
			for i := range 4 {
				output <- i
			}
		})

		validateChannel(t, 0, true, output)
		validateChannel(t, 1, true, output)
		validateChannel(t, 2, true, output)
		validateChannel(t, 3, true, output)
		validateChannel(t, 0, false, output)
	})

	// in any situation where the production process might produce more records
	// than the consumer will accept (eg: the consumer exits early for some
	// reason) the producing code will hang indefinitely, waiting for the queue
	// to drain.
	//
	// to avoid this memory leak, a common practice should be to use a signal
	// (often a context) to safely abort the producer which will then exit the
	// function and trigger cleanup.
	//
	// downstream consumers are much simpler since they can either listen to the
	// same signal, or just drain the rest of the queue and finish normally.
	t.Run("cancelable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		output := Stream(func(output chan<- int) {
			for i := range 10 {
				select {
				case <-ctx.Done():
					return
				case output <- i:
				}
			}
		})

		validateChannel(t, 0, true, output)
		validateChannel(t, 1, true, output)
		cancel()
		// guarantee the select completes for consistency
		time.Sleep(1 * time.Millisecond)
		validateChannel(t, 0, false, output)
	})
}
