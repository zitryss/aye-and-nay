//go:build unit

package service

import (
	"context"
	"testing"
	"time"

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
	start := time.Now()
	album, err := pq.poll(ctx)
	d := time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if album != 0x89C1 {
		t.Error("album != 0x89C1")
	}
	if !(380*time.Millisecond < d && d < 420*time.Millisecond) {
		t.Error("!(380*time.Millisecond < d && d < 420*time.Millisecond)")
	}
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if album != 0x85D5 {
		t.Error("album != 0x85D5")
	}
	if !(80*time.Millisecond < d && d < 120*time.Millisecond) {
		t.Error("!(80*time.Millisecond < d && d < 120*time.Millisecond)")
	}
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if album != 0x97D3 {
		t.Error("album != 0x97D3")
	}
	if !(180*time.Millisecond < d && d < 220*time.Millisecond) {
		t.Error("!(180*time.Millisecond < d && d < 220*time.Millisecond)")
	}
}
