package stream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiplex(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		channels := []<-chan int{
			Stream(func(output chan<- int) {
				output <- 1
				output <- 2
			}),
			Stream(func(output chan<- int) {
				output <- 3
				output <- 4
			}),
		}

		outputs := []int{}
		for value := range Multiplex(channels) {
			outputs = append(outputs, value)
		}

		require.ElementsMatch(t, []int{1, 2, 3, 4}, outputs)
	})
}
