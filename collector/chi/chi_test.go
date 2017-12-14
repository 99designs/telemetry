package chi

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/99designs/telemetry"
	"github.com/99designs/telemetry/sink"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	router := chi.NewRouter()
	ts := sink.Test()
	c := telemetry.New()
	c.AddSink(ts)

	router.Use(Middleware(c))
	router.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	router.Route("/nested", func(router chi.Router) {
		router.HandleFunc("/{status}", func(w http.ResponseWriter, r *http.Request) {
			v, _ := strconv.Atoi(chi.URLParam(r, "status"))
			w.WriteHeader(v)
		})
	})

	t.Run("logs successful requests", func(t *testing.T) {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ok", nil))
		require.NotEqual(t, 0, ts["app.request.duration"].Value)
		require.Equal(t, "route:/ok", ts["app.request.duration"].Tags[0])
		require.Equal(t, "status:200", ts["app.request.duration"].Tags[1])
	})

	t.Run("logs unsuccessful requests", func(t *testing.T) {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/not-found", nil))
		require.Equal(t, "route:unknown", ts["app.request.duration"].Tags[0])
		require.Equal(t, "status:404", ts["app.request.duration"].Tags[1])
	})

	t.Run("logs nested requests", func(t *testing.T) {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/nested/500", nil))
		require.Equal(t, "route:/nested/:status:", ts["app.request.duration"].Tags[0])
		require.Equal(t, "status:500", ts["app.request.duration"].Tags[1])
	})
}
