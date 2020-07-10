package http

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func newMiddleware() middleware {
	conf := newMiddlewareConfig()
	return middleware{conf}
}

type middleware struct {
	conf middlewareConfig
}

func (m *middleware) chain(h http.Handler) http.Handler {
	return m.recover(m.limit(h))
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
			now := time.Now()
			sm.Lock()
			for k, v := range sm.data {
				if now.Sub(v.seen) >= m.conf.limiterTimeToLive {
					delete(sm.data, k)
				}
			}
			sm.Unlock()
			time.Sleep(m.conf.limiterCleanupInterval)
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
