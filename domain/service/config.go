package service

import (
	"github.com/spf13/viper"
)

func newServiceConfig() serviceConfig {
	return serviceConfig{
		numberOfWorkers: viper.GetInt("service.numberOfWorkers"),
		albumIdLength:   viper.GetInt("service.albumIdLength"),
		imageIdLength:   viper.GetInt("service.imageIdLength"),
		tokenIdLength:   viper.GetInt("service.tokenIdLength"),
	}
}

type serviceConfig struct {
	numberOfWorkers int
	albumIdLength   int
	imageIdLength   int
	tokenIdLength   int
}
