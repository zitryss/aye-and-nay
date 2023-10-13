package http

import (
	"time"
)

const (
	kb = 1 << (10 * 1)
)

type ServerConfig struct {
	Domain          string           `mapstructure:"SERVER_DOMAIN"`
	Host            string           `mapstructure:"SERVER_HOST"`
	Port            string           `mapstructure:"SERVER_PORT"             validate:"required"`
	H2C             bool             `mapstructure:"SERVER_H2C"`
	ReadTimeout     time.Duration    `mapstructure:"SERVER_READ_TIMEOUT"     validate:"required"`
	WriteTimeout    time.Duration    `mapstructure:"SERVER_WRITE_TIMEOUT"    validate:"required"`
	IdleTimeout     time.Duration    `mapstructure:"SERVER_IDLE_TIMEOUT"     validate:"required"`
	ShutdownTimeout time.Duration    `mapstructure:"SERVER_SHUTDOWN_TIMEOUT" validate:"required"`
	Controller      ControllerConfig `mapstructure:",squash"`
}

type MiddlewareConfig struct {
	CorsAllowOrigin string        `mapstructure:"MIDDLEWARE_CORS_ALLOW_ORIGIN" validate:"required"`
	WriteTimeout    time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"         validate:"required"`
	MaxFileSize     int64         `mapstructure:"CONTROLLER_MAX_FILE_SIZE"     validate:"required"`
	Debug           bool          `mapstructure:"MIDDLEWARE_DEBUG"`
}

type ControllerConfig struct {
	MaxNumberOfFiles int   `mapstructure:"CONTROLLER_MAX_NUMBER_OF_FILES" validate:"required"`
	MaxFileSize      int64 `mapstructure:"CONTROLLER_MAX_FILE_SIZE"       validate:"required"`
}

var (
	DefaultServerConfig = ServerConfig{
		Domain:          "",
		Host:            "localhost",
		Port:            "8001",
		H2C:             false,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     10 * time.Second,
		ShutdownTimeout: 1 * time.Second,
		Controller:      DefaultControllerConfig,
	}
	DefaultMiddlewareConfig = MiddlewareConfig{
		CorsAllowOrigin: "",
		WriteTimeout:    10 * time.Second,
		MaxFileSize:     512 * kb,
		Debug:           false,
	}
	DefaultControllerConfig = ControllerConfig{
		MaxNumberOfFiles: 3,
		MaxFileSize:      512 * kb,
	}
)
