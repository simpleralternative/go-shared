package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDistribute(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		channels := Distribute(
			Stream(func(_ context.Context, output chan<- int) {
				output <- 1
				output <- 2
			}),
			2,
		)

		var actual []int
		for value := range channels[0] {
			actual = append(actual, value)
		}

		for value := range channels[1] {
			actual = append(actual, value)
		}

		require.ElementsMatch(t, []int{1, 2}, actual)
	})

	t.Run("cancelable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// this stream has a large buffer and default context.
		defaultStream := Stream(func(ctx context.Context, output chan<- int) {
			for i := range 10 {
				output <- i
			}
		})

		channels := Distribute(defaultStream,
			2,
			WithContext(ctx),  // can be cancelled.
			WithBufferSize(0), // generates unbuffered, synchronous queues.
		)

		// the synchronous buffers can only send when something listens to them
		// so this example can receieve from any combination of the channels.
		actual := []int{<-channels[0], <-channels[0], <-channels[1]}

		cancel()                         // signal via context.Done()
		time.Sleep(1 * time.Millisecond) // wait for selects to resolve
		validateChannel(t, 0, false, channels[0])
		validateChannel(t, 0, false, channels[1])

		require.ElementsMatch(t, []int{0, 1, 2}, actual)
	})
}
