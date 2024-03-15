# API trace example

The example demonstrates how to use the GoVPP API trace functionality. The trace is bound to a connection.
The `core.Connection` includes the API trace object capable of recording API messages processed during the VPP API
message exchange. Traced messages are called `records`.

Each record contains information about the API message, its direction, time, and the channel ID.

The trace is disabled by default. In order to enable it, call `NewTrace`:

```go
c, err := govpp.Connect(*sockAddr)
if err != nil {
// handler error
}
size := 10
trace := core.NewTrace(c, size)
defer trace.Close()
```

The code above initializes the new tracer which records a certain number of messages defined by the `size`.

The following methods are available to call on the trace:

* `GetRecords() []*Record` returns all records beginning with the initialization of the trace till the point of the
  method call. The size also restricts the maximum number of records. All records received after the tracer is full are
  discarded.
* `GetRecordsForChannel(chId uint16) []*Record` works the same as the method above, but filters messages per channel.
* `Clear()` resets the tracer and allows to reuse it with (the size remains the same).
* `Close()` closes the tracer.