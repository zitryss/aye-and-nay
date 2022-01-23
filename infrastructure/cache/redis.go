package cache

import (
	"context"
	"strings"
	"time"

	redisdb "github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/pkg/base64"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

var (
	_ domain.Cacher = (*Redis)(nil)
)

func NewRedis(ctx context.Context, conf RedisConfig) (*Redis, error) {
	client := redisdb.NewClient(&redisdb.Options{Addr: conf.Host + ":" + conf.Port})
	r := &Redis{}
	r.conf = conf
	r.client = client
	ctx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	err := retry.Do(conf.RetryTimes, conf.RetryPause, func() error {
		_, err := r.Health(ctx)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return &Redis{}, errors.Wrap(err)
	}
	r.limiter = redis_rate.NewLimiter(client)
	r.limit = redis_rate.PerSecond(conf.LimiterRequestsPerSecond)
	return r, nil
}

type Redis struct {
	conf    RedisConfig
	client  *redisdb.Client
	limiter *redis_rate.Limiter
	limit   redis_rate.Limit
}

func (r *Redis) Allow(ctx context.Context, ip uint64) (bool, error) {
	ipB64 := base64.FromUint64(ip)
	key := "ip:" + ipB64
	res, err := r.limiter.Allow(ctx, key, r.limit)
	if err != nil {
		return false, errors.Wrap(err)
	}
	return res.Allowed > 0, nil
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
	pipe.Expire(ctx, key, r.conf.TimeToLive)
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
	_ = r.client.Expire(ctx, key, r.conf.TimeToLive)
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
	err = r.client.Set(ctx, key, albumB64+":"+imageB64, r.conf.TimeToLive).Err()
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

func (r *Redis) Close(_ context.Context) error {
	err := r.client.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
