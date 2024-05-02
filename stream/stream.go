package stream

import "context"

// Stream is a convenience function that performs the basic channel ceremonies
// for a standard producer-consumer paradigm.
//
// it takes a function that will be given a send-only channel. that channel is
// immediately returned and the function is executed in a goroutine where it may
// then do whatever logic is needed to produce the values that are put into the
// channel.
//
// note: the processing functions in the stream package work on Result packets,
// so your producer may need to wrap the data accordingly.
//
// when the provided function completes, the output channel is closed and
// consumers can drain the channel as normal.
//
// See the tests for examples and benchmarks.
func Stream[T any](
	src func(ctx context.Context, output chan<- T),
	opts ...Option,
) <-chan T {
	options := &options{
		ctx:        context.Background(),
		bufferSize: defaultBufferSize,
	}
	for _, opt := range opts {
		opt(options)
	}

	output := make(chan T, options.bufferSize)

	go func() {
		defer close(output)
		src(options.ctx, output)
	}()

	return output
}
