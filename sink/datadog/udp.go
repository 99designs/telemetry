package datadog

import (
	"log"

	"github.com/99designs/telemetry"
	"github.com/theckman/godspeed-og"
)

type DatadogSink struct {
	god *godspeed.Godspeed
}

// UDP pushes metrics to local dogstatsd / datadog agent
func UDP(host string, port int) (telemetry.Sink, error) {
	god, err := godspeed.New(host, port, false)
	if err != nil {
		return nil, err
	}

	return &DatadogSink{
		god: god,
	}, nil
}

// Count adds a value to a stat
func (d *DatadogSink) Count(c *telemetry.Context, stat string, count float64) {
	err := d.god.Count(stat, count, c.Tags())
	if err != nil {
		log.Printf("Unable to send stats: " + err.Error())
	}
}

// Gague sends an absolute value. Useful for tracking things like memory.
func (d *DatadogSink) Gauge(c *telemetry.Context, stat string, value float64) {
	err := d.god.Gauge(stat, value, c.Tags())
	if err != nil {
		log.Printf("Unable to send stats: " + err.Error())
	}
}

// Histogram measures the statistical distribution of a set of values. eg query time
func (d *DatadogSink) Histogram(c *telemetry.Context, stat string, value float64) {
	err := d.god.Histogram(stat, value, c.Tags())
	if err != nil {
		log.Printf("Unable to send stats: " + err.Error())
	}
}

// Timing is a special subclass of Histgram for timing information.
func (d *DatadogSink) Timing(c *telemetry.Context, stat string, value float64) {
	err := d.god.Timing(stat, value, c.Tags())
	if err != nil {
		log.Printf("Unable to send stats: " + err.Error())
	}
}

// Set counts unique values. Send user id to monitor unique users.
func (d *DatadogSink) Set(c *telemetry.Context, stat string, value float64) {
	err := d.god.Set(stat, value, c.Tags())
	if err != nil {
		log.Printf("Unable to send stats: " + err.Error())
	}
}
