// +build integration

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/domain/domain"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestRedisAllow(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		if testing.Short() {
			t.Skip("short flag is set")
		}
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		rpm := redis.conf.limiterRequestsPerMinute
		for j := 0; j < rpm; j++ {
			allowed, err := redis.Allow(context.Background(), 0xDEAD)
			if err != nil {
				t.Error(err)
			}
			if !allowed {
				t.Error("!allowed")
			}
		}
		time.Sleep(60 * time.Second)
		for j := 0; j < rpm; j++ {
			allowed, err := redis.Allow(context.Background(), 0xDEAD)
			if err != nil {
				t.Error(err)
			}
			if !allowed {
				t.Error("!allowed")
			}
		}
	})
	t.Run("Negative", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		rps := redis.conf.limiterRequestsPerMinute
		for i := 0; i < rps; i++ {
			allowed, err := redis.Allow(context.Background(), 0xBEEF)
			if err != nil {
				t.Error(err)
			}
			if !allowed {
				t.Error("!allowed")
			}
		}
		allowed, err := redis.Allow(context.Background(), 0xBEEF)
		if err != nil {
			t.Error(err)
		}
		if allowed {
			t.Error("allowed")
		}
	})
}

func TestRedisQueue(t *testing.T) {
	redis, err := NewRedis()
	if err != nil {
		t.Fatal(err)
	}
	n, err := redis.Size(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	err = redis.Add(context.Background(), 0x5D6D, 0x1ED1)
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), 0x5D6D, 0x1ED1)
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), 0x5D6D, 0xF612)
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), 0x5D6D, 0x1A83)
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), 0x5D6D, 0xF612)
	if err != nil {
		t.Error(err)
	}
	n, err = redis.Size(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if n != 3 {
		t.Error("n != 3")
	}
	album, err := redis.Poll(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if album != 0x1ED1 {
		t.Error("album != 0x1ED1")
	}
	n, err = redis.Size(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if n != 2 {
		t.Error("n != 2")
	}
	album, err = redis.Poll(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if album != 0xF612 {
		t.Error("album != 0xF612")
	}
	album, err = redis.Poll(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if album != 0x1A83 {
		t.Error("album != 0x1A83")
	}
	n, err = redis.Size(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	album, err = redis.Poll(context.Background(), 0x5D6D)
	if err == nil {
		t.Error(err)
	}
	if album != 0x0 {
		t.Error("album != \"0x0\"")
	}
	n, err = redis.Size(context.Background(), 0x5D6D)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	_, err = redis.Poll(context.Background(), 0x5D6D)
	if !errors.Is(err, domain.ErrUnknown) {
		t.Error(err)
	}
}

func TestRedisPQueue(t *testing.T) {
	redis, err := NewRedis()
	if err != nil {
		t.Fatal(err)
	}
	n, err := redis.PSize(context.Background(), 0x7D31)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	err = redis.PAdd(context.Background(), 0x7D31, 0xE976, time.Unix(904867200, 0))
	if err != nil {
		t.Error(err)
	}
	err = redis.PAdd(context.Background(), 0x7D31, 0xEC0E, time.Unix(1075852800, 0))
	if err != nil {
		t.Error(err)
	}
	err = redis.PAdd(context.Background(), 0x7D31, 0x4CAF, time.Unix(681436800, 0))
	if err != nil {
		t.Error(err)
	}
	n, err = redis.PSize(context.Background(), 0x7D31)
	if err != nil {
		t.Error(err)
	}
	if n != 3 {
		t.Error("n != 3")
	}
	album, expires, err := redis.PPoll(context.Background(), 0x7D31)
	if err != nil {
		t.Error(err)
	}
	if album != 0x4CAF {
		t.Error("album != 0x4CAF")
	}
	if !expires.Equal(time.Unix(681436800, 0)) {
		t.Error("!expires.Equal(time.Unix(681436800, 0))")
	}
	n, err = redis.PSize(context.Background(), 0x7D31)
	if err != nil {
		t.Error(err)
	}
	if n != 2 {
		t.Error("n != 2")
	}
	album, expires, err = redis.PPoll(context.Background(), 0x7D31)
	if err != nil {
		t.Error(err)
	}
	if album != 0xE976 {
		t.Error("album != 0xE976")
	}
	if !expires.Equal(time.Unix(904867200, 0)) {
		t.Error("!expires.Equal(time.Unix(904867200, 0))")
	}
	album, expires, err = redis.PPoll(context.Background(), 0x7D31)
	if err != nil {
		t.Error(err)
	}
	if album != 0xEC0E {
		t.Error("album != 0xEC0E")
	}
	if !expires.Equal(time.Unix(1075852800, 0)) {
		t.Error("!expires.Equal(time.Unix(1075852800, 0))")
	}
	n, err = redis.PSize(context.Background(), 0x7D31)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	_, _, err = redis.PPoll(context.Background(), 0x7D31)
	if !errors.Is(err, domain.ErrUnknown) {
		t.Error(err)
	}
}

func TestRedisPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image1 := uint64(0x3E3D)
		image2 := uint64(0xB399)
		err = redis.Push(context.Background(), 0x23D2, [][2]uint64{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		image3, image4, err := redis.Pop(context.Background(), 0x23D2)
		if err != nil {
			t.Error(err)
		}
		if image1 != image3 {
			t.Error("image1 != image3")
		}
		if image2 != image4 {
			t.Error("image2 != image4")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = redis.Pop(context.Background(), 0x73BF)
		if !errors.Is(err, domain.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image1 := uint64(0x44DC)
		image2 := uint64(0x721B)
		err = redis.Push(context.Background(), 0x1AE9, [][2]uint64{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		_, _, err = redis.Pop(context.Background(), 0x1AE9)
		if err != nil {
			t.Error(err)
		}
		_, _, err = redis.Pop(context.Background(), 0x1AE9)
		if !errors.Is(err, domain.ErrPairNotFound) {
			t.Error(err)
		}
	})
}

func TestRedisToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image1 := uint64(0x52BD)
		token := uint64(0xB41C)
		err = redis.Set(context.Background(), 0xC2E7, token, image1)
		if err != nil {
			t.Error(err)
		}
		image2, err := redis.Get(context.Background(), 0xC2E7, token)
		if err != nil {
			t.Error(err)
		}
		if image1 != image2 {
			t.Error("image1 != image2")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image := uint64(0x583C)
		token := uint64(0xF0EE)
		err = redis.Set(context.Background(), 0x1C4A, token, image)
		if err != nil {
			t.Error(err)
		}
		err = redis.Set(context.Background(), 0x1C4A, token, image)
		if !errors.Is(err, domain.ErrTokenAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		_, err = redis.Get(context.Background(), 0x1C4A, 0xC4F8)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image := uint64(0x7C45)
		token := uint64(0xC67F)
		err = redis.Set(context.Background(), 0xEB96, token, image)
		if err != nil {
			t.Error(err)
		}
		_, err = redis.Get(context.Background(), 0xEB96, token)
		if err != nil {
			t.Error(err)
		}
		_, err = redis.Get(context.Background(), 0xEB96, token)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}
