# stream
This package provides simple helpers that enforce safe, concurrent data flow
patterns.

## producer-consumer conventions
A producer should own the lifetime of a channel. All channel-source functions
return an output-only channel, and also close it when data has been exhausted
or a cancellation signal is received. This prevents any scenario where other
code causes an error by sending a value to a closed channel.

A consumer will simply read values from a channel until signaled to stop. Both a
`for value := range channel` and the second form of equals
`value, ok := <- channel` can be used to safely recognise when the channel has
been closed, indicating there is no more work to do.

## functional options
This model enables simple runtime configuration of a process. Each function has
a simple default configuration and a set of matching options that modify its
behaviour.

## fan-out, fan-in
When a data stream benefits from being processed in parallel, then you can
distribute the contents across multiple channels and work on them in parallel.

Parallel data sources or processes are often combined into a single process.
The multiplexer consolidates multiple channels to a single channel.