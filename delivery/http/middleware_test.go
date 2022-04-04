package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/zitryss/aye-and-nay/internal/context"
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
	if !*unit {
		t.Skip()
	}
	t.Run("Positive", func(t *testing.T) {
		t.Parallel()
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
		AssertStatusCode(t, w, 418)
		AssertHeader(t, w, "Content-Type", "text/plain; charset=utf-8")
		AssertBody(t, w, `I'm a teapot`)
	})
	t.Run("Negative", func(t *testing.T) {
		t.Parallel()
		fn := func(w http.ResponseWriter, r *http.Request) {
			panic("don't")
		}
		lim := limiterMockPos{}
		middle := NewMiddleware(DefaultMiddlewareConfig, lim)
		handler := middle.recover(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		handler.ServeHTTP(w, r)
		AssertStatusCode(t, w, 500)
		AssertHeader(t, w, "Content-Type", "application/json; charset=utf-8")
		AssertBody(t, w, `{"error":{"code":22,"msg":"internal server error"}}`+"\n")
	})
}

func TestMiddlewareLimit(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	t.Run("Positive", func(t *testing.T) {
		t.Parallel()
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
		AssertStatusCode(t, w, 418)
		AssertHeader(t, w, "Content-Type", "text/plain; charset=utf-8")
		AssertBody(t, w, `I'm a teapot`)
	})
	t.Run("Negative", func(t *testing.T) {
		t.Parallel()
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
		AssertStatusCode(t, w, 429)
		AssertHeader(t, w, "Content-Type", "application/json; charset=utf-8")
		AssertBody(t, w, `{"error":{"code":1,"msg":"too many requests"}}`+"\n")
	})
}

func TestIP(t *testing.T) {
	if !*unit {
		t.Skip()
	}
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
		tt := tt
		r := httptest.NewRequest(http.MethodGet, "/api/health/", http.NoBody)
		r.Header.Set("X-Forwarded-For", tt.args.xff)
		r.RemoteAddr = tt.args.remoteAddr
		t.Run("", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, ip(r))
		})
	}
}

func TestMiddlewareRequestId(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	t.Parallel()
	ids := []uint64{0}
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := GetRequestId(ctx)
		assert.NotZero(t, id)
		assert.Positive(t, id)
		assert.Greater(t, id, ids[len(ids)-1])
		ids = append(ids, id)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(418)
		_, _ = io.WriteString(w, "I'm a teapot")
	}
	lim := limiterMockPos{}
	middle := NewMiddleware(DefaultMiddlewareConfig, lim)
	handler := middle.requestId(http.HandlerFunc(fn))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	handler.ServeHTTP(w, r)
	AssertStatusCode(t, w, 418)
	AssertHeader(t, w, "Content-Type", "text/plain; charset=utf-8")
	AssertBody(t, w, `I'm a teapot`)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	handler.ServeHTTP(w, r)
	AssertStatusCode(t, w, 418)
	AssertHeader(t, w, "Content-Type", "text/plain; charset=utf-8")
	AssertBody(t, w, `I'm a teapot`)
}

func TestMiddlewareHeaders(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	t.Parallel()
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(418)
		_, _ = io.WriteString(w, "I'm a teapot")
	}
	lim := limiterMockPos{}
	middle := NewMiddleware(DefaultMiddlewareConfig, lim)
	handler := middle.headers(http.HandlerFunc(fn))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	handler.ServeHTTP(w, r)
	AssertStatusCode(t, w, 418)
	AssertHeader(t, w, "Content-Type", "text/plain; charset=utf-8")
	AssertHeader(t, w, "X-Content-Type-Options", "nosniff")
	AssertBody(t, w, `I'm a teapot`)
}
