package database

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/internal/log"
)

func New(ctx context.Context, conf DatabaseConfig) (domain.Databaser, error) {
	switch conf.Database {
	case "mongo":
		log.Info(context.Background(), "msg", "connecting to database")
		return NewMongo(ctx, conf.Mongo)
	case "badger":
		log.Info(context.Background(), "msg", "connecting to embedded database")
		b, err := NewBadger(conf.Badger)
		if err != nil {
			return nil, err
		}
		b.Monitor(ctx)
		return b, nil
	case "mem":
		return NewMem(conf.Mem), nil
	default:
		return NewMem(conf.Mem), nil
	}
}
