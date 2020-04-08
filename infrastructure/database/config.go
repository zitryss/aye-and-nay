package database

import (
	"time"

	"github.com/spf13/viper"
)

func newMongoConfig() mongoConfig {
	return mongoConfig{
		host:    viper.GetString("mongo.host"),
		port:    viper.GetString("mongo.port"),
		times:   viper.GetInt("mongo.retry.times"),
		pause:   viper.GetDuration("mongo.retry.pause"),
		timeout: viper.GetDuration("mongo.retry.timeout"),
	}
}

type mongoConfig struct {
	host    string
	port    string
	times   int
	pause   time.Duration
	timeout time.Duration
}

func newRedisConfig() redisConfig {
	return redisConfig{
		host:       viper.GetString("redis.host"),
		port:       viper.GetString("redis.port"),
		times:      viper.GetInt("redis.retry.times"),
		pause:      viper.GetDuration("redis.retry.pause"),
		expiration: viper.GetDuration("redis.expiration"),
	}
}

type redisConfig struct {
	host       string
	port       string
	times      int
	pause      time.Duration
	expiration time.Duration
}
