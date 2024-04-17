package stream

type Result[T any] struct {
	Value T
	Error error
}

func NewResult[T any](value T, err error) *Result[T] {
	return &Result[T]{value, err}
}

func (res *Result[T]) Destructure() (T, error) {
	return res.Value, res.Error
}
