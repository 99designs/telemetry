package chi

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"strings"

	"github.com/99designs/telemetry"
	"github.com/99designs/telemetry/collector"
	"github.com/go-chi/chi"
)

func Middleware(c *telemetry.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wWrapper := collector.NewInterceptor(w)

			routeContext := c.SubContext()
			r = r.WithContext(telemetry.WithTelemetry(r.Context(), routeContext))

			next.ServeHTTP(wWrapper, r)
			duration := time.Now().Sub(start).Seconds()

			routeContext.AddTags(
				"route:"+getRouteName(r),
				"status:"+strconv.Itoa(wWrapper.Code),
			)

			routeContext.Incr("app.request.count")
			routeContext.Histogram("app.request.duration", duration)
		})
	}
}

func getRouteName(r *http.Request) string {
	buf := &bytes.Buffer{}
	patterns := chi.RouteContext(r.Context()).RoutePatterns
	for i, r := range patterns {
		// Strip the wildcard/pattern match off everything bust the last prefix match.
		r = strings.NewReplacer(
			"{", ":",
			"}", ":",
		).Replace(r)
		if i != len(patterns)-1 {
			buf.WriteString(filepath.Dir(r))
		} else {
			buf.WriteString(r)
		}
	}

	if buf.Len() > 0 {
		return buf.String()
	} else {
		return "unknown"
	}
}
