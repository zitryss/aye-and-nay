//go:build unit

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/domain/domain"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestMemPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		err := mem.Push(context.Background(), 0x23D2, [][2]uint64{{0x3E3D, 0xB399}})
		if err != nil {
			t.Error(err)
		}
		image1, image2, err := mem.Pop(context.Background(), 0x23D2)
		if err != nil {
			t.Error(err)
		}
		if image1 != 0x3E3D {
			t.Error("image1 != 0x3E3D")
		}
		if image2 != 0xB399 {
			t.Error("image2 != 0xB399")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		_, _, err := mem.Pop(context.Background(), 0x73BF)
		if !errors.Is(err, domain.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		err := mem.Push(context.Background(), 0x1AE9, [][2]uint64{{0x44DC, 0x721B}})
		if err != nil {
			t.Error(err)
		}
		_, _, err = mem.Pop(context.Background(), 0x1AE9)
		if err != nil {
			t.Error(err)
		}
		_, _, err = mem.Pop(context.Background(), 0x1AE9)
		if !errors.Is(err, domain.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		heartbeatPair := make(chan interface{})
		mem := NewMem(DefaultMemConfig, WithHeartbeatPair(heartbeatPair))
		mem.Monitor()
		err := mem.Push(context.Background(), 0xF51A, [][2]uint64{{0x4BB0, 0x3A87}})
		if err != nil {
			t.Error(err)
		}
		time.Sleep(mem.conf.TimeToLive)
		CheckChannel(t, heartbeatPair)
		CheckChannel(t, heartbeatPair)
		_, _, err = mem.Pop(context.Background(), 0xF51A)
		if !errors.Is(err, domain.ErrPairNotFound) {
			t.Error(err)
		}
	})
}

func TestMemToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xC2E7)
		album1 := uint64(0xB41C)
		image1 := uint64(0x52BD)
		err := mem.Set(context.Background(), token, album1, image1)
		if err != nil {
			t.Error(err)
		}
		album2, image2, err := mem.Get(context.Background(), token)
		if err != nil {
			t.Error(err)
		}
		if album1 != album2 {
			t.Error("album1 != album2")
		}
		if image1 != image2 {
			t.Error("image1 != image2")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0x1C4A)
		album := uint64(0xF0EE)
		image := uint64(0x583C)
		err := mem.Set(context.Background(), token, album, image)
		if err != nil {
			t.Error(err)
		}
		err = mem.Set(context.Background(), token, album, image)
		if !errors.Is(err, domain.ErrTokenAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xC4F8)
		_, _, err := mem.Get(context.Background(), token)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xEB96)
		album := uint64(0xC67F)
		image := uint64(0x7C45)
		err := mem.Set(context.Background(), token, album, image)
		if err != nil {
			t.Error(err)
		}
		_, _, err = mem.Get(context.Background(), token)
		if err != nil {
			t.Error(err)
		}
		err = mem.Del(context.Background(), token)
		if err != nil {
			t.Error(err)
		}
		err = mem.Del(context.Background(), token)
		if err != nil {
			t.Error(err)
		}
		_, _, err = mem.Get(context.Background(), token)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative4", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xD3BF)
		err := mem.Del(context.Background(), token)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Negative5", func(t *testing.T) {
		heartbeatToken := make(chan interface{})
		mem := NewMem(DefaultMemConfig, WithHeartbeatToken(heartbeatToken))
		mem.Monitor()
		token := uint64(0xE0AF)
		album := uint64(0xCF1E)
		image := uint64(0xDD0A)
		err := mem.Set(context.Background(), token, album, image)
		if err != nil {
			t.Error(err)
		}
		time.Sleep(mem.conf.TimeToLive)
		CheckChannel(t, heartbeatToken)
		CheckChannel(t, heartbeatToken)
		_, _, err = mem.Get(context.Background(), token)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}
