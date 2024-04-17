package stream

import "context"

// Stream is a convenience function that performs the basic channel ceremonies
// for a standard producer-consumer paradigm.
//
// it takes a function that will be given a send-only channel. the function may
// then do whatever logic is needed to produce the values that are put into the
// channel.
//
// when the provided function completes, the output channel is closed and
// consumers can drain the channel as normal.
//
// Examples are provided in the tests.
func Stream[T any](
	src func(ctx context.Context, output chan<- T),
	opts ...Option,
) <-chan T {
	options := &Options{
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
