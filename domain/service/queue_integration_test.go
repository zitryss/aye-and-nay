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
	pq.Monitor(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := pq.add(context.Background(), "ac/dc", time.Now().Add(400*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
		err = pq.add(context.Background(), "doors", time.Now().Add(200*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
		err = pq.add(context.Background(), "abba", time.Now().Add(400*time.Millisecond))
		if err != nil {
			t.Error(err)
		}
	}()
	album, err := pq.poll(context.Background())
	if err != nil {
		t.Error(err)
	}
	if album != "doors" {
		t.Error("album != \"doors\"")
	}
	album, err = pq.poll(context.Background())
	if err != nil {
		t.Error(err)
	}
	if album != "ac/dc" {
		t.Error("album != \"ac/dc\"")
	}
	album, err = pq.poll(context.Background())
	if err != nil {
		t.Error(err)
	}
	if album != "abba" {
		t.Error("album != \"abba\"")
	}
}
