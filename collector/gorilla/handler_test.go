package gorilla

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/telemetry"
	"github.com/99designs/telemetry/sink"
	"github.com/gorilla/mux"
)

func TestRouteContext(t *testing.T) {
	router := mux.NewRouter()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContextForRequest(r).Gauge("test_metric", 2)
	})

	s := sink.Test()
	handler = Handler(telemetry.NewContext(s), router, handler)
	router.Handle("/", handler).Name("test.route")

	doTestGet(router, "/")

	if len(s) != 3 {
		t.Error("Not enough metrics logged")
	}

	if s["test_metric"].Tags[0] != "route:test.route" {
		t.Error("Test metric was not logged with route tag")
	}

	if s["app.request.duration"].Tags[1] != "status:200" {
		t.Errorf("app request duration was not logged with a status")
	}
}

func TestRouteContextWithUnknownRequest(t *testing.T) {
	c := ContextForRequest(&http.Request{})

	if c == nil {
		t.Error("RequestContext should always be set")
	}
}

func doTestGet(router *mux.Router, path string) {
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		panic(err)
	}
	router.ServeHTTP(&httptest.ResponseRecorder{}, r)
}
