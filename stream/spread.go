package stream

import "context"

// Spread is a pipeline "flattener". provide it a channel of slices of data and
// it will return a channel of individual items.
func Spread[T any](
	input <-chan []T,
	opts ...Option,
) <-chan T {
	return Stream(func(ctx context.Context, output chan<- T) {
		for items := range input {
			for _, item := range items {
				output <- item
			}
		}
	}, opts...)
}
