package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

var (
	ErrServerClosed = http.ErrServerClosed
)

func NewServer(
	serv model.Servicer,
	serverWait chan<- error,
) *Server {
	conf := newServerConfig()
	contr := newController(serv)
	router := newRouter(contr)
	middle := newMiddleware()
	handler := middle.chain(router)
	https := newHttps(conf, handler)
	return &Server{conf, https, serverWait}
}

func newHttps(conf serverConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         conf.host + ":" + conf.port,
		Handler:      h2c.NewHandler(http.TimeoutHandler(handler, conf.writeTimeout, ""), &http2.Server{}),
		ReadTimeout:  conf.readTimeout,
		WriteTimeout: conf.writeTimeout + 1*time.Second,
		IdleTimeout:  conf.idleTimeout,
	}
}

type Server struct {
	conf       serverConfig
	https      *http.Server
	serverWait chan<- error
}

func (s *Server) Monitor() {
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		s.shutdown()
	}()
}

func (s *Server) Start() error {
	err := s.https.ListenAndServe()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *Server) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.shutdownTimeout)
	defer cancel()
	err := s.https.Shutdown(ctx)
	s.serverWait <- err
}
