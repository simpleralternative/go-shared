# stream
This is the short version. The reasoning and rationals for why you might want to
use this package is provided in [README-rationale.md](README-rationale.md).

Everything in this package is about adding convenience to standard Go channels.
You can use Stream to generate one, or start with your own, then use any of the
library functions to process that channel and consume it directly or via another
library function.

Channels do add some overhead, and this library adds extra overhead with
built-in sanity checks and other library boilerplate, but if you're using
channels, this can help you standardize the way you interact with them and
minimize footguns for minimal performance impact. 

### install
`> go get github.com/simpleralternative/go-shared/stream`

### very basic usage
```go
// provides control: you can add values, cancel signals, and deadlines.
outerCtx := context.Background()

// Stream creates, backgrounds, and automatically closes the channel when done.
results := stream.Stream(
    // this could also be a named function that you use as a component.
    func(ctx context.Context, output <-chan *result.Result[myStruct]) {
        // any kind of data generation process. database, file, etc.
        for range 1000 {
            // provide as many or as few results as needed
            output <- stream.NewResult(makeMyStruct(ctx), nil)
            // Trace adds a wrapping error with this callsite noted.
            output <- stream.NewResult(myStruct{}, stream.Trace(ErrMine))
        }
    },
    // an optional configuration. by default a generic context is passed in.
    WithContext(outerCtx), 
)

parallel := stream.Distribute(results, 5)
processed := stream.Process(parallel, processingFunctionFromMyLibrary)
serial := stream.Multiplex(processed)

for _, value := range serial {
    mine, err := value.Destructure()
    // etc
}
```

Consult the rationale and tests for more.
