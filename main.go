package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/delivery/http"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

var (
	ballast []byte
)

func main() {
	path := ""
	flag.StringVar(&path, "config", "./config.env", "filepath to a config file")
	flag.Parse()

	reload := true
	for reload {
		reload = false

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

		conf, err := config.New(path)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "critical:", err)
			reload = true
			stop()
			continue
		}

		if conf.Reload {
			conf.OnChange(ctx, func() {
				reload = true
				stop()
			})
		}

		ballast = make([]byte, conf.App.Ballast)

		log.SetOutput(os.Stderr)
		log.SetPrefix(conf.App.Name)
		log.SetLevel(conf.App.Log)
		log.Info("logging initialized:", "log level:", conf.App.Log)

		cach, err := cache.New(ctx, conf.Cache)
		if err != nil {
			log.Critical(err)
			reload = true
			stop()
			continue
		}

		comp, err := compressor.New(ctx, conf.Compressor)
		if err != nil {
			log.Critical(err)
			reload = true
			stop()
			continue
		}

		data, err := database.New(ctx, conf.Database)
		if err != nil {
			log.Critical(err)
			reload = true
			stop()
			continue
		}

		stor, err := storage.New(ctx, conf.Storage)
		if err != nil {
			log.Critical(err)
			reload = true
			stop()
			continue
		}

		qCalc := service.NewQueueCalc(cach)
		qCalc.Monitor(ctx)

		qComp := &service.QueueComp{}
		if !conf.Compressor.IsMock() {
			qComp = service.NewQueueComp(cach)
			qComp.Monitor(ctx)
		}

		qDel := service.NewQueueDel(cach)
		qDel.Monitor(ctx)

		serv := service.New(conf.Service, comp, stor, data, cach, qCalc, qComp, qDel)
		err = serv.CleanUp(ctx)
		if err != nil {
			log.Critical(err)
			reload = true
			stop()
			continue
		}

		gCalc, ctxCalc := errgroup.WithContext(ctx)
		log.Info("starting calculation worker pool")
		serv.StartWorkingPoolCalc(ctxCalc, gCalc)

		gComp := (*errgroup.Group)(nil)
		ctxComp := context.Context(nil)
		if !conf.Compressor.IsMock() {
			gComp, ctxComp = errgroup.WithContext(ctx)
			log.Info("starting compression worker pool")
			serv.StartWorkingPoolComp(ctxComp, gComp)
		}

		gDel, ctxDel := errgroup.WithContext(ctx)
		log.Info("starting deletion worker pool")
		serv.StartWorkingPoolDel(ctxDel, gDel)

		middle := http.NewMiddleware(conf.Middleware, cach)
		srvWait := make(chan error, 1)
		srv, err := http.NewServer(conf.Server, middle.Chain, serv, srvWait)
		if err != nil {
			log.Critical(err)
			reload = true
			stop()
			continue
		}
		srv.Monitor(ctx)
		log.Info("starting web server")
		err = srv.Start()

		log.Info("stopping web server")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(err)
		}
		err = <-srvWait
		if err != nil {
			log.Error(err)
		}

		log.Info("stopping deletion worker pool")
		err = gDel.Wait()
		if err != nil {
			log.Error(err)
		}

		if !conf.Compressor.IsMock() {
			log.Info("stopping compression worker pool")
			err = gComp.Wait()
			if err != nil {
				log.Error(err)
			}
		}

		log.Info("stopping calculation worker pool")
		err = gCalc.Wait()
		if err != nil {
			log.Error(err)
		}

		r, ok := cach.(*cache.Redis)
		if ok {
			err = r.Close(ctx)
			if err != nil {
				log.Error(err)
			}
		}

		m, ok := data.(*database.Mongo)
		if ok {
			err = m.Close(ctx)
			if err != nil {
				log.Error(err)
			}
		}

		b, ok := data.(*database.Badger)
		if ok {
			err = b.Close(ctx)
			if err != nil {
				log.Error(err)
			}
		}

		stop()
	}
}
