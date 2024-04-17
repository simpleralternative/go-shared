package stream

import (
	"context"
)

// Transform composes a providing channel with a premptive error check and a
// function that performs any action on a value and returns the result.
func Transform[T, U any](
	input <-chan *Result[T],
	step func(ctx context.Context, res *Result[T]) *Result[U],
	opts ...Option,
) <-chan *Result[U] {
	return Stream(func(ctx context.Context, output chan<- *Result[U]) {
		for result := range input {
			if result.Error != nil {
				output <- NewResult[U](*new(U), result.Error)
			}

			output <- step(ctx, result)
		}
	}, opts...)
}
