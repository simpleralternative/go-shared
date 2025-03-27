If the regular README didn't cover the bases well enough, here's a slightly more
in-depth rendition:

# stream
This package provides simple helpers that focus on safe, concurrent data flow
patterns with minimal performance impacts. It intends to simplify expressing
data flows in our own processes and handles real-world use cases that we have
repeatedly encountered. There is no magic. Your project can do all of these
things without this package, and there are examples provided in the tests.

The main value of the model is automatic handling of channel-based ceremonies,
while enabling data flows as small, easily testable components.

## the argument AGAINST channels
Channels and goroutines are not free. Go's compile time call-site inlining makes
simple loops increadibly fast. You might be able to do 10_000_000_000 iterations
in a pure, nearly empty `for` loop but "only" a few tens of millions of
iterations over a channel per second. This package impacts them further by
requiring the use of a Result wrapper while doing row-by-row error and
done-signal checking in the transformers.

So why would you ever use channels, let alone `Stream`? What makes them worth
considering?

## the argument FOR channels
"Concurrency is not parallelism." Concurrency is structuring the code so that
the individual processes can be thought of as independent, using signals to
receive and transmit work. Breaking the code up into easy to understand - and
validate - components makes code much easier to make correct and maintainable.

Whether channel-based processes run serially, or in parallel, is then a matter
of environment. If it is able, by the presense of a capable system and enabled
by configuration, then concurrent processes will automatically begin to be
executed in parallel, trading a tiny amount of channel and goroutine overhead
for potentially large performance gains with no additional programming effort
required.

This package automates channel concurrency ceremonies to reduce errors. Channels
are automatically closed, errors are automatically forwarded down the stream,
and error handling can be consolidated.

## context and conventions

### producer-consumer
A producer should own the lifetime of a channel. All channel-source functions
return an output-only channel, and also close it when data has been exhausted
or a cancellation signal is received. This prevents any scenario where other
code causes an error by sending a value to a closed channel.

A consumer will simply read values from a channel until signaled to stop. Both
of these patterns are valid for receiving from channels, open AND closed:
```go
for value := range channel {
    println(value)
}

value, ok := <- channel
```
The loop will simply exit if there is nothing on the channel and it is closed,
while reading directly from an empty-and-closed channel will return the "zero"
value of the channel's data type. Since the zero-type may be a valid value for
your use case, adding the `ok` return value lets you check if the receive is
valid, indicating whether there is more work to do.

### configuration by "functional options"
This model enables simple runtime configuration of a process. Each function has
a reasonable default configuration and a set of matching options that modify its
behaviour.

### fan-out, fan-in
- **Fan-out**: Distribute work across multiple goroutines.
- **Fan-in**: Gather results from multiple sources into a single channel.

When a data stream benefits from being processed in parallel, then you can
`Distribute` the contents across multiple channels and work on them at the same
time.

Parallel data sources or processes are often combined into a single process for
downstream consumption. The `Multiplex` function consolidates multiple channels
into a single channel for further processing.

### transform
The Transform function extends the paradigm by using pure channels to embrace
Railway Oriented Programming. We can compose small, well-tested functions with
automatic error handling. Any error simply bypasses the rest of the operations
in a transform chain and can be read at the end.

Transform's internal function interfaces are just your value types and errors.
Data exchanged with a transforming process requires a Result type for error
checking. Examples are provided in the tests.

### batching
Many iterative workloads can have improved performance by batching the data.

### pipelining
Very few processes will have no cost beyond the loop itself. A unit of work that
takes 1 second will completely negate an extremely fast iterator. Pipelining is
using concurrency to logically divide the work so that different parts of it can
be processed at the same time. If the system is not allowed to be parallel,
the same code will behave as well as the direct loop. If it can - generally
available, but can be modified by configured by environment variables, process
code, and the hardware we run on - we gain the benefits of parallelism with
identical code.

A simple example would be if that 5 second of work was 5 steps that each took
1 seconds. The single-threaded case is straightforward with or without
concurrency:
```
unit 1: a(1) + b(1) + c(1) + d(1) + e(1) = 5s
unit 2: a(1) + b(1) + c(1) + d(1) + e(1) = 5s
```

With concurrency and a parallel-capable environment, you could pipeline the
process so that each of those 5 steps was processed by a different goroutine,
with the working data exchanged over channels. The process looks the same for
that a single unit, but as the first unit finishes the first subprocess, a
second unit of work can be started while the first unit moves on.
```
unit 1: a(1) + b(1) + c(1) + d(1) + e(1)
unit 2:        a(1) + b(1) + c(1) + d(1) + e(1)
unit 3:               a(1) + b(1) + c(1) + d(1) + e(1)
unit 4:                      a(1) + b(1) + c(1) + d(1) + e(1)
unit 5:                             a(1) + b(1) + c(1) + d(1) + e(1)
```

By the end of the first second, the first unit of work has completed as normal,
but 4 other processes are in flight and the average throughput is now 5/s with
no further code changes.

### parallelism
As shown above, concurrency can enable pipelining, but we can also divide and
distribute work at any stage.

In the previous example, if the work in each subprocess is not dependent on the
previous, we could split that work and do them all that the same time.
```
                     a(1)
                     b(1)
unit 1: Distribute < c(1) > Multiplex = 1s
                     d(1)
                     e(1)
```

In most cases, you'd use a combination. A stream is initiated where some source
of data is iterated over - query result, file contents, etc - and results are
sent to be worked downstream.

eg:
```
                    read file contents     extract data
iterate file list < read file contents > < extract data > accumulate and batch insert to sql
                    read file contents     extract data
```
