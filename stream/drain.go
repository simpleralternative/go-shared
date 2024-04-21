package stream

// Drain is a convenience function to remove all messages from a channel.
func Drain[T any](input <-chan T) {
	for range input {
	}
}
