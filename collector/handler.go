package collector

import (
	"net/http"
	"strconv"
	"time"
)

import (
	"github.com/99designs/telemetry"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

const ContextKey = "telemetry_context"

// Handler returns middleware to collect request duration/exit status
func Gorilla(c *telemetry.Context, router *mux.Router, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		routeContext := c.SubContext(
			"route:" + getRouteName(router, r),
		)

		context.Set(r, ContextKey, routeContext)

		wWrapper := &responseWriterWithCode{w, http.StatusOK}
		next.ServeHTTP(wWrapper, r)

		completedRouteContext := routeContext.SubContext(
			"status:" + strconv.Itoa(wWrapper.code),
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

// There appears to be no good way determine the http code
// from the writer.
type responseWriterWithCode struct {
	next http.ResponseWriter
	code int
}

func (w *responseWriterWithCode) Header() http.Header {
	return w.next.Header()
}

func (w *responseWriterWithCode) Write(b []byte) (int, error) {
	return w.next.Write(b)
}
func (w *responseWriterWithCode) WriteHeader(code int) {
	w.code = code
	w.next.WriteHeader(code)
}
