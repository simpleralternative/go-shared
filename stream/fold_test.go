package stream

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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
		total, err := Fold(
			data,
			0,
			func(_ context.Context, total int, value int) (int, error) {
				return total + value, nil
			},
		)
		require.NoError(t, err)
		require.Equal(t, 5, total)
	})

	t.Run("side effect", func(t *testing.T) {
		data := Stream(
			func(ctx context.Context, output chan<- *Result[string]) {
				output <- NewResult("a", nil)
				output <- NewResult("b", nil)
				output <- NewResult("a", nil)
				output <- NewResult("b", nil)
				output <- NewResult("c", nil)
			},
		)

		m := map[string]int{}
		Fold(
			data,
			m,
			func(
				_ context.Context,
				total map[string]int,
				value string,
			) (map[string]int, error) {
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
		data := Stream(
			func(ctx context.Context, output chan<- *Result[string]) {
				output <- NewResult("a", nil)
				output <- NewResult("b", nil)
				output <- NewResult("c", nil)
				output <- NewResult("d", nil)
				output <- NewResult("e", nil)
			},
		)
		builder := &strings.Builder{}
		_, err := Fold(
			data,
			struct{}{},
			func(
				_ context.Context,
				_ struct{},
				value string,
			) (struct{}, error) {
				_, err := builder.WriteString(value)
				return struct{}{}, err
			},
		)
		require.NoError(t, err)
		require.Equal(t, "abcde", builder.String())
	})

	t.Run("stream error", func(t *testing.T) {
		e := errors.New("failure")
		data := Stream(
			func(ctx context.Context, output chan<- *Result[string]) {
				output <- NewResult("a", nil)
				output <- NewResult("b", nil)
				output <- NewResult("c", nil)
				output <- NewResult("d", e)
				output <- NewResult("e", nil)
			},
		)
		_, err := Fold(
			data,
			&strings.Builder{},
			func(
				_ context.Context,
				total *strings.Builder,
				value string,
			) (*strings.Builder, error) {
				_, err := total.WriteString(value)
				return total, err
			},
		)
		require.Equal(t, "failure", err.Error())
	})

	t.Run("fold error", func(t *testing.T) {
		e := errors.New("failure")
		data := Stream(
			func(ctx context.Context, output chan<- *Result[string]) {
				output <- NewResult("a", nil)
				output <- NewResult("b", nil)
				output <- NewResult("c", nil)
				output <- NewResult("d", nil)
				output <- NewResult("e", nil)
			},
		)
		_, err := Fold(
			data,
			&strings.Builder{},
			func(
				_ context.Context,
				total *strings.Builder,
				value string,
			) (*strings.Builder, error) {
				return nil, e
			},
		)
		require.Equal(t, "failure", err.Error())
	})

	t.Run("context canceled", func(t *testing.T) {
		e := errors.New("timeout")
		data := Stream(
			func(ctx context.Context, output chan<- *Result[string]) {
				output <- NewResult("a", nil)
				output <- NewResult("b", nil)
				output <- NewResult("c", nil)
				output <- NewResult("d", nil)
				output <- NewResult("e", nil)
			},
		)
		ctx, cancel := context.WithDeadline(
			context.Background(),
			time.Now().Add(time.Second),
		)
		defer cancel()

		total, err := Fold(
			data,
			&strings.Builder{},
			func(
				ctx context.Context,
				total *strings.Builder,
				value string,
			) (*strings.Builder, error) {
				select {
				case <-ctx.Done():
					return total, e
				case <-time.After(2 * time.Second):
					total.WriteString("mistake")
					return total, nil
				}
			},
			WithContext(ctx),
		)
		require.Equal(t, "timeout", err.Error())
		require.Equal(t, "", total.String())
	})
}
