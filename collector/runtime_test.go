package collector

import (
	"github.com/99designs/telemetry"
	"github.com/99designs/telemetry/sink"

	"testing"
	"time"
)

func TestRuntimeCollector(t *testing.T) {
	s := sink.Test()

	rt := Runtime(telemetry.NewContext(s))
	time.Sleep(50 * time.Millisecond)
	rt.Stop()

	if _, ok := s["app.runtime.goroutines"]; !ok {
		t.Error("Missing metric app.runtime.goroutines")
	}
}
