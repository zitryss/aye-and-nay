package config

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/radovskyb/watcher"
	"github.com/spf13/viper"

	"github.com/zitryss/aye-and-nay/delivery/http"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(path string) (Config, error) {
	viper.Reset()
	conf := Config{}
	conf.path = path
	err := readConfig(path, &conf)
	if err != nil {
		return Config{}, errors.Wrap(err)
	}
	fillGaps(&conf)
	err = validator.New().Struct(conf)
	if err != nil {
		return Config{}, errors.Wrap(err)
	}
	return conf, nil
}

type Config struct {
	path           string
	Reload         bool                        `mapstructure:"CONFIG_RELOAD"`
	ReloadInterval time.Duration               `mapstructure:"CONFIG_RELOAD_INTERVAL" validate:"required"`
	App            AppConfig                   `mapstructure:",squash"`
	Server         http.ServerConfig           `mapstructure:",squash"`
	Middleware     http.MiddlewareConfig       `mapstructure:",squash"`
	Service        service.ServiceConfig       `mapstructure:",squash"`
	Cache          cache.CacheConfig           `mapstructure:",squash"`
	Compressor     compressor.CompressorConfig `mapstructure:",squash"`
	Database       database.DatabaseConfig     `mapstructure:",squash"`
	Storage        storage.StorageConfig       `mapstructure:",squash"`
}

type AppConfig struct {
	Name    string `mapstructure:"APP_NAME"     validate:"required"`
	Ballast int64  `mapstructure:"APP_BALLAST"`
	Log     string `mapstructure:"APP_LOG"      validate:"required"`
}

func (c *Config) OnChange(ctx context.Context, fn func()) {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)
	err := w.Add(c.path)
	if err != nil {
		log.Error(errors.Wrap(err))
	}
	go func() {
		for {
			select {
			case event := <-w.Event:
				log.Debugf("watcher: event: %s\n", event)
				fn()
			case err := <-w.Error:
				log.Error(errors.Wrap(err))
			case <-w.Closed:
				return
			case <-ctx.Done():
				w.Wait()
				w.Close()
				return
			}
		}
	}()
	go func() {
		err := w.Start(c.ReloadInterval)
		if err != nil {
			log.Error(errors.Wrap(err))
		}
	}()
}

func readConfig(path string, conf *Config) error {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Error(errors.Wrap(err))
	}
	if len(viper.AllSettings()) == 0 {
		bindEnv(reflect.TypeOf(*conf))
	}
	if len(viper.AllSettings()) == 0 {
		return errors.Wrap(errors.New("no configuration is provided"))
	}
	err = viper.Unmarshal(conf)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func bindEnv(t reflect.Type) {
	if t.Kind() != reflect.Struct {
		return
	}
	for _, field := range reflect.VisibleFields(t) {
		bindEnv(field.Type)
		tag := field.Tag.Get("mapstructure")
		if field.IsExported() && tag != "" && tag != ",squash" {
			err := viper.BindEnv(strings.ToLower(tag), tag)
			if err != nil {
				log.Error(errors.Wrap(err))
			}
		}
	}
}

func fillGaps(conf *Config) {
	if conf.Compressor.IsMock() {
		conf.Database.Mem.Compressed = true
		conf.Database.Mongo.Compressed = true
		conf.Database.Badger.Compressed = true
	}
}
