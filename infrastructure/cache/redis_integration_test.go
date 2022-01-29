//go:build integration

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitryss/aye-and-nay/domain/domain"
)

func TestRedisAllow(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		if testing.Short() {
			t.Skip("short flag is set")
		}
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		rpm := redis.conf.LimiterRequestsPerSecond
		for j := 0; j < rpm; j++ {
			allowed, err := redis.Allow(context.Background(), 0xDEAD)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
		time.Sleep(1 * time.Second)
		for j := 0; j < rpm; j++ {
			allowed, err := redis.Allow(context.Background(), 0xDEAD)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
	})
	t.Run("Negative", func(t *testing.T) {
		t.Skip("flaky test")
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		rps := redis.conf.LimiterRequestsPerSecond
		for i := 0; i < rps; i++ {
			allowed, err := redis.Allow(context.Background(), 0xBEEF)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
		allowed, err := redis.Allow(context.Background(), 0xBEEF)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestRedisQueue(t *testing.T) {
	redis, err := NewRedis(context.Background(), DefaultRedisConfig)
	require.NoError(t, err)
	n, err := redis.Size(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	err = redis.Add(context.Background(), 0x5D6D, 0x1ED1)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), 0x5D6D, 0x1ED1)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), 0x5D6D, 0xF612)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), 0x5D6D, 0x1A83)
	assert.NoError(t, err)
	err = redis.Add(context.Background(), 0x5D6D, 0xF612)
	assert.NoError(t, err)
	n, err = redis.Size(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	album, err := redis.Poll(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x1ED1), album)
	n, err = redis.Size(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	album, err = redis.Poll(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0xF612), album)
	album, err = redis.Poll(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x1A83), album)
	n, err = redis.Size(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	album, err = redis.Poll(context.Background(), 0x5D6D)
	assert.Error(t, err)
	assert.Equal(t, uint64(0x0), album)
	n, err = redis.Size(context.Background(), 0x5D6D)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	_, err = redis.Poll(context.Background(), 0x5D6D)
	assert.ErrorIs(t, err, domain.ErrUnknown)
}

func TestRedisPQueue(t *testing.T) {
	redis, err := NewRedis(context.Background(), DefaultRedisConfig)
	require.NoError(t, err)
	n, err := redis.PSize(context.Background(), 0x7D31)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	err = redis.PAdd(context.Background(), 0x7D31, 0xE976, time.Unix(904867200, 0))
	assert.NoError(t, err)
	err = redis.PAdd(context.Background(), 0x7D31, 0xEC0E, time.Unix(1075852800, 0))
	assert.NoError(t, err)
	err = redis.PAdd(context.Background(), 0x7D31, 0x4CAF, time.Unix(681436800, 0))
	assert.NoError(t, err)
	n, err = redis.PSize(context.Background(), 0x7D31)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	album, expires, err := redis.PPoll(context.Background(), 0x7D31)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x4CAF), album)
	assert.True(t, expires.Equal(time.Unix(681436800, 0)))
	n, err = redis.PSize(context.Background(), 0x7D31)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	album, expires, err = redis.PPoll(context.Background(), 0x7D31)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0xE976), album)
	assert.True(t, expires.Equal(time.Unix(904867200, 0)))
	album, expires, err = redis.PPoll(context.Background(), 0x7D31)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0xEC0E), album)
	assert.True(t, expires.Equal(time.Unix(1075852800, 0)))
	n, err = redis.PSize(context.Background(), 0x7D31)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	_, _, err = redis.PPoll(context.Background(), 0x7D31)
	assert.ErrorIs(t, err, domain.ErrUnknown)
}

func TestRedisPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		image1 := uint64(0x3E3D)
		image2 := uint64(0xB399)
		err = redis.Push(context.Background(), 0x23D2, [][2]uint64{{image1, image2}})
		assert.NoError(t, err)
		image3, image4, err := redis.Pop(context.Background(), 0x23D2)
		assert.NoError(t, err)
		assert.Equal(t, image1, image3)
		assert.Equal(t, image2, image4)
	})
	t.Run("Negative1", func(t *testing.T) {
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		_, _, err = redis.Pop(context.Background(), 0x73BF)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		image1 := uint64(0x44DC)
		image2 := uint64(0x721B)
		err = redis.Push(context.Background(), 0x1AE9, [][2]uint64{{image1, image2}})
		assert.NoError(t, err)
		_, _, err = redis.Pop(context.Background(), 0x1AE9)
		assert.NoError(t, err)
		_, _, err = redis.Pop(context.Background(), 0x1AE9)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
}

func TestRedisToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := uint64(0xC2E7)
		album1 := uint64(0xB41C)
		image1 := uint64(0x52BD)
		err = redis.Set(context.Background(), token, album1, image1)
		assert.NoError(t, err)
		album2, image2, err := redis.Get(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, album1, album2)
		assert.Equal(t, image1, image2)
	})
	t.Run("Negative1", func(t *testing.T) {
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := uint64(0x1C4A)
		album := uint64(0xF0EE)
		image := uint64(0x583C)
		err = redis.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		err = redis.Set(context.Background(), token, album, image)
		assert.ErrorIs(t, err, domain.ErrTokenAlreadyExists)
	})
	t.Run("Negative2", func(t *testing.T) {
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := uint64(0xC4F8)
		_, _, err = redis.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := uint64(0xEB96)
		album := uint64(0xC67F)
		image := uint64(0x7C45)
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
		redis, err := NewRedis(context.Background(), DefaultRedisConfig)
		require.NoError(t, err)
		token := uint64(0xD3BF)
		err = redis.Del(context.Background(), token)
		assert.NoError(t, err)
	})
}
