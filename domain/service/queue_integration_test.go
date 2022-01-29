//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitryss/aye-and-nay/infrastructure/cache"
)

func TestPQueueIntegration(t *testing.T) {
	redis, err := cache.NewRedis(context.Background(), cache.DefaultRedisConfig)
	require.NoError(t, err)
	pq := newPQueue(0xFE28, redis)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pq.Monitor(ctx)
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := pq.add(ctx, 0x85D5, time.Now().Add(400*time.Millisecond))
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, 0x89C1, time.Now().Add(200*time.Millisecond))
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, 0x97D3, time.Now().Add(400*time.Millisecond))
		assert.NoError(t, err)
	}()
	album, err := pq.poll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x89C1), album)
	album, err = pq.poll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x85D5), album)
	album, err = pq.poll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x97D3), album)
}
