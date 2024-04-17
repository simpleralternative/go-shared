package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMultiplex(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		channels := []<-chan int{
			Stream(func(_ context.Context, output chan<- int) {
				for i := range 8 {
					output <- i
				}
			}),
			Stream(func(_ context.Context, output chan<- int) {
				output <- 1
				output <- 2
			}),
		}

		outputs := []int{}
		for value := range Multiplex(channels) {
			outputs = append(outputs, value)
		}

		require.ElementsMatch(t, []int{0, 1, 1, 2, 2, 3, 4, 5, 6, 7}, outputs)
	})

	t.Run("cancelable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// the source streams will produce content as fast as they can, but use
		// unbuffered channels.
		channels := []<-chan int{
			Stream(func(_ context.Context, output chan<- int) {
				for i := range 8 {
					time.Sleep(1 * time.Millisecond)
					output <- i
				}
			}, WithBufferSize(0)),
			Stream(func(_ context.Context, output chan<- int) {
				output <- 10
				output <- 11
			}, WithBufferSize(0)),
		}

		output := Multiplex(channels, WithContext(ctx), WithBufferSize(0))

		require.Equal(t, 10, <-output) // the second stream outruns the first
		require.Equal(t, 11, <-output)
		require.Equal(t, 0, <-output) // the first stream produces
		require.Equal(t, 1, <-output)

		cancel()                         // signal via context.Done()
		time.Sleep(2 * time.Millisecond) // wait for select to resolve
		validateChannel(t, 0, false, output)
	})
}
