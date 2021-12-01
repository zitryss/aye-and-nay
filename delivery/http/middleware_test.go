//go:build unit

package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

type limiterMockPos struct{}

func (l limiterMockPos) Allow(_ context.Context, _ uint64) (bool, error) {
	return true, nil
}

type limiterMockNeg struct{}

func (l limiterMockNeg) Allow(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}

func TestMiddlewareRecover(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		lim := limiterMockPos{}
		middle := NewMiddleware(lim)
		handler := middle.recover(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		handler.ServeHTTP(w, r)
		CheckStatusCode(t, w, 418)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `I'm a teapot`)
	})
	t.Run("Negative", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			panic("don't")
		}
		lim := limiterMockPos{}
		middle := NewMiddleware(lim)
		handler := middle.recover(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		handler.ServeHTTP(w, r)
		CheckStatusCode(t, w, 500)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"code":17,"msg":"internal server error"}}`+"\n")
	})
}

func TestMiddlewareLimit(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		lim := limiterMockPos{}
		middle := NewMiddleware(lim)
		handler := middle.limit(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		handler.ServeHTTP(w, r)
		CheckStatusCode(t, w, 418)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `I'm a teapot`)
	})
	t.Run("Negative", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		lim := limiterMockNeg{}
		middle := NewMiddleware(lim)
		handler := middle.limit(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		handler.ServeHTTP(w, r)
		CheckStatusCode(t, w, 429)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"code":0,"msg":"too many requests"}}`+"\n")
	})
}
