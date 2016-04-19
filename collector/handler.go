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
	"net"
	"strings"
)

const ContextKey = "telemetry_context"

// Handler returns middleware to collect request duration/exit status
func Gorilla(c *telemetry.Context, router *mux.Router, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// IPv6 addresses use colons, but they are the tag separator. Lets use dots instead.
		ip := strings.Replace(getIp(r), ":", ".", -1)

		routeContext := c.SubContext(
			"route:" + getRouteName(router, r),
			"ip:" + ip,
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

// getIp returns the last non-private X-Forwarded-For header otherwise the requests remote address
func getIp(r *http.Request) string {
	for _, xff := range r.Header["X-Forwarded-For"] {
		addresses := strings.Split(strings.Replace(xff, " ", ",", -1), ",")

		for i := len(addresses) -1 ; i >= 0; i-- {
			xffIp := net.ParseIP(addresses[i])
			if xffIp != nil && !privateIP(xffIp) {
				return xffIp.String()
			}
		}
	}

	return net.ParseIP(strings.Split(r.RemoteAddr, ":")[0]).String()
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
