package config

import (
	"github.com/spf13/viper"

	"github.com/zitryss/aye-and-nay/pkg/unit"
)

func init() {
	viper.Set("middleware.limiter.requestsPerSecond", 100)
	viper.Set("middleware.limiter.burst", 1)
	viper.Set("middleware.limiter.timeToLive", "0ms")
	viper.Set("controller.maxNumberOfFiles", 3)
	viper.Set("controller.maxFileSize", 512*unit.KB)
	viper.Set("service.numberOfWorkersCalc", 2)
	viper.Set("service.numberOfWorkersComp", 2)
	viper.Set("service.albumIdLength", 8)
	viper.Set("service.imageIdLength", 8)
	viper.Set("service.tokenIdLength", 8)
	viper.Set("cache.redis.host", "localhost")
	viper.Set("cache.redis.port", 6379)
	viper.Set("cache.redis.retry.times", 4)
	viper.Set("cache.redis.retry.pause", "5s")
	viper.Set("cache.redis.retry.timeout", "30s")
	viper.Set("cache.redis.timeToLive", "1s")
	viper.Set("compressor.use", "notamock")
	viper.Set("compressor.imaginary.host", "localhost")
	viper.Set("compressor.imaginary.port", 9000)
	viper.Set("compressor.imaginary.retry.times", 4)
	viper.Set("compressor.imaginary.retry.pause", "5s")
	viper.Set("compressor.imaginary.retry.timeout", "30s")
	viper.Set("compressor.shortpixel.uploadTimeout", "250ms")
	viper.Set("compressor.shortpixel.downloadTimeout", "250ms")
	viper.Set("database.mongo.host", "localhost")
	viper.Set("database.mongo.port", 27017)
	viper.Set("database.mongo.retry.times", 4)
	viper.Set("database.mongo.retry.pause", "5s")
	viper.Set("database.mongo.retry.timeout", "30s")
	viper.Set("storage.minio.host", "localhost")
	viper.Set("storage.minio.port", 9000)
	viper.Set("storage.minio.accessKey", "12345678")
	viper.Set("storage.minio.secretKey", "qwertyui")
	viper.Set("storage.minio.token", "")
	viper.Set("storage.minio.secure", false)
	viper.Set("storage.minio.retry.times", 4)
	viper.Set("storage.minio.retry.pause", "5s")
	viper.Set("storage.minio.retry.timeout", "30s")
	viper.Set("storage.minio.location", "eu-central-1")
}
