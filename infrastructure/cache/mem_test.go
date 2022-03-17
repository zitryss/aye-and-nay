package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zitryss/aye-and-nay/domain/domain"
	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMemTestSuite(t *testing.T) {
	suite.Run(t, &MemTestSuite{})
}

type MemTestSuite struct {
	suite.Suite
	ctx              context.Context
	cancel           context.CancelFunc
	conf             MemConfig
	heartbeatCleanup chan interface{}
	heartbeatPair    chan interface{}
	heartbeatToken   chan interface{}
	cache            domain.Cacher
	setupTestFn      func()
}

func (suite *MemTestSuite) SetupSuite() {
	if !*unit {
		suite.T().Skip()
	}
	ctx, cancel := context.WithCancel(context.Background())
	conf := DefaultMemConfig
	hc := make(chan interface{})
	hp := make(chan interface{})
	ht := make(chan interface{})
	mem := NewMem(conf, WithHeartbeatCleanup(hc), WithHeartbeatPair(hp), WithHeartbeatToken(ht))
	mem.Monitor(ctx)
	suite.ctx = ctx
	suite.cancel = cancel
	suite.conf = conf
	suite.heartbeatCleanup = hc
	suite.heartbeatPair = hp
	suite.heartbeatToken = ht
	suite.cache = mem
	suite.setupTestFn = suite.SetupTest
}

func (suite *MemTestSuite) SetupTest() {
	err := suite.cache.(*Mem).Reset()
	require.NoError(suite.T(), err)
}

func (suite *MemTestSuite) TearDownTest() {

}

func (suite *MemTestSuite) TearDownSuite() {
	err := suite.cache.(*Mem).Reset()
	require.NoError(suite.T(), err)
	suite.cancel()
}

func (suite *MemTestSuite) TestPair() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		album := id()
		pairs := [][2]uint64{{id(), id()}}
		err := suite.cache.Push(suite.ctx, album, pairs)
		assert.NoError(t, err)
		image1, image2, err := suite.cache.Pop(suite.ctx, album)
		assert.NoError(t, err)
		assert.Equal(t, ids.Uint64(1), image1)
		assert.Equal(t, ids.Uint64(2), image2)
	})
	suite.T().Run("Negative1", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		album := id()
		_, _, err := suite.cache.Pop(suite.ctx, album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	suite.T().Run("Negative2", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		album := id()
		pairs := [][2]uint64{{id(), id()}}
		err := suite.cache.Push(suite.ctx, album, pairs)
		assert.NoError(t, err)
		_, _, err = suite.cache.Pop(suite.ctx, album)
		assert.NoError(t, err)
		_, _, err = suite.cache.Pop(suite.ctx, album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
	suite.T().Run("Negative3", func(t *testing.T) {
		suite.setupTestFn()
		_, ok := suite.cache.(*Redis)
		if testing.Short() && ok {
			t.Skip("short flag is set")
		}
		id, _ := GenId()
		album := id()
		pairs := [][2]uint64{{id(), id()}}
		err := suite.cache.Push(suite.ctx, album, pairs)
		assert.NoError(t, err)
		time.Sleep(suite.conf.TimeToLive * 2)
		AssertChannel(t, suite.heartbeatPair)
		AssertChannel(t, suite.heartbeatPair)
		_, _, err = suite.cache.Pop(suite.ctx, album)
		assert.ErrorIs(t, err, domain.ErrPairNotFound)
	})
}

func (suite *MemTestSuite) TestToken() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		token := id()
		album1 := id()
		image1 := id()
		err := suite.cache.Set(suite.ctx, token, album1, image1)
		assert.NoError(t, err)
		album2, image2, err := suite.cache.Get(suite.ctx, token)
		assert.NoError(t, err)
		assert.Equal(t, album1, album2)
		assert.Equal(t, image1, image2)
	})
	suite.T().Run("Negative1", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		token := id()
		album := id()
		image := id()
		err := suite.cache.Set(suite.ctx, token, album, image)
		assert.NoError(t, err)
		err = suite.cache.Set(suite.ctx, token, album, image)
		assert.ErrorIs(t, err, domain.ErrTokenAlreadyExists)
	})
	suite.T().Run("Negative2", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		token := id()
		_, _, err := suite.cache.Get(suite.ctx, token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	suite.T().Run("Negative3", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		token := id()
		album := id()
		image := id()
		err := suite.cache.Set(suite.ctx, token, album, image)
		assert.NoError(t, err)
		_, _, err = suite.cache.Get(suite.ctx, token)
		assert.NoError(t, err)
		err = suite.cache.Del(suite.ctx, token)
		assert.NoError(t, err)
		err = suite.cache.Del(suite.ctx, token)
		assert.NoError(t, err)
		_, _, err = suite.cache.Get(suite.ctx, token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
	suite.T().Run("Negative4", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		token := id()
		err := suite.cache.Del(suite.ctx, token)
		assert.NoError(t, err)
	})
	suite.T().Run("Negative5", func(t *testing.T) {
		suite.setupTestFn()
		_, ok := suite.cache.(*Redis)
		if testing.Short() && ok {
			t.Skip("short flag is set")
		}
		id, _ := GenId()
		token := id()
		album := id()
		image := id()
		err := suite.cache.Set(suite.ctx, token, album, image)
		assert.NoError(t, err)
		time.Sleep(suite.conf.TimeToLive * 2)
		AssertChannel(t, suite.heartbeatToken)
		AssertChannel(t, suite.heartbeatToken)
		_, _, err = suite.cache.Get(suite.ctx, token)
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}
