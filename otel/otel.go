package otel

import (
	"context"
	"errors"
)

var ErrConfigurationRequired = errors.New("configuration not provided")

type shutdownFunc func(context.Context) error

// ShutdownFuncs consolidates closures from the reference code to enable
// refactoring the code for clarity.
type ShutdownFuncs struct {
	funcs []shutdownFunc
}

func (sdf *ShutdownFuncs) append(f shutdownFunc) {
	sdf.funcs = append(sdf.funcs, f)
}

// Shutdown calls cleanup functions registered via shutdownFuncs.
// The errors from the calls are joined.
// Each registered cleanup will be invoked once.
func (sdf *ShutdownFuncs) Shutdown(ctx context.Context) error {

	var err error
	for _, fn := range sdf.funcs {
		err = errors.Join(err, fn(ctx))
	}
	sdf.funcs = nil
	return err
}

// handleErr is a convenience function that calls shutdown for cleanup and makes
// sure that all errors are returned.
func (sdf *ShutdownFuncs) handleErr(
	ctx context.Context,
	err error,
) (context.Context, shutdownFunc, error) {
	return ctx, sdf.Shutdown, errors.Join(err, sdf.Shutdown(ctx))
}
