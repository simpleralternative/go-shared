package stream

import "context"

// defaultBufferSize provides a reasonable value for channel buffer sizes.
const defaultBufferSize uint16 = 1000

// options defines the possible configurations that may be modified via the
// functional options pattern.
type options struct {
	ctx        context.Context
	bufferSize uint16
}

// Option is a function that modifies the Options values.
type Option func(*options)

// WithContext adds the supplied context to the options.
func WithContext(ctx context.Context) Option {
	return func(opts *options) {
		opts.ctx = ctx
	}
}

// WithBufferSize sets the buffer size of the output channels.
func WithBufferSize(size uint16) Option {
	return func(opts *options) {
		opts.bufferSize = size
	}
}
