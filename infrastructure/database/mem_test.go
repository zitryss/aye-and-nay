package database

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMemTestSuite(t *testing.T) {
	suite.Run(t, &MemTestSuite{})
}

type MemTestSuite struct {
	suite.Suite
	ctx         context.Context
	cancel      context.CancelFunc
	db          domain.Databaser
	setupTestFn func()
}

func (suite *MemTestSuite) SetupSuite() {
	if !*unit {
		suite.T().Skip()
	}
	ctx, cancel := context.WithCancel(context.Background())
	mem := NewMem(DefaultMemConfig)
	suite.ctx = ctx
	suite.cancel = cancel
	suite.db = mem
	suite.setupTestFn = suite.SetupTest
}

func (suite *MemTestSuite) SetupTest() {
	err := suite.db.(*Mem).Reset()
	require.NoError(suite.T(), err)
}

func (suite *MemTestSuite) TearDownTest() {

}

func (suite *MemTestSuite) TearDownSuite() {
	err := suite.db.(*Mem).Reset()
	require.NoError(suite.T(), err)
	suite.cancel()
}

func (suite *MemTestSuite) saveAlbum(id IdGenFunc, ids Ids) model.Album {
	suite.T().Helper()
	alb := AlbumFactory(id, ids)
	err := suite.db.SaveAlbum(suite.ctx, alb)
	assert.NoError(suite.T(), err)
	return alb
}

func (suite *MemTestSuite) TestAlbum() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		alb := suite.saveAlbum(id, ids)
		edgs, err := suite.db.GetEdges(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, alb.Edges, edgs)
	})
	suite.T().Run("Negative1", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		alb := suite.saveAlbum(id, ids)
		err := suite.db.SaveAlbum(suite.ctx, alb)
		assert.ErrorIs(t, err, domain.ErrAlbumAlreadyExists)
	})
	suite.T().Run("Negative2", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		_, err := suite.db.GetImagesIds(suite.ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative3", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		_, err := suite.db.GetEdges(suite.ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func (suite *MemTestSuite) TestCount() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		n, err := suite.db.CountImages(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		n, err = suite.db.CountImagesCompressed(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 0, n)
		err = suite.db.UpdateCompressionStatus(suite.ctx, ids.Uint64(0), ids.Uint64(1))
		assert.NoError(t, err)
		n, err = suite.db.CountImages(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		n, err = suite.db.CountImagesCompressed(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		err = suite.db.UpdateCompressionStatus(suite.ctx, ids.Uint64(0), ids.Uint64(2))
		assert.NoError(t, err)
		n, err = suite.db.CountImagesCompressed(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
	})
	suite.T().Run("Negative1", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		err := suite.db.UpdateCompressionStatus(suite.ctx, ids.Uint64(0), ids.Uint64(1))
		assert.NoError(t, err)
		err = suite.db.UpdateCompressionStatus(suite.ctx, ids.Uint64(0), ids.Uint64(1))
		assert.NoError(t, err)
		n, err := suite.db.CountImagesCompressed(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	})
	suite.T().Run("Negative2", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		_, err := suite.db.CountImages(suite.ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative3", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		_, err := suite.db.CountImagesCompressed(suite.ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative4", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		err := suite.db.UpdateCompressionStatus(suite.ctx, id(), ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative5", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		err := suite.db.UpdateCompressionStatus(suite.ctx, ids.Uint64(0), id())
		assert.ErrorIs(t, err, domain.ErrImageNotFound)
	})
}

func (suite *MemTestSuite) TestImage() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		src, err := suite.db.GetImageSrc(suite.ctx, ids.Uint64(0), ids.Uint64(4))
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(4), src)
	})
	suite.T().Run("Negative1", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_, err := suite.db.GetImageSrc(suite.ctx, id(), ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative2", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		_, err := suite.db.GetImageSrc(suite.ctx, ids.Uint64(0), id())
		assert.ErrorIs(t, err, domain.ErrImageNotFound)
	})
}

