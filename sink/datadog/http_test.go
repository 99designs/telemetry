package datadog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/99designs/telemetry"
	"github.com/stretchr/testify/require"
)

type tripperFunc func(*http.Request) (*http.Response, error)

func (t tripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return t(r) }

func TestSink(t *testing.T) {
	sink := testSink()
	ctx := telemetry.NewContext(sink).SubContext("tag:test")

	t.Run("test rate", func(t *testing.T) {
		sink.Count(ctx, "test.rate", 1)
		sink.Count(ctx, "test.rate", 1)
		sink.Count(ctx, "test.rate", -1)
		sink.Count(ctx, "test.rate", 1)
		pumpAll(sink)

		require.Equal(t, "test.rate", sink.buildBatch().Series[0].Metric)
		require.Equal(t, 2.0, sink.buildBatch().Series[0].Points[time.Now().Unix()])
	})

	sink.Reset()

	t.Run("test gauge", func(t *testing.T) {
		sink.Gauge(ctx, "test.gauge", 12.3)
		sink.Gauge(ctx, "test.gauge", 12.4)
		pumpAll(sink)

		require.Equal(t, "test.gauge", sink.buildBatch().Series[0].Metric)
		require.Equal(t, 12.4, sink.buildBatch().Series[0].Points[time.Now().Unix()])
	})

	sink.Reset()

	t.Run("test set", func(t *testing.T) {
		sink.Set(ctx, "test.set", 1)
		sink.Set(ctx, "test.set", 2)
		sink.Set(ctx, "test.set", 2)
		sink.Set(ctx, "test.set", 5)
		pumpAll(sink)

		require.Equal(t, "test.set", sink.buildBatch().Series[0].Metric)
		require.Equal(t, 3.0, sink.buildBatch().Series[0].Points[time.Now().Unix()])
	})

	sink.Reset()

	t.Run("test hist", func(t *testing.T) {
		sink.Histogram(ctx, "test.hist", 1)
		sink.Histogram(ctx, "test.hist", 2)
		sink.Histogram(ctx, "test.hist", 2)
		sink.Histogram(ctx, "test.hist", 3)
		pumpAll(sink)

		series := sink.buildBatch().Series

		require.Equal(t, "test.hist.min", series[0].Metric)
		require.Equal(t, 1.0, series[0].Points[time.Now().Unix()])
		require.Equal(t, "test.hist.max", series[1].Metric)
		require.Equal(t, 3.0, series[1].Points[time.Now().Unix()])

		require.Equal(t, "test.hist.count", series[2].Metric)
		require.Equal(t, 4.0, series[2].Points[time.Now().Unix()])
		require.Equal(t, "test.hist.avg", series[3].Metric)
		require.Equal(t, 2.0, series[3].Points[time.Now().Unix()])

	})

	sink.Reset()

	t.Run("test serialize", func(t *testing.T) {
		sink.Count(ctx, "test.rate", 1)
		sink.Count(ctx, "test.rate", 1)
		sink.Count(ctx, "test.rate", -1)
		sink.Count(ctx, "test.rate", 1)

		sink.Set(ctx, "test.set", 1)
		sink.Set(ctx, "test.set", 2)
		sink.Set(ctx, "test.set", 2)
		sink.Set(ctx, "test.set", 5)

		sink.Gauge(ctx, "test.gauge", 12.3)
		sink.Gauge(ctx, "test.gauge", 12.4)

		sink.Histogram(ctx, "test.hist", 1)
		sink.Histogram(ctx, "test.hist", 2)
		sink.Histogram(ctx, "test.hist", 2)
		sink.Histogram(ctx, "test.hist", 3)

		sink.Timing(ctx, "test.timing", 1)
		sink.Timing(ctx, "test.timing", 2)
		sink.Timing(ctx, "test.timing", 2)
		sink.Timing(ctx, "test.timing", 3)

		pumpAll(sink)

		sink.Client.Transport = tripperFunc(func(r *http.Request) (*http.Response, error) {
			b, _ := ioutil.ReadAll(r.Body)

			bs := BatchSeries{}
			err := json.Unmarshal(b, &bs)
			if err != nil {
				panic(err)
			}

			now := time.Now().Unix()

			expected := BatchSeries{
				Series: []Series{
					{
						Metric: "test.timing.min",
						Points: Points{now: 1},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.timing.max",
						Points: Points{now: 3},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.timing.count",
						Points: Points{now: 4},
						Type:   "rate",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.timing.avg",
						Points: Points{now: 2},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.set",
						Points: Points{now: 3},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.rate",
						Points: Points{now: 2},
						Type:   "rate",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.hist.min",
						Points: Points{now: 1},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.hist.max",
						Points: Points{now: 3},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.hist.count",
						Points: Points{now: 4},
						Type:   "rate",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.hist.avg",
						Points: Points{now: 2},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					}, {
						Metric: "test.gauge",
						Points: Points{now: 12.4},
						Type:   "gauge",
						Tags:   []string{"tag:test"},
						Host:   hostname,
					},
				},
			}

			require.EqualValues(t, expected, bs)

			return &http.Response{
				Body: ioutil.NopCloser(&bytes.Buffer{}),
			}, nil
		})

		sink.send()
	})
}

func testSink() *DirectSink {
	sink := &DirectSink{
		metrics: make(chan singlePoint, 1000),
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
	sink.Reset()

	return sink
}

func pumpAll(ds *DirectSink) {
	n := 0
	for {
		select {
		case point := <-ds.metrics:
			ds.addPoint(point)
			n++
		default:
			if n == 0 {
				panic("nothing in channel to pump!")
			}
			return
		}
	}
}
