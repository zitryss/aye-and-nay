package service

import (
	"github.com/spf13/viper"
)

func newServiceConfig() serviceConfig {
	return serviceConfig{
		numberOfWorkersCalc: viper.GetInt("service.numberOfWorkersCalc"),
		numberOfWorkersComp: viper.GetInt("service.numberOfWorkersComp"),
		albumIdLength:       viper.GetInt("service.albumIdLength"),
		imageIdLength:       viper.GetInt("service.imageIdLength"),
		tokenIdLength:       viper.GetInt("service.tokenIdLength"),
	}
}

type serviceConfig struct {
	numberOfWorkersCalc int
	numberOfWorkersComp int
	albumIdLength       int
	imageIdLength       int
	tokenIdLength       int
}
