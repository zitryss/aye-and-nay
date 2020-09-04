package storage

import (
	"time"

	"github.com/spf13/viper"
)

func newMinioConfig() minioConfig {
	return minioConfig{
		host:      viper.GetString("minio.host"),
		port:      viper.GetString("minio.port"),
		accessKey: viper.GetString("minio.accessKey"),
		secretKey: viper.GetString("minio.secretKey"),
		token:     viper.GetString("minio.token"),
		secure:    viper.GetBool("minio.secure"),
		times:     viper.GetInt("minio.retry.times"),
		pause:     viper.GetDuration("minio.retry.pause"),
		timeout:   viper.GetDuration("minio.retry.timeout"),
		location:  viper.GetString("minio.location"),
		prefix:    viper.GetString("minio.prefix"),
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
