package database

import (
	"time"

	"github.com/spf13/viper"
)

func newMemConfig() memConfig {
	return memConfig{
		compressed: !viper.GetBool("shortpixel.use"),
	}
}

type memConfig struct {
	compressed bool
}

func newMongoConfig() mongoConfig {
	return mongoConfig{
		host:       viper.GetString("mongo.host"),
		port:       viper.GetString("mongo.port"),
		times:      viper.GetInt("mongo.retry.times"),
		pause:      viper.GetDuration("mongo.retry.pause"),
		timeout:    viper.GetDuration("mongo.retry.timeout"),
		compressed: !viper.GetBool("shortpixel.use"),
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
