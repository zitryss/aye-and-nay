package compressor

import (
	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(s string) (domain.Compresser, error) {
	switch s {
	case "imaginary":
		log.Info("connecting to imaginary")
		return NewImaginary()
	case "shortpixel":
		log.Info("connecting to compressor")
		sp := NewShortPixel()
		err := sp.Ping()
		if err != nil {
			return nil, err
		}
		sp.Monitor()
		return sp, nil
	case "mock":
		return NewMock(), nil
	default:
		return NewMock(), nil
	}
}
