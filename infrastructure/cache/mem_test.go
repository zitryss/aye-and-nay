package cache

import (
	"context"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestMemPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		image1 := "RcBj3m9vuYPbntAE"
		image2 := "Q3NafBGuDH9PAtS4"
		err := mem.Push(context.Background(), "Pa6YTumLBRMFa7cX", [][2]string{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		image3, image4, err := mem.Pop(context.Background(), "Pa6YTumLBRMFa7cX")
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
		mem := NewMem()
		_, _, err := mem.Pop(context.Background(), "hP4tQHZr55JXMdnG")
		if !errors.Is(err, model.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem()
		image1 := "5t2AMJ7NWAxBDDe4"
		image2 := "cPp7xeV4EMka5SpM"
		err := mem.Push(context.Background(), "5dVZ5tVm7QKtRjVA", [][2]string{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		_, _, err = mem.Pop(context.Background(), "5dVZ5tVm7QKtRjVA")
		if err != nil {
			t.Error(err)
		}
		_, _, err = mem.Pop(context.Background(), "5dVZ5tVm7QKtRjVA")
		if !errors.Is(err, model.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		heartbeatPair := make(chan interface{})
		mem := NewMem(WithHeartbeatPair(heartbeatPair))
		mem.Monitor()
		image1 := "RYvhkVCK3WAhULBa"
		image2 := "2EWKZXVVuh27sRkL"
		err := mem.Push(context.Background(), "rtxyfrCFm6LYcVwF", [][2]string{{image1, image2}})
		if err != nil {
			t.Error(err)
		}
		time.Sleep(mem.conf.timeToLive)
		CheckChannel(t, heartbeatPair)
		CheckChannel(t, heartbeatPair)
		_, _, err = mem.Pop(context.Background(), "rtxyfrCFm6LYcVwF")
		if !errors.Is(err, model.ErrPairNotFound) {
			t.Error(err)
		}
	})
}

func TestMemToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		image := "gTwdSTUDmz9LBerC"
		token := "kqsEDug6rK6BcHHy"
		err := mem.Set(context.Background(), "A55vmoMMLWX0g1KW", token, image)
		if err != nil {
			t.Error(err)
		}
		image2, err := mem.Get(context.Background(), "A55vmoMMLWX0g1KW", token)
		if err != nil {
			t.Error(err)
		}
		if image != image2 {
			t.Error("image != image2")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem()
		image := "FvEfGeXG7xEuLREm"
		token := "a3MmBWHGMDC7LeN9"
		err := mem.Set(context.Background(), "b919qD42qhC4201o", token, image)
		if err != nil {
			t.Error(err)
		}
		err = mem.Set(context.Background(), "b919qD42qhC4201o", token, image)
		if !errors.Is(err, model.ErrTokenAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem()
		token := "wmnAznYhVg6e8jHk"
		_, err := mem.Get(context.Background(), "b919qD42qhC4201o", token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		mem := NewMem()
		image := "QWfqTS8S4Hp2BzKn"
		token := "PK4dWeYgnY9vunmp"
		err := mem.Set(context.Background(), "0nq95EBOTH8I79LR", token, image)
		if err != nil {
			t.Error(err)
		}
		_, err = mem.Get(context.Background(), "0nq95EBOTH8I79LR", token)
		if err != nil {
			t.Error(err)
		}
		_, err = mem.Get(context.Background(), "0nq95EBOTH8I79LR", token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative4", func(t *testing.T) {
		heartbeatToken := make(chan interface{})
		mem := NewMem(WithHeartbeatToken(heartbeatToken))
		mem.Monitor()
		image := "mHz9nH5nwCfZHn5C"
		token := "ebqwp4yEHuH2eB2U"
		err := mem.Set(context.Background(), "xZALPEN7kt7RS5rz", token, image)
		if err != nil {
			t.Error(err)
		}
		time.Sleep(mem.conf.timeToLive)
		CheckChannel(t, heartbeatToken)
		CheckChannel(t, heartbeatToken)
		_, err = mem.Get(context.Background(), "xZALPEN7kt7RS5rz", token)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}
