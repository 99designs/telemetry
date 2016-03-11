package telemetry

type Context struct {
	sinks []Sink
	tags  []string
}

func NewContext(sinks ...Sink) *Context {
	return &Context{
		sinks: sinks,
	}
}

// Subcontext creates a new context with the given tags
func (c *Context) SubContext(tags ...string) *Context {
	return &Context{
		sinks: c.sinks,
		tags:  append(c.tags, tags...),
	}
}

func (c *Context) Tags() []string {
	return c.tags
}

// AddSink adds a new sink destination
func (c *Context) AddSink(sinks ...Sink) {
	c.sinks = append(c.sinks, sinks...)
}

// Count adds a value to a stat
func (c *Context) Count(stat string, count float64) {
	for _, sink := range c.sinks {
		sink.Count(c, stat, count)
	}
}

// Incr Adds one to a stat
func (c *Context) Incr(stat string) {
	for _, sink := range c.sinks {
		sink.Count(c, stat, 1)
	}
}

// Decr removes one from a stat
func (c *Context) Decr(stat string) {
	for _, sink := range c.sinks {
		sink.Count(c, stat, -1)
	}
}

// Gague sends an absolute value. Useful for tracking things like memory.
func (c *Context) Gauge(stat string, value float64) {
	for _, sink := range c.sinks {
		sink.Gauge(c, stat, value)
	}
}

// Histogram measures the statistical distribution of a set of values. eg query time
func (c *Context) Histogram(stat string, value float64) {
	for _, sink := range c.sinks {
		sink.Histogram(c, stat, value)
	}
}

// Timing is a special subclass of Histgram for timing information.
func (c *Context) Timing(stat string, value float64) {
	for _, sink := range c.sinks {
		sink.Timing(c, stat, value)
	}
}

// Set counts unique values. Send user id to monitor unique users.
func (c *Context) Set(stat string, value float64) {
	for _, sink := range c.sinks {
		sink.Set(c, stat, value)
	}
}
