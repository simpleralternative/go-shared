# the argument for channels
Channels and goroutines are not free. Go's compile time call-site inlining makes
simple loops increadibly fast. You might be able to do 10_000_000_000 iterations
in a pure for loop but "only" 50_000_000 iterations over a channel per second.
So what makes them worth considering?

"Concurrency is not parallelism." Concurrency is structuring the code so that
the individual processes can be thought of as independent, using signals to
receive and transmit work. Whether they process serially or in parallel, is then
a matter of environment. If it is able, by the presense of a capable system and
enabled by configuration, then concurrent processes will default to also being
parallel processes, trading a tiny amount of channel and goroutine overhead for
potentially large performance gains.

### pipelining
Very few processes will have no cost beyond the loop itself. A unit of work that
takes 1 second will completely negate that extremely fast iterator. Pipelining
is using concurrency to logically divide the work so that different parts of it
can be processed at the same time. If the system is not allowed to be parallel,
the same code will behave as well as the direct loop, but if it can we gain the
benefits of parallelism with identical code.

A simple example would be if that 1 second of work was 5 steps that each took
0.2 seconds. The single-threaded case is straightforward with or without
concurrency:
```
unit 1: a(0.2) + b(0.2) + c(0.2) + d(0.2) + e(0.2) = 1s
unit 2: a(0.2) + b(0.2) + c(0.2) + d(0.2) + e(0.2) = 1s
...
```

With concurrency and a parallel-capable environment, you could pipeline the
process so that each of those 5 steps was processed by a different goroutine,
with the working data exchanged over channels. The process looks the same for
that a single unit, but as the first unit finishes the first subprocess, a
second unit of work can be started while the first unit moves on.
```
unit 1: a(0.2) + b(0.2) + c(0.2) + d(0.2) + e(0.2)
unit 2:          a(0.2) + b(0.2) + c(0.2) + d(0.2) + e(0.2)
unit 3:                   a(0.2) + b(0.2) + c(0.2) + d(0.2) + e(0.2)
unit 4:                            a(0.2) + b(0.2) + c(0.2) + d(0.2) + e(0.2)
unit 5:                                     a(0.2) + b(0.2) + c(0.2) + d(0.2) + e(0.2)
...
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
                     a(0.2)
                     b(0.2)
unit 1: distribute < c(0.2) > multiplex = 0.2s
                     d(0.2)
                     e(0.2)
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

# transform
The Transform function extends the paradigm by using pure channels to embrace
Railway Oriented Programming. We can compose small, well-tested functions with
automatic error handling. Any error simply bypasses the rest of the operations
in a transform chain.

Transform's internal function interfaces are just your value types and errors.
Data exchanged with a transforming process requires a Result type for error
checking. Check the tests for examples.
