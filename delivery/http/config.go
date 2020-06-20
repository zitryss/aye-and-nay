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
		certFile:        viper.GetString("server.certFile"),
		keyFile:         viper.GetString("server.keyFile"),
		shutdownTimeout: viper.GetDuration("server.shutdownTimeout"),
	}
}

type serverConfig struct {
	host            string
	port            string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	certFile        string
	keyFile         string
	shutdownTimeout time.Duration
}

func newMiddlewareConfig() middlewareConfig {
	return middlewareConfig{
		limiterRequestsPerSecond: viper.GetFloat64("middleware.limiter.requestsPerSecond"),
		limiterBurst:             viper.GetInt("middleware.limiter.burst"),
		limiterTimeToLive:        viper.GetDuration("middleware.limiter.timeToLive"),
		limiterCleanupInterval:   viper.GetDuration("middleware.limiter.cleanupInterval"),
	}
}

type middlewareConfig struct {
	limiterRequestsPerSecond float64
	limiterBurst             int
	limiterTimeToLive        time.Duration
	limiterCleanupInterval   time.Duration
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
