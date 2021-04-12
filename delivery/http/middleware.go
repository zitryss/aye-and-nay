package http

import (
	"hash/fnv"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/cors"
	"golang.org/x/time/rate"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMiddleware(opts ...options) *Middleware {
	conf := newMiddlewareConfig()
	m := Middleware{conf: conf}
	for _, opt := range opts {
		opt(&m)
	}
	return &m
}

type options func(*Middleware)

func WithHeartbeat(ch chan<- interface{}) options {
	return func(m *Middleware) {
		m.heartbeat = ch
	}
}

type Middleware struct {
	conf      middlewareConfig
	heartbeat chan<- interface{}
}

func (m *Middleware) Chain(h http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{m.conf.corsAllowOrigin},
		AllowedMethods: []string{"GET", "POST", "PATCH"},
	})
	return c.Handler(m.recover(m.limit(h)))
}

func (m *Middleware) recover(h http.Handler) http.Handler {
	return handleHttpError(
		func(w http.ResponseWriter, r *http.Request) (e error) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				err, ok := v.(error)
				if ok {
					e = errors.Wrap(err)
				} else {
					e = errors.Wrapf(model.ErrUnknown, "%v", v)
				}
			}()
			h.ServeHTTP(w, r)
			return
		},
	)
}

func (m *Middleware) limit(h http.Handler) http.Handler {
	type visitor struct {
		limiter *rate.Limiter
		seen    time.Time
	}
	type syncVisitors struct {
		sync.Mutex
		visitors map[uint64]*visitor
	}
	sv := syncVisitors{visitors: map[uint64]*visitor{}}
	go func() {
		for {
			if m.heartbeat != nil {
				m.heartbeat <- struct{}{}
			}
			now := time.Now()
			sv.Lock()
			for k, v := range sv.visitors {
				if now.Sub(v.seen) >= m.conf.limiterTimeToLive {
					delete(sv.visitors, k)
				}
			}
			sv.Unlock()
			time.Sleep(m.conf.limiterCleanupInterval)
			if m.heartbeat != nil {
				m.heartbeat <- struct{}{}
			}
		}
	}()
	return handleHttpError(
		func(w http.ResponseWriter, r *http.Request) error {
			ip := r.RemoteAddr
			hash := fnv.New64a()
			_, err := io.WriteString(hash, ip)
			if err != nil {
				return errors.Wrap(err)
			}
			id := hash.Sum64()
			sv.Lock()
			v, ok := sv.visitors[id]
			if !ok {
				l := rate.NewLimiter(rate.Limit(m.conf.limiterRequestsPerSecond), m.conf.limiterBurst)
				v = &visitor{limiter: l}
				sv.visitors[id] = v
			}
			v.seen = time.Now()
			sv.Unlock()
			if !v.limiter.Allow() {
				return errors.Wrap(model.ErrTooManyRequests)
			}
			h.ServeHTTP(w, r)
			return nil
		},
	)
}
