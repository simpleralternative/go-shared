package stream

// Fold is a generic pipeline accumulator that performs the aggregator function
// over the contents of the input channel, accumulating the results each row,
// beginning with the initialValue.
//
// The accumulated value, or an error passed through the stream, will be
// returned when complete.
func Fold[T, U any](
	input <-chan *Result[T],
	initialValue U,
	aggregator func(accumulator U, value T) (U, error),
) (U, error) {
	// a successful process will simply have nothing to drain.
	defer func() {
		go Drain(input)
	}()

	accumulator := initialValue
	var err error
	for result := range input {
		if result.Error != nil {
			return accumulator, result.Error
		}
		accumulator, err = aggregator(accumulator, result.Value)
		if err != nil {
			return accumulator, err
		}
	}

	return accumulator, nil
}
