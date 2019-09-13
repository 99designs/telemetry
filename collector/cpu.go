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

func CPU(tel *telemetry.Context) {
	go func() {
		for {
			line, err := ioutil.ReadFile("/proc/loadavg")
			if err != nil {
				log.Printf("unable to read /proc/loadavg")
				return
			}

			values := strings.Fields(string(line))

			tel.Gauge("telemetry.cpu.load1", parseAvg(values[0]))
			tel.Gauge("telemetry.cpu.load5", parseAvg(values[1]))
			tel.Gauge("telemetry.cpu.load15", parseAvg(values[2]))

			time.Sleep(5 * time.Second)
		}
	}()
}

func parseAvg(avg string) float64 {
	load, err := strconv.ParseFloat(avg, 64)
	if err != nil {
		log.Printf("cannot parse loadavg %s: %s", avg, err.Error())
	}
	return load
}
