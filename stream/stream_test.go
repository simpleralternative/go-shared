package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validateChannel[T any](
	t *testing.T,
	expected T,
	okay bool,
	src <-chan T,
) {
	actual, ok := <-src
	require.Equal(t, okay, ok)
	require.Equal(t, expected, actual)
}

func TestStream(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		output := Stream(func(_ context.Context, output chan<- int) {
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
		output := Stream(func(_ context.Context, output chan<- int) {
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
		output := Stream(func(ctx context.Context, output chan<- int) {
			for i := range 10 {
				select {
				case <-ctx.Done():
					return
				case output <- i:
				}
			}
		},
			WithContext(ctx),
			WithBufferSize(0), // prevent preloading all the values
		)

		validateChannel(t, 0, true, output)
		validateChannel(t, 1, true, output)
		cancel()
		// guarantee the select completes for consistency
		time.Sleep(1 * time.Millisecond)
		validateChannel(t, 0, false, output)
	})
}

// note, benchmark values vary slightly from run to run. these no-load results
// are representative enough to show the performance of Stream is effectively
// the same as a vanilla producer.
//
// using a reasonable buffer has a dramatic beneficial effect on the throughput.
// using the context cancel channel almost halves the throughput, but usecases
// may benefit or demand it.
//
// goos: linux
// goarch: amd64
// pkg: github.com/simpleralternative/go-shared/stream
// cpu: AMD Ryzen 7 7840U w/ Radeon 780M Graphics
// BenchmarkStream/vanilla_unbuffered_channel-16         12110923   96.22 ns/op  0 B/op  0 allocs/op
// BenchmarkStream/Stream-sourced_unbuffered_channel-16  12515575   97.38 ns/op  0 B/op  0 allocs/op
// BenchmarkStream/vanilla_10k_buffered_channel-16       58020580   20.79 ns/op  0 B/op  0 allocs/op
// BenchmarkStream/Stream-sourced_channel-16             56895408   20.69 ns/op  0 B/op  0 allocs/op
// BenchmarkStream/vanilla_channel_with_select-16        35377322   33.66 ns/op   0 B/op  0 allocs/op
// BenchmarkStream/Stream-sourced_channel_with_select-16 35859765   33.93 ns/op  0 B/op  0 allocs/op
// BenchmarkStream/vanilla_verifier-16                   11284138  104.1 ns/op   0 B/op  0 allocs/op
func BenchmarkStream(b *testing.B) {
	b.Run("vanilla unbuffered channel", func(b *testing.B) {
		channel := make(chan int)
		go func() {
			defer close(channel)
			for i := 0; ; i++ {
				channel <- i
			}
		}()

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("Stream-sourced unbuffered channel", func(b *testing.B) {
		channel := Stream(func(_ context.Context, output chan<- int) {
			for i := 0; ; i++ {
				output <- i
			}
		}, WithBufferSize(0))

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("vanilla buffered channel", func(b *testing.B) {
		channel := make(chan int, defaultBufferSize)
		go func() {
			defer close(channel)
			for i := 0; ; i++ {
				channel <- i
			}
		}()

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("Stream-sourced channel", func(b *testing.B) {
		channel := Stream(func(_ context.Context, output chan<- int) {
			for i := 0; ; i++ {
				output <- i
			}
		})

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("vanilla channel with select", func(b *testing.B) {
		channel := make(chan int, defaultBufferSize)
		ctx := context.Background()
		go func() {
			defer close(channel)
			for i := 0; ; i++ {
				select {
				case <-ctx.Done():
					return
				case channel <- i:
				}
			}
		}()

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("Stream-sourced channel with select", func(b *testing.B) {
		channel := Stream(func(ctx context.Context, output chan<- int) {
			for i := 0; ; i++ {
				select {
				case <-ctx.Done():
					return
				case output <- i:
				}
			}
		})

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("vanilla verifier", func(b *testing.B) {
		channel := make(chan int64)
		go func() {
			defer close(channel)
			for i := 0; ; i++ {
				channel <- int64(i)
			}
		}()

		ct, sum := 0, int64(0)
		for i := 0; i < b.N; i++ {
			sum += <-channel
			ct++
		}
		b.Log(ct, sum)
	})

	b.Run("Stream-sourced verifier", func(b *testing.B) {
		channel := Stream(func(_ context.Context, output chan<- int64) {
			for i := 0; ; i++ {
				output <- int64(i)
			}
		})

		ct, sum := 0, int64(0)
		for i := 0; i < b.N; i++ {
			sum += <-channel
			ct++
		}
		b.Log(ct, sum)
	})
}
