package database

import (
	"time"

	"github.com/spf13/viper"
)

func newMemConfig() memConfig {
	return memConfig{
		timeToLive:      viper.GetDuration("redis.timeToLive"),
		cleanupInterval: viper.GetDuration("redis.cleanupInterval"),
		compressed:      !viper.GetBool("shortpixel.use"),
	}
}

type memConfig struct {
	timeToLive      time.Duration
	cleanupInterval time.Duration
	compressed      bool
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

func newRedisConfig() redisConfig {
	return redisConfig{
		host:       viper.GetString("redis.host"),
		port:       viper.GetString("redis.port"),
		times:      viper.GetInt("redis.retry.times"),
		pause:      viper.GetDuration("redis.retry.pause"),
		timeout:    viper.GetDuration("redis.retry.timeout"),
		timeToLive: viper.GetDuration("redis.timeToLive"),
	}
}

type redisConfig struct {
	host       string
	port       string
	times      int
	pause      time.Duration
	timeout    time.Duration
	timeToLive time.Duration
}
