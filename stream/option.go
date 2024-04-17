package stream

import "context"

// defaultBufferSize provides a reasonable value for channel buffer sizes.
const defaultBufferSize uint16 = 1000

// Options defines the possible configurations that may be modified via the
// functional options pattern.
type Options struct {
	ctx        context.Context
	bufferSize uint16
}

// Option is a function that modifies the Options values.
type Option func(*Options)

// WithContext adds the supplied context to the options.
func WithContext(ctx context.Context) Option {
	return func(opts *Options) {
		opts.ctx = ctx
	}
}

// WithBufferSize sets the buffer size of the output channels.
func WithBufferSize(size uint16) Option {
	return func(opts *Options) {
		opts.bufferSize = size
	}
}
