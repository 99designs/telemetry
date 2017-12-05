package datadog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/99designs/telemetry"
	"github.com/99designs/telemetry/collector"
)

// This test starts logging some fake metrics, but never terminates. Go check the metrics in datadog and manually kill the test.
func TestHTTPIntegration(t *testing.T) {
	key := os.Getenv("DD_KEY")
	if key == "" {
		t.Skip("DD_KEY must be specified in an env var to run integration tests")
	}

	sink := HTTP(key).(*DirectSink)
	sink.Client.Transport = tripperFunc(func(r *http.Request) (*http.Response, error) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		fmt.Println("Sending: " + string(b))

		return http.DefaultTransport.RoundTrip(r)
	})
	ctx := telemetry.NewContext(sink).SubContext("env:test", "app:telemetry")

	collector.Runtime(ctx)
	collector.CPU(ctx)
	collector.Mem(ctx)
	collector.Disk(ctx, "/")

	for {
		ctx.Histogram("telemetry.test.hist", rand.Float64())
		ctx.Timing("telemetry.test.timing", rand.Float64())
		ctx.Count("telemetry.test.count", float64(rand.Int()%3-1))
		ctx.Gauge("telemetry.test.gauge", rand.Float64())
		ctx.Set("telemetry.test.set", float64(rand.Int()&30))

		time.Sleep(100 * time.Millisecond)
	}
}
