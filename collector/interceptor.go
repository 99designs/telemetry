package collector

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"reflect"
)

// Interceptor fetches the response code from the claler
type Interceptor struct {
	next http.ResponseWriter
	Code int
}

func NewInterceptor(next http.ResponseWriter) *Interceptor {
	return &Interceptor{
		next: next,
		Code: 200,
	}
}

func (w *Interceptor) Header() http.Header {
	return w.next.Header()
}

func (w *Interceptor) Write(b []byte) (int, error) {
	return w.next.Write(b)
}
func (w *Interceptor) WriteHeader(code int) {
	w.Code = code
	w.next.WriteHeader(code)
}

func (w *Interceptor) Flush() {
	if f, ok := w.next.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *Interceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if f, ok := w.next.(http.Hijacker); ok {
		return f.Hijack()
	}

	return nil, nil, errors.New("Hijack not supported on " + reflect.TypeOf(w.next).String())
}
