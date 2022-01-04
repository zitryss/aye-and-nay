package compressor

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(ctx context.Context, conf CompressorConfig) (domain.Compresser, error) {
	switch conf.Compressor {
	case "shortpixel":
		log.Info("connecting to compressor")
		sp := NewShortpixel(conf.Shortpixel)
		err := sp.Ping(ctx)
		if err != nil {
			return nil, err
		}
		sp.Monitor()
		return sp, nil
	case "imaginary":
		log.Info("connecting to imaginary")
		return NewImaginary(ctx, conf.Imaginary)
	case "mock":
		return NewMock(), nil
	default:
		return NewMock(), nil
	}
}
