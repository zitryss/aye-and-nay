package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	. "github.com/zitryss/aye-and-nay/internal/generator"
)

func TestServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, &ServiceIntegrationTestSuite{})
}

type ServiceIntegrationTestSuite struct {
	suite.Suite
	base ServiceTestSuite
}

func (suite *ServiceIntegrationTestSuite) SetupSuite() {
	if !*integration {
		suite.T().Skip()
	}
	suite.base = ServiceTestSuite{}
	suite.base.SetT(suite.T())
	ctx, cancel := context.WithCancel(context.Background())
	comp, err := compressor.New(ctx, compressor.CompressorConfig{Compressor: "mock"})
	require.NoError(suite.T(), err)
	stor, err := storage.New(ctx, storage.StorageConfig{Storage: "minio", Minio: storage.DefaultMinioConfig})
	require.NoError(suite.T(), err)
	data, err := database.New(ctx, database.DatabaseConfig{Database: "mongo", Mongo: database.DefaultMongoConfig})
	require.NoError(suite.T(), err)
	cach, err := cache.New(ctx, cache.CacheConfig{Cache: "redis", Redis: cache.DefaultRedisConfig})
	require.NoError(suite.T(), err)
	qCalc := NewQueueCalc(cach)
	qCalc.Monitor(ctx)
	qComp := NewQueueComp(cach)
	qComp.Monitor(ctx)
	qDel := NewQueueDel(cach)
	qDel.Monitor(ctx)
	fnShuffle := func(n int, swap func(i int, j int)) {}
	heartbeatComp := make(chan interface{})
	heartbeatCalc := make(chan interface{})
	heartbeatDel := make(chan interface{})
	serv := New(DefaultServiceConfig, comp, stor, data, cach, qCalc, qComp, qDel,
		WithRandShuffle(fnShuffle),
		WithHeartbeatComp(heartbeatComp),
		WithHeartbeatCalc(heartbeatCalc),
		WithHeartbeatDel(heartbeatDel),
	)
	gComp, ctxComp := errgroup.WithContext(ctx)
	serv.StartWorkingPoolComp(ctxComp, gComp)
	gCalc, ctxCalc := errgroup.WithContext(ctx)
	serv.StartWorkingPoolCalc(ctxCalc, gCalc)
	gDel, ctxDel := errgroup.WithContext(ctx)
	serv.StartWorkingPoolDel(ctxDel, gDel)
	suite.base.ctx = ctx
	suite.base.cancel = cancel
	suite.base.heartbeatComp = heartbeatComp
	suite.base.heartbeatCalc = heartbeatCalc
	suite.base.heartbeatDel = heartbeatDel
	suite.base.serv = serv
	suite.base.gComp = gComp
	suite.base.gCalc = gComp
	suite.base.gDel = gComp
	suite.base.setupTestFn = suite.SetupTest
}

func (suite *ServiceIntegrationTestSuite) SetupTest() {
	id, ids := GenId()
	fnId := func() func() (uint64, error) {
		return func() (uint64, error) {
			return id(), nil
		}
	}()
	suite.base.id = id
	suite.base.ids = ids
	suite.base.serv.rand.id = fnId
	err := suite.base.serv.stor.(*storage.Minio).Reset()
	require.NoError(suite.T(), err)
	err = suite.base.serv.pers.(*database.Mongo).Reset()
	require.NoError(suite.T(), err)
	err = suite.base.serv.cache.(*cache.Redis).Reset()
	require.NoError(suite.T(), err)
}

func (suite *ServiceIntegrationTestSuite) TearDownTest() {

}

func (suite *ServiceIntegrationTestSuite) TearDownSuite() {
	suite.base.cancel()
	err := suite.base.gDel.Wait()
	require.NoError(suite.T(), err)
	err = suite.base.gCalc.Wait()
	require.NoError(suite.T(), err)
	err = suite.base.gComp.Wait()
	require.NoError(suite.T(), err)
	err = suite.base.serv.stor.(*storage.Minio).Reset()
	require.NoError(suite.T(), err)
	err = suite.base.serv.pers.(*database.Mongo).Reset()
	require.NoError(suite.T(), err)
	err = suite.base.serv.cache.(*cache.Redis).Reset()
	require.NoError(suite.T(), err)
	err = suite.base.serv.pers.(*database.Mongo).Close(suite.base.ctx)
	require.NoError(suite.T(), err)
	err = suite.base.serv.cache.(*cache.Redis).Close(suite.base.ctx)
	require.NoError(suite.T(), err)
}

func (suite *ServiceIntegrationTestSuite) TestServiceIntegrationAlbum() {
	suite.base.TestServiceAlbum()
}
func (suite *ServiceIntegrationTestSuite) TestServiceIntegrationPair() {
	suite.base.TestServicePair()
}
func (suite *ServiceIntegrationTestSuite) TestServiceIntegrationImage() {
	suite.base.TestServiceImage()
}
func (suite *ServiceIntegrationTestSuite) TestServiceIntegrationVote() {
	suite.base.TestServiceVote()
}
func (suite *ServiceIntegrationTestSuite) TestServiceIntegrationTop() {
	suite.base.TestServiceTop()
}
func (suite *ServiceIntegrationTestSuite) TestServiceIntegrationDelete() {
	suite.base.TestServiceDelete()
}
func (suite *ServiceIntegrationTestSuite) TestServiceIntegrationHealth() {
	suite.base.TestServiceHealth()
}
