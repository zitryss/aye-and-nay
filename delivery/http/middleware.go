package http

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/rand"
)

func newMiddleware() middleware {
	conf := newMiddlewareConfig()
	return middleware{conf}
}

type middleware struct {
	conf middlewareConfig
}

func (m *middleware) chain(h http.Handler) http.Handler {
	return m.recover(m.limit(m.restrict(m.authorize(h))))
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

func (m *middleware) restrict(h http.Handler) http.Handler {
	return handleHttpError(
		func(w http.ResponseWriter, r *http.Request) error {
			if strings.HasPrefix(r.URL.Path, "/static/") && strings.HasSuffix(r.URL.Path, "/") {
				return errors.Wrap(model.ErrForbinden)
			}
			h.ServeHTTP(w, r)
			return nil
		},
	)
}

func (m *middleware) authorize(h http.Handler) http.Handler {
	type syncmap struct {
		sync.RWMutex
		data map[string]time.Time
	}
	sm := syncmap{data: map[string]time.Time{}}
	go func() {
		for {
			now := time.Now()
			sm.Lock()
			for k, v := range sm.data {
				if now.After(v) {
					delete(sm.data, k)
				}
			}
			sm.Unlock()
			time.Sleep(m.conf.sessionCleanupInterval)
		}
	}()
	cookieFactory := func() (http.Cookie, error) {
		id, err := rand.Id(m.conf.sessionIdLength)
		if err != nil {
			return http.Cookie{}, errors.Wrap(err)
		}
		now := time.Now()
		sm.Lock()
		sm.data[id] = now.Add(m.conf.sessionTimeToLive)
		sm.Unlock()
		cookie := http.Cookie{}
		cookie.Name = "session"
		cookie.Value = id
		cookie.Expires = now.Add(m.conf.sessionTimeToLive)
		cookie.MaxAge = int(m.conf.sessionTimeToLive.Seconds())
		cookie.HttpOnly = true
		return cookie, nil
	}
	checkCookie := func(w http.ResponseWriter, r *http.Request) error {
		c, err := r.Cookie("session")
		if errors.Is(err, http.ErrNoCookie) {
			return errors.Wrap(model.ErrForbinden)
		}
		if err != nil {
			return errors.Wrap(err)
		}
		sm.RLock()
		_, ok := sm.data[c.Value]
		sm.RUnlock()
		if !ok {
			return errors.Wrap(model.ErrForbinden)
		}
		return nil
	}
	writeCookie := func(w http.ResponseWriter, r *http.Request) error {
		c, err := r.Cookie("session")
		if errors.Is(err, http.ErrNoCookie) {
			c, err := cookieFactory()
			if err != nil {
				return errors.Wrap(err)
			}
			http.SetCookie(w, &c)
			return nil
		}
		if err != nil {
			return errors.Wrap(err)
		}
		sm.RLock()
		_, ok := sm.data[c.Value]
		sm.RUnlock()
		if !ok {
			c, err := cookieFactory()
			if err != nil {
				return errors.Wrap(err)
			}
			http.SetCookie(w, &c)
		}
		return nil
	}
	return handleHttpError(
		func(w http.ResponseWriter, r *http.Request) error {
			if strings.HasPrefix(r.URL.Path, "/api") {
				err := checkCookie(w, r)
				if err != nil {
					return errors.Wrap(err)
				}
			} else {
				err := writeCookie(w, r)
				if err != nil {
					return errors.Wrap(err)
				}
			}
			h.ServeHTTP(w, r)
			return nil
		},
	)
}
