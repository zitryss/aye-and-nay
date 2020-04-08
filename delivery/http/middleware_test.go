package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/rand"
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
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Internal Server Error`+"\n")
	})
}

func TestMiddlewareLimit(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
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
		time.Sleep(m.conf.limiterTimeToLive + m.conf.limiterCleanupInterval)
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

func TestMiddlewareRestrict(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.restrict(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/static/img/favicon.ico", nil)
		h.ServeHTTP(w, r)
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
		m := newMiddleware()
		h := m.restrict(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/static/js/", nil)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 403)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Forbidden`+"\n")
	})
}

func TestMiddlewareAuthorize(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "5zNEZRda9xqd8ssJ"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.authorize(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/suEq6FHp3DSmsqwx1/", nil)
		c := http.Cookie{}
		c.Name = "session"
		c.Value = "5zNEZRda9xqd8ssJ1"
		r.AddCookie(&c)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 418)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `I'm a teapot`)
	})
	t.Run("Negative1", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.authorize(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/albums/suEq6FHp3DSmsqwx1/", nil)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 403)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Forbidden`+"\n")
	})
	t.Run("Negative2", func(t *testing.T) {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.authorize(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/albums/suEq6FHp3DSmsqwx1/", nil)
		c := http.Cookie{}
		c.Name = "session"
		c.Value = "CeUREELG32wskC861"
		r.AddCookie(&c)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 403)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Forbidden`+"\n")
	})
	t.Run("Negative3", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "2NAzZs3a4fk6uwYB"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.authorize(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		time.Sleep(m.conf.sessionTimeToLive + m.conf.sessionCleanupInterval)
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/suEq6FHp3DSmsqwx1/", nil)
		c := http.Cookie{}
		c.Name = "session"
		c.Value = "2NAzZs3a4fk6uwYB1"
		r.AddCookie(&c)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 403)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `Forbidden`+"\n")
	})
	t.Run("Negative4", func(t *testing.T) {
		rand.Id = func() func(int) (string, error) {
			id := "BjgrR2kaPHYAknZs"
			i := 0
			return func(length int) (string, error) {
				i++
				return id + strconv.Itoa(i), nil
			}
		}()
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(418)
			_, _ = io.WriteString(w, "I'm a teapot")
		}
		m := newMiddleware()
		h := m.authorize(http.HandlerFunc(fn))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		c := http.Cookie{}
		c.Name = "session"
		c.Value = "KUHyhJc9CAW4qyKd"
		r.AddCookie(&c)
		h.ServeHTTP(w, r)
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/api/albums/suEq6FHp3DSmsqwx1/", nil)
		c = http.Cookie{}
		c.Name = "session"
		c.Value = "BjgrR2kaPHYAknZs1"
		r.AddCookie(&c)
		h.ServeHTTP(w, r)
		CheckStatusCode(t, w, 418)
		CheckContentType(t, w, "text/plain; charset=utf-8")
		CheckBody(t, w, `I'm a teapot`)
	})
}
