package cache

import (
	"context"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New(ctx context.Context, conf CacheConfig) (domain.Cacher, error) {
	switch conf.Cache {
	case "redis":
		log.Info("connecting to cache")
		return NewRedis(ctx, conf.Redis)
	case "mem":
		mem := NewMem(conf.Mem)
		mem.Monitor()
		return mem, nil
	default:
		mem := NewMem(conf.Mem)
		mem.Monitor()
		return mem, nil
	}
}
