package http

import (
	"context"
	"net/http"
	"time"

	"github.com/caddyserver/certmagic"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

var (
	ErrServerClosed = http.ErrServerClosed
)

func NewServer(
	conf ServerConfig,
	middle func(http.Handler) http.Handler,
	serv domain.Servicer,
	serverWait chan<- error,
) (*Server, error) {
	contr := newController(conf.Controller, serv)
	router := newRouter(contr)
	handler := middle(router)
	srv, err := newServer(conf, handler)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &Server{conf, srv, serverWait}, nil
}

func newServer(conf ServerConfig, handler http.Handler) (*http.Server, error) {
	srv := &http.Server{
		Addr:         conf.Host + ":" + conf.Port,
		Handler:      handler,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout + 1*time.Second,
		IdleTimeout:  conf.IdleTimeout,
	}
	if conf.Domain != "" {
		tlsConfig, err := certmagic.TLS([]string{conf.Domain})
		if err != nil {
			return nil, errors.Wrap(err)
		}
		srv.TLSConfig = tlsConfig
	}
	if conf.H2C {
		srv.Handler = h2c.NewHandler(srv.Handler, &http2.Server{})
	}
	return srv, nil
}

type Server struct {
	conf       ServerConfig
	srv        *http.Server
	serverWait chan<- error
}

func (s *Server) Monitor(ctx context.Context) {
	go func() {
		<-ctx.Done()
		s.shutdown()
	}()
}

func (s *Server) Start() error {
	err := s.srv.ListenAndServe()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *Server) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.ShutdownTimeout)
	defer cancel()
	err := s.srv.Shutdown(ctx)
	s.serverWait <- err
}
