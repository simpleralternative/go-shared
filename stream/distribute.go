package stream

// Distribute standardizes the fan-out case, where a single producer generates
// data to be handled by multiple concurrent downstream processes.
func Distribute[T any](input <-chan T, count int) []<-chan T {
	outputs := make([]<-chan T, count)

	for i := range count {
		outputs[i] = Stream(func(output chan<- T) {
			for value := range input {
				output <- value
			}
		})
	}

	return outputs
}
