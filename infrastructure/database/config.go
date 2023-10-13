package database

import (
	"time"
)

type DatabaseConfig struct {
	Database string       `mapstructure:"APP_DATABASE" validate:"required"`
	Mem      MemConfig    `mapstructure:",squash"`
	Mongo    MongoConfig  `mapstructure:",squash"`
	Badger   BadgerConfig `mapstructure:",squash"`
}

type MemConfig struct {
	Compressed bool
}

type MongoConfig struct {
	Host       string        `mapstructure:"DATABASE_MONGO_HOST"        validate:"required"`
	Port       string        `mapstructure:"DATABASE_MONGO_PORT"        validate:"required"`
	RetryTimes int           `mapstructure:"DATABASE_MONGO_RETRY_TIMES" validate:"required"`
	RetryPause time.Duration `mapstructure:"DATABASE_MONGO_RETRY_PAUSE" validate:"required"`
	Timeout    time.Duration `mapstructure:"DATABASE_MONGO_TIMEOUT"     validate:"required"`
	LRU        int           `mapstructure:"DATABASE_MONGO_LRU"         validate:"required"`
	Compressed bool
}

type BadgerConfig struct {
	InMemory        bool          `mapstructure:"DATABASE_BADGER_IN_MEMORY"`
	GcRatio         float64       `mapstructure:"DATABASE_BADGER_GC_RATIO"         validate:"required"`
	CleanupInterval time.Duration `mapstructure:"DATABASE_BADGER_CLEANUP_INTERVAL" validate:"required"`
	LRU             int           `mapstructure:"DATABASE_BADGER_LRU"              validate:"required"`
	Compressed      bool
}

var (
	DefaultMemConfig = MemConfig{
		Compressed: false,
	}
	DefaultMongoConfig = MongoConfig{
		Host:       "localhost",
		Port:       "27017",
		RetryTimes: 4,
		RetryPause: 5 * time.Second,
		Timeout:    30 * time.Second,
		LRU:        1,
		Compressed: false,
	}
	DefaultBadgerConfig = BadgerConfig{
		InMemory:        true,
		GcRatio:         0,
		CleanupInterval: 0,
		LRU:             1,
		Compressed:      false,
	}
)
