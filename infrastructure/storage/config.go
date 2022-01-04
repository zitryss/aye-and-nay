package storage

import (
	"time"
)

type StorageConfig struct {
	Storage string      `mapstructure:"STORAGE" validate:"required"`
	Minio   MinioConfig `mapstructure:",squash"`
}

type MinioConfig struct {
	Host       string        `mapstructure:"STORAGE_MINIO_HOST"        validate:"required"`
	Port       string        `mapstructure:"STORAGE_MINIO_PORT"        validate:"required"`
	AccessKey  string        `mapstructure:"STORAGE_MINIO_ACCESS_KEY"  validate:"required"`
	SecretKey  string        `mapstructure:"STORAGE_MINIO_SECRET_KEY"  validate:"required"`
	Token      string        `mapstructure:"STORAGE_MINIO_TOKEN"`
	Secure     bool          `mapstructure:"STORAGE_MINIO_SECURE"`
	RetryTimes int           `mapstructure:"STORAGE_MINIO_RETRY_TIMES" validate:"required"`
	RetryPause time.Duration `mapstructure:"STORAGE_MINIO_RETRY_PAUSE" validate:"required"`
	Timeout    time.Duration `mapstructure:"STORAGE_MINIO_TIMEOUT"     validate:"required"`
	Location   string        `mapstructure:"STORAGE_MINIO_LOCATION"    validate:"required"`
	Prefix     string        `mapstructure:"STORAGE_MINIO_PREFIX"`
}

var (
	DefaultMinioConfig = MinioConfig{
		Host:       "localhost",
		Port:       "9000",
		AccessKey:  "12345678",
		SecretKey:  "qwertyui",
		Token:      "",
		Secure:     false,
		RetryTimes: 4,
		RetryPause: 5 * time.Second,
		Timeout:    30 * time.Second,
		Location:   "eu-central-1",
		Prefix:     "",
	}
)
