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
)

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, &RedisTestSuite{})
}

type RedisTestSuite struct {
	suite.Suite
	base        MemTestSuite
	setupTestFn func()
}

func (suite *RedisTestSuite) SetupSuite() {
	if !*integration {
		suite.T().Skip()
	}
	suite.base = MemTestSuite{}
	suite.base.SetT(suite.T())
	ctx, cancel := context.WithCancel(context.Background())
	conf := DefaultRedisConfig
	redis, err := NewRedis(ctx, conf)
	require.NoError(suite.T(), err)
	suite.base.ctx = ctx
	suite.base.cancel = cancel
	suite.base.conf.LimiterRequestsPerSecond = float64(conf.LimiterRequestsPerSecond)
	suite.base.conf.TimeToLive = conf.TimeToLive
	suite.base.cache = redis
	suite.base.setupTestFn = suite.SetupTest
	suite.setupTestFn = suite.SetupTest
}

func (suite *RedisTestSuite) SetupTest() {
	err := suite.base.cache.(*Redis).Reset()
	require.NoError(suite.T(), err)
}

func (suite *RedisTestSuite) TearDownTest() {

}

func (suite *RedisTestSuite) TearDownSuite() {
	err := suite.base.cache.(*Redis).Reset()
	require.NoError(suite.T(), err)
	err = suite.base.cache.(*Redis).Close(suite.base.ctx)
	require.NoError(suite.T(), err)
	suite.base.cancel()
}

func (suite *RedisTestSuite) TestRedisAllow() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		rpm := suite.base.conf.LimiterRequestsPerSecond
		ip := id()
		for j := float64(0); j < rpm; j++ {
			allowed, err := suite.base.cache.Allow(suite.base.ctx, ip)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
		time.Sleep(1 * time.Second)
		for j := float64(0); j < rpm; j++ {
			allowed, err := suite.base.cache.Allow(suite.base.ctx, ip)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		rps := suite.base.conf.LimiterRequestsPerSecond
		ip := id()
		for i := float64(0); i < rps; i++ {
			allowed, err := suite.base.cache.Allow(suite.base.ctx, ip)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
		allowed, err := suite.base.cache.Allow(suite.base.ctx, ip)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}

func (suite *RedisTestSuite) TestRedisQueue() {
	id, _ := GenId()
	queue := id()
	albumExp1 := id()
	albumExp2 := id()
	albumExp3 := id()
	n, err := suite.base.cache.Size(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, n)
	err = suite.base.cache.Add(suite.base.ctx, queue, albumExp1)
	assert.NoError(suite.T(), err)
	err = suite.base.cache.Add(suite.base.ctx, queue, albumExp1)
	assert.NoError(suite.T(), err)
	err = suite.base.cache.Add(suite.base.ctx, queue, albumExp2)
	assert.NoError(suite.T(), err)
	err = suite.base.cache.Add(suite.base.ctx, queue, albumExp3)
	assert.NoError(suite.T(), err)
	err = suite.base.cache.Add(suite.base.ctx, queue, albumExp2)
	assert.NoError(suite.T(), err)
	n, err = suite.base.cache.Size(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, n)
	album, err := suite.base.cache.Poll(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), albumExp1, album)
	n, err = suite.base.cache.Size(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, n)
	album, err = suite.base.cache.Poll(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), albumExp2, album)
	album, err = suite.base.cache.Poll(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), albumExp3, album)
	n, err = suite.base.cache.Size(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, n)
	album, err = suite.base.cache.Poll(suite.base.ctx, queue)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), uint64(0x0), album)
	n, err = suite.base.cache.Size(suite.base.ctx, queue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, n)
	_, err = suite.base.cache.Poll(suite.base.ctx, queue)
	assert.ErrorIs(suite.T(), err, domain.ErrUnknown)
}

func (suite *RedisTestSuite) TestRedisPQueue() {
	id, _ := GenId()
	pqueue := id()
	albumExp1 := id()
	albumExp2 := id()
	albumExp3 := id()
	n, err := suite.base.cache.PSize(suite.base.ctx, pqueue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, n)
	err = suite.base.cache.PAdd(suite.base.ctx, pqueue, albumExp1, time.Unix(904867200, 0))
	assert.NoError(suite.T(), err)
	err = suite.base.cache.PAdd(suite.base.ctx, pqueue, albumExp2, time.Unix(1075852800, 0))
	assert.NoError(suite.T(), err)
	err = suite.base.cache.PAdd(suite.base.ctx, pqueue, albumExp3, time.Unix(681436800, 0))
	assert.NoError(suite.T(), err)
	n, err = suite.base.cache.PSize(suite.base.ctx, pqueue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, n)
	album, expires, err := suite.base.cache.PPoll(suite.base.ctx, pqueue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), albumExp3, album)
	assert.True(suite.T(), expires.Equal(time.Unix(681436800, 0)))
	n, err = suite.base.cache.PSize(suite.base.ctx, pqueue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, n)
	album, expires, err = suite.base.cache.PPoll(suite.base.ctx, pqueue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), albumExp1, album)
	assert.True(suite.T(), expires.Equal(time.Unix(904867200, 0)))
	album, expires, err = suite.base.cache.PPoll(suite.base.ctx, pqueue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), albumExp2, album)
	assert.True(suite.T(), expires.Equal(time.Unix(1075852800, 0)))
	n, err = suite.base.cache.PSize(suite.base.ctx, pqueue)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, n)
	_, _, err = suite.base.cache.PPoll(suite.base.ctx, pqueue)
	assert.ErrorIs(suite.T(), err, domain.ErrUnknown)
}

func (suite *RedisTestSuite) TestRedisPair() {
	suite.base.TestPair()
}

func (suite *RedisTestSuite) TestRedisToken() {
	suite.base.TestToken()
}
