//go:build unit

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/zitryss/aye-and-nay/domain/domain"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestMemPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
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
		mem := NewMem()
		_, _, err := mem.Pop(context.Background(), 0x73BF)
		if !errors.Is(err, domain.ErrPairNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem()
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
		mem := NewMem(WithHeartbeatPair(heartbeatPair))
		mem.Monitor()
		err := mem.Push(context.Background(), 0xF51A, [][2]uint64{{0x4BB0, 0x3A87}})
		if err != nil {
			t.Error(err)
		}
		time.Sleep(mem.conf.timeToLive)
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
		mem := NewMem()
		err := mem.Set(context.Background(), 0xC2E7, 0xB41C, 0x52BD)
		if err != nil {
			t.Error(err)
		}
		image, err := mem.Get(context.Background(), 0xC2E7, 0xB41C)
		if err != nil {
			t.Error(err)
		}
		if image != 0x52BD {
			t.Error("image != 0x52BD")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem()
		err := mem.Set(context.Background(), 0x1C4A, 0xF0EE, 0x583C)
		if err != nil {
			t.Error(err)
		}
		err = mem.Set(context.Background(), 0x1C4A, 0xF0EE, 0x583C)
		if !errors.Is(err, domain.ErrTokenAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem()
		_, err := mem.Get(context.Background(), 0x1C4A, 0xC4F8)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		mem := NewMem()
		err := mem.Set(context.Background(), 0xEB96, 0xC67F, 0x7C45)
		if err != nil {
			t.Error(err)
		}
		_, err = mem.Get(context.Background(), 0xEB96, 0xC67F)
		if err != nil {
			t.Error(err)
		}
		_, err = mem.Get(context.Background(), 0xEB96, 0xC67F)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative4", func(t *testing.T) {
		heartbeatToken := make(chan interface{})
		mem := NewMem(WithHeartbeatToken(heartbeatToken))
		mem.Monitor()
		err := mem.Set(context.Background(), 0xE0AF, 0xCF1E, 0xDD0A)
		if err != nil {
			t.Error(err)
		}
		time.Sleep(mem.conf.timeToLive)
		CheckChannel(t, heartbeatToken)
		CheckChannel(t, heartbeatToken)
		_, err = mem.Get(context.Background(), 0xE0AF, 0xCF1E)
		if !errors.Is(err, domain.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}
