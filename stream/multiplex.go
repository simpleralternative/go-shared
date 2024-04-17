package stream

import (
	"context"
	"sync"
)

// Multiplex standardizes the fan-in case, where multiple process outputs are
// merged into a single channel for simple processing.
func Multiplex[T any](inputs []<-chan T, opts ...Option) <-chan T {
	return Stream(func(ctx context.Context, output chan<- T) {
		var wg sync.WaitGroup
		wg.Add(len(inputs))
		for _, input := range inputs {
			go func() {
				defer wg.Done()
				for value := range input {
					select {
					case <-ctx.Done():
						return
					case output <- value:
					}
				}
			}()
		}
		wg.Wait()
	}, opts...)
}
