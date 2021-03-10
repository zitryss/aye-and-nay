package service

import (
	"github.com/spf13/viper"
)

func newServiceConfig() serviceConfig {
	return serviceConfig{
		numberOfWorkersCalc: viper.GetInt("service.numberOfWorkersCalc"),
		numberOfWorkersComp: viper.GetInt("service.numberOfWorkersComp"),
	}
}

type serviceConfig struct {
	numberOfWorkersCalc int
	numberOfWorkersComp int
}
