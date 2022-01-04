package storage

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(ctx context.Context, conf StorageConfig) (domain.Storager, error) {
	switch conf.Storage {
	case "minio":
		log.Info("connecting to storage")
		return NewMinio(ctx, conf.Minio)
	case "mock":
		return NewMock(), nil
	default:
		return NewMock(), nil
	}
}
