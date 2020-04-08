package database

import (
	"context"
	"strings"

	redisdb "github.com/go-redis/redis/v7"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewRedis(ctx context.Context) (redis, error) {
	conf := newRedisConfig()
	client := redisdb.NewClient(&redisdb.Options{Addr: conf.host + ":" + conf.port})
	client = client.WithContext(ctx)
	err := retry.Do(conf.times, conf.pause, func() error {
		err := client.Ping().Err()
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return redis{}, errors.Wrap(err)
	}
	return redis{conf, client}, nil
}

type redis struct {
	conf   redisConfig
	client *redisdb.Client
}

func (r *redis) PopPair(_ context.Context, album string) (string, string, error) {
	key := "album:" + album + ":pairs"
	n, err := r.client.LLen(key).Result()
	if err != nil {
		return "", "", errors.Wrap(err)
	}
	if n == 0 {
		return "", "", errors.Wrap(model.ErrPairNotFound)
	}
	val, err := r.client.LPop(key).Result()
	if err != nil {
		return "", "", errors.Wrap(err)
	}
	images := strings.Split(val, ":")
	return images[0], images[1], nil
}

func (r *redis) PushPair(_ context.Context, album string, pairs [][2]string) error {
	key := "album:" + album + ":pairs"
	pipe := r.client.Pipeline()
	for _, images := range pairs {
		pipe.RPush(key, images[0]+":"+images[1])
	}
	pipe.Expire(key, r.conf.expiration)
	_, err := pipe.Exec()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *redis) GetImageId(_ context.Context, album string, token string) (string, error) {
	key := "album:" + album + ":token:" + token + ":image"
	image, err := r.client.Get(key).Result()
	if errors.Is(err, redisdb.Nil) {
		return "", errors.Wrap(model.ErrTokenNotFound)
	}
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = r.client.Del(key).Err()
	if err != nil {
		return "", errors.Wrap(err)
	}
	return image, nil
}

func (r *redis) SetToken(_ context.Context, album string, token string, image string) error {
	key := "album:" + album + ":token:" + token + ":image"
	n, err := r.client.Exists(key).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 1 {
		return errors.Wrap(model.ErrTokenAlreadyExists)
	}
	err = r.client.Set(key, image, r.conf.expiration).Err()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
