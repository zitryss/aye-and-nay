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
	middle func(http.Handler) http.Handler,
	serv domain.Servicer,
	serverWait chan<- error,
) (*Server, error) {
	conf := newServerConfig()
	contr := newController(serv)
	router := newRouter(contr)
	handler := middle(router)
	srv, err := newServer(conf, handler)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &Server{conf, srv, serverWait}, nil
}

func newServer(conf serverConfig, handler http.Handler) (*http.Server, error) {
	srv := &http.Server{
		Addr:         conf.host + ":" + conf.port,
		Handler:      http.TimeoutHandler(handler, conf.writeTimeout, ""),
		ReadTimeout:  conf.readTimeout,
		WriteTimeout: conf.writeTimeout + 1*time.Second,
		IdleTimeout:  conf.idleTimeout,
	}
	if conf.domain != "" {
		tlsConfig, err := certmagic.TLS([]string{conf.domain})
		if err != nil {
			return nil, errors.Wrap(err)
		}
		srv.TLSConfig = tlsConfig
	}
	if conf.h2c {
		srv.Handler = h2c.NewHandler(srv.Handler, &http2.Server{})
	}
	return srv, nil
}

type Server struct {
	conf       serverConfig
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
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.shutdownTimeout)
	defer cancel()
	err := s.srv.Shutdown(ctx)
	s.serverWait <- err
}