func (suite *MemTestSuite) TestVote() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		err := suite.db.SaveVote(suite.ctx, ids.Uint64(0), ids.Uint64(3), ids.Uint64(5))
		assert.NoError(t, err)
		err = suite.db.SaveVote(suite.ctx, ids.Uint64(0), ids.Uint64(3), ids.Uint64(5))
		assert.NoError(t, err)
		edgs, err := suite.db.GetEdges(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 2, edgs[ids.Uint64(3)][ids.Uint64(5)])
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		err := suite.db.SaveVote(suite.ctx, id(), id(), id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func (suite *MemTestSuite) TestSort() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		imgs1, err := suite.db.GetImagesOrdered(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		img1 := model.Image{Id: ids.Uint64(4), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(4), Rating: 0.77920413}
		img2 := model.Image{Id: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1), Rating: 0.48954984}
		img3 := model.Image{Id: ids.Uint64(3), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(3), Rating: 0.41218211}
		img4 := model.Image{Id: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2), Rating: 0.19186324}
		img5 := model.Image{Id: ids.Uint64(5), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(5), Rating: 0.13278389}
		imgs2 := []model.Image{img1, img2, img3, img4, img5}
		assert.Equal(t, imgs2, imgs1)
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		id, _ := GenId()
		_, err := suite.db.GetImagesOrdered(suite.ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func (suite *MemTestSuite) TestRatings() {
	suite.T().Run("Positive", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		img1 := model.Image{Id: ids.Uint64(1), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1), Rating: 0.54412788}
		img2 := model.Image{Id: ids.Uint64(2), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2), Rating: 0.32537162}
		img3 := model.Image{Id: ids.Uint64(3), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(3), Rating: 0.43185491}
		img4 := model.Image{Id: ids.Uint64(4), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(4), Rating: 0.57356209}
		img5 := model.Image{Id: ids.Uint64(5), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(5), Rating: 0.61438023}
		imgs1 := []model.Image{img1, img2, img3, img4, img5}
		vector := map[uint64]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err := suite.db.UpdateRatings(suite.ctx, ids.Uint64(0), vector)
		assert.NoError(t, err)
		imgs2, err := suite.db.GetImagesOrdered(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		sort.Slice(imgs1, func(i, j int) bool { return imgs1[i].Rating > imgs1[j].Rating })
		assert.Equal(t, imgs1, imgs2)
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		album := id()
		img1 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1), Rating: 0.54412788}
		img2 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2), Rating: 0.32537162}
		img3 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(3), Rating: 0.43185491}
		img4 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(4), Rating: 0.57356209}
		img5 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(5), Rating: 0.61438023}
		vector := map[uint64]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err := suite.db.UpdateRatings(suite.ctx, album, vector)
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}

func (suite *MemTestSuite) TestDelete() {
	suite.T().Run("Positive1", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		alb := AlbumFactory(id, ids)
		_, err := suite.db.CountImages(suite.ctx, ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
		err = suite.db.SaveAlbum(suite.ctx, alb)
		assert.NoError(t, err)
		n, err := suite.db.CountImages(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		albums, err := suite.db.AlbumsToBeDeleted(suite.ctx)
		assert.NoError(t, err)
		assert.Len(t, albums, 0)
		err = suite.db.DeleteAlbum(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		_, err = suite.db.CountImages(suite.ctx, ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Positive2", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		alb := AlbumFactory(id, ids)
		alb.Expires = time.Now().Add(-1 * time.Hour)
		err := suite.db.SaveAlbum(suite.ctx, alb)
		assert.NoError(t, err)
		albums, err := suite.db.AlbumsToBeDeleted(suite.ctx)
		assert.NoError(t, err)
		assert.True(t, len(albums) == 1 && albums[0].Id == alb.Id && !albums[0].Expires.IsZero())
		err = suite.db.DeleteAlbum(suite.ctx, ids.Uint64(0))
		assert.NoError(t, err)
		_, err = suite.db.CountImages(suite.ctx, ids.Uint64(0))
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
	suite.T().Run("Negative", func(t *testing.T) {
		suite.setupTestFn()
		id, ids := GenId()
		_ = suite.saveAlbum(id, ids)
		err := suite.db.DeleteAlbum(suite.ctx, id())
		assert.ErrorIs(t, err, domain.ErrAlbumNotFound)
	})
}
