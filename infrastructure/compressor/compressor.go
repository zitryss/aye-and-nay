package compressor

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/internal/log"
)

func New(ctx context.Context, conf CompressorConfig) (domain.Compresser, error) {
	switch conf.Compressor {
	case "shortpixel":
		log.Info(context.Background(), "connecting to compressor")
		sp := NewShortpixel(conf.Shortpixel)
		err := sp.Ping(ctx)
		if err != nil {
			return nil, err
		}
		sp.Monitor(ctx)
		return sp, nil
	case "imaginary":
		log.Info(context.Background(), "connecting to imaginary")
		return NewImaginary(ctx, conf.Imaginary)
	case "mock":
		return NewMock(), nil
	default:
		return NewMock(), nil
	}
}
