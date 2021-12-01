package cache

import (
	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(s string) (domain.Cacher, error) {
	switch s {
	case "redis":
		log.Info("connecting to cache")
		return NewRedis()
	case "mem":
		mem := NewMem()
		mem.Monitor()
		return mem, nil
	default:
		mem := NewMem()
		mem.Monitor()
		return mem, nil
	}
}
