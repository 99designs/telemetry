package gorilla

import (
	"net/http"
	"strconv"
	"time"

	"github.com/99designs/telemetry"
	"github.com/99designs/telemetry/collector"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

const ContextKey = "telemetry_context"

// Handler returns middleware to collect request duration/exit status
func Handler(c *telemetry.Context, router *mux.Router, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		routeContext := c.SubContext(
			"route:" + getRouteName(router, r),
		)

		context.Set(r, ContextKey, routeContext)

		wWrapper := collector.NewInterceptor(w)
		next.ServeHTTP(wWrapper, r)

		completedRouteContext := routeContext.SubContext(
			"status:" + strconv.Itoa(wWrapper.Code),
		)

		diff := time.Now().Sub(start).Seconds()
		completedRouteContext.Incr("app.request.count")
		completedRouteContext.Histogram("app.request.duration", diff)
	})
}

func ContextForRequest(r *http.Request) *telemetry.Context {
	c, ok := context.Get(r, ContextKey).(*telemetry.Context)
	if !ok {
		return &telemetry.Context{}
	}

	return c
}

func getRouteName(router *mux.Router, r *http.Request) string {
	routeMatch := &mux.RouteMatch{}
	router.Match(r, routeMatch)

	if routeMatch != nil && routeMatch.Route != nil && routeMatch.Route.GetName() != "" {
		return routeMatch.Route.GetName()
	} else {
		return "unknown"
	}
}
