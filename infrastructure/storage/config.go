package storage

import (
	"time"

	"github.com/spf13/viper"
)

func newMinioConfig() minioConfig {
	return minioConfig{
		host:      viper.GetString("storage.minio.host"),
		port:      viper.GetString("storage.minio.port"),
		accessKey: viper.GetString("storage.minio.accessKey"),
		secretKey: viper.GetString("storage.minio.secretKey"),
		token:     viper.GetString("storage.minio.token"),
		secure:    viper.GetBool("storage.minio.secure"),
		times:     viper.GetInt("storage.minio.retry.times"),
		pause:     viper.GetDuration("storage.minio.retry.pause"),
		timeout:   viper.GetDuration("storage.minio.retry.timeout"),
		location:  viper.GetString("storage.minio.location"),
		prefix:    viper.GetString("storage.minio.prefix"),
	}
}

type minioConfig struct {
	host      string
	port      string
	accessKey string
	secretKey string
	token     string
	secure    bool
	times     int
	pause     time.Duration
	timeout   time.Duration
	location  string
	prefix    string
}
