package cache

import (
	"time"
)

type CacheConfig struct {
	Cache string      `mapstructure:"APP_CACHE" validate:"required"`
	Mem   MemConfig   `mapstructure:",squash"`
	Redis RedisConfig `mapstructure:",squash"`
}

type MemConfig struct {
	CleanupInterval          time.Duration `mapstructure:"CACHE_MEM_CLEANUP_INTERVAL"             validate:"required"`
	LimiterRequestsPerSecond float64       `mapstructure:"MIDDLEWARE_LIMITER_REQUESTS_PER_SECOND" validate:"required"`
	LimiterBurst             int           `mapstructure:"MIDDLEWARE_LIMITER_BURST"               validate:"required"`
	TimeToLive               time.Duration `mapstructure:"CACHE_REDIS_TIME_TO_LIVE"               validate:"required"`
}

type RedisConfig struct {
	Host                     string        `mapstructure:"CACHE_REDIS_HOST"                       validate:"required"`
	Port                     string        `mapstructure:"CACHE_REDIS_PORT"                       validate:"required"`
	RetryTimes               int           `mapstructure:"CACHE_REDIS_RETRY_TIMES"                validate:"required"`
	RetryPause               time.Duration `mapstructure:"CACHE_REDIS_RETRY_PAUSE"                validate:"required"`
	Timeout                  time.Duration `mapstructure:"CACHE_REDIS_TIMEOUT"                    validate:"required"`
	LimiterRequestsPerSecond int           `mapstructure:"MIDDLEWARE_LIMITER_REQUESTS_PER_SECOND" validate:"required"`
	LimiterBurst             int64         `mapstructure:"MIDDLEWARE_LIMITER_BURST"               validate:"required"`
	TimeToLive               time.Duration `mapstructure:"CACHE_REDIS_TIME_TO_LIVE"               validate:"required"`
	TxRetries                int           `mapstructure:"CACHE_REDIS_TX_RETRIES"                 validate:"required"`
}

var (
	DefaultMemConfig = MemConfig{
		CleanupInterval:          0,
		LimiterRequestsPerSecond: 30000,
		LimiterBurst:             300,
		TimeToLive:               0,
	}
	DefaultRedisConfig = RedisConfig{
		Host:                     "localhost",
		Port:                     "6379",
		RetryTimes:               4,
		RetryPause:               5 * time.Second,
		Timeout:                  30 * time.Second,
		LimiterRequestsPerSecond: 1,
		LimiterBurst:             1,
		TimeToLive:               3 * time.Second,
		TxRetries:                1,
	}
)
