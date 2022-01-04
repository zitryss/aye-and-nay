package service

type ServiceConfig struct {
	TempLinks           bool    `mapstructure:"SERVICE_TEMP_LINKS"`
	NumberOfWorkersCalc int     `mapstructure:"SERVICE_NUMBER_OF_WORKERS_CALC" validate:"required"`
	NumberOfWorkersComp int     `mapstructure:"SERVICE_NUMBER_OF_WORKERS_COMP" validate:"required"`
	Accuracy            float64 `mapstructure:"SERVICE_ACCURACY"               validate:"required"`
}

var (
	DefaultServiceConfig = ServiceConfig{
		TempLinks:           true,
		NumberOfWorkersCalc: 2,
		NumberOfWorkersComp: 2,
		Accuracy:            0.625,
	}
)
