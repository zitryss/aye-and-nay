package service

import (
	"github.com/spf13/viper"
)

func newServiceConfig() serviceConfig {
	return serviceConfig{
		numberOfWorkersCalc: viper.GetInt("service.numberOfWorkersCalc"),
		numberOfWorkersComp: viper.GetInt("service.numberOfWorkersComp"),
		accuracy:            viper.GetFloat64("service.accuracy"),
	}
}

type serviceConfig struct {
	numberOfWorkersCalc int
	numberOfWorkersComp int
	accuracy            float64
}
