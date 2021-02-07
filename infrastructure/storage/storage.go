package storage

import (
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(s string) (model.Storager, error) {
	switch s {
	case "minio":
		log.Info("connecting to storage")
		return NewMinio()
	case "mock":
		return NewMock(), nil
	default:
		return NewMock(), nil
	}
}
