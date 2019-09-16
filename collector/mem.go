// +build !js !wasm

package collector

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/telemetry"
)

func Mem(tel *telemetry.Context) {
	go func() {
		for {
			info, err := ioutil.ReadFile("/proc/meminfo")
			if err != nil {
				log.Printf("unable to read /proc/meminfo")
				return
			}

			meminfo := map[string]float64{}
			for _, line := range strings.Split(string(info), "\n") {
				if line == "" {
					continue
				}
				parts := strings.Split(line, ":")
				valueParts := strings.Fields(parts[1])

				value, err := strconv.ParseFloat(valueParts[0], 64)
				if err != nil {
					log.Printf("cannot parse mem value %s: %s", valueParts[0], err.Error())
				}

				meminfo[parts[0]] = value * 1024 // meminfo is in bk, but bytes are better in datadog
			}

			tel.Gauge("telemetry.mem.total", meminfo["MemTotal"])
			tel.Gauge("telemetry.mem.free", meminfo["MemFree"])
			tel.Gauge("telemetry.mem.available", meminfo["MemAvailable"])
			tel.Gauge("telemetry.mem.buffers", meminfo["Buffers"])
			tel.Gauge("telemetry.mem.cached", meminfo["Cached"])

			time.Sleep(10 * time.Second)
		}
	}()
}
