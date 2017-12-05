package collector

import (
	"syscall"
	"time"

	"github.com/99designs/telemetry"
)

func Disk(tel *telemetry.Context, path string) {
	go func() {
		for {
			fs := syscall.Statfs_t{}
			err := syscall.Statfs(path, &fs)
			if err != nil {
				return
			}

			ctx := tel.SubContext("disk:" + path)

			ctx.Gauge("telemetry.disk.space.used", float64(fs.Blocks*uint64(fs.Bsize)))
			ctx.Gauge("telemetry.disk.space.free", float64(fs.Bavail*uint64(fs.Bsize)))
			ctx.Gauge("telemetry.disk.inodes.used", float64(fs.Files))
			ctx.Gauge("telemetry.disk.inodes.free", float64(fs.Ffree))

			time.Sleep(10 * time.Second)
		}
	}()
}
