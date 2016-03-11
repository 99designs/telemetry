package sink

import "github.com/99designs/telemetry"

type TestSink map[string]TestMetric

func Test() TestSink {
	return TestSink{}
}

type TestMetric struct {
	Stat  string
	Value float64
	Tags  []string
}

// Count adds a value to a stat
func (d TestSink) Count(c *telemetry.Context, stat string, count float64) {
	d[stat] = TestMetric{"Count", count, c.Tags()}
}

// Gague sends an absolute value. Useful for tracking things like memory.
func (d TestSink) Gauge(c *telemetry.Context, stat string, value float64) {
	d[stat] = TestMetric{"Gauge", value, c.Tags()}
}

// Histogram measures the statistical distribution of a set of values. eg query time
func (d TestSink) Histogram(c *telemetry.Context, stat string, value float64) {
	d[stat] = TestMetric{"Histogram", value, c.Tags()}
}

// Timing is a special subclass of Histgram for timing information.
func (d TestSink) Timing(c *telemetry.Context, stat string, value float64) {
	d[stat] = TestMetric{"Timing", value, c.Tags()}
}

// Set counts unique values. Send user id to monitor unique users.
func (d TestSink) Set(c *telemetry.Context, stat string, value float64) {
	d[stat] = TestMetric{"Set", value, c.Tags()}
}
