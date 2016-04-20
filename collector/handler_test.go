package collector

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
	handler = Gorilla(telemetry.NewContext(s), router, handler)
	router.Handle("/", handler).Name("test.route")

	r := mustRequest("/")
	r.Header.Add("X-Forwarded-For", "8.8.8.8")
	router.ServeHTTP(&httptest.ResponseRecorder{}, r)

	if len(s) != 3 {
		t.Error("Not enough metrics logged")
	}

	if s["test_metric"].Tags[0] != "route:test.route" {
		t.Error("Test metric was not logged with route tag")
	}

	if s["app.request.duration"].Tags[1] != "ip:8.8.8.8" {
		t.Errorf("app request was not logged with an ip address")
	}

	if s["app.request.duration"].Tags[2] != "status:200" {
		t.Errorf("app request duration was not logged with a status")
	}
}

func TestGetIp(t *testing.T) {
	checks := map[string]string {
		"8.8.8.8,10.0.0.1": "8.8.8.8",
		"8.8.8.8,8.8.4.4": "8.8.4.4",
		"8.8.8.8 8.8.4.4": "8.8.4.4",
	}

	for xff, expected := range checks {
		r := mustRequest("/")
		r.Header.Add("X-Forwarded-for", xff)

		if getIp(r) != expected {
			t.Errorf("Incorrect ip returned, expected %s got %s", expected, getIp(r))
		}
	}
}

func TestRouteContextWithUnknownRequest(t *testing.T) {
	c := ContextForRequest(&http.Request{})

	if c == nil {
		t.Error("RequestContext should always be set")
	}
}

func mustRequest(path string) *http.Request {
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		panic(err)
	}
	return r
}
