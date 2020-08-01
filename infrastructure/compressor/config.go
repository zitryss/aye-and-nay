package compressor

import (
	"time"

	"github.com/spf13/viper"
)

func newShortPixelConfig() shortPixelConfig {
	return shortPixelConfig{
		url:             viper.GetString("shortpixel.url"),
		url2:            viper.GetString("shortpixel.url2"),
		apiKey:          viper.GetString("shortpixel.apiKey"),
		times:           viper.GetInt("shortpixel.retry.times"),
		pause:           viper.GetDuration("shortpixel.retry.pause"),
		wait:            viper.GetString("shortpixel.wait"),
		uploadTimeout:   viper.GetDuration("shortpixel.uploadTimeout"),
		downloadTimeout: viper.GetDuration("shortpixel.downloadTimeout"),
		repeatIn:        viper.GetDuration("shortpixel.repeatIn"),
		restartIn:       viper.GetDuration("shortpixel.restartIn"),
	}
}

type shortPixelConfig struct {
	url             string
	url2            string
	apiKey          string
	times           int
	pause           time.Duration
	wait            string
	uploadTimeout   time.Duration
	downloadTimeout time.Duration
	repeatIn        time.Duration
	restartIn       time.Duration
}
