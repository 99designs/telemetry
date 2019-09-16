// +build js,wasm

package collector

import "github.com/99designs/telemetry"

func CPU(tel *telemetry.Context) {}

func Disk(tel *telemetry.Context, path string) {}

func Mem(tel *telemetry.Context) {}
