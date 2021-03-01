package service

import (
	"context"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/infrastructure/cache"
)

func TestPQueue(t *testing.T) {
	mem := cache.NewMem()
	pq := newPQueue("WM5BtzjdncQtExgY", mem)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pq.Monitor(ctx)
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := pq.add(ctx, "ac/dc", time.Now().Add(400*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, "doors", time.Now().Add(200*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
		err = pq.add(ctx, "abba", time.Now().Add(400*time.Millisecond))
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
	if album != "doors" {
		t.Error("album != \"doors\"")
	}
	if !(390*time.Millisecond < d && d < 410*time.Millisecond) {
		t.Error("!(390*time.Millisecond < d && d < 410*time.Millisecond)")
	}
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if album != "ac/dc" {
		t.Error("album != \"ac/dc\"")
	}
	if !(90*time.Millisecond < d && d < 110*time.Millisecond) {
		t.Error("!(90*time.Millisecond < d && d < 110*time.Millisecond)")
	}
	start = time.Now()
	album, err = pq.poll(ctx)
	d = time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if album != "abba" {
		t.Error("album != \"abba\"")
	}
	if !(190*time.Millisecond < d && d < 210*time.Millisecond) {
		t.Error("!(190*time.Millisecond < d && d < 210*time.Millisecond)")
	}
}
