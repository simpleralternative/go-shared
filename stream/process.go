package stream

import "context"

// Processor distributes a workload over a slice of channels.
//
// the process parameter is a middleware-style function that injects the
// distribution index and returns a function matching the Transform signature.
func Processor[T, U any](
	inputs []<-chan *Result[T],
	process func(id int) func(ctx context.Context, input T) (U, error),
	opts ...Option,
) []<-chan *Result[U] {
	outputs := make([]<-chan *Result[U], len(inputs))

	for id, input := range inputs {
		outputs[id] = Transform(input, process(id), opts...)
	}

	return outputs
}
