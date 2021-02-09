package compressor

import (
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(s string) (model.Compresser, error) {
	switch s {
	case "shortpixel":
		log.Info("connecting to compressor")
		sp := NewShortPixel()
		err := sp.Ping()
		if err != nil {
			return nil, err
		}
		sp.Monitor()
		return sp, nil
	case "imaginary":
		return NewImaginary()
	case "mock":
		return NewMock(), nil
	default:
		return NewMock(), nil
	}
}
