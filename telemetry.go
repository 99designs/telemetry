// Pushes metrics to one or more sinks
package telemetry

type Sink interface {
	// Count adds a value to a metric
	Count(context *Context, stat string, count float64)

	// Gauge sends an absolute value. Useful for tracking things like memory.
	Gauge(context *Context, stat string, value float64)

	// Histogram measures the statistical distribution of a set of values. eg query time
	Histogram(context *Context, stat string, value float64)

	// Timing is a special subclass of Histgram for timing information.
	Timing(context *Context, stat string, value float64)

	// Set counts unique values. Send user id to monitor unique users.
	Set(context *Context, stat string, value float64)
}

func New(tags ...string) *Context {
	return &Context{
		sinks: []Sink{},
		tags:  tags,
	}
}
