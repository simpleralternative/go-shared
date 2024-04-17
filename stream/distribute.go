package stream

import "context"

// Distribute standardizes the fan-out case, where a single producer generates
// data to be handled by channelCount concurrent downstream processes.
func Distribute[T any](
	input <-chan T,
	channelCount int,
	opts ...Option,
) []<-chan T {
	outputs := make([]<-chan T, channelCount)

	for i := range channelCount {
		outputs[i] = Stream(func(ctx context.Context, output chan<- T) {
			for value := range input {
				select {
				case <-ctx.Done():
					return
				case output <- value:
				}
			}
		}, opts...)
	}

	return outputs
}
