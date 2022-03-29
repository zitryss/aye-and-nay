package service

import (
	"context"
	"flag"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/internal/dockertest"
	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *ci || !*integration {
		code := m.Run()
		os.Exit(code)
	}
	log.SetOutput(os.Stderr)
	log.SetLevel(log.CRITICAL)
	docker := dockertest.New()
	host := &cache.DefaultRedisConfig.Host
	port := &cache.DefaultRedisConfig.Port
	docker.RunRedis(host, port)
	host = &compressor.DefaultImaginaryConfig.Host
	port = &compressor.DefaultImaginaryConfig.Port
	docker.RunImaginary(host, port)
	host = &database.DefaultMongoConfig.Host
	port = &database.DefaultMongoConfig.Port
	docker.RunMongo(host, port)
	host = &storage.DefaultMinioConfig.Host
	port = &storage.DefaultMinioConfig.Port
	accessKey := storage.DefaultMinioConfig.AccessKey
	secretKey := storage.DefaultMinioConfig.SecretKey
	docker.RunMinio(host, port, accessKey, secretKey)
	log.SetOutput(io.Discard)
	code := m.Run()
	docker.Purge()
	os.Exit(code)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &ServiceTestSuite{})
}

type ServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	cancel        context.CancelFunc
	id            IdGenFunc
	ids           *IdLogBook
	heartbeatComp chan any
	heartbeatCalc chan any
	heartbeatDel  chan any
	serv          *Service
	gComp         *errgroup.Group
	gCalc         *errgroup.Group
	gDel          *errgroup.Group
	setupTestFn   func()
}

func (suite *ServiceTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	comp := compressor.NewMock()
	stor := storage.NewMock()
	data := database.NewMem(database.DefaultMemConfig)
	cach := cache.NewMem(cache.DefaultMemConfig)
	qCalc := NewQueueCalc(cach)
	qCalc.Monitor(ctx)
	qComp := NewQueueComp(cach)
	qComp.Monitor(ctx)
	qDel := NewQueueDel(cach)
	qDel.Monitor(ctx)
	fnShuffle := func(n int, swap func(i int, j int)) {}
	heartbeatComp := make(chan any)
	heartbeatCalc := make(chan any)
	heartbeatDel := make(chan any)
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
	suite.ctx = ctx
	suite.cancel = cancel
	suite.heartbeatComp = heartbeatComp
	suite.heartbeatCalc = heartbeatCalc
	suite.heartbeatDel = heartbeatDel
	suite.serv = serv
	suite.gComp = gComp
	suite.gCalc = gComp
	suite.gDel = gComp
	suite.setupTestFn = suite.SetupTest
}

func (suite *ServiceTestSuite) SetupTest() {
	id, ids := GenId()
	fnId := func() func() (uint64, error) {
		return func() (uint64, error) {
			return id(), nil
		}
	}()
	suite.id = id
	suite.ids = ids
	suite.serv.rand.id = fnId
	err := suite.serv.pers.(*database.Mem).Reset()
	require.NoError(suite.T(), err)
	err = suite.serv.cache.(*cache.Mem).Reset()
	require.NoError(suite.T(), err)
}

func (suite *ServiceTestSuite) TearDownTest() {

}

func (suite *ServiceTestSuite) TearDownSuite() {
	suite.cancel()
	err := suite.gDel.Wait()
	require.NoError(suite.T(), err)
	err = suite.gCalc.Wait()
	require.NoError(suite.T(), err)
	err = suite.gComp.Wait()
	require.NoError(suite.T(), err)
	err = suite.serv.pers.(*database.Mem).Reset()
	require.NoError(suite.T(), err)
	err = suite.serv.cache.(*cache.Mem).Reset()
	require.NoError(suite.T(), err)
}

func (suite *ServiceTestSuite) TestServiceAlbum() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		_, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		v := AssertChannel(t, suite.heartbeatComp)
		p, ok := v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 0.5, p, TOLERANCE)
		v = AssertChannel(t, suite.heartbeatComp)
		p, ok = v.(float64)
		assert.True(t, ok)
		assert.InDelta(t, 1, p, TOLERANCE)
	})
}

