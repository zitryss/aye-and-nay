// +build integration

package cache

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/internal/dockertest"
	"github.com/zitryss/aye-and-nay/pkg/env"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func TestMain(m *testing.M) {
	_, err := env.Lookup("CONTINUOUS_INTEGRATION")
	if err != nil {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.Lcritical)
		docker := dockertest.New()
		docker.RunRedis()
		log.SetOutput(ioutil.Discard)
		code := m.Run()
		docker.Purge()
		os.Exit(code)
	}
	code := m.Run()
	os.Exit(code)
}

func TestRedisQueue(t *testing.T) {
	redis, err := NewRedis()
	if err != nil {
		t.Fatal(err)
	}
	n, err := redis.Size(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	err = redis.Add(context.Background(), "8wwEdmRqQnQ6Yhjy", "MMJ9P9r7qbbMrjmx")
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), "8wwEdmRqQnQ6Yhjy", "MMJ9P9r7qbbMrjmx")
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), "8wwEdmRqQnQ6Yhjy", "YrEQ85fcDzzTd5fS")
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), "8wwEdmRqQnQ6Yhjy", "58ZNTHsAErKuU7Sk")
	if err != nil {
		t.Error(err)
	}
	err = redis.Add(context.Background(), "8wwEdmRqQnQ6Yhjy", "YrEQ85fcDzzTd5fS")
	if err != nil {
		t.Error(err)
	}
	n, err = redis.Size(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if n != 3 {
		t.Error("n != 3")
	}
	album, err := redis.Poll(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if album != "MMJ9P9r7qbbMrjmx" {
		t.Error("album != \"MMJ9P9r7qbbMrjmx\"")
	}
	n, err = redis.Size(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if n != 2 {
		t.Error("n != 2")
	}
	album, err = redis.Poll(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if album != "YrEQ85fcDzzTd5fS" {
		t.Error("album != \"YrEQ85fcDzzTd5fS\"")
	}
	album, err = redis.Poll(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if album != "58ZNTHsAErKuU7Sk" {
		t.Error("album != \"58ZNTHsAErKuU7Sk\"")
	}
	n, err = redis.Size(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	album, err = redis.Poll(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err == nil {
		t.Error(err)
	}
	if album != "" {
		t.Error("album != \"\"")
	}
	n, err = redis.Size(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	_, err = redis.Poll(context.Background(), "8wwEdmRqQnQ6Yhjy")
	if !errors.Is(err, model.ErrUnknown) {
		t.Error(err)
	}
}

func TestRedisPQueue(t *testing.T) {
	redis, err := NewRedis()
	if err != nil {
		t.Fatal(err)
	}
	n, err := redis.PSize(context.Background(), "d9YtN3xaf3z569Pa")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	err = redis.PAdd(context.Background(), "d9YtN3xaf3z569Pa", "3SNvbjeg5uuEK9yz", time.Unix(904867200, 0))
	if err != nil {
		t.Error(err)
	}
	err = redis.PAdd(context.Background(), "d9YtN3xaf3z569Pa", "uneKPF2Fy43yj8yz", time.Unix(1075852800, 0))
	if err != nil {
		t.Error(err)
	}
	err = redis.PAdd(context.Background(), "d9YtN3xaf3z569Pa", "EHJajVMUAu3ewR5B", time.Unix(681436800, 0))
	if err != nil {
		t.Error(err)
	}
	n, err = redis.PSize(context.Background(), "d9YtN3xaf3z569Pa")
	if err != nil {
		t.Error(err)
	}
	if n != 3 {
		t.Error("n != 3")
	}
	album, expires, err := redis.PPoll(context.Background(), "d9YtN3xaf3z569Pa")
	if err != nil {
		t.Error(err)
	}
	if album != "EHJajVMUAu3ewR5B" {
		t.Error("album != \"EHJajVMUAu3ewR5B\"")
	}
	if !expires.Equal(time.Unix(681436800, 0)) {
		t.Error("!expires.Equal(time.Unix(681436800, 0))")
	}
	n, err = redis.PSize(context.Background(), "d9YtN3xaf3z569Pa")
	if err != nil {
		t.Error(err)
	}
	if n != 2 {
		t.Error("n != 2")
	}
	album, expires, err = redis.PPoll(context.Background(), "d9YtN3xaf3z569Pa")
	if err != nil {
		t.Error(err)
	}
	if album != "3SNvbjeg5uuEK9yz" {
		t.Error("album != \"3SNvbjeg5uuEK9yz\"")
	}
	if !expires.Equal(time.Unix(904867200, 0)) {
		t.Error("!expires.Equal(time.Unix(904867200, 0))")
	}
	album, expires, err = redis.PPoll(context.Background(), "d9YtN3xaf3z569Pa")
	if err != nil {
		t.Error(err)
	}
	if album != "uneKPF2Fy43yj8yz" {
		t.Error("album != \"uneKPF2Fy43yj8yz\"")
	}
	if !expires.Equal(time.Unix(1075852800, 0)) {
		t.Error("!expires.Equal(time.Unix(1075852800, 0))")
	}
	n, err = redis.PSize(context.Background(), "d9YtN3xaf3z569Pa")
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Error("n != 0")
	}
	_, _, err = redis.PPoll(context.Background(), "d9YtN3xaf3z569Pa")
	if !errors.Is(err, model.ErrUnknown) {
		t.Error(err)
	}
}

func TestRedisPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image1 := "RcBj3m9vuYPbntAE"
		image2 := "Q3NafBGuDH9PAtS4"
		err = redis.Push(context.Background(), "Pa6YTumLBRMFa7cX", [][2]string{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		image3, image4, err := redis.Pop(context.Background(), "Pa6YTumLBRMFa7cX")
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
		_, _, err = redis.Pop(context.Background(), "hP4tQHZr55JXMdnG")
		if !errors.Is(err, model.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image1 := "5t2AMJ7NWAxBDDe4"
		image2 := "cPp7xeV4EMka5SpM"
		err = redis.Push(context.Background(), "5dVZ5tVm7QKtRjVA", [][2]string{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		_, _, err = redis.Pop(context.Background(), "5dVZ5tVm7QKtRjVA")
		if err != nil {
			t.Error(err)
		}
		_, _, err = redis.Pop(context.Background(), "5dVZ5tVm7QKtRjVA")
		if !errors.Is(err, model.ErrPairNotFound) {
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
		image1 := "gTwdSTUDmz9LBerC"
		token := "kqsEDug6rK6BcHHy"
		err = redis.Set(context.Background(), "A55vmoMMLWX0g1KW", token, image1)
		if err != nil {
			t.Error(err)
		}
		image2, err := redis.Get(context.Background(), "A55vmoMMLWX0g1KW", token)
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
		image := "FvEfGeXG7xEuLREm"
		token := "a3MmBWHGMDC7LeN9"
		err = redis.Set(context.Background(), "b919qD42qhC4201o", token, image)
		if err != nil {
			t.Error(err)
		}
		err = redis.Set(context.Background(), "b919qD42qhC4201o", token, image)
		if !errors.Is(err, model.ErrTokenAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		token := "wmnAznYhVg6e8jHk"
		_, err = redis.Get(context.Background(), "b919qD42qhC4201o", token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		redis, err := NewRedis()
		if err != nil {
			t.Fatal(err)
		}
		image := "QWfqTS8S4Hp2BzKn"
		token := "PK4dWeYgnY9vunmp"
		err = redis.Set(context.Background(), "0nq95EBOTH8I79LR", token, image)
		if err != nil {
			t.Error(err)
		}
		_, err = redis.Get(context.Background(), "0nq95EBOTH8I79LR", token)
		if err != nil {
			t.Error(err)
		}
		_, err = redis.Get(context.Background(), "0nq95EBOTH8I79LR", token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}
