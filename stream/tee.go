package stream

import (
	"context"
)

// Tee copies messages to two output channels.
// It only does a simple copy, so pointers, and nested pointers, will
// both reference the same original memory.
// Standard caveats apply.
//
// NOTE: the output channels may be processed at different rates and the slowest
// process - source, target 1, or target 2, may govern the overall throughput.
// Set buffersizes to match expectations.
//
// Trivia: Tee refers to a 90degree split in pipes, as incorporated into Linux.
func Tee[T any](input <-chan T, opts ...Option) (<-chan T, <-chan T) {
	options := &options{
		ctx:        context.Background(),
		bufferSize: defaultBufferSize,
	}
	for _, opt := range opts {
		opt(options)
	}

	out1 := make(chan T, options.bufferSize)
	out2 := make(chan T, options.bufferSize)
	go func() {
		defer close(out1)
		defer close(out2)
		for msg := range input {
			select {
			case <-options.ctx.Done():
				return
			default:
				out1 <- msg
				out2 <- msg
			}

		}
	}()

	return out1, out2
}
