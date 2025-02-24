package stream

import (
	"errors"
	"fmt"
	"runtime"
)

// Trace wraps an incoming error with an error noting the source of the error or
// returns the unmodified input value.
func Trace[T any](value T, err error) (T, error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return value, errors.Join(
			err,
			fmt.Errorf("stream.Trace - %s:%d", file, line),
		)
	}

	return value, nil
}
