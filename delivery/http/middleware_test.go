package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMiddlewareRecover(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.recover(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 418)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `I'm a teapot`)
	})
	t.Run("Negative", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			panic("don't")
		}
		m := newMiddleware()
		h := m.recover(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 500)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"msg":"Internal Server Error"}}`+"\n")
	})
}

func TestMiddlewareLimit(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		heartbeatMiddle := make(chan interface{})
		m := newMiddleware(WithHeartbeat(heartbeatMiddle))
		h := m.limit(http.HandlerFunc(fn))
		rps := int(m.conf.limiterRequestsPerSecond)
		for i := 0; i < rps; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)
			CheckStatusCode(t, w, 418)
			CheckContentType(t, w, "text/plain; charset=utf-8")
			CheckBody(t, w, `I'm a teapot`)
			time.Sleep(time.Duration(1/float64(rps)*1000000000) * time.Nanosecond)
		}
		time.Sleep(m.conf.limiterTimeToLive)
		CheckChannel(t, heartbeatMiddle)
		CheckChannel(t, heartbeatMiddle)
		for i := 0; i < rps; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)
			CheckStatusCode(t, w, 418)
			CheckContentType(t, w, "text/plain; charset=utf-8")
			CheckBody(t, w, `I'm a teapot`)
			time.Sleep(time.Duration(1/float64(rps)*1000000000) * time.Nanosecond)
		}
	})
	t.Run("Negative", func(t *testing.T) {
		t.Skip("flaky test")
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.limit(http.HandlerFunc(fn))
		for i := 0; i < m.conf.limiterBurst; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)
			CheckStatusCode(t, w, 418)
			CheckContentType(t, w, "text/plain; charset=utf-8")
			CheckBody(t, w, `I'm a teapot`)
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 429)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Too Many Requests`+"\n")
	})
}
