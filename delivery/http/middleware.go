package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/rs/cors"
	"golang.org/x/time/rate"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func newMiddleware(opts ...options) middleware {
	conf := newMiddlewareConfig()
	m := middleware{conf: conf}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

type options func(*middleware)

func WithHeartbeat(ch chan<- interface{}) options {
	return func(m *middleware) {
		m.heartbeat = ch
	}
}

type middleware struct {
	conf      middlewareConfig
	heartbeat chan<- interface{}
}

func (m *middleware) chain(h http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{m.conf.corsAllowOrigin},
		AllowedMethods: []string{"GET", "POST", "PATCH"},
	})
	return c.Handler(m.recover(m.limit(h)))
}

func (m *middleware) recover(h http.Handler) http.Handler {
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

func (m *middleware) limit(h http.Handler) http.Handler {
	type visitor struct {
		limiter *rate.Limiter
		seen    time.Time
	}
	type syncmap struct {
		sync.Mutex
		data map[string]*visitor
	}
	sm := syncmap{data: map[string]*visitor{}}
	go func() {
		for {
			if m.heartbeat != nil {
				m.heartbeat <- struct{}{}
			}
			now := time.Now()
			sm.Lock()
			for k, v := range sm.data {
				if now.Sub(v.seen) >= m.conf.limiterTimeToLive {
					delete(sm.data, k)
				}
			}
			sm.Unlock()
			time.Sleep(m.conf.limiterCleanupInterval)
			if m.heartbeat != nil {
				m.heartbeat <- struct{}{}
			}
		}
	}()
	return handleHttpError(
		func(w http.ResponseWriter, r *http.Request) error {
			ip := r.RemoteAddr
			sm.Lock()
			v, ok := sm.data[ip]
			if !ok {
				l := rate.NewLimiter(rate.Limit(m.conf.limiterRequestsPerSecond), m.conf.limiterBurst)
				v = &visitor{limiter: l}
				sm.data[ip] = v
			}
			v.seen = time.Now()
			if !v.limiter.Allow() {
				sm.Unlock()
				return errors.Wrap(model.ErrTooManyRequests)
			}
			sm.Unlock()
			h.ServeHTTP(w, r)
			return nil
		},
	)
}
