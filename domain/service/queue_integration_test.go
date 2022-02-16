package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	. "github.com/zitryss/aye-and-nay/internal/generator"
)

func TestPQueueIntegration(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	id, _ := GenId()
	redis, err := cache.NewRedis(context.Background(), cache.DefaultRedisConfig)
	require.NoError(t, err)
	pqueue := id()
	album1 := id()
	album2 := id()
	album3 := id()
	pq := newPQueue(pqueue, redis)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pq.Monitor(ctx)
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := pq.add(ctx, album1, time.Now().Add(400*time.Millisecond))
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, album2, time.Now().Add(200*time.Millisecond))
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, album3, time.Now().Add(400*time.Millisecond))
		assert.NoError(t, err)
	}()
	album, err := pq.poll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, album2, album)
	album, err = pq.poll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, album1, album)
	album, err = pq.poll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, album3, album)
}
