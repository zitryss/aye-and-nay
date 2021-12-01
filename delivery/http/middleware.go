package http

import (
	"hash/fnv"
	"io"
	"net/http"

	"github.com/rs/cors"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMiddleware(lim domain.Limiter) *Middleware {
	conf := newMiddlewareConfig()
	return &Middleware{conf, lim}
}

type Middleware struct {
	conf middlewareConfig
	lim  domain.Limiter
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
					e = errors.Wrapf(domain.ErrUnknown, "%v", v)
				}
			}()
			h.ServeHTTP(w, r)
			return
		},
	)
}

func (m *Middleware) limit(h http.Handler) http.Handler {
	return handleHttpError(
		func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()
			ip := r.RemoteAddr
			hash := fnv.New64a()
			_, err := io.WriteString(hash, ip)
			if err != nil {
				return errors.Wrap(err)
			}
			allowed, err := m.lim.Allow(ctx, hash.Sum64())
			if err != nil {
				return errors.Wrap(err)
			}
			if !allowed {
				return errors.Wrap(domain.ErrTooManyRequests)
			}
			h.ServeHTTP(w, r)
			return nil
		},
	)
}
