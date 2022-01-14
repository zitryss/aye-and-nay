package http

import (
	"hash/fnv"
	"io"
	"net/http"
	"strings"

	"github.com/rs/cors"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMiddleware(
	conf MiddlewareConfig,
	lim domain.Limiter,
) *Middleware {
	return &Middleware{conf, lim}
}

type Middleware struct {
	conf MiddlewareConfig
	lim  domain.Limiter
}

func (m *Middleware) Chain(h http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{m.conf.CorsAllowOrigin},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch},
	})
	return m.recover(m.limit(c.Handler(m.headers(h))))
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
			hash := fnv.New64a()
			_, err := io.WriteString(hash, ip(r))
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

func (m *Middleware) headers(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
			w.Header().Set("X-Content-Type-Options", "nosniff")
		},
	)
}

func ip(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.Split(xff, ", ")[0]
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
