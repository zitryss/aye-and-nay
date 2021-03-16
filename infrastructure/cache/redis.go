package cache

import (
	"context"
	"strings"
	"time"

	redisdb "github.com/go-redis/redis/v8"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/base64"
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
	queueB64 := base64.FromUint64(queue)
	albumB64 := base64.FromUint64(album)
	key1 := "queue:" + queueB64 + ":set"
	ok, err := r.client.SIsMember(ctx, key1, albumB64).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	if ok {
		return nil
	}
	_, err = r.client.SAdd(ctx, key1, albumB64).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	key2 := "queue:" + queueB64 + ":list"
	_, err = r.client.RPush(ctx, key2, albumB64).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Poll(ctx context.Context, queue uint64) (uint64, error) {
	queueB64 := base64.FromUint64(queue)
	key1 := "queue:" + queueB64 + ":list"
	albumB64, err := r.client.LPop(ctx, key1).Result()
	if errors.Is(err, redisdb.Nil) {
		return 0x0, errors.Wrap(model.ErrUnknown)
	}
	key2 := "queue:" + queueB64 + ":set"
	_, err = r.client.SRem(ctx, key2, albumB64).Result()
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	album, err := base64.ToUint64(albumB64)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	return album, nil
}

func (r *Redis) Size(ctx context.Context, queue uint64) (int, error) {
	queueB64 := base64.FromUint64(queue)
	key := "queue:" + queueB64 + ":set"
	n, err := r.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return int(n), nil
}

func (r *Redis) PAdd(ctx context.Context, pqueue uint64, album uint64, expires time.Time) error {
	pqueueB64 := base64.FromUint64(pqueue)
	albumB64 := base64.FromUint64(album)
	key := "pqueue:" + pqueueB64 + ":sortedset"
	err := r.client.ZAdd(ctx, key, &redisdb.Z{Score: float64(expires.UnixNano()), Member: albumB64}).Err()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) PPoll(ctx context.Context, pqueue uint64) (uint64, time.Time, error) {
	pqueueB64 := base64.FromUint64(pqueue)
	key := "pqueue:" + pqueueB64 + ":sortedset"
	val, err := r.client.ZPopMin(ctx, key).Result()
	if err != nil {
		return 0x0, time.Time{}, errors.Wrap(err)
	}
	if len(val) == 0 {
		return 0x0, time.Time{}, errors.Wrap(model.ErrUnknown)
	}
	album, err := base64.ToUint64(val[0].Member.(string))
	if err != nil {
		return 0x0, time.Time{}, errors.Wrap(err)
	}
	expires := time.Unix(0, int64(val[0].Score))
	return album, expires, nil
}

func (r *Redis) PSize(ctx context.Context, pqueue uint64) (int, error) {
	pqueueB64 := base64.FromUint64(pqueue)
	key := "pqueue:" + pqueueB64 + ":sortedset"
	n, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return int(n), nil
}

func (r *Redis) Push(ctx context.Context, album uint64, pairs [][2]uint64) error {
	albumB64 := base64.FromUint64(album)
	key := "album:" + albumB64 + ":pairs"
	pipe := r.client.Pipeline()
	for _, images := range pairs {
		image0B64 := base64.FromUint64(images[0])
		image1B64 := base64.FromUint64(images[1])
		pipe.RPush(ctx, key, image0B64+":"+image1B64)
	}
	pipe.Expire(ctx, key, r.conf.timeToLive)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Pop(ctx context.Context, album uint64) (uint64, uint64, error) {
	albumB64 := base64.FromUint64(album)
	key := "album:" + albumB64 + ":pairs"
	n, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	if n == 0 {
		return 0x0, 0x0, errors.Wrap(model.ErrPairNotFound)
	}
	val, err := r.client.LPop(ctx, key).Result()
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	imagesB64 := strings.Split(val, ":")
	if len(imagesB64) != 2 {
		return 0x0, 0x0, errors.Wrap(model.ErrUnknown)
	}
	image0, err := base64.ToUint64(imagesB64[0])
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	image1, err := base64.ToUint64(imagesB64[1])
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	return image0, image1, nil
}

func (r *Redis) Set(ctx context.Context, album uint64, token uint64, image uint64) error {
	albumB64 := base64.FromUint64(album)
	tokenB64 := base64.FromUint64(token)
	imageB64 := base64.FromUint64(image)
	key := "album:" + albumB64 + ":token:" + tokenB64 + ":image"
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 1 {
		return errors.Wrap(model.ErrTokenAlreadyExists)
	}
	err = r.client.Set(ctx, key, imageB64, r.conf.timeToLive).Err()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, album uint64, token uint64) (uint64, error) {
	albumB64 := base64.FromUint64(album)
	tokenB64 := base64.FromUint64(token)
	key := "album:" + albumB64 + ":token:" + tokenB64 + ":image"
	imageB64, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redisdb.Nil) {
		return 0x0, errors.Wrap(model.ErrTokenNotFound)
	}
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	err = r.client.Del(ctx, key).Err()
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	image, err := base64.ToUint64(imageB64)
	if err != nil {
		return 0x0, errors.Wrap(err)
	}
	return image, nil
}
