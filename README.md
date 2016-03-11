telemetry
=========

Telemetry is a collection of performance monitoring collectors and sinks. It
is inspired by https://github.com/gocraft/health but written to take advantage
of datadogs tagging.



Sinks
=====
The only supported sink currently is datadog:
```go
t := telemetry.New("app:myapp")
t.AddSink(sink.Datadog("mydatadog.local", 1234))
```


Collectors
==========
Runtime collector fetches stats from the golang runtime environment, eg memory
usage and active goroutines.

runtime
```go
collector.Runtime(t)
```

gorilla
```go
// Middleware to collect timing for named routes from gorilla mux
router := mux.NewRouter()
router.Handle("/", handler).Name("test.route")

handler := collector.Gorilla(t, router, func(w http.ResponseWriter, r *http.Request) {
    // You can log your own metrics from any handler after this.
    ContextForRequest(r).Gauge("test_metric", 2)
})
```
