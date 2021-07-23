package http

import (
	"time"

	"github.com/spf13/viper"
)

func newServerConfig() serverConfig {
	return serverConfig{
		domain:          viper.GetString("server.domain"),
		host:            viper.GetString("server.host"),
		port:            viper.GetString("server.port"),
		h2c:             viper.GetBool("server.h2c"),
		readTimeout:     viper.GetDuration("server.readTimeout"),
		writeTimeout:    viper.GetDuration("server.writeTimeout"),
		idleTimeout:     viper.GetDuration("server.idleTimeout"),
		shutdownTimeout: viper.GetDuration("server.shutdownTimeout"),
	}
}

type serverConfig struct {
	domain          string
	host            string
	port            string
	h2c             bool
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
