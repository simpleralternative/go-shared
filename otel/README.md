# otel - open telemetry
This library is a basic wrapper for common usages of the OTel SDK. The code found herein is derived from [https://opentelemetry.io/docs/languages/go/getting-started/](https://opentelemetry.io/docs/languages/go/getting-started/). 

**NOTE**: the documentation specifies that the SDK API isn't completely stable, so there may be times when any consumer of it must update their code. This library may fall behind and stop working until we notice and fix it.

## direct usage
The `otel/examples/` directory shows how to use the direct `Setup*` functions, along with the required cleanup function calls. If you want to make your functions available globally, refer to the link above for the functions to publish the 

## abstracted
We use the libraries by attaching them to the context, scoping them to the current process in a concurrent application. We have published these in named packages under the `otel/context/` directory.
