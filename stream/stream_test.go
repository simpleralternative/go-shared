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

	t.Run("simple With Result", func(t *testing.T) {
		output := Stream(func(_ context.Context, output chan<- *Result[int]) {
			output <- NewResult(1, nil)
			output <- NewResult(2, nil)
			output <- &Result[int]{Value: 3}
		})

		validateChannel(t, &Result[int]{Value: 1}, true, output)
		validateChannel(t, NewResult(2, nil), true, output)
		validateChannel(t, NewResult(3, nil), true, output)
		validateChannel(t, nil, false, output)
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

	t.Run("loop with Result", func(t *testing.T) {
		output := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for i := range 4 {
				output <- NewResult(i, nil)
			}
		})

		validateChannel(t, NewResult(0, nil), true, output)
		validateChannel(t, NewResult(1, nil), true, output)
		validateChannel(t, NewResult(2, nil), true, output)
		validateChannel(t, NewResult(3, nil), true, output)
		validateChannel(t, nil, false, output)
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
		output := Stream(func(ctx context.Context, output chan<- *Result[int]) {
			for i := range 10 {
				select {
				case <-ctx.Done():
					return
				case output <- NewResult(i, nil):
				}
			}
		},
			WithContext(ctx),
			WithBufferSize(0), // prevent preloading all the values
		)

		validateChannel(t, NewResult(0, nil), true, output)
		validateChannel(t, NewResult(1, nil), true, output)
		cancel()
		// guarantee the select completes for consistency
		time.Sleep(1 * time.Millisecond)
		validateChannel(t, nil, false, output)
	})
}

// note, benchmark values vary slightly from run to run. these no-load results
// are representative enough to show the performance of Stream is effectively
// the same as a vanilla producer for the same process and payload.
//
// the biggest takeaway is that concurrency isn't about iteration performance.
// the reference loops demonstrate that a pure for loop, shown with a function
// callsite for effect, with or without a Result object, is orders of magnitude
// faster than using channels. as the readme describes, we use them to improve
// our code's readability when we can afford the performance of only a few
// million iterations per second.
//
// use a reasonable buffer (such as the default). it has a dramatic beneficial
// effect on the throughput.
//
// using the context cancel channel imposes a large performance penalty, but
// provides control. use it when you need it.
//
// goos: linux
// goarch: amd64
// pkg: github.com/simpleralternative/go-shared/stream
// cpu: AMD Ryzen 7 7840U w/ Radeon 780M Graphics
// BenchmarkStream/nonconcurrent_reference-16               1000000000    0.2209 ns/op  0 B/op  0 allocs/op
// BenchmarkStream/nonconcurrent_reference_with_Result-16   1000000000    0.8165 ns/op  0 B/op  0 allocs/op
//
// BenchmarkStream/vanilla_unbuffered_channel-16              11734498  101.3 ns/op    0 B/op  0 allocs/op
// BenchmarkStream/vanilla_unbuffered_channel_with_Result-16   8046616  157.1 ns/op   24 B/op  1 allocs/op
// BenchmarkStream/Stream-sourced_unbuffered_channel-16        8177119  141.2 ns/op   24 B/op  1 allocs/op
// BenchmarkStream/vanilla_buffered_channel-16                53098714   23.12 ns/op   0 B/op  0 allocs/op
// BenchmarkStream/vanilla_buffered_channel_with_Result-16    22307397   53.24 ns/op  24 B/op  1 allocs/op
// BenchmarkStream/Stream-sourced_channel-16                  22733671   51.03 ns/op  24 B/op  1 allocs/op
// BenchmarkStream/vanilla_channel_with_select-16             33689454   36.32 ns/op   0 B/op  0 allocs/op
// BenchmarkStream/vanilla_channel_with_select_and_Result-16  18533895   65.91 ns/op  24 B/op  1 allocs/op
// BenchmarkStream/Stream-sourced_channel_with_select-16      19718461   64.38 ns/op  24 B/op  1 allocs/op
// BenchmarkStream/vanilla_verifier-16                        11238660  112.5 ns/op    0 B/op  0 allocs/op
// BenchmarkStream/Stream-sourced_verifier-16                 22809950   56.12 ns/op  24 B/op  1 allocs/op
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

	b.Run("vanilla unbuffered channel with Result", func(b *testing.B) {
		channel := make(chan *Result[int])
		go func() {
			defer close(channel)
			for i := 0; ; i++ {
				channel <- &Result[int]{Value: i}
			}
		}()

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("Stream-sourced unbuffered channel", func(b *testing.B) {
		channel := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for i := 0; ; i++ {
				output <- NewResult(i, nil)
			}
		}, WithBufferSize(0))

		b.ResetTimer()
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

	b.Run("vanilla buffered channel with Result", func(b *testing.B) {
		channel := make(chan *Result[int], defaultBufferSize)
		go func() {
			defer close(channel)
			for i := 0; ; i++ {
				channel <- NewResult(i, nil)
			}
		}()

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("Stream-sourced channel", func(b *testing.B) {
		channel := Stream(func(_ context.Context, output chan<- *Result[int]) {
			for i := 0; ; i++ {
				output <- NewResult(i, nil)
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

	b.Run("vanilla channel with select and Result", func(b *testing.B) {
		channel := make(chan *Result[int], defaultBufferSize)
		ctx := context.Background()
		go func() {
			defer close(channel)
			for i := 0; ; i++ {
				select {
				case <-ctx.Done():
					return
				case channel <- NewResult(i, nil):
				}
			}
		}()

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("Stream-sourced channel with select", func(b *testing.B) {
		channel := Stream(func(ctx context.Context, output chan<- *Result[int]) {
			for i := 0; ; i++ {
				select {
				case <-ctx.Done():
					return
				case output <- NewResult(i, nil):
				}
			}
		})

		for i := 0; i < b.N; i++ {
			<-channel
		}
	})

	b.Run("nonconcurrent reference", func(b *testing.B) {
		ct, sum := 0, int64(0)
		for i := 0; i < b.N; i++ {
			sum += func() int64 {
				return int64(i)
			}()
			ct++
		}
		b.Log(ct, sum)
	})

	b.Run("nonconcurrent reference with Result", func(b *testing.B) {
		ct, sum := 0, int64(0)
		for i := 0; i < b.N; i++ {
			sum += (func() *Result[int64] {
				return NewResult(int64(i), nil)
			}()).Value
			ct++
		}
		b.Log(ct, sum)
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
		channel := Stream(
			func(_ context.Context, output chan<- *Result[int64]) {
				for i := 0; ; i++ {
					output <- NewResult(int64(i), nil)
				}
			},
		)

		ct, sum := 0, int64(0)
		for i := 0; i < b.N; i++ {
			sum += (<-channel).Value
			ct++
		}
		b.Log(ct, sum)
	})
}
