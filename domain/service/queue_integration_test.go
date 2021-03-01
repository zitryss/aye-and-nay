// +build integration

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
	pq := newPQueue("WM5BtzjdncQtExgY", redis)
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
	album, err := pq.poll(ctx)
	if err != nil {
		t.Error(err)
	}
	if album != "doors" {
		t.Error("album != \"doors\"")
	}
	album, err = pq.poll(ctx)
	if err != nil {
		t.Error(err)
	}
	if album != "ac/dc" {
		t.Error("album != \"ac/dc\"")
	}
	album, err = pq.poll(ctx)
	if err != nil {
		t.Error(err)
	}
	if album != "abba" {
		t.Error("album != \"abba\"")
	}
}
