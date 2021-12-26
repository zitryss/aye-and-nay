package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	redisdb "github.com/go-redis/redis/v8"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/base64"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

func NewRedis() (*Redis, error) {
	conf := newRedisConfig()
	client := redisdb.NewClient(&redisdb.Options{Addr: conf.host + ":" + conf.port})
	r := &Redis{conf, client}
	ctx, cancel := context.WithTimeout(context.Background(), conf.timeout)
	defer cancel()
	err := retry.Do(conf.times, conf.pause, func() error {
		_, err := r.Health(ctx)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return &Redis{}, errors.Wrap(err)
	}
	return r, nil
}

type Redis struct {
	conf   redisConfig
	client *redisdb.Client
}

func (r *Redis) Allow(ctx context.Context, ip uint64) (bool, error) {
	ipB64 := base64.FromUint64(ip)
	key := "ip:" + ipB64
	value, err := r.client.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redisdb.Nil) {
		return false, errors.Wrap(err)
	}
	if errors.Is(err, redisdb.Nil) {
		value = "-1"
	}
	count, err := strconv.Atoi(value)
	if err != nil {
		return false, errors.Wrap(err)
	}
	if count >= r.conf.limiterRequestsPerMinute {
		return false, nil
	}
	pipe := r.client.Pipeline()
	pipe.IncrBy(ctx, key, r.conf.limiterBurst)
	pipe.Expire(ctx, key, 59*time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, errors.Wrap(err)
	}
	return true, nil
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
		return 0x0, errors.Wrap(domain.ErrUnknown)
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
		return 0x0, time.Time{}, errors.Wrap(domain.ErrUnknown)
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
		return 0x0, 0x0, errors.Wrap(domain.ErrPairNotFound)
	}
	val, err := r.client.LPop(ctx, key).Result()
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	_ = r.client.Expire(ctx, key, r.conf.timeToLive)
	imagesB64 := strings.Split(val, ":")
	if len(imagesB64) != 2 {
		return 0x0, 0x0, errors.Wrap(domain.ErrUnknown)
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

func (r *Redis) Set(ctx context.Context, token uint64, album uint64, image uint64) error {
	tokenB64 := base64.FromUint64(token)
	albumB64 := base64.FromUint64(album)
	imageB64 := base64.FromUint64(image)
	key := "token:" + tokenB64
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 1 {
		return errors.Wrap(domain.ErrTokenAlreadyExists)
	}
	err = r.client.Set(ctx, key, albumB64+":"+imageB64, r.conf.timeToLive).Err()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, token uint64) (uint64, uint64, error) {
	tokenB64 := base64.FromUint64(token)
	key := "token:" + tokenB64
	s, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redisdb.Nil) {
		return 0x0, 0x0, errors.Wrap(domain.ErrTokenNotFound)
	}
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	ss := strings.Split(s, ":")
	if len(ss) != 2 {
		return 0x0, 0x0, errors.Wrap(domain.ErrUnknown)
	}
	album, err := base64.ToUint64(ss[0])
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	image, err := base64.ToUint64(ss[1])
	if err != nil {
		return 0x0, 0x0, errors.Wrap(err)
	}
	return album, image, nil
}

func (r *Redis) Del(ctx context.Context, token uint64) error {
	tokenB64 := base64.FromUint64(token)
	key := "token:" + tokenB64
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *Redis) Health(ctx context.Context) (bool, error) {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthCache, "%s", err)
	}
	return true, nil
}

func (r *Redis) Close() error {
	err := r.client.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
