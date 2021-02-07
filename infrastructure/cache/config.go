package cache

import (
	"time"

	"github.com/spf13/viper"
)

func newMemConfig() memConfig {
	return memConfig{
		timeToLive:      viper.GetDuration("cache.redis.timeToLive"),
		cleanupInterval: viper.GetDuration("cache.redis.cleanupInterval"),
	}
}

type memConfig struct {
	timeToLive      time.Duration
	cleanupInterval time.Duration
}

func newRedisConfig() redisConfig {
	return redisConfig{
		host:       viper.GetString("cache.redis.host"),
		port:       viper.GetString("cache.redis.port"),
		times:      viper.GetInt("cache.redis.retry.times"),
		pause:      viper.GetDuration("cache.redis.retry.pause"),
		timeout:    viper.GetDuration("cache.redis.retry.timeout"),
		timeToLive: viper.GetDuration("cache.redis.timeToLive"),
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
