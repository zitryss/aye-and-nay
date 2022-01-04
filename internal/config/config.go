package config

import (
	"reflect"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator"
	"github.com/spf13/viper"

	"github.com/zitryss/aye-and-nay/delivery/http"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func New(path string) (Config, error) {
	viper.Reset()
	conf := Config{}
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

func OnChange(run func()) {
	viper.OnConfigChange(func(_ fsnotify.Event) {
		run()
	})
	viper.WatchConfig()
}

type Config struct {
	App        AppConfig                   `mapstructure:",squash"`
	Server     http.ServerConfig           `mapstructure:",squash"`
	Middleware http.MiddlewareConfig       `mapstructure:",squash"`
	Service    service.ServiceConfig       `mapstructure:",squash"`
	Cache      cache.CacheConfig           `mapstructure:",squash"`
	Compressor compressor.CompressorConfig `mapstructure:",squash"`
	Database   database.DatabaseConfig     `mapstructure:",squash"`
	Storage    storage.StorageConfig       `mapstructure:",squash"`
}

type AppConfig struct {
	Name    string `mapstructure:"APP_NAME"    validate:"required"`
	Ballast int64  `mapstructure:"APP_BALLAST"`
	Log     string `mapstructure:"APP_LOG"     validate:"required"`
}

func readConfig(path string, conf *Config) error {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
	if len(viper.AllSettings()) == 0 {
		bindEnv(reflect.TypeOf(*conf))
	}
	if len(viper.AllSettings()) == 0 {
		return errors.Wrap(errors.New("no configuration is provided"))
	}
	err := viper.Unmarshal(conf)
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
			_ = viper.BindEnv(strings.ToLower(tag), tag)
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
