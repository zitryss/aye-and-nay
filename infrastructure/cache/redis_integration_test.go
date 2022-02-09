//go:build integration

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitryss/aye-and-nay/domain/domain"
	. "github.com/zitryss/aye-and-nay/internal/generator"
)

func TestRedisAllow(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		rpm := redis.conf.LimiterRequestsPerSecond
		ip := id()
		for j := 0; j < rpm; j++ {
			allowed, err := redis.Allow(context.Background(), ip)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
		time.Sleep(1 * time.Second)
		for j := 0; j < rpm; j++ {
			allowed, err := redis.Allow(context.Background(), ip)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
	})
	t.Run("Negative", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		rps := redis.conf.LimiterRequestsPerSecond
		ip := id()
		for i := 0; i < rps; i++ {
			allowed, err := redis.Allow(context.Background(), ip)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
		allowed, err := redis.Allow(context.Background(), ip)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestRedisQueue(t *testing.T) {
	id, _ := GenId()
	redis, err := NewRedis(context.Background(), DefaultRedisConfig)
	require.NoError(t, err)
	queue := id()
	albumExp1 := id()
	albumExp2 := id()
	albumExp3 := id()
	n, err := redis.Size(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	err = redis.Add(context.Background(), queue, albumExp1)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), queue, albumExp1)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), queue, albumExp2)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), queue, albumExp3)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), queue, albumExp2)
	assert.NoError(t, err)
	n, err = redis.Size(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	album, err := redis.Poll(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, albumExp1, album)
	n, err = redis.Size(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	album, err = redis.Poll(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, albumExp2, album)
	album, err = redis.Poll(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, albumExp3, album)
	n, err = redis.Size(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	album, err = redis.Poll(context.Background(), queue)
	assert.Error(t, err)
	assert.Equal(t, uint64(0x0), album)
	n, err = redis.Size(context.Background(), queue)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	_, err = redis.Poll(context.Background(), queue)
	assert.ErrorIs(t, err, domain.ErrUnknown)
}

func TestRedisPQueue(t *testing.T) {
	id, _ := GenId()
	redis, err := NewRedis(context.Background(), DefaultRedisConfig)
	require.NoError(t, err)
	pqueue := id()
	albumExp1 := id()
	albumExp2 := id()
	albumExp3 := id()
	n, err := redis.PSize(context.Background(), pqueue)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	err = redis.PAdd(context.Background(), pqueue, albumExp1, time.Unix(904867200, 0))
	assert.NoError(t, err)
	err = redis.PAdd(context.Background(), pqueue, albumExp2, time.Unix(1075852800, 0))
	assert.NoError(t, err)
	err = redis.PAdd(context.Background(), pqueue, albumExp3, time.Unix(681436800, 0))
	assert.NoError(t, err)
	n, err = redis.PSize(context.Background(), pqueue)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	album, expires, err := redis.PPoll(context.Background(), pqueue)
	assert.NoError(t, err)
	assert.Equal(t, albumExp3, album)
	assert.True(t, expires.Equal(time.Unix(681436800, 0)))
	n, err = redis.PSize(context.Background(), pqueue)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	album, expires, err = redis.PPoll(context.Background(), pqueue)
	assert.NoError(t, err)
	assert.Equal(t, albumExp1, album)
	assert.True(t, expires.Equal(time.Unix(904867200, 0)))
	album, expires, err = redis.PPoll(context.Background(), pqueue)
	assert.NoError(t, err)
	assert.Equal(t, albumExp2, album)
	assert.True(t, expires.Equal(time.Unix(1075852800, 0)))
	n, err = redis.PSize(context.Background(), pqueue)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	_, _, err = redis.PPoll(context.Background(), pqueue)
	assert.ErrorIs(t, err, domain.ErrUnknown)
}

func TestRedisPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		album := id()
		image1 := id()
		image2 := id()
		pairs := [][2]uint64{{image1, image2}}
		err = redis.Push(context.Background(), album, pairs)
		assert.NoError(t, err)
		image3, image4, err := redis.Pop(context.Background(), album)
		assert.NoError(t, err)
		assert.Equal(t, image1, image3)
		assert.Equal(t, image2, image4)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		album := id()
		_, _, err = redis.Pop(context.Background(), album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		album := id()
		image1 := id()
		image2 := id()
		pairs := [][2]uint64{{image1, image2}}
		err = redis.Push(context.Background(), album, pairs)
		assert.NoError(t, err)
		_, _, err = redis.Pop(context.Background(), album)
		assert.NoError(t, err)
		_, _, err = redis.Pop(context.Background(), album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
}

func TestRedisToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := id()
		albumExp := id()
		imageExp := id()
		err = redis.Set(context.Background(), token, albumExp, imageExp)
		assert.NoError(t, err)
		album, image, err := redis.Get(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, albumExp, album)
		assert.Equal(t, imageExp, image)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := id()
		album := id()
		image := id()
		err = redis.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		err = redis.Set(context.Background(), token, album, image)
		assert.ErrorIs(t, err, domain.ErrTokenAlreadyExists)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := id()
		_, _, err = redis.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := id()
		album := id()
		image := id()
		err = redis.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		_, _, err = redis.Get(context.Background(), token)
		assert.NoError(t, err)
		err = redis.Del(context.Background(), token)
		assert.NoError(t, err)
		err = redis.Del(context.Background(), token)
		assert.NoError(t, err)
		_, _, err = redis.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	t.Run("Negative4", func(t *testing.T) {
		id, _ := GenId()
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := id()
		err = redis.Del(context.Background(), token)
		assert.NoError(t, err)
	})
}
