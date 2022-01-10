//go:build unit

package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		middle := NewMiddleware(DefaultMiddlewareConfig, lim)
		handler := middle.recover(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
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
		middle := NewMiddleware(DefaultMiddlewareConfig, lim)
		handler := middle.recover(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
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
		middle := NewMiddleware(DefaultMiddlewareConfig, lim)
		handler := middle.limit(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
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
		middle := NewMiddleware(DefaultMiddlewareConfig, lim)
		handler := middle.limit(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		handler.ServeHTTP(w, r)
		CheckStatusCode(t, w, 429)
		CheckContentType(t, w, "application/json; charset=utf-8")
		CheckBody(t, w, `{"error":{"code":0,"msg":"too many requests"}}`+"\n")
	})
}

func TestIP(t *testing.T) {
	type args struct {
		xff        string
		remoteAddr string
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args: args{
				xff:        "203.0.113.195",
				remoteAddr: "192.168.1.2:65530",
			},
			want: "203.0.113.195",
		},
		{
			args: args{
				xff:        "203.0.113.195, 70.41.3.18, 150.172.238.178",
				remoteAddr: "192.168.1.2:65530",
			},
			want: "203.0.113.195",
		},
		{
			args: args{
				xff:        "2001:db8:85a3:8d3:1319:8a2e:370:7348",
				remoteAddr: "192.168.1.2:65530",
			},
			want: "2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			args: args{
				xff:        "",
				remoteAddr: "192.168.1.2:65530",
			},
			want: "192.168.1.2",
		},
		{
			args: args{
				xff:        "",
				remoteAddr: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/api/health/", http.NoBody)
		r.Header.Set("X-Forwarded-For", tt.args.xff)
		r.RemoteAddr = tt.args.remoteAddr
		t.Run("", func(t *testing.T) {
			if got := ip(r); got != tt.want {
				t.Errorf("ip() = %v, want %v", got, tt.want)
			}
		})
	}
}
