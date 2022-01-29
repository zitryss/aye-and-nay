//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/infrastructure/cache"
)

func TestPQueue(t *testing.T) {
	mem := cache.NewMem(cache.DefaultMemConfig)
	pq := newPQueue(0xFE28, mem)
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
	start := time.Now()
	album, err := pq.poll(ctx)
	d := time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x89C1), album)
	assert.True(t, 380*time.Millisecond < d && d < 420*time.Millisecond)
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x85D5), album)
	assert.True(t, 80*time.Millisecond < d && d < 120*time.Millisecond)
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x97D3), album)
	assert.True(t, 180*time.Millisecond < d && d < 220*time.Millisecond)
}
