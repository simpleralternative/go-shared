package stream

import "context"

// Batch is a pipeline accumulator.
//
// many operations - io and cpu, from sql queries, to file operations, to
// network calls - are more efficient when performed on grouped data. this
// function enables streams to take advantage of that by accumulating
// incoming data and returning a channel of results of slices of the data.
//
// note: to balance between performance and control, a select is included at the
// send, but nowhere else.
func Batch[T any](
	input <-chan *Result[T],
	batchSize int,
	opts ...Option,
) <-chan *Result[[]T] {
	return Stream(func(ctx context.Context, output chan<- *Result[[]T]) {
		var acc []T
		for result := range input {
			if result.Error != nil {
				output <- NewResult[[]T](nil, result.Error)
			} else {
				if len(acc) >= batchSize {
					select {
					case <-ctx.Done():
						return
					case output <- NewResult(acc, nil):
					}
					acc = nil
				}
				acc = append(acc, result.Value)
			}
		}
		if len(acc) > 0 {
			output <- NewResult(acc, nil)
		}
	}, opts...)
}
