// +build integration

package database

import (
	"context"
	"testing"

	"github.com/zitryss/aye-and-nay/domain/model"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestRedisPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		cache, err := NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		image1 := "RcBj3m9vuYPbntAE"
		image2 := "Q3NafBGuDH9PAtS4"
		err = cache.PushPair(context.Background(), "Pa6YTumLBRMFa7cX", [][2]string{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		image3, image4, err := cache.PopPair(context.Background(), "Pa6YTumLBRMFa7cX")
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
		cache, err := NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = cache.PopPair(context.Background(), "hP4tQHZr55JXMdnG")
		if !errors.Is(err, model.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		cache, err := NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		image1 := "5t2AMJ7NWAxBDDe4"
		image2 := "cPp7xeV4EMka5SpM"
		err = cache.PushPair(context.Background(), "5dVZ5tVm7QKtRjVA", [][2]string{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		_, _, err = cache.PopPair(context.Background(), "5dVZ5tVm7QKtRjVA")
		if err != nil {
			t.Error(err)
		}
		_, _, err = cache.PopPair(context.Background(), "5dVZ5tVm7QKtRjVA")
		if !errors.Is(err, model.ErrPairNotFound) {
			t.Error(err)
		}
	})
}

func TestRedisToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		cache, err := NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		image1 := "gTwdSTUDmz9LBerC"
		token := "kqsEDug6rK6BcHHy"
		err = cache.SetToken(context.Background(), "A55vmoMMLWX0g1KW", token, image1)
		if err != nil {
			t.Error(err)
		}
		image2, err := cache.GetImageId(context.Background(), "A55vmoMMLWX0g1KW", token)
		if err != nil {
			t.Error(err)
		}
		if image1 != image2 {
			t.Error("image1 != image2")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		cache, err := NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		image := "FvEfGeXG7xEuLREm"
		token := "a3MmBWHGMDC7LeN9"
		err = cache.SetToken(context.Background(), "b919qD42qhC4201o", token, image)
		if err != nil {
			t.Error(err)
		}
		err = cache.SetToken(context.Background(), "b919qD42qhC4201o", token, image)
		if !errors.Is(err, model.ErrTokenAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		cache, err := NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		token := "wmnAznYhVg6e8jHk"
		_, err = cache.GetImageId(context.Background(), "b919qD42qhC4201o", token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		cache, err := NewRedis(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		image := "QWfqTS8S4Hp2BzKn"
		token := "PK4dWeYgnY9vunmp"
		err = cache.SetToken(context.Background(), "0nq95EBOTH8I79LR", token, image)
		if err != nil {
			t.Error(err)
		}
		_, err = cache.GetImageId(context.Background(), "0nq95EBOTH8I79LR", token)
		if err != nil {
			t.Error(err)
		}
		_, err = cache.GetImageId(context.Background(), "0nq95EBOTH8I79LR", token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}
