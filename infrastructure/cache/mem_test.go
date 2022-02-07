//go:build unit

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/domain/domain"
	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMemPair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, ids := GenId()
		mem := NewMem(DefaultMemConfig)
		album := id()
		pairs := [][2]uint64{{id(), id()}}
		err := mem.Push(context.Background(), album, pairs)
		assert.NoError(t, err)
		image1, image2, err := mem.Pop(context.Background(), album)
		assert.NoError(t, err)
		assert.Equal(t, ids.Uint64(1), image1)
		assert.Equal(t, ids.Uint64(2), image2)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, _ := GenId()
		mem := NewMem(DefaultMemConfig)
		album := id()
		_, _, err := mem.Pop(context.Background(), album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, _ := GenId()
		mem := NewMem(DefaultMemConfig)
		album := id()
		pairs := [][2]uint64{{id(), id()}}
		err := mem.Push(context.Background(), album, pairs)
		assert.NoError(t, err)
		_, _, err = mem.Pop(context.Background(), album)
		assert.NoError(t, err)
		_, _, err = mem.Pop(context.Background(), album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		id, _ := GenId()
		heartbeatPair := make(chan interface{})
		mem := NewMem(DefaultMemConfig, WithHeartbeatPair(heartbeatPair))
		mem.Monitor()
		album := id()
		pairs := [][2]uint64{{id(), id()}}
		err := mem.Push(context.Background(), album, pairs)
		assert.NoError(t, err)
		time.Sleep(mem.conf.TimeToLive)
		AssertChannel(t, heartbeatPair)
		AssertChannel(t, heartbeatPair)
		_, _, err = mem.Pop(context.Background(), album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
}

func TestMemToken(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		id, _ := GenId()
		mem := NewMem(DefaultMemConfig)
		token := id()
		album1 := id()
		image1 := id()
		err := mem.Set(context.Background(), token, album1, image1)
		assert.NoError(t, err)
		album2, image2, err := mem.Get(context.Background(), token)
		assert.NoError(t, err)
		assert.Equal(t, album1, album2)
		assert.Equal(t, image1, image2)
	})
	t.Run("Negative1", func(t *testing.T) {
		id, _ := GenId()
		mem := NewMem(DefaultMemConfig)
		token := id()
		album := id()
		image := id()
		err := mem.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		err = mem.Set(context.Background(), token, album, image)
		assert.ErrorIs(t, err, domain.ErrTokenAlreadyExists)
	})
	t.Run("Negative2", func(t *testing.T) {
		id, _ := GenId()
		mem := NewMem(DefaultMemConfig)
		token := id()
		_, _, err := mem.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	t.Run("Negative3", func(t *testing.T) {
		id, _ := GenId()
		mem := NewMem(DefaultMemConfig)
		token := id()
		album := id()
		image := id()
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
		id, _ := GenId()
		mem := NewMem(DefaultMemConfig)
		token := id()
		err := mem.Del(context.Background(), token)
		assert.NoError(t, err)
	})
	t.Run("Negative5", func(t *testing.T) {
		id, _ := GenId()
		heartbeatToken := make(chan interface{})
		mem := NewMem(DefaultMemConfig, WithHeartbeatToken(heartbeatToken))
		mem.Monitor()
		token := id()
		album := id()
		image := id()
		err := mem.Set(context.Background(), token, album, image)
		assert.NoError(t, err)
		time.Sleep(mem.conf.TimeToLive)
		AssertChannel(t, heartbeatToken)
		AssertChannel(t, heartbeatToken)
		_, _, err = mem.Get(context.Background(), token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}
