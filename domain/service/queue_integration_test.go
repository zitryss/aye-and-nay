//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/infrastructure/cache"
)

func TestPQueueIntegration(t *testing.T) {
	redis, err := cache.NewRedis()
	if err != nil {
		t.Fatal(err)
	}
	pq := newPQueue(0xFE28, redis)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pq.Monitor(ctx)
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := pq.add(ctx, 0x85D5, time.Now().Add(400*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, 0x89C1, time.Now().Add(200*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, 0x97D3, time.Now().Add(400*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
	}()
	album, err := pq.poll(ctx)
	if err != nil {
		t.Error(err)
	}
	if album != 0x89C1 {
		t.Error("album != 0x89C1")
	}
	album, err = pq.poll(ctx)
	if err != nil {
		t.Error(err)
	}
	if album != 0x85D5 {
		t.Error("album != 0x85D5")
	}
	album, err = pq.poll(ctx)
	if err != nil {
		t.Error(err)
	}
	if album != 0x97D3 {
		t.Error("album != 0x97D3")
	}
}
