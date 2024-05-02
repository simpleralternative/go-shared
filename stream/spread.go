package stream

import "context"

// Spread is a pipeline "flattener". provide it a channel of slices of data and
// it will return a channel of individual items.
func Spread[T any](
	input <-chan *Result[[]T],
	opts ...Option,
) <-chan *Result[T] {
	return Stream(func(ctx context.Context, output chan<- *Result[T]) {
		for results := range input {
			if results.Error != nil {
				output <- NewResult(*new(T), results.Error)
			} else {
				for _, item := range results.Value {
					output <- NewResult(item, nil)
				}
			}
		}
	}, opts...)
}
