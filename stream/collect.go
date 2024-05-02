package stream

import "context"

// Collect is a complete pipeline accumulator.
//
// Similar to Batch, it will build a slice of outputs. Unlike Batch, it is all
// or nothing. An error will start a background Drain and send an error Result.
//
// In the interest of performance,
// Do not use Collect on unbounded streams.
func Collect[T any](
	input <-chan *Result[T],
	opts ...Option,
) <-chan *Result[[]T] {
	return Stream(func(ctx context.Context, output chan<- *Result[[]T]) {
		var acc []T
		for result := range input {
			select {
			case <-ctx.Done():
				return
			default:
				if result.Error != nil {
					output <- NewResult[[]T](nil, result.Error)
					go Drain(input)
					return
				}
				acc = append(acc, result.Value)
			}
		}
		output <- NewResult(acc, nil)
	}, opts...)
}
