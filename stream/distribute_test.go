package stream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDistribute(t *testing.T) {
	channels := Distribute(Stream(func(output chan<- int) {
		output <- 1
		output <- 2
	}), 2)

	actual := []int{<-channels[0], <-channels[1]}
	require.ElementsMatch(t, []int{1, 2}, actual)
}
