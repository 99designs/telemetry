package collector

import (
	"runtime"
	"runtime/debug"
	"time"
)

import "github.com/99designs/telemetry"

type RuntimeCollector struct {
	context *telemetry.Context
	running bool
}

// Runtime collects stats from the active runtime environment.
func Runtime(context *telemetry.Context) *RuntimeCollector {
	c := RuntimeCollector{
		context: context,
	}

	c.Start()
	return &c
}

// Start fires off a goroutine to periodically collect stats from the go runtime
func (d *RuntimeCollector) Start() {
	if d.running {
		return
	}

	d.running = true

	go func() {
		for d.running {
			mem := runtime.MemStats{}
			runtime.ReadMemStats(&mem)

			d.context.Gauge("app.runtime.goroutines", float64(runtime.NumGoroutine()))
			d.context.Gauge("app.runtime.cgo_calls", float64(runtime.NumCgoCall()))

			d.context.Gauge("app.mem.gc.pause_total_ns", float64(mem.PauseTotalNs))
			d.context.Gauge("app.mem.gc.num", float64(mem.NumGC))
			d.context.Gauge("app.mem.gc.next", float64(mem.NextGC))
			d.context.Gauge("app.mem.gc.cpu_fraction", mem.GCCPUFraction)

			gc := debug.GCStats{}
			gc.PauseQuantiles = make([]time.Duration, 3)
			debug.ReadGCStats(&gc)
			d.context.Gauge("app.mem.gc.gc_pause_quantile_50", float64(gc.PauseQuantiles[1]/1000)/1000.0)
			d.context.Gauge("app.mem.gc.gc_pause_quantile_max", float64(gc.PauseQuantiles[2]/1000)/1000.0)

			d.context.Gauge("app.mem.alloc", float64(mem.Alloc))
			d.context.Gauge("app.mem.heap_objects", float64(mem.HeapObjects))
			d.context.Gauge("app.mem.sys", float64(mem.Sys))
			d.context.Gauge("app.mem.active_allocs", float64(mem.Mallocs-mem.Frees))

			time.Sleep(5 * time.Second)
		}
	}()
}

func (d *RuntimeCollector) Stop() {
	d.running = false
}
