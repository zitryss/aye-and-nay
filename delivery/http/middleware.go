package http

import (
	"hash/fnv"
	"io"
	"net/http"
	"strings"

	"github.com/rs/cors"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/internal/context"
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
	handler :=
		m.recover(
			m.limit(
				c.Handler(
					http.TimeoutHandler(
						http.MaxBytesHandler(
							m.requestId(
								m.headers(h),
							),
							m.conf.MaxFileSize,
						),
						m.conf.WriteTimeout,
						"",
					),
				),
			),
		)
	return handler
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

func ip(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		before, _, _ := strings.Cut(xff, ", ")
		return before
	}
	before, _, _ := strings.Cut(r.RemoteAddr, ":")
	return before
}

func (m *Middleware) requestId(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithRequestId(r.Context())
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		},
	)
}

func (m *Middleware) headers(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			h.ServeHTTP(w, r)
		},
	)
}
