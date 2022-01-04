package compressor

import (
	"time"
)

type CompressorConfig struct {
	Compressor string           `mapstructure:"COMPRESSOR" validate:"required"`
	Shortpixel ShortpixelConfig `mapstructure:",squash"`
	Imaginary  ImaginaryConfig  `mapstructure:",squash"`
}

type ShortpixelConfig struct {
	Url             string        `mapstructure:"COMPRESSOR_SHORTPIXEL_URL"              validate:"required"`
	Url2            string        `mapstructure:"COMPRESSOR_SHORTPIXEL_URL2"             validate:"required"`
	ApiKey          string        `mapstructure:"COMPRESSOR_SHORTPIXEL_API_KEY"          validate:"required"`
	RetryTimes      int           `mapstructure:"COMPRESSOR_SHORTPIXEL_RETRY_TIMES"      validate:"required"`
	RetryPause      time.Duration `mapstructure:"COMPRESSOR_SHORTPIXEL_RETRY_PAUSE"      validate:"required"`
	Timeout         time.Duration `mapstructure:"COMPRESSOR_SHORTPIXEL_TIMEOUT"          validate:"required"`
	Wait            string        `mapstructure:"COMPRESSOR_SHORTPIXEL_WAIT"             validate:"required"`
	UploadTimeout   time.Duration `mapstructure:"COMPRESSOR_SHORTPIXEL_UPLOAD_TIMEOUT"   validate:"required"`
	DownloadTimeout time.Duration `mapstructure:"COMPRESSOR_SHORTPIXEL_DOWNLOAD_TIMEOUT" validate:"required"`
	RepeatIn        time.Duration `mapstructure:"COMPRESSOR_SHORTPIXEL_REPEAT_IN"        validate:"required"`
	RestartIn       time.Duration `mapstructure:"COMPRESSOR_SHORTPIXEL_RESTART_IN"       validate:"required"`
}

type ImaginaryConfig struct {
	Host       string        `mapstructure:"COMPRESSOR_IMAGINARY_HOST"        validate:"required"`
	Port       string        `mapstructure:"COMPRESSOR_IMAGINARY_PORT"        validate:"required"`
	RetryTimes int           `mapstructure:"COMPRESSOR_IMAGINARY_RETRY_TIMES" validate:"required"`
	RetryPause time.Duration `mapstructure:"COMPRESSOR_IMAGINARY_RETRY_PAUSE" validate:"required"`
	Timeout    time.Duration `mapstructure:"COMPRESSOR_IMAGINARY_TIMEOUT"     validate:"required"`
}

func (c CompressorConfig) IsMock() bool {
	return c.Compressor == "mock"
}

var (
	DefaultShortpixelConfig = ShortpixelConfig{
		Url:             "",
		Url2:            "",
		ApiKey:          "",
		RetryTimes:      0,
		RetryPause:      0,
		Timeout:         0,
		Wait:            "",
		UploadTimeout:   250 * time.Millisecond,
		DownloadTimeout: 250 * time.Millisecond,
		RepeatIn:        0,
		RestartIn:       0,
	}
	DefaultImaginaryConfig = ImaginaryConfig{
		Host:       "localhost",
		Port:       "9001",
		RetryTimes: 4,
		RetryPause: 5 * time.Second,
		Timeout:    30 * time.Second,
	}
)
