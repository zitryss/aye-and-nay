package main

import (
	"context"
	"flag"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/delivery/http"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func main() {
	conf := ""
	flag.StringVar(&conf, "config", ".", "relative path to config file")
	flag.Parse()

	viper.SetConfigName("config")
	viper.AddConfigPath(conf)
	err := viper.ReadInConfig()
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	lvl := viper.GetString("app.log")
	log.SetOutput(os.Stderr)
	log.SetLevel(lvl)
	log.Info("logging initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	comp := model.Compresser(nil)
	if viper.GetBool("shortpixel.use") {
		log.Info("connecting to shortpixel")
		sp := compressor.NewShortPixel()
		err = sp.Ping()
		if err != nil {
			log.Critical(err)
			os.Exit(1)
		}
		comp = &sp
	} else {
		mock := compressor.NewMock()
		comp = &mock
	}

	stor := model.Storager(nil)
	if viper.GetBool("minio.use") {
		log.Info("connecting to minio")
		minio, err := storage.NewMinio()
		if err != nil {
			log.Critical(err)
			os.Exit(1)
		}
		stor = &minio
	} else {
		mock := storage.NewMock()
		stor = &mock
	}

	pers := model.Persister(nil)
	if viper.GetBool("mongo.use") {
		log.Info("connecting to mongo")
		mongo, err := database.NewMongo()
		if err != nil {
			log.Critical(err)
			os.Exit(1)
		}
		pers = &mongo
	} else {
		mem := database.NewMem()
		pers = &mem
	}

	cache := model.Cacher(nil)
	if viper.GetBool("redis.use") {
		log.Info("connecting to redis")
		redis, err := database.NewRedis(ctx)
		if err != nil {
			log.Critical(err)
			os.Exit(1)
		}
		cache = &redis
	} else {
		mem := database.NewMem()
		cache = &mem
	}

	sched := service.NewScheduler()
	sched.Monitor(ctx)

	serv := service.NewService(comp, stor, pers, cache, &sched)

	g, ctx := errgroup.WithContext(ctx)
	heartbeat := chan struct{}(nil)
	log.Info("starting worker pool")
	serv.StartWorkingPool(ctx, g, heartbeat)

	srvWait := make(chan error, 1)
	srv, err := http.NewServer(&serv, cancel, srvWait)
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}
	srv.Monitor()
	log.Info("starting web server")
	err = srv.Start()

	log.Info("stopping web server")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error(err)
		return
	}
	err = <-srvWait
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("stopping worker pool")
	err = g.Wait()
	if err != nil {
		log.Error(err)
		return
	}
}
