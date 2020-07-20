package config

import (
	"github.com/spf13/viper"

	"github.com/zitryss/aye-and-nay/pkg/unit"
)

func init() {
	viper.Set("middleware.limiter.requestsPerSecond", 100)
	viper.Set("middleware.limiter.burst", 1)
	viper.Set("middleware.limiter.timeToLive", "500ms")
	viper.Set("middleware.limiter.cleanupInterval", "500ms")
	viper.Set("controller.maxNumberOfFiles", 3)
	viper.Set("controller.maxFileSize", 512*unit.KB)
	viper.Set("service.numberOfWorkersCalc", 2)
	viper.Set("service.numberOfWorkersComp", 2)
	viper.Set("shortpixel.uploadTimeout", "100ms")
	viper.Set("shortpixel.downloadTimeout", "100ms")
	viper.Set("minio.host", "localhost")
	viper.Set("minio.port", 9000)
	viper.Set("minio.accessKey", "12345678")
	viper.Set("minio.secretKey", "qwertyui")
	viper.Set("minio.secure", false)
	viper.Set("minio.retry.times", 12)
	viper.Set("minio.retry.pause", "5s")
	viper.Set("minio.retry.timeout", "5s")
	viper.Set("minio.location", "eu-central-1")
	viper.Set("mongo.host", "localhost")
	viper.Set("mongo.port", 27017)
	viper.Set("mongo.retry.times", 4)
	viper.Set("mongo.retry.pause", "5s")
	viper.Set("mongo.retry.timeout", "10s")
	viper.Set("redis.host", "localhost")
	viper.Set("redis.port", 6379)
	viper.Set("redis.retry.times", 4)
	viper.Set("redis.retry.pause", "5s")
	viper.Set("redis.timeToLive", "500ms")
	viper.Set("redis.cleanupInterval", "500ms")
}
