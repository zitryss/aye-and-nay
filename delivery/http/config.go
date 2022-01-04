package http

import (
	"time"

	"github.com/zitryss/aye-and-nay/pkg/unit"
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
	CorsAllowOrigin string `mapstructure:"MIDDLEWARE_CORS_ALLOW_ORIGIN" validate:"required"`
}

type ControllerConfig struct {
	MaxNumberOfFiles int   `mapstructure:"CONTROLLER_MAX_NUMBER_OF_FILES" validate:"required"`
	MaxFileSize      int64 `mapstructure:"CONTROLLER_MAX_FILE_SIZE"       validate:"required"`
}

var (
	DefaultMiddlewareConfig = MiddlewareConfig{
		CorsAllowOrigin: "",
	}
	DefaultControllerConfig = ControllerConfig{
		MaxNumberOfFiles: 3,
		MaxFileSize:      512 * unit.KB,
	}
)
