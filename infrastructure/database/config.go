package database

import (
	"time"

	"github.com/spf13/viper"
)

func newMemConfig() memConfig {
	return memConfig{
		compressed: viper.GetString("compressor.use") != "shortpixel",
	}
}

type memConfig struct {
	compressed bool
}

func newMongoConfig() mongoConfig {
	return mongoConfig{
		host:       viper.GetString("database.mongo.host"),
		port:       viper.GetString("database.mongo.port"),
		times:      viper.GetInt("database.mongo.retry.times"),
		pause:      viper.GetDuration("database.mongo.retry.pause"),
		timeout:    viper.GetDuration("database.mongo.retry.timeout"),
		compressed: viper.GetString("compressor.use") != "shortpixel",
	}
}

type mongoConfig struct {
	host       string
	port       string
	times      int
	pause      time.Duration
	timeout    time.Duration
	compressed bool
}
