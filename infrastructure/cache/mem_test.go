//go:build unit

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/domain/domain"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMemPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		err := mem.Push(context.Background(), 0x23D2, [][2]uint64{{0x3E3D, 0xB399}})
		assert.NoError(t, err)
		image1, image2, err := mem.Pop(context.Background(), 0x23D2)
		assert.NoError(t, err)
		assert.Equal(t, uint64(0x3E3D), image1)
		assert.Equal(t, uint64(0xB399), image2)
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		_, _, err := mem.Pop(context.Background(), 0x73BF)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		err := mem.Push(context.Background(), 0x1AE9, [][2]uint64{{0x44DC, 0x721B}})
		assert.NoError(t, err)
		_, _, err = mem.Pop(context.Background(), 0x1AE9)
		assert.NoError(t, err)
		_, _, err = mem.Pop(context.Background(), 0x1AE9)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		heartbeatPair := make(chan interface{})
		mem := NewMem(DefaultMemConfig, WithHeartbeatPair(heartbeatPair))
		mem.Monitor()
		err := mem.Push(context.Background(), 0xF51A, [][2]uint64{{0x4BB0, 0x3A87}})
		assert.NoError(t, err)
		time.Sleep(mem.conf.TimeToLive)
		AssertChannel(t, heartbeatPair)
		AssertChannel(t, heartbeatPair)
		_, _, err = mem.Pop(context.Background(), 0xF51A)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
}

func TestMemToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xC2E7)
		album1 := uint64(0xB41C)
		image1 := uint64(0x52BD)
		err := mem.Set(context.Background(), token, album1, image1)
		assert.NoError(t, err)
		album2, image2, err := mem.Get(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, album1, album2)
		assert.Equal(t, image1, image2)
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0x1C4A)
		album := uint64(0xF0EE)
		image := uint64(0x583C)
		err := mem.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		err = mem.Set(context.Background(), token, album, image)
		assert.ErrorIs(t, err, domain.ErrTokenAlreadyExists)
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xC4F8)
		_, _, err := mem.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xEB96)
		album := uint64(0xC67F)
		image := uint64(0x7C45)
		err := mem.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		_, _, err = mem.Get(context.Background(), token)
		assert.NoError(t, err)
		err = mem.Del(context.Background(), token)
		assert.NoError(t, err)
		err = mem.Del(context.Background(), token)
		assert.NoError(t, err)
		_, _, err = mem.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	t.Run("Negative4", func(t *testing.T) {
		mem := NewMem(DefaultMemConfig)
		token := uint64(0xD3BF)
		err := mem.Del(context.Background(), token)
		assert.NoError(t, err)
	})
	t.Run("Negative5", func(t *testing.T) {
		heartbeatToken := make(chan interface{})
		mem := NewMem(DefaultMemConfig, WithHeartbeatToken(heartbeatToken))
		mem.Monitor()
		token := uint64(0xE0AF)
		album := uint64(0xCF1E)
		image := uint64(0xDD0A)
		err := mem.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		time.Sleep(mem.conf.TimeToLive)
		AssertChannel(t, heartbeatToken)
		AssertChannel(t, heartbeatToken)
		_, _, err = mem.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}
