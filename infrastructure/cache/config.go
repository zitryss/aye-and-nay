package cache

import (
	"time"

	"github.com/spf13/viper"
)

func newMemConfig() memConfig {
	return memConfig{
		timeToLive:      viper.GetDuration("redis.timeToLive"),
		cleanupInterval: viper.GetDuration("redis.cleanupInterval"),
	}
}

type memConfig struct {
	timeToLive      time.Duration
	cleanupInterval time.Duration
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