func (suite *ServiceTestSuite) TestServicePair() {
	suite.T().Run("Positive1", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img7, img8, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		img1 := model.Image{Id: suite.ids.Uint64(1), Token: suite.ids.Uint64(3), Src: "/api/images/" + suite.ids.Base64(3) + "/"}
		img2 := model.Image{Id: suite.ids.Uint64(2), Token: suite.ids.Uint64(4), Src: "/api/images/" + suite.ids.Base64(4) + "/"}
		imgs1 := []model.Image{img1, img2}
		assert.NotEqual(t, img7, img8)
		assert.Contains(t, imgs1, img7)
		assert.Contains(t, imgs1, img8)
		img9, img10, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		img3 := model.Image{Id: suite.ids.Uint64(2), Token: suite.ids.Uint64(5), Src: "/api/images/" + suite.ids.Base64(5) + "/"}
		img4 := model.Image{Id: suite.ids.Uint64(1), Token: suite.ids.Uint64(6), Src: "/api/images/" + suite.ids.Base64(6) + "/"}
		imgs2 := []model.Image{img3, img4}
		assert.NotEqual(t, img9, img10)
		assert.Contains(t, imgs2, img9)
		assert.Contains(t, imgs2, img10)
		img11, img12, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: suite.ids.Uint64(1), Token: suite.ids.Uint64(7), Src: "/api/images/" + suite.ids.Base64(7) + "/"}
		img6 := model.Image{Id: suite.ids.Uint64(2), Token: suite.ids.Uint64(8), Src: "/api/images/" + suite.ids.Base64(8) + "/"}
		imgs3 := []model.Image{img5, img6}
		assert.NotEqual(t, img11, img12)
		assert.Contains(t, imgs3, img11)
		assert.Contains(t, imgs3, img12)
	})
	suite.T().Run("Positive2", func(t *testing.T) {
		suite.setupTestFn()
		oldTempLinks := suite.serv.conf.TempLinks
		suite.serv.conf.TempLinks = false
		defer func() { suite.serv.conf.TempLinks = oldTempLinks }()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img7, img8, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		img1 := model.Image{Id: suite.ids.Uint64(1), Token: suite.ids.Uint64(1), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(1)}
		img2 := model.Image{Id: suite.ids.Uint64(2), Token: suite.ids.Uint64(2), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(2)}
		imgs1 := []model.Image{img1, img2}
		assert.NotEqual(t, img7, img8)
		assert.Contains(t, imgs1, img7)
		assert.Contains(t, imgs1, img8)
		img9, img10, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		img3 := model.Image{Id: suite.ids.Uint64(2), Token: suite.ids.Uint64(2), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(2)}
		img4 := model.Image{Id: suite.ids.Uint64(1), Token: suite.ids.Uint64(1), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(1)}
		imgs2 := []model.Image{img3, img4}
		assert.NotEqual(t, img9, img10)
		assert.Contains(t, imgs2, img9)
		assert.Contains(t, imgs2, img10)
		img11, img12, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: suite.ids.Uint64(1), Token: suite.ids.Uint64(1), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(1)}
		img6 := model.Image{Id: suite.ids.Uint64(2), Token: suite.ids.Uint64(2), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(2)}
		imgs3 := []model.Image{img5, img6}
		assert.NotEqual(t, img11, img12)
		assert.Contains(t, imgs3, img11)
		assert.Contains(t, imgs3, img12)
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		_, _, err := suite.serv.Pair(suite.ctx, suite.id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func (suite *ServiceTestSuite) TestServiceImage() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		f, err := suite.serv.Image(suite.ctx, img1.Token)
		assert.NoError(t, err)
		assert.NotNil(t, f.Reader)
		f, err = suite.serv.Image(suite.ctx, img2.Token)
		assert.NoError(t, err)
		assert.NotNil(t, f.Reader)
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		_, err := suite.serv.Image(suite.ctx, suite.id())
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func (suite *ServiceTestSuite) TestServiceVote() {
	suite.T().Run("Positive1", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		err = suite.serv.Vote(suite.ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
	})
	suite.T().Run("Positive2", func(t *testing.T) {
		suite.setupTestFn()
		oldTempLinks := suite.serv.conf.TempLinks
		suite.serv.conf.TempLinks = false
		defer func() { suite.serv.conf.TempLinks = oldTempLinks }()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		err = suite.serv.Vote(suite.ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
	})
	suite.T().Run("Negative1", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		err = suite.serv.Vote(suite.ctx, suite.id(), img1.Token, img2.Token)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative2", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		_, _, err = suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		err = suite.serv.Vote(suite.ctx, album, suite.id(), suite.id())
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func (suite *ServiceTestSuite) TestServiceTop() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		album, err := suite.serv.Album(suite.ctx, files, 0*time.Millisecond)
		assert.NoError(t, err)
		img1, img2, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		err = suite.serv.Vote(suite.ctx, album, img1.Token, img2.Token)
		assert.NoError(t, err)
		AssertChannel(t, suite.heartbeatCalc)
		img3, img4, err := suite.serv.Pair(suite.ctx, album)
		assert.NoError(t, err)
		err = suite.serv.Vote(suite.ctx, album, img3.Token, img4.Token)
		assert.NoError(t, err)
		AssertChannel(t, suite.heartbeatCalc)
		imgs1, err := suite.serv.Top(suite.ctx, album)
		assert.NoError(t, err)
		img5 := model.Image{Id: suite.ids.Uint64(1), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(1), Rating: 0.5, Compressed: false}
		img6 := model.Image{Id: suite.ids.Uint64(2), Src: "/aye-and-nay/albums/" + suite.ids.Base64(0) + "/images/" + suite.ids.Base64(2), Rating: 0.5, Compressed: false}
		imgs2 := []model.Image{img5, img6}
		assert.Equal(t, imgs2, imgs1)
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		_, err := suite.serv.Top(suite.ctx, suite.id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func (suite *ServiceTestSuite) TestServiceDelete() {
	suite.T().Run("Positive1", func(t *testing.T) {
		suite.setupTestFn()
		id1, ids1 := GenId()
		alb1 := AlbumFactory(id1, ids1)
		alb1.Expires = time.Now().Add(-1 * time.Hour)
		err := suite.serv.pers.SaveAlbum(suite.ctx, alb1)
		assert.NoError(t, err)
		id2, ids2 := GenId()
		alb2 := AlbumFactory(id2, ids2)
		alb2.Expires = time.Now().Add(1 * time.Hour)
		err = suite.serv.pers.SaveAlbum(suite.ctx, alb2)
		assert.NoError(t, err)
		err = suite.serv.CleanUp(suite.ctx)
		assert.NoError(t, err)
		v := AssertChannel(t, suite.heartbeatDel)
		album, ok := v.(uint64)
		assert.True(t, ok)
		assert.Equal(t, ids1.Uint64(0), album)
	})
	suite.T().Run("Positive2", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		dur := 100 * time.Millisecond
		album, err := suite.serv.Album(suite.ctx, files, dur)
		assert.NoError(t, err)
		AssertChannel(t, suite.heartbeatDel)
		_, err = suite.serv.Top(suite.ctx, album)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		files := []model.File{Png(), Png()}
		dur := 0 * time.Second
		album, err := suite.serv.Album(suite.ctx, files, dur)
		assert.NoError(t, err)
		AssertNotChannel(t, suite.heartbeatDel)
		_, err = suite.serv.Top(suite.ctx, album)
		assert.NoError(t, err)
	})
}

func (suite *ServiceTestSuite) TestServiceHealth() {
	_, err := suite.serv.Health(suite.ctx)
	assert.NoError(suite.T(), err)
}
