package http

import (
	"context"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/cors"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/internal/log"
	"github.com/zitryss/aye-and-nay/internal/requestid"
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
		MaxAge:         86400, // Firefox caps the value at 86400 (24 hours) while all Chromium-based browsers cap it at 7200 (2 hours)
	})
	if m.conf.Debug {
		h = m.debug(h)
	}
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
		func(w http.ResponseWriter, r *http.Request) (ctx context.Context, e error) {
			ctx = r.Context()
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
		func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
			ctx := r.Context()
			hash := fnv.New64a()
			_, err := io.WriteString(hash, ip(r))
			if err != nil {
				return ctx, errors.Wrap(err)
			}
			allowed, err := m.lim.Allow(ctx, hash.Sum64())
			if err != nil {
				return ctx, errors.Wrap(err)
			}
			if !allowed {
				return ctx, errors.Wrap(domain.ErrTooManyRequests)
			}
			h.ServeHTTP(w, r)
			return ctx, nil
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
			ctx := requestid.Set(r.Context())
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

func (m *Middleware) debug(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log.Debug(ctx,
				"incoming request",
				"time", time.Now(),
				"r.RemoteAddr", r.RemoteAddr,
				"r.Host", r.Host,
				"r.Proto", r.Proto,
				"r.TLS", r.TLS != nil,
				"r.Method", r.Method,
				"r.RequestURI", r.RequestURI,
				"r.URL.Path", r.URL.Path,
				"r.URL.RawQuery", r.URL.RawQuery,
				"req.Header", r.Header,
			)
			h.ServeHTTP(w, r)
		},
	)
}
