package stream

import (
	"context"
)

// Transform composes a providing channel with a premptive error check and a
// function that performs any action on a value and returns the result.
//
// This model, combined with basic channels as the interface, allows simple
// composition of functions as demonstrated in the test cases.
func Transform[T, U any](
	input <-chan *Result[T],
	transform func(ctx context.Context, input T) (U, error),
	opts ...Option,
) <-chan *Result[U] {
	return Stream(func(ctx context.Context, output chan<- *Result[U]) {
		for result := range input {
			if result.Error != nil {
				output <- NewResult[U](*new(U), result.Error)
			} else {
				output <- NewResult[U](transform(ctx, result.Value))
			}
		}
	}, opts...)
}
