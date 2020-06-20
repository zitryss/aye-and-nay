package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

var (
	ErrServerClosed = http.ErrServerClosed
)

func NewServer(
	serv model.Servicer,
	cancel context.CancelFunc,
	serverWait chan<- error,
) server {
	conf := newServerConfig()
	contr := newController(serv)
	router := newRouter(contr)
	middle := newMiddleware()
	handler := middle.chain(router)
	http := newHttp(conf)
	https := newHttps(conf, handler)
	https.RegisterOnShutdown(cancel)
	svr := server{conf, http, https, serverWait}
	return svr
}

func newHttp(conf serverConfig) http.Server {
	return http.Server{
		Addr:         conf.host + ":" + conf.port,
		Handler:      redirect(conf.portRedirect),
		ReadTimeout:  conf.readTimeout,
		WriteTimeout: conf.writeTimeout,
		IdleTimeout:  conf.idleTimeout,
	}
}

func newHttps(conf serverConfig, handler http.Handler) http.Server {
	return http.Server{
		Addr:         conf.host + ":" + conf.portTls,
		Handler:      http.TimeoutHandler(handler, conf.writeTimeout, ""),
		ReadTimeout:  conf.readTimeout,
		WriteTimeout: conf.writeTimeout + 1*time.Second,
		IdleTimeout:  conf.idleTimeout,
	}
}

func redirect(portRedirect string) http.HandlerFunc {
	return handleHttpError(
		func(w http.ResponseWriter, r *http.Request) error {
			host := strings.Split(r.Host, ":")[0]
			http.Redirect(w, r, "https://"+host+":"+portRedirect+r.RequestURI, http.StatusMovedPermanently)
			return nil
		},
	)
}

type server struct {
	conf       serverConfig
	http       http.Server
	https      http.Server
	serverWait chan<- error
}

func (s *server) Monitor() {
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		s.shutdown()
	}()
}

func (s *server) Start() error {
	go func() {
		_ = s.http.ListenAndServe()
	}()
	err := s.https.ListenAndServeTLS(s.conf.certFile, s.conf.keyFile)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *server) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.shutdownTimeout)
	defer cancel()
	err := s.https.Shutdown(ctx)
	s.serverWait <- err
}
