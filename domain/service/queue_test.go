//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	. "github.com/zitryss/aye-and-nay/internal/generator"
)

func TestPQueue(t *testing.T) {
	id, _ := GenId()
	mem := cache.NewMem(cache.DefaultMemConfig)
	pqueue := id()
	album1 := id()
	album2 := id()
	album3 := id()
	pq := newPQueue(pqueue, mem)
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
	start := time.Now()
	album, err := pq.poll(ctx)
	d := time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, album2, album)
	assert.True(t, 380*time.Millisecond < d && d < 420*time.Millisecond)
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, album1, album)
	assert.True(t, 80*time.Millisecond < d && d < 120*time.Millisecond)
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, album3, album)
	assert.True(t, 180*time.Millisecond < d && d < 220*time.Millisecond)
}
