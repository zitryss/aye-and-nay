package database

import (
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(s string) (model.Databaser, error) {
	switch s {
	case "mongo":
		log.Info("connecting to database")
		return NewMongo()
	case "badger":
		log.Info("connecting to embedded database")
		b, err := NewBadger(disk)
		if err != nil {
			return nil, err
		}
		return b, nil
	case "mem":
		return NewMem(), nil
	default:
		return NewMem(), nil
	}
}
