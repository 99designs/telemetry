package datadog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"os"

	"github.com/99designs/telemetry"
)

const MAX_POINTS = 100000

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

type DirectSink struct {
	metrics chan singlePoint
	key     string
	Client  http.Client
	points  int64 `json:"-"`

	// [metric name][tags]points
	buffer map[string]map[string]points
}

type points struct {
	// [time]: []samples
	Values map[int64][]float64
	Type   string
}

type singlePoint struct {
	Metric string
	Time   int64
	Value  float64
	Type   string
	Tags   []string
}

// HTTP pushes metrics directly to the datadog http apis. No agent is required.
func HTTP(key string) telemetry.Sink {
	sink := &DirectSink{
		metrics: make(chan singlePoint, 1000),
		key:     key,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
	sink.Reset()

	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := sink.send(); err != nil {
					log.Printf(err.Error())
				}

			case point := <-sink.metrics:
				sink.addPoint(point)
			}
		}
	}()

	return sink
}

// Count adds a value to a stat
func (s *DirectSink) Count(c *telemetry.Context, stat string, count float64) {
	s.writePoint(c, stat, count, "rate")
}

// Gague sends an absolute value. Useful for tracking things like memory.
func (s *DirectSink) Gauge(c *telemetry.Context, stat string, value float64) {
	s.writePoint(c, stat, value, "gauge")
}

// Histogram measures the statistical distribution of a set of values. eg query time
func (s *DirectSink) Histogram(c *telemetry.Context, stat string, value float64) {
	s.writePoint(c, stat, value, "hist")
}

// Timing is a special subclass of Histgram for timing information.
func (s *DirectSink) Timing(c *telemetry.Context, stat string, value float64) {
	s.writePoint(c, stat, value, "timing")
}

// Set counts unique values. Send user id to monitor unique users.
func (s *DirectSink) Set(c *telemetry.Context, stat string, value float64) {
	s.writePoint(c, stat, value, "set")
}

func (s *DirectSink) buildBatch() *BatchSeries {
	bs := BatchSeries{}

	for metric, tags := range s.buffer {
		for tagstr, inSeries := range tags {
			outSeries := Series{
				Type:   inSeries.Type,
				Metric: metric,
				Tags:   strings.Split(tagstr, ","),
				Points: Points{},
				Host:   hostname,
			}

			switch inSeries.Type {
			case "rate": // take the sum of all points
				outSeries.Type = "rate"
				for time, points := range inSeries.Values {
					var val float64 = 0
					for _, v := range points {
						val += v
					}
					outSeries.Points[time] = val
				}

				bs.Series = append(bs.Series, outSeries)

			case "gauge": // take the last value
				outSeries.Type = "gauge"
				for time, points := range inSeries.Values {
					outSeries.Points[time] = points[len(points)-1]
				}

				bs.Series = append(bs.Series, outSeries)

			case "set": // count the distinct entries
				outSeries.Type = "gauge"
				for time, points := range inSeries.Values {
					counts := map[float64]int{}
					for _, v := range points {
						counts[v]++
					}
					outSeries.Points[time] = float64(len(counts))
					bs.Series = append(bs.Series, outSeries)
				}

			case "hist", "timing": // send aggregate stats
				outSeries.Type = "gauge"

				avgSeries := outSeries.Copy(".avg")
				minSeries := outSeries.Copy(".min")
				maxSeries := outSeries.Copy(".max")
				countSeries := outSeries.Copy(".count")
				countSeries.Type = "rate"

				for time, points := range inSeries.Values {
					var sum, count, min, max float64 = 0, 0, math.Inf(1), math.Inf(-1)
					for _, v := range points {
						sum += v
						count += 1
						if v < min {
							min = v
						}
						if v > max {
							max = v
						}
					}
					avgSeries.Points[time] = sum / count
					countSeries.Points[time] = count
					minSeries.Points[time] = min
					maxSeries.Points[time] = max
				}

				bs.Series = append(bs.Series, avgSeries)
				bs.Series = append(bs.Series, countSeries)
				bs.Series = append(bs.Series, minSeries)
				bs.Series = append(bs.Series, maxSeries)
			}
		}
	}

	sort.Slice(bs.Series, func(i, j int) bool {
		return strings.Compare(bs.Series[i].Metric, bs.Series[j].Metric) > 0
	})

	return &bs
}

func (s *DirectSink) send() error {
	if len(s.buffer) == 0 {
		return nil
	}

	bs := s.buildBatch()
	body, err := json.Marshal(bs)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "https://app.datadoghq.com/api/v1/series?api_key="+s.key, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return errors.New("error posting metrics to datadog: " + err.Error())
	}

	if resp.StatusCode != http.StatusAccepted {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error posting metrics to datadog status=%d: %s", resp.StatusCode, string(b))
	}

	err = resp.Body.Close()
	if err != nil {
		panic(err)
	}

	s.Reset()

	return nil
}

func (s *DirectSink) Reset() {
	s.points = 0
	s.buffer = map[string]map[string]points{}
}

func (s *DirectSink) addPoint(point singlePoint) {
	if s.points >= MAX_POINTS {
		if s.points == MAX_POINTS {
			log.Printf("datadog hit MAX_POINTS. is datadog down? dropping metrics to keep app running.")
		}
		return
	}
	s.points++

	if s.buffer[point.Metric] == nil {
		s.buffer[point.Metric] = map[string]points{}
	}

	tagstr := strings.Join(point.Tags, ",")
	points := s.buffer[point.Metric][tagstr]
	points.Type = point.Type
	if points.Values == nil {
		points.Values = map[int64][]float64{}
	}
	points.Values[point.Time] = append(points.Values[point.Time], point.Value)
	s.buffer[point.Metric][tagstr] = points
}

func (s *DirectSink) writePoint(c *telemetry.Context, stat string, value float64, typ string) {
	select {
	case s.metrics <- singlePoint{
		Tags:   c.Tags(),
		Type:   typ,
		Metric: stat,
		Time:   time.Now().Unix(),
		Value:  value,
	}:
	default:
		log.Printf("dropping metric %s due channel overflow. DD outage?", stat)
	}
}
