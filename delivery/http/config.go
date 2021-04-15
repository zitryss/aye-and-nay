package http

import (
	"time"

	"github.com/spf13/viper"
)

func newServerConfig() serverConfig {
	return serverConfig{
		host:            viper.GetString("server.host"),
		port:            viper.GetString("server.port"),
		readTimeout:     viper.GetDuration("server.readTimeout"),
		writeTimeout:    viper.GetDuration("server.writeTimeout"),
		idleTimeout:     viper.GetDuration("server.idleTimeout"),
		shutdownTimeout: viper.GetDuration("server.shutdownTimeout"),
	}
}

type serverConfig struct {
	host            string
	port            string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
}

func newMiddlewareConfig() middlewareConfig {
	return middlewareConfig{
		corsAllowOrigin: viper.GetString("middleware.cors.allowOrigin"),
	}
}

type middlewareConfig struct {
	corsAllowOrigin string
}

func newContrConfig() contrConfig {
	return contrConfig{
		maxNumberOfFiles: viper.GetInt("controller.maxNumberOfFiles"),
		maxFileSize:      viper.GetInt64("controller.maxFileSize"),
	}
}

type contrConfig struct {
	maxNumberOfFiles int
	maxFileSize      int64
}
