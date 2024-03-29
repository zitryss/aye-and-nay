package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/delivery/http"
	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/internal/gctuner"
	"github.com/zitryss/aye-and-nay/internal/log"
	"github.com/zitryss/aye-and-nay/internal/ulimit"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func main() {
	err := ulimit.SetMax()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "critical:", err)
		os.Exit(1)
	}

	path := ""
	flag.StringVar(&path, "config", "./config.env", "filepath to a config file")
	flag.Parse()

	cach := domain.Cacher(nil)
	comp := domain.Compresser(nil)
	data := domain.Databaser(nil)
	stor := domain.Storager(nil)

	reload := true
	for reload {
		reload = false

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

		conf, err := config.New(path)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "critical:", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
			continue
		}

		if conf.Reload {
			conf.OnChange(ctx, func() {
				reload = true
				stop()
			})
		}

		err = log.New(conf.App.Log, conf.App.Name)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "critical:", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
			continue
		}

		log.Info(context.Background(), "logging initialized", "log level", conf.App.Log)

		if conf.App.GcTuner == "custom" {
			err = gctuner.Start(ctx, conf.App.MemTotal, conf.App.MemLimitRatio)
			if err != nil {
				log.Critical(context.Background(), "err", "stacktrace", err)
				reload = true
				stop()
				time.Sleep(2 * time.Second)
				continue
			}
		} else if conf.App.GcTuner == "go" {
			debug.SetMemoryLimit(int64(float64(conf.App.MemTotal) * conf.App.MemLimitRatio))
		}

		cach, err = cache.New(ctx, conf.Cache)
		if err != nil {
			log.Critical(context.Background(), "err", "stacktrace", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
			continue
		}

		comp, err = compressor.New(ctx, conf.Compressor)
		if err != nil {
			log.Critical(context.Background(), "err", "stacktrace", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
			continue
		}

		data, err = database.New(ctx, conf.Database)
		if err != nil {
			log.Critical(context.Background(), "err", "stacktrace", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
			continue
		}

		stor, err = storage.New(ctx, conf.Storage)
		if err != nil {
			log.Critical(context.Background(), "err", "stacktrace", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
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
			log.Critical(context.Background(), "err", "stacktrace", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
			continue
		}

		gCalc, ctxCalc := errgroup.WithContext(ctx)
		log.Info(context.Background(), "starting calculation worker pool")
		serv.StartWorkingPoolCalc(ctxCalc, gCalc)

		gComp := (*errgroup.Group)(nil)
		ctxComp := context.Context(nil)
		if !conf.Compressor.IsMock() {
			gComp, ctxComp = errgroup.WithContext(ctx)
			log.Info(context.Background(), "starting compression worker pool")
			serv.StartWorkingPoolComp(ctxComp, gComp)
		}

		gDel, ctxDel := errgroup.WithContext(ctx)
		log.Info(context.Background(), "starting deletion worker pool")
		serv.StartWorkingPoolDel(ctxDel, gDel)

		middle := http.NewMiddleware(conf.Middleware, cach)
		srvWait := make(chan error, 1)
		srv, err := http.NewServer(conf.Server, middle.Chain, serv, srvWait)
		if err != nil {
			log.Critical(context.Background(), "err", "stacktrace", err)
			reload = true
			stop()
			time.Sleep(2 * time.Second)
			continue
		}
		srv.Monitor(ctx)
		log.Info(context.Background(), "starting web server")
		err = srv.Start()

		log.Info(context.Background(), "stopping web server")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(context.Background(), "err", "stacktrace", err)
		}
		err = <-srvWait
		if err != nil {
			log.Error(context.Background(), "err", "stacktrace", err)
		}

		log.Info(context.Background(), "stopping deletion worker pool")
		err = gDel.Wait()
		if err != nil {
			log.Error(context.Background(), "err", "stacktrace", err)
		}

		if !conf.Compressor.IsMock() {
			log.Info(context.Background(), "stopping compression worker pool")
			err = gComp.Wait()
			if err != nil {
				log.Error(context.Background(), "err", "stacktrace", err)
			}
		}

		log.Info(context.Background(), "stopping calculation worker pool")
		err = gCalc.Wait()
		if err != nil {
			log.Error(context.Background(), "err", "stacktrace", err)
		}

		stop()

		b, ok := data.(*database.Badger)
		if ok {
			err = b.Close(context.Background())
			if err != nil {
				log.Error(context.Background(), "err", "stacktrace", err)
			}
		}
	}

	r, ok := cach.(*cache.Redis)
	if ok {
		err = r.Close(context.Background())
		if err != nil {
			log.Error(context.Background(), "err", "stacktrace", err)
		}
	}

	m, ok := data.(*database.Mongo)
	if ok {
		err = m.Close(context.Background())
		if err != nil {
			log.Error(context.Background(), "err", "stacktrace", err)
		}
	}
}
