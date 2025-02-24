package stream

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFold(t *testing.T) {
	t.Run("basic math", func(t *testing.T) {
		data := Stream(func(ctx context.Context, output chan<- *Result[int]) {
			output <- NewResult(1, nil)
			output <- NewResult(2, nil)
			output <- NewResult(1, nil)
			output <- NewResult(2, nil)
			output <- NewResult(-1, nil)
		})
		total, err := Fold(data, 0, func(total int, value int) (int, error) {
			return total + value, nil
		})
		require.NoError(t, err)
		require.Equal(t, 5, total)
	})

	t.Run("side effect", func(t *testing.T) {
		data := Stream(func(ctx context.Context, output chan<- *Result[string]) {
			output <- NewResult("a", nil)
			output <- NewResult("b", nil)
			output <- NewResult("a", nil)
			output <- NewResult("b", nil)
			output <- NewResult("c", nil)
		})

		m := map[string]int{}
		Fold(
			data,
			m,
			func(total map[string]int, value string) (map[string]int, error) {
				if _, exists := total[value]; !exists {
					total[value] = 0
				}
				total[value]++
				return total, nil
			},
		)
		require.Equal(t, map[string]int{"a": 2, "b": 2, "c": 1}, m)
	})

	t.Run("closure", func(t *testing.T) {
		data := Stream(func(ctx context.Context, output chan<- *Result[string]) {
			output <- NewResult("a", nil)
			output <- NewResult("b", nil)
			output <- NewResult("c", nil)
			output <- NewResult("d", nil)
			output <- NewResult("e", nil)
		})
		builder := &strings.Builder{}
		_, err := Fold(
			data,
			struct{}{},
			func(_ struct{}, value string) (struct{}, error) {
				_, err := builder.WriteString(value)
				return struct{}{}, err
			},
		)
		require.NoError(t, err)
		require.Equal(t, "abcde", builder.String())
	})

	t.Run("stream error", func(t *testing.T) {
		e := errors.New("failure")
		data := Stream(func(ctx context.Context, output chan<- *Result[string]) {
			output <- NewResult("a", nil)
			output <- NewResult("b", nil)
			output <- NewResult("c", nil)
			output <- NewResult("d", e)
			output <- NewResult("e", nil)
		})
		_, err := Fold(
			data,
			&strings.Builder{},
			func(total *strings.Builder, value string) (*strings.Builder, error) {
				_, err := total.WriteString(value)
				return total, err
			},
		)
		require.Equal(t, "failure", err.Error())
	})

	t.Run("fold error", func(t *testing.T) {
		e := errors.New("failure")
		data := Stream(func(ctx context.Context, output chan<- *Result[string]) {
			output <- NewResult("a", nil)
			output <- NewResult("b", nil)
			output <- NewResult("c", nil)
			output <- NewResult("d", nil)
			output <- NewResult("e", nil)
		})
		_, err := Fold(
			data,
			&strings.Builder{},
			func(total *strings.Builder, value string) (*strings.Builder, error) {
				return nil, e
			},
		)
		require.Equal(t, "failure", err.Error())
	})
}
