package cache

import (
	"context"
	"strings"
	"time"

	redisdb "github.com/go-redis/redis/v8"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewRedis() (*Redis, error) {
	conf := newRedisConfig()
	client := redisdb.NewClient(&redisdb.Options{Addr: conf.host + ":" + conf.port})
	ctx, cancel := context.WithTimeout(context.Background(), conf.timeout)
	defer cancel()
	err := retry.Do(conf.times, conf.pause, func() error {
		err := client.Ping(ctx).Err()
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return &Redis{}, errors.Wrap(err)
	}
	return &Redis{conf, client}, nil
}

type Redis struct {
	conf   redisConfig
	client *redisdb.Client
}

func (r *Redis) Add(ctx context.Context, queue uint64, album uint64) error {
	key1 := "queue:" + queue + ":set"
	ok, err := r.client.SIsMember(ctx, key1, album).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	if ok {
		return nil
	}
	_, err = r.client.SAdd(ctx, key1, album).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	key2 := "queue:" + queue + ":list"
	_, err = r.client.RPush(ctx, key2, album).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Poll(ctx context.Context, queue uint64) (uint64, error) {
	key1 := "queue:" + queue + ":list"
	album, err := r.client.LPop(ctx, key1).Result()
	if errors.Is(err, redisdb.Nil) {
		return "", errors.Wrap(model.ErrUnknown)
	}
	key2 := "queue:" + queue + ":set"
	_, err = r.client.SRem(ctx, key2, album).Result()
	if err != nil {
		return "", errors.Wrap(err)
	}
	return album, nil
}

func (r *Redis) Size(ctx context.Context, queue uint64) (int, error) {
	key := "queue:" + queue + ":set"
	n, err := r.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return int(n), nil
}

func (r *Redis) PAdd(ctx context.Context, pqueue uint64, album uint64, expires time.Time) error {
	key := "pqueue:" + pqueue + ":sortedset"
	err := r.client.ZAdd(ctx, key, &redisdb.Z{Score: float64(expires.UnixNano()), Member: album}).Err()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) PPoll(ctx context.Context, pqueue uint64) (uint64, time.Time, error) {
	key := "pqueue:" + pqueue + ":sortedset"
	val, err := r.client.ZPopMin(ctx, key).Result()
	if err != nil {
		return "", time.Time{}, errors.Wrap(err)
	}
	if len(val) == 0 {
		return "", time.Time{}, errors.Wrap(model.ErrUnknown)
	}
	album := val[0].Member.(string)
	expires := time.Unix(0, int64(val[0].Score))
	return album, expires, nil
}

func (r *Redis) PSize(ctx context.Context, pqueue uint64) (int, error) {
	key := "pqueue:" + pqueue + ":sortedset"
	n, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return int(n), nil
}

func (r *Redis) Push(ctx context.Context, album uint64, pairs [][2]uint64) error {
	key := "album:" + album + ":pairs"
	pipe := r.client.Pipeline()
	for _, images := range pairs {
		pipe.RPush(ctx, key, images[0]+":"+images[1])
	}
	pipe.Expire(ctx, key, r.conf.timeToLive)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Pop(ctx context.Context, album uint64) (uint64, uint64, error) {
	key := "album:" + album + ":pairs"
	n, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return "", "", errors.Wrap(err)
	}
	if n == 0 {
		return "", "", errors.Wrap(model.ErrPairNotFound)
	}
	val, err := r.client.LPop(ctx, key).Result()
	if err != nil {
		return "", "", errors.Wrap(err)
	}
	images := strings.Split(val, ":")
	if len(images) != 2 {
		return "", "", errors.Wrap(model.ErrUnknown)
	}
	return images[0], images[1], nil
}

func (r *Redis) Set(ctx context.Context, album uint64, token uint64, image uint64) error {
	key := "album:" + album + ":token:" + token + ":image"
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 1 {
		return errors.Wrap(model.ErrTokenAlreadyExists)
	}
	err = r.client.Set(ctx, key, image, r.conf.timeToLive).Err()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, album uint64, token uint64) (uint64, error) {
	key := "album:" + album + ":token:" + token + ":image"
	image, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redisdb.Nil) {
		return "", errors.Wrap(model.ErrTokenNotFound)
	}
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = r.client.Del(ctx, key).Err()
	if err != nil {
		return "", errors.Wrap(err)
	}
	return image, nil
}
