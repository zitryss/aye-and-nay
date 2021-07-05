package database

import (
	"time"

	"github.com/spf13/viper"
)

func newMemConfig() memConfig {
	return memConfig{
		compressed: viper.GetString("compressor.use") == "mock",
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
		compressed: viper.GetString("compressor.use") == "mock",
		lru:        viper.GetInt("database.mongo.lru"),
	}
}

type mongoConfig struct {
	host       string
	port       string
	times      int
	pause      time.Duration
	timeout    time.Duration
	compressed bool
	lru        int
}

func newBadgerConfig() badgerConfig {
	return badgerConfig{
		gcRatio:         viper.GetFloat64("database.badger.gcRatio"),
		cleanupInterval: viper.GetDuration("database.badger.cleanupInterval"),
		compressed:      viper.GetString("compressor.use") == "mock",
		lru:             viper.GetInt("database.badger.lru"),
	}
}

type badgerConfig struct {
	gcRatio         float64
	cleanupInterval time.Duration
	compressed      bool
	lru             int
}
