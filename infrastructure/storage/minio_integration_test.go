package storage

import (
	"context"
	"testing"

	minioS3 "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zitryss/aye-and-nay/domain/domain"
	. "github.com/zitryss/aye-and-nay/internal/generator"
	. "github.com/zitryss/aye-and-nay/internal/testing"
)

func TestMinioTestSuite(t *testing.T) {
	suite.Run(t, &MinioTestSuite{})
}

type MinioTestSuite struct {
	suite.Suite
	ctx     context.Context
	cancel  context.CancelFunc
	storage domain.Storager
}

func (suite *MinioTestSuite) SetupSuite() {
	if !*integration {
		suite.T().Skip()
	}
	ctx, cancel := context.WithCancel(context.Background())
	minio, err := NewMinio(ctx, DefaultMinioConfig)
	require.NoError(suite.T(), err)
	suite.ctx = ctx
	suite.cancel = cancel
	suite.storage = minio
}

func (suite *MinioTestSuite) SetupTest() {
	err := suite.storage.(*Minio).Reset()
	require.NoError(suite.T(), err)
}

func (suite *MinioTestSuite) TearDownTest() {

}

func (suite *MinioTestSuite) TearDownSuite() {
	err := suite.storage.(*Minio).Reset()
	require.NoError(suite.T(), err)
	suite.cancel()
}

func (suite *MinioTestSuite) TestMinio() {
	suite.T().Run("", func(t *testing.T) {
		id, ids := GenId()
		album := id()
		image := id()
		f, err := suite.storage.Get(suite.ctx, album, image)
		e := minioS3.ErrorResponse{}
		assert.ErrorAs(t, err, &e)
		assert.Nil(t, f.Reader)
		src, err := suite.storage.Put(suite.ctx, album, image, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(1), src)
		f, err = suite.storage.Get(suite.ctx, album, image)
		assert.NoError(t, err)
		AssertEqualFile(t, f, Png())
		err = suite.storage.Remove(suite.ctx, album, image)
		assert.NoError(t, err)
		f, err = suite.storage.Get(suite.ctx, album, image)
		e = minioS3.ErrorResponse{}
		assert.ErrorAs(t, err, &e)
		assert.Nil(t, f.Reader)
	})
	suite.T().Run("", func(t *testing.T) {
		id, ids := GenId()
		album := id()
		image := id()
		src, err := suite.storage.Put(suite.ctx, album, image, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(1), src)
		f, err := suite.storage.Get(suite.ctx, album, image)
		assert.NoError(t, err)
		AssertEqualFile(t, f, Png())
		err = suite.storage.Remove(suite.ctx, album, image)
		assert.NoError(t, err)
		src, err = suite.storage.Put(suite.ctx, album, image, Png())
		assert.NoError(t, err)
		assert.Equal(t, "/aye-and-nay/albums/"+ids.Base64(0)+"/images/"+ids.Base64(1), src)
	})
}

func (suite *MinioTestSuite) TestMinioHealth() {
	_, err := suite.storage.Health(suite.ctx)
	assert.NoError(suite.T(), err)
}
