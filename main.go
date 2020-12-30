package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/delivery/http"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

var (
	ballast []byte
)

func main() {
	conf := ""
	flag.StringVar(&conf, "config", ".", "relative path to config file")
	flag.Parse()

	viper.SetConfigName("config")
	viper.AddConfigPath(conf)
	err := viper.ReadInConfig()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "critical:", err)
		os.Exit(1)
	}

	ballast = make([]byte, viper.GetInt64("app.ballast"))

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
		sp.Monitor()
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

	pers := model.Databaser(nil)
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

	temp := model.Cacher(nil)
	if viper.GetBool("redis.use") {
		log.Info("connecting to redis")
		redis, err := cache.NewRedis()
		if err != nil {
			log.Critical(err)
			os.Exit(1)
		}
		temp = &redis
	} else {
		mem := cache.NewMem()
		mem.Monitor()
		temp = &mem
	}

	qCalc := service.NewQueueCalc("calculation", temp)
	qCalc.Monitor(ctx)

	qComp := &service.QueueComp{}
	if viper.GetBool("shortpixel.use") {
		qComp = service.NewQueueComp("compression", temp)
		qComp.Monitor(ctx)
	}

	qDel := service.NewQueueDel("deletion", temp)
	qDel.Monitor(ctx)

	serv := service.NewService(comp, stor, pers, temp, qCalc, qComp, qDel)

	g1, ctx1 := errgroup.WithContext(ctx)
	log.Info("starting calculation worker pool")
	serv.StartWorkingPoolCalc(ctx1, g1)

	g2 := (*errgroup.Group)(nil)
	ctx2 := context.Context(nil)
	if viper.GetBool("shortpixel.use") {
		g2, ctx2 = errgroup.WithContext(ctx)
		log.Info("starting compression worker pool")
		serv.StartWorkingPoolComp(ctx2, g2)
	}

	g3, ctx3 := errgroup.WithContext(ctx)
	log.Info("starting deletion worker pool")
	serv.StartWorkingPoolDel(ctx3, g3)

	srvWait := make(chan error, 1)
	srv := http.NewServer(&serv, cancel, srvWait)
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

	log.Info("stopping deletion worker pool")
	err = g3.Wait()
	if err != nil {
		log.Error(err)
		return
	}

	if viper.GetBool("shortpixel.use") {
		log.Info("stopping compression worker pool")
		err = g2.Wait()
		if err != nil {
			log.Error(err)
			return
		}
	}

	log.Info("stopping calculation worker pool")
	err = g1.Wait()
	if err != nil {
		log.Error(err)
		return
	}
}
